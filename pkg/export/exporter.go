package export

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

// ExportDocument exports a document to the specified format
func (e *Exporter) ExportDocument(documentID string, manifest *types.Manifest, style *types.Style, pandocConfig *types.PandocConfig, options *types.ExportOptions) (string, error) {
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

	// Execute pandoc command with timeout
	ctx, cancel := context.WithTimeout(context.Background(), e.config.ExportTimeout)
	defer cancel()

	cmd.SysProcAttr = nil // Ensure clean execution
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start pandoc: %w", err)
	}

	// Wait for completion with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
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
		args = append(args, "--pdf-engine", pandocConfig.PDFEngine)
		if style != nil {
			// Add font and margin settings
			args = append(args, "-V", fmt.Sprintf("fontsize=%s", style.FontSize))
			args = append(args, "-V", fmt.Sprintf("mainfont=%s", style.FontFamily))
			args = append(args, "-V", fmt.Sprintf("geometry:margin=%s", style.Margins.Top))
		}
	case types.ExportFormatDOCX:
		// Add DOCX-specific options
		args = append(args, "--reference-doc", "reference.docx") // Could be configurable
	case types.ExportFormatHTML:
		args = append(args, "--standalone")
		args = append(args, "--css", "style.css") // Could be configurable
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

	return exec.Command(e.config.PandocPath, args...)
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
	if _, err := exec.LookPath(e.config.PandocPath); err != nil {
		report.Errors = append(report.Errors, fmt.Sprintf("Pandoc not found at path: %s", e.config.PandocPath))
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
func (e *Exporter) PreviewChapter(documentID string, chapterNum types.ChapterNumber, format types.ExportFormat) (string, error) {
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

	cmd := exec.Command(e.config.PandocPath, args...)

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

	yaml.WriteString(fmt.Sprintf("title: %s\n", doc.Title))
	yaml.WriteString(fmt.Sprintf("author: %s\n", doc.Author))
	yaml.WriteString(fmt.Sprintf("date: %s\n", time.Now().Format("2006-01-02")))

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
		yaml.WriteString(fmt.Sprintf("fontsize: %s\n", style.FontSize))
		yaml.WriteString(fmt.Sprintf("fontfamily: %s\n", style.FontFamily))
		if style.Margins.Top != "" {
			yaml.WriteString(fmt.Sprintf("geometry: margin=%s\n", style.Margins.Top))
		}
	}

	return yaml.String()
}