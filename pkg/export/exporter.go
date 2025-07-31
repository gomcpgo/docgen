package export

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gomcpgo/docgen/pkg/config"
	"github.com/gomcpgo/docgen/pkg/types"
)

// Exporter handles document export operations using Pandoc
type Exporter struct {
	config *config.Config
}

// NewExporter creates a new exporter instance
func NewExporter(cfg *config.Config) *Exporter {
	return &Exporter{
		config: cfg,
	}
}

// ChapterRebuildFunc defines a function that rebuilds chapter markdown from sections
type ChapterRebuildFunc func(docID types.DocumentID, chapterNum types.ChapterNumber) error

// ExportDocument exports a document to the specified format
func (e *Exporter) ExportDocument(documentID string, manifest *types.Manifest, style *types.Style, pandocConfig *types.PandocConfig, options *types.ExportOptions, rebuildFunc ChapterRebuildFunc) (string, error) {
	// Rebuild all chapter markdown files from section files to ensure they're current
	if rebuildFunc != nil {
		for _, chapter := range manifest.Document.Chapters {
			if err := rebuildFunc(types.DocumentID(documentID), chapter.Number); err != nil {
				return "", fmt.Errorf("failed to rebuild chapter %d markdown: %w", chapter.Number, err)
			}
		}
	}

	// Validate document first
	report := e.ValidateDocument(documentID, manifest)
	if !report.Valid {
		return "", fmt.Errorf("document validation failed: %v", report.Errors)
	}

	// Generate combined markdown
	markdown, err := e.GenerateMarkdown(documentID, manifest, options)
	if err != nil {
		return "", fmt.Errorf("failed to generate markdown: %w", err)
	}

	// Create temporary input file
	tempInputFile := filepath.Join(os.TempDir(), fmt.Sprintf("%s-input.md", documentID))
	if err := os.WriteFile(tempInputFile, []byte(markdown), 0644); err != nil {
		return "", fmt.Errorf("failed to write temporary input file: %w", err)
	}
	defer os.Remove(tempInputFile)

	// Generate output file path
	outputFile := e.config.ExportPath(documentID, string(options.Format))
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate pandoc command
	cmd := e.GeneratePandocCommand(documentID, tempInputFile, outputFile, manifest, style, pandocConfig, options)

	// Command is ready for execution

	// Execute pandoc command with timeout
	ctx, cancel := context.WithTimeout(context.Background(), e.config.ExportTimeout)
	defer cancel()

	// Capture stderr for better error reporting
	cmd.Stderr = nil // We'll capture it manually
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	cmd.SysProcAttr = nil // Ensure clean execution
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start pandoc: %w", err)
	}

	// Read stderr in background
	stderrCh := make(chan string, 1)
	go func() {
		stderrBytes, err := io.ReadAll(stderrPipe)
		if err == nil {
			stderrCh <- string(stderrBytes)
		} else {
			stderrCh <- ""
		}
	}()

	// Wait for completion with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		stderr := <-stderrCh
		if err != nil {
			if stderr != "" {
				return "", fmt.Errorf("pandoc execution failed: %w. Stderr: %s", err, stderr)
			}
			return "", fmt.Errorf("pandoc execution failed: %w", err)
		}
	case <-ctx.Done():
		cmd.Process.Kill()
		return "", fmt.Errorf("pandoc execution timed out after %v", e.config.ExportTimeout)
	}

	// Verify output file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		return "", fmt.Errorf("output file was not created: %s", outputFile)
	}

	return outputFile, nil
}

// GenerateMarkdown combines all chapters into a single markdown document
func (e *Exporter) GenerateMarkdown(documentID string, manifest *types.Manifest, options *types.ExportOptions) (string, error) {
	var content strings.Builder

	// Add YAML metadata header
	yaml := generateYAMLMetadata(&manifest.Document, nil)
	content.WriteString("---\n")
	content.WriteString(yaml)
	content.WriteString("---\n\n")

	// Determine which chapters to include
	chaptersToInclude := options.Chapters
	if len(chaptersToInclude) == 0 {
		// Include all chapters if none specified
		for _, chapter := range manifest.Document.Chapters {
			chaptersToInclude = append(chaptersToInclude, chapter.Number)
		}
	}

	// Process each chapter
	for _, chapterNum := range chaptersToInclude {
		// Find the chapter in manifest
		var chapter *types.Chapter
		for i := range manifest.Document.Chapters {
			if manifest.Document.Chapters[i].Number == chapterNum {
				chapter = &manifest.Document.Chapters[i]
				break
			}
		}

		if chapter == nil {
			return "", fmt.Errorf("chapter %d not found", chapterNum)
		}

		// Read chapter content
		chapterContent, err := e.loadChapterContent(documentID, int(chapterNum))
		if err != nil {
			return "", fmt.Errorf("failed to load chapter %d content: %w", chapterNum, err)
		}

		// Add chapter to combined content
		content.WriteString(fmt.Sprintf("\\newpage\n\n"))
		content.WriteString(chapterContent)
		content.WriteString("\n\n")
	}

	return content.String(), nil
}

// GeneratePandocCommand creates the pandoc command with all necessary options
func (e *Exporter) GeneratePandocCommand(documentID, inputFile, outputFile string, manifest *types.Manifest, style *types.Style, pandocConfig *types.PandocConfig, options *types.ExportOptions) *exec.Cmd {
	args := []string{
		inputFile,
		"-o", outputFile,
	}

	// Add format-specific options
	switch options.Format {
	case types.ExportFormatPDF:
		// Determine PDF engine based on style
		pdfEngine := determinePDFEngine(style, pandocConfig)
		args = append(args, "--pdf-engine", pdfEngine)
		
		if style != nil {
			// Generate and include LaTeX header for advanced styling
			latexHeader := generateLaTeXHeader(style, manifest)
			if latexHeader != "" {
				// Create temporary LaTeX header file
				tempHeaderFile := filepath.Join(os.TempDir(), fmt.Sprintf("%s-header.tex", documentID))
				if err := os.WriteFile(tempHeaderFile, []byte(latexHeader), 0644); err == nil {
					args = append(args, "-H", tempHeaderFile)
				}
			}
			
			// Add basic font and margin settings
			if style.Body.FontSize != "" {
				args = append(args, "-V", fmt.Sprintf("fontsize=%s", style.Body.FontSize))
			}
			if style.Body.FontFamily != "" {
				args = append(args, "-V", fmt.Sprintf("mainfont=%s", style.Body.FontFamily))
			}
			if style.Margins.Top != "" {
				args = append(args, "-V", fmt.Sprintf("geometry:margin=%s", style.Margins.Top))
			}
		}
		
	case types.ExportFormatDOCX:
		// Use custom reference document if specified
		referenceDoc := "reference.docx" // default
		if style != nil && style.ReferenceDocx != "" {
			// Resolve reference document path
			if filepath.IsAbs(style.ReferenceDocx) {
				referenceDoc = style.ReferenceDocx
			} else {
				// Relative to document directory
				docDir := filepath.Dir(e.config.ManifestPath(documentID))
				referenceDoc = filepath.Join(docDir, style.ReferenceDocx)
			}
		}
		args = append(args, "--reference-doc", referenceDoc)
		
	case types.ExportFormatHTML:
		args = append(args, "--standalone")
		
		// Use custom CSS if specified
		cssFile := "style.css" // default
		if style != nil && style.StyleCSS != "" {
			// Resolve CSS file path
			if filepath.IsAbs(style.StyleCSS) {
				cssFile = style.StyleCSS
			} else {
				// Relative to document directory
				docDir := filepath.Dir(e.config.ManifestPath(documentID))
				cssFile = filepath.Join(docDir, style.StyleCSS)
			}
		} else if style != nil {
			// Generate CSS file with style specifications
			cssContent := generateHTMLCSS(style, manifest)
			if cssContent != "" {
				// Create temporary CSS file
				tempCSSFile := filepath.Join(os.TempDir(), fmt.Sprintf("%s-style.css", documentID))
				if err := os.WriteFile(tempCSSFile, []byte(cssContent), 0644); err == nil {
					cssFile = tempCSSFile
				}
			}
		}
		args = append(args, "--css", cssFile)
	}

	// Add table of contents if enabled
	if pandocConfig.TOC {
		args = append(args, "--toc")
		if pandocConfig.TOCDepth > 0 {
			args = append(args, "--toc-depth", fmt.Sprintf("%d", pandocConfig.TOCDepth))
		}
	}

	// Add any additional arguments
	args = append(args, pandocConfig.Args...)

	// Add variables
	for key, value := range pandocConfig.Variables {
		args = append(args, "-V", fmt.Sprintf("%s=%s", key, value))
	}

	// Resolve pandoc path
	pandocPath, err := findPandocPath(e.config.PandocPath)
	if err != nil {
		// This should have been caught in validation, but handle gracefully
		pandocPath = e.config.PandocPath
	}

	return exec.Command(pandocPath, args...)
}

// ValidateDocument validates that a document is ready for export
func (e *Exporter) ValidateDocument(documentID string, manifest *types.Manifest) *types.ValidationReport {
	report := &types.ValidationReport{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Check that all chapters have content files
	for _, chapter := range manifest.Document.Chapters {
		chapterPath := e.config.ChapterContentPath(documentID, int(chapter.Number))
		if _, err := os.Stat(chapterPath); os.IsNotExist(err) {
			report.Errors = append(report.Errors, fmt.Sprintf("Chapter %d content file not found: %s", chapter.Number, chapterPath))
			report.Valid = false
		}
	}

	// Check that required files exist
	manifestPath := e.config.ManifestPath(documentID)
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		report.Errors = append(report.Errors, fmt.Sprintf("Manifest file not found: %s", manifestPath))
		report.Valid = false
	}

	// Check pandoc availability
	_, err := findPandocPath(e.config.PandocPath)
	if err != nil {
		report.Errors = append(report.Errors, err.Error())
		report.Valid = false
	}

	// Add warnings for missing figures or broken references
	for _, chapter := range manifest.Document.Chapters {
		for _, figure := range chapter.Figures {
			imagePath := filepath.Join(e.config.AssetsPath(documentID), filepath.Base(figure.ImagePath))
			if _, err := os.Stat(imagePath); os.IsNotExist(err) {
				report.Warnings = append(report.Warnings, fmt.Sprintf("Figure image not found: %s", imagePath))
			}
		}
	}

	return report
}

// PreviewChapter generates a preview of a single chapter
func (e *Exporter) PreviewChapter(documentID string, chapterNum types.ChapterNumber, format types.ExportFormat, rebuildFunc ChapterRebuildFunc) (string, error) {
	// Rebuild chapter markdown from section files to ensure it's current
	if rebuildFunc != nil {
		if err := rebuildFunc(types.DocumentID(documentID), chapterNum); err != nil {
			return "", fmt.Errorf("failed to rebuild chapter %d markdown: %w", chapterNum, err)
		}
	}

	// Create a minimal manifest with just one chapter
	chapterContent, err := e.loadChapterContent(documentID, int(chapterNum))
	if err != nil {
		return "", fmt.Errorf("failed to load chapter content: %w", err)
	}

	// Create temporary input file
	tempInputFile := filepath.Join(os.TempDir(), fmt.Sprintf("%s-chapter-%d-preview.md", documentID, chapterNum))
	if err := os.WriteFile(tempInputFile, []byte(chapterContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write temporary input file: %w", err)
	}
	defer os.Remove(tempInputFile)

	// Generate output file path
	outputFile := filepath.Join(e.config.ExportsDir, fmt.Sprintf("%s-chapter-%d-preview.%s", documentID, chapterNum, format))
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create simple pandoc command for preview
	args := []string{
		tempInputFile,
		"-o", outputFile,
		"--standalone",
	}

	if format == types.ExportFormatPDF {
		args = append(args, "--pdf-engine", "pdflatex")
	}

	// Resolve pandoc path
	pandocPath, err := findPandocPath(e.config.PandocPath)
	if err != nil {
		return "", fmt.Errorf("pandoc not found: %w", err)
	}

	cmd := exec.Command(pandocPath, args...)

	// Execute with timeout
	ctx, cancel := context.WithTimeout(context.Background(), e.config.ExportTimeout)
	defer cancel()

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("pandoc execution failed: %w", err)
	}

	select {
	case <-ctx.Done():
		return "", fmt.Errorf("pandoc execution timed out")
	default:
	}

	return outputFile, nil
}

// loadChapterContent loads the content of a specific chapter
func (e *Exporter) loadChapterContent(documentID string, chapterNumber int) (string, error) {
	contentPath := e.config.ChapterContentPath(documentID, chapterNumber)
	content, err := os.ReadFile(contentPath)
	if err != nil {
		return "", fmt.Errorf("failed to read chapter content: %w", err)
	}
	return string(content), nil
}

// generateYAMLMetadata generates YAML metadata for the document
func generateYAMLMetadata(doc *types.Document, style *types.Style) string {
	var yaml strings.Builder

	yaml.WriteString(fmt.Sprintf("title: %q\n", doc.Title))
	yaml.WriteString(fmt.Sprintf("author: %q\n", doc.Author))
	yaml.WriteString(fmt.Sprintf("date: %q\n", time.Now().Format("2006-01-02")))

	// Document class based on type
	switch doc.Type {
	case types.DocumentTypeBook:
		yaml.WriteString("documentclass: book\n")
	case types.DocumentTypeReport:
		yaml.WriteString("documentclass: report\n")
	case types.DocumentTypeArticle:
		yaml.WriteString("documentclass: article\n")
	default:
		yaml.WriteString("documentclass: article\n")
	}

	// Add style information if provided
	if style != nil {
		if style.Body.FontSize != "" {
			yaml.WriteString(fmt.Sprintf("fontsize: %q\n", style.Body.FontSize))
		}
		if style.Body.FontFamily != "" {
			yaml.WriteString(fmt.Sprintf("fontfamily: %q\n", style.Body.FontFamily))
		}
		if style.Margins.Top != "" {
			yaml.WriteString(fmt.Sprintf("geometry: %q\n", fmt.Sprintf("margin=%s", style.Margins.Top)))
		}
	}

	return yaml.String()
}

// findPandocPath finds the pandoc executable using system commands
func findPandocPath(configPath string) (string, error) {
	// If config path is an absolute path, use it directly
	if filepath.IsAbs(configPath) {
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}
		return "", fmt.Errorf("pandoc not found at configured path: %s", configPath)
	}

	// Use system command to find pandoc in PATH
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("where", configPath)
	} else {
		cmd = exec.Command("which", configPath)
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("pandoc not found in PATH. Please install pandoc or set PANDOC_PATH environment variable")
	}

	path := strings.TrimSpace(string(output))
	// On Windows, 'where' might return multiple paths, take the first one
	if runtime.GOOS == "windows" {
		lines := strings.Split(path, "\n")
		if len(lines) > 0 {
			path = strings.TrimSpace(lines[0])
		}
	}

	return path, nil
}

// generateLaTeXHeader creates a LaTeX header file with advanced styling
func generateLaTeXHeader(style *types.Style, manifest *types.Manifest) string {
	if style == nil {
		return ""
	}

	var header strings.Builder

	// Font settings (requires XeLaTeX or LuaLaTeX)
	if needsXeLaTeX(style) {
		header.WriteString("% Font settings (requires XeLaTeX or LuaLaTeX)\n")
		header.WriteString("\\usepackage{fontspec}\n")
		
		// Set main font (body text)
		if style.Body.FontFamily != "" {
			header.WriteString(fmt.Sprintf("\\setmainfont{%s}\n", style.Body.FontFamily))
		}
		
		// Set heading font
		if style.Heading.FontFamily != "" && style.Heading.FontFamily != style.Body.FontFamily {
			header.WriteString(fmt.Sprintf("\\newfontfamily\\headingfont{%s}\n", style.Heading.FontFamily))
		}
		
		// Set monospace font
		if style.Monospace.FontFamily != "" {
			header.WriteString(fmt.Sprintf("\\newfontfamily\\monospacefont{%s}\n", style.Monospace.FontFamily))
			header.WriteString("\\setmonofont{" + style.Monospace.FontFamily + "}\n")
		}
	}

	// Font sizes and section styling
	header.WriteString("\n% Font sizes and section styling\n")
	header.WriteString("\\usepackage{sectsty}\n")
	
	// Apply heading font to all sections if different from body
	if style.Heading.FontFamily != "" && style.Heading.FontFamily != style.Body.FontFamily {
		header.WriteString("\\allsectionsfont{\\headingfont}\n")
	}

	// Color settings
	if hasColors(style) {
		header.WriteString("\n% Color settings\n")
		header.WriteString("\\usepackage{xcolor}\n")
		
		// Define colors
		if style.Body.Color != "" {
			bodyColorHex := convertColorToHex(style.Body.Color)
			header.WriteString(fmt.Sprintf("\\definecolor{bodycolor}{HTML}{%s}\n", bodyColorHex))
		}
		
		if style.Heading.Color != "" {
			headingColorHex := convertColorToHex(style.Heading.Color)
			header.WriteString(fmt.Sprintf("\\definecolor{headingcolor}{HTML}{%s}\n", headingColorHex))
		}
		
		if style.LinkColor != "" {
			linkColorHex := convertColorToHex(style.LinkColor)
			header.WriteString(fmt.Sprintf("\\definecolor{linkcolor}{HTML}{%s}\n", linkColorHex))
		}

		// Apply colors
		header.WriteString("\n% Apply colors\n")
		if style.Heading.Color != "" {
			header.WriteString("\\allsectionsfont{\\color{headingcolor}}\n")
		}
		if style.Body.Color != "" {
			header.WriteString("\\AtBeginDocument{\\color{bodycolor}}\n")
		}
		if style.LinkColor != "" {
			header.WriteString("\\usepackage{hyperref}\n")
			header.WriteString("\\hypersetup{colorlinks=true,linkcolor=linkcolor,urlcolor=linkcolor}\n")
		}
	}

	// Header/Footer setup
	if style.HeaderFooter.HeaderTemplate != "" || style.HeaderFooter.FooterTemplate != "" {
		header.WriteString("\n% Header/Footer\n")
		header.WriteString("\\usepackage{fancyhdr}\n")
		header.WriteString("\\usepackage{lastpage}\n")
		header.WriteString("\\pagestyle{fancy}\n")
		header.WriteString("\\fancyhf{}\n")

		// Process templates for LaTeX
		vars := CreateTemplateVariables(manifest)
		
		if style.HeaderFooter.HeaderTemplate != "" {
			headerContent := ProcessTemplateForPDF(style.HeaderFooter.HeaderTemplate, vars)
			header.WriteString(fmt.Sprintf("\\fancyhead[C]{%s}\n", headerContent))
		}
		
		if style.HeaderFooter.FooterTemplate != "" {
			footerContent := ProcessTemplateForPDF(style.HeaderFooter.FooterTemplate, vars)
			header.WriteString(fmt.Sprintf("\\fancyfoot[C]{%s}\n", footerContent))
		}
	}

	// Line spacing
	if style.LineSpacing != "" && style.LineSpacing != "1" {
		header.WriteString(fmt.Sprintf("\n%% Line spacing\n\\linespread{%s}\n", style.LineSpacing))
	}

	// Add any custom LaTeX header content
	if style.LaTeXHeader != "" {
		header.WriteString("\n% Custom LaTeX header\n")
		header.WriteString(style.LaTeXHeader)
		header.WriteString("\n")
	}

	return header.String()
}

// needsXeLaTeX determines if XeLaTeX is required for the given style
func needsXeLaTeX(style *types.Style) bool {
	if style == nil {
		return false
	}

	defaultFonts := []string{"Times New Roman", "Computer Modern", "Latin Modern", ""}
	
	// Check if any custom fonts are specified
	if !contains(defaultFonts, style.Body.FontFamily) {
		return true
	}
	if !contains(defaultFonts, style.Heading.FontFamily) {
		return true
	}
	if style.Monospace.FontFamily != "" && !contains(defaultFonts, style.Monospace.FontFamily) {
		return true
	}
	
	return false
}

// hasColors checks if any colors are specified in the style
func hasColors(style *types.Style) bool {
	if style == nil {
		return false
	}
	
	return style.Body.Color != "" || style.Heading.Color != "" || 
		   style.Monospace.Color != "" || style.LinkColor != ""
}

// convertColorToHex converts color format to hex (removes # prefix for LaTeX)
func convertColorToHex(color string) string {
	if color == "" {
		return "000000"
	}
	
	// Remove # prefix if present
	if strings.HasPrefix(color, "#") {
		return strings.TrimPrefix(color, "#")
	}
	
	// Handle rgb() format
	if strings.HasPrefix(color, "rgb(") {
		// For now, return black as fallback - could implement RGB to hex conversion
		return "000000"
	}
	
	// Assume it's already in hex format without #
	return color
}

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// determinePDFEngine determines which PDF engine to use based on style
func determinePDFEngine(style *types.Style, pandocConfig *types.PandocConfig) string {
	// User override takes precedence
	if pandocConfig != nil && pandocConfig.PDFEngine != "" {
		return pandocConfig.PDFEngine
	}
	
	// Check if custom fonts are used
	if needsXeLaTeX(style) {
		return "xelatex"
	}
	
	return "pdflatex"
}

// generateHTMLCSS creates CSS content for HTML export
func generateHTMLCSS(style *types.Style, manifest *types.Manifest) string {
	if style == nil {
		return ""
	}

	var css strings.Builder

	// Google Fonts imports (if Google Fonts are detected)
	if needsGoogleFonts(style) {
		css.WriteString("@import url('https://fonts.googleapis.com/css2?family=")
		
		fonts := make(map[string]bool)
		if isGoogleFont(style.Body.FontFamily) {
			fonts[strings.ReplaceAll(style.Body.FontFamily, " ", "+")] = true
		}
		if isGoogleFont(style.Heading.FontFamily) {
			fonts[strings.ReplaceAll(style.Heading.FontFamily, " ", "+")] = true
		}
		if isGoogleFont(style.Monospace.FontFamily) {
			fonts[strings.ReplaceAll(style.Monospace.FontFamily, " ", "+")] = true
		}
		
		var fontNames []string
		for font := range fonts {
			fontNames = append(fontNames, font)
		}
		css.WriteString(strings.Join(fontNames, "&family="))
		css.WriteString("&display=swap');\n\n")
	}

	// Body styles
	css.WriteString("body {\n")
	if style.Body.FontFamily != "" {
		css.WriteString(fmt.Sprintf("    font-family: '%s', serif;\n", style.Body.FontFamily))
	}
	if style.Body.FontSize != "" {
		css.WriteString(fmt.Sprintf("    font-size: %s;\n", style.Body.FontSize))
	}
	if style.Body.Color != "" {
		css.WriteString(fmt.Sprintf("    color: %s;\n", style.Body.Color))
	}
	if style.LineSpacing != "" {
		css.WriteString(fmt.Sprintf("    line-height: %s;\n", style.LineSpacing))
	}
	css.WriteString("}\n\n")

	// Heading styles
	css.WriteString("h1, h2, h3, h4, h5, h6 {\n")
	if style.Heading.FontFamily != "" {
		css.WriteString(fmt.Sprintf("    font-family: '%s', sans-serif;\n", style.Heading.FontFamily))
	}
	if style.Heading.Color != "" {
		css.WriteString(fmt.Sprintf("    color: %s;\n", style.Heading.Color))
	}
	css.WriteString("}\n\n")

	// Monospace styles
	if style.Monospace.FontFamily != "" || style.Monospace.FontSize != "" || style.Monospace.Color != "" {
		css.WriteString("code, pre, .code {\n")
		if style.Monospace.FontFamily != "" {
			css.WriteString(fmt.Sprintf("    font-family: '%s', monospace;\n", style.Monospace.FontFamily))
		}
		if style.Monospace.FontSize != "" {
			css.WriteString(fmt.Sprintf("    font-size: %s;\n", style.Monospace.FontSize))
		}
		if style.Monospace.Color != "" {
			css.WriteString(fmt.Sprintf("    color: %s;\n", style.Monospace.Color))
		}
		css.WriteString("}\n\n")
	}

	// Link styles
	if style.LinkColor != "" {
		css.WriteString("a {\n")
		css.WriteString(fmt.Sprintf("    color: %s;\n", style.LinkColor))
		css.WriteString("}\n\n")
	}

	// Print-specific styles for headers/footers (if templates are specified)
	if style.HeaderFooter.HeaderTemplate != "" || style.HeaderFooter.FooterTemplate != "" {
		vars := CreateTemplateVariables(manifest)
		
		css.WriteString("@media print {\n")
		
		if style.HeaderFooter.HeaderTemplate != "" {
			headerContent := ProcessTemplateForHTML(style.HeaderFooter.HeaderTemplate, vars)
			css.WriteString("    @page {\n")
			css.WriteString(fmt.Sprintf("        @top-center { content: '%s'; }\n", headerContent))
			css.WriteString("    }\n")
		}
		
		if style.HeaderFooter.FooterTemplate != "" {
			footerContent := ProcessTemplateForHTML(style.HeaderFooter.FooterTemplate, vars)
			css.WriteString("    @page {\n")
			css.WriteString(fmt.Sprintf("        @bottom-center { content: '%s'; }\n", footerContent))
			css.WriteString("    }\n")
		}
		
		css.WriteString("}\n")
	}

	return css.String()
}

// needsGoogleFonts checks if any Google Fonts are used
func needsGoogleFonts(style *types.Style) bool {
	return isGoogleFont(style.Body.FontFamily) || 
		   isGoogleFont(style.Heading.FontFamily) || 
		   isGoogleFont(style.Monospace.FontFamily)
}

// isGoogleFont checks if a font name is likely a Google Font
func isGoogleFont(fontName string) bool {
	if fontName == "" {
		return false
	}
	
	// Basic heuristic: Google Fonts typically have specific naming patterns
	// For now, we'll include common ones and exclude system fonts
	systemFonts := []string{
		"Times New Roman", "Arial", "Helvetica", "Georgia", "Verdana",
		"Courier New", "Comic Sans MS", "Impact", "Trebuchet MS",
		"Computer Modern", "Latin Modern",
	}
	
	for _, sysFont := range systemFonts {
		if strings.EqualFold(fontName, sysFont) {
			return false
		}
	}
	
	// If it's not a known system font, assume it might be a Google Font
	return true
}
