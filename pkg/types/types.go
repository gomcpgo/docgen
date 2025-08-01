package types

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// DocumentID represents a unique document identifier
type DocumentID string

// ChapterNumber represents a chapter number (1, 2, 3...)
type ChapterNumber int

// SectionNumber represents a section number (1.1, 1.2.1, etc.)
type SectionNumber []int

// FigureID represents a figure identifier (fig-1.1, fig-2.3...)
type FigureID string

// TableID represents a table identifier (table-1.1, table-2.3...)
type TableID string

// DocumentType represents the type of document
type DocumentType string

const (
	DocumentTypeBook   DocumentType = "book"
	DocumentTypeReport DocumentType = "report"
	DocumentTypeArticle DocumentType = "article"
	DocumentTypeLetter DocumentType = "letter"
)

// ExportFormat represents the export format
type ExportFormat string

const (
	ExportFormatPDF  ExportFormat = "pdf"
	ExportFormatDOCX ExportFormat = "docx"
	ExportFormatHTML ExportFormat = "html"
)

// ImagePosition represents image positioning options
type ImagePosition string

const (
	PositionHere   ImagePosition = "here"
	PositionTop    ImagePosition = "top"
	PositionBottom ImagePosition = "bottom"
	PositionPage   ImagePosition = "page"
	PositionFloat  ImagePosition = "float"
)

// ImageAlignment represents image alignment options
type ImageAlignment string

const (
	AlignLeft   ImageAlignment = "left"
	AlignCenter ImageAlignment = "center"
	AlignRight  ImageAlignment = "right"
)

// Document represents a complete document
type Document struct {
	ID          DocumentID    `yaml:"id"`
	Title       string        `yaml:"title"`
	Author      string        `yaml:"author"`
	Type        DocumentType  `yaml:"type"`
	CreatedAt   time.Time     `yaml:"created_at"`
	UpdatedAt   time.Time     `yaml:"updated_at"`
	Chapters    []Chapter     `yaml:"chapters"`
}

// Chapter represents a document chapter
type Chapter struct {
	Number    ChapterNumber `yaml:"number"`
	Title     string        `yaml:"title"`
	Content   string        `yaml:"-"` // Stored in separate file
	Sections  []Section     `yaml:"sections"`
	Figures   []Figure      `yaml:"figures"`
	Tables    []Table       `yaml:"tables"`
	CreatedAt time.Time     `yaml:"created_at"`
	UpdatedAt time.Time     `yaml:"updated_at"`
}

// Section represents a document section
type Section struct {
	Number    SectionNumber `yaml:"number"`
	Title     string        `yaml:"title"`
	Content   string        `yaml:"content"`
	Level     int           `yaml:"level"` // 1, 2, 3 for different heading levels
	CreatedAt time.Time     `yaml:"created_at"`
	UpdatedAt time.Time     `yaml:"updated_at"`
}

// Figure represents an image figure
type Figure struct {
	ID        FigureID       `yaml:"id"`
	Chapter   ChapterNumber  `yaml:"chapter"`
	Sequence  int            `yaml:"sequence"`
	Caption   string         `yaml:"caption"`
	ImagePath string         `yaml:"image_path"`
	Position  ImagePosition  `yaml:"position"`
	Width     string         `yaml:"width,omitempty"`
	Alignment ImageAlignment `yaml:"alignment"`
	CreatedAt time.Time      `yaml:"created_at"`
	UpdatedAt time.Time      `yaml:"updated_at"`
}

// Table represents a document table
type Table struct {
	ID        TableID       `yaml:"id"`
	Chapter   ChapterNumber `yaml:"chapter"`
	Sequence  int           `yaml:"sequence"`
	Caption   string        `yaml:"caption"`
	Content   string        `yaml:"content"` // Markdown table content
	Format    string        `yaml:"format"`  // "markdown" for MVP
	CreatedAt time.Time     `yaml:"created_at"`
	UpdatedAt time.Time     `yaml:"updated_at"`
}

// ChapterCount tracks counts for a chapter
type ChapterCount struct {
	Sections int `yaml:"sections"`
	Figures  int `yaml:"figures"`
	Tables   int `yaml:"tables"`
}

// Manifest represents the document manifest
type Manifest struct {
	Document      Document                    `yaml:"document"`
	ChapterCounts map[ChapterNumber]ChapterCount `yaml:"chapter_counts"`
	CreatedAt     time.Time                   `yaml:"created_at"`
	UpdatedAt     time.Time                   `yaml:"updated_at"`
}

// TextStyle represents font and color settings for text elements
type TextStyle struct {
	FontFamily string `yaml:"font_family"`
	FontSize   string `yaml:"font_size,omitempty"`
	Color      string `yaml:"color,omitempty"`
}

// Style represents document styling configuration
type Style struct {
	// Text styles
	Body          TextStyle      `yaml:"body"`
	Heading       TextStyle      `yaml:"heading"`
	Monospace     TextStyle      `yaml:"monospace,omitempty"`
	
	// Global styles
	LinkColor     string         `yaml:"link_color,omitempty"`
	Margins       Margins        `yaml:"margins"`
	LineSpacing   string         `yaml:"line_spacing,omitempty"`
	
	// Header/Footer with template support
	HeaderFooter  HeaderFooter   `yaml:"header_footer"`
	
	// Other settings
	NumberingStyle NumberingStyle `yaml:"numbering_style"`
	
	// Output-specific templates
	ReferenceDocx string         `yaml:"reference_docx,omitempty"`
	StyleCSS      string         `yaml:"style_css,omitempty"`
	LaTeXHeader   string         `yaml:"latex_header,omitempty"`
}

// Margins represents document margins
type Margins struct {
	Top    string `yaml:"top"`
	Bottom string `yaml:"bottom"`
	Left   string `yaml:"left"`
	Right  string `yaml:"right"`
}

// HeaderFooter represents header and footer configuration
type HeaderFooter struct {
	HeaderTemplate string `yaml:"header_template,omitempty"`
	FooterTemplate string `yaml:"footer_template,omitempty"`
}

// NumberingStyle represents numbering preferences
type NumberingStyle struct {
	Chapters bool `yaml:"chapters"`
	Sections bool `yaml:"sections"`
	Figures  bool `yaml:"figures"`
	Tables   bool `yaml:"tables"`
}

// PandocConfig represents pandoc-specific configuration
type PandocConfig struct {
	PDFEngine     string            `yaml:"pdf_engine"`
	TOC           bool              `yaml:"toc"`
	TOCDepth      int               `yaml:"toc_depth"`
	CitationStyle string            `yaml:"citation_style"`
	Args          []string          `yaml:"args"`
	Variables     map[string]string `yaml:"variables"`
}

// ValidationReport represents document validation results
type ValidationReport struct {
	Valid    bool     `yaml:"valid"`
	Errors   []string `yaml:"errors"`
	Warnings []string `yaml:"warnings"`
}

// ExportOptions represents export configuration
type ExportOptions struct {
	Format   ExportFormat  `yaml:"format"`
	Chapters []ChapterNumber `yaml:"chapters,omitempty"`
	Template string        `yaml:"template,omitempty"`
}

// Validate validates a DocumentID
func (id DocumentID) Validate() error {
	if len(id) == 0 {
		return fmt.Errorf("document ID cannot be empty")
	}
	if len(id) > 50 {
		return fmt.Errorf("document ID too long (max 50 characters)")
	}
	
	// Allow alphanumeric characters, hyphens, and underscores
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, string(id))
	if !matched {
		return fmt.Errorf("document ID contains invalid characters (use only letters, numbers, hyphens, and underscores)")
	}
	
	return nil
}

// String returns the string representation of a ChapterNumber
func (c ChapterNumber) String() string {
	return strconv.Itoa(int(c))
}

// String returns the string representation of a SectionNumber
func (s SectionNumber) String() string {
	if len(s) == 0 {
		return ""
	}
	
	parts := make([]string, len(s))
	for i, num := range s {
		parts[i] = strconv.Itoa(num)
	}
	return strings.Join(parts, ".")
}

// Validate validates a FigureID
func (id FigureID) Validate() error {
	// Expected format: fig-{chapter}.{sequence}
	matched, _ := regexp.MatchString(`^fig-\d+\.\d+$`, string(id))
	if !matched {
		return fmt.Errorf("invalid figure ID format (expected: fig-{chapter}.{sequence})")
	}
	return nil
}

// Validate validates a TableID
func (id TableID) Validate() error {
	// Expected format: table-{chapter}.{sequence}
	matched, _ := regexp.MatchString(`^table-\d+\.\d+$`, string(id))
	if !matched {
		return fmt.Errorf("invalid table ID format (expected: table-{chapter}.{sequence})")
	}
	return nil
}

// AddChapter adds a chapter to the document
func (d *Document) AddChapter(chapter Chapter) {
	d.Chapters = append(d.Chapters, chapter)
	d.UpdatedAt = time.Now()
}

// GetChapter retrieves a chapter by number
func (d *Document) GetChapter(number ChapterNumber) (*Chapter, error) {
	for i := range d.Chapters {
		if d.Chapters[i].Number == number {
			return &d.Chapters[i], nil
		}
	}
	return nil, fmt.Errorf("chapter %d not found", number)
}

// TotalSections returns the total number of sections across all chapters
func (m *Manifest) TotalSections() int {
	total := 0
	for _, count := range m.ChapterCounts {
		total += count.Sections
	}
	return total
}

// TotalFigures returns the total number of figures across all chapters
func (m *Manifest) TotalFigures() int {
	total := 0
	for _, count := range m.ChapterCounts {
		total += count.Figures
	}
	return total
}

// TotalTables returns the total number of tables across all chapters
func (m *Manifest) TotalTables() int {
	total := 0
	for _, count := range m.ChapterCounts {
		total += count.Tables
	}
	return total
}

// GenerateFigureID generates a figure ID for a chapter and sequence
func GenerateFigureID(chapter ChapterNumber, sequence int) FigureID {
	return FigureID(fmt.Sprintf("fig-%d.%d", chapter, sequence))
}

// GenerateTableID generates a table ID for a chapter and sequence
func GenerateTableID(chapter ChapterNumber, sequence int) TableID {
	return TableID(fmt.Sprintf("table-%d.%d", chapter, sequence))
}

// NewSectionNumber creates a new section number
func NewSectionNumber(parts ...int) SectionNumber {
	return SectionNumber(parts)
}

// DefaultStyle returns default document styling
func DefaultStyle() Style {
	return Style{
		Body: TextStyle{
			FontFamily: "Times New Roman",
			FontSize:   "12pt",
			Color:      "#000000",
		},
		Heading: TextStyle{
			FontFamily: "Times New Roman",
			Color:      "#000000",
		},
		Monospace: TextStyle{
			FontFamily: "Courier New",
			FontSize:   "10pt",
			Color:      "#000000",
		},
		LinkColor:   "#0000FF",
		LineSpacing: "1.5",
		Margins: Margins{
			Top:    "1in",
			Bottom: "1in",
			Left:   "1in",
			Right:  "1in",
		},
		HeaderFooter: HeaderFooter{
			HeaderTemplate: "",
			FooterTemplate: "Page {page}",
		},
		NumberingStyle: NumberingStyle{
			Chapters: true,
			Sections: true,
			Figures:  true,
			Tables:   true,
		},
		ReferenceDocx: "",
		StyleCSS:      "",
		LaTeXHeader:   "",
	}
}

// DefaultPandocConfig returns default pandoc configuration
func DefaultPandocConfig() PandocConfig {
	return PandocConfig{
		PDFEngine:     "pdflatex",
		TOC:           true,
		TOCDepth:      3,
		CitationStyle: "apa",
		Args:          []string{},
		Variables:     make(map[string]string),
	}
}

// StyleValidation represents style validation results
type StyleValidation struct {
	Errors   []string `yaml:"errors"`   // Fatal errors that prevent export
	Warnings []string `yaml:"warnings"` // Non-fatal issues with fallbacks
}

// IsValid returns true if there are no fatal errors
func (v *StyleValidation) IsValid() bool {
	return len(v.Errors) == 0
}

// AddError adds a fatal error to the validation result
func (v *StyleValidation) AddError(message string) {
	v.Errors = append(v.Errors, message)
}

// AddWarning adds a warning to the validation result
func (v *StyleValidation) AddWarning(message string) {
	v.Warnings = append(v.Warnings, message)
}

// ValidateStyle validates a style configuration
func ValidateStyle(style *Style) *StyleValidation {
	if style == nil {
		return &StyleValidation{}
	}

	validation := &StyleValidation{}

	// Validate colors
	validateColor(style.Body.Color, "body color", validation)
	validateColor(style.Heading.Color, "heading color", validation)
	validateColor(style.Monospace.Color, "monospace color", validation)
	validateColor(style.LinkColor, "link color", validation)

	// Validate font sizes
	validateFontSize(style.Body.FontSize, "body font size", validation)
	validateFontSize(style.Heading.FontSize, "heading font size", validation)
	validateFontSize(style.Monospace.FontSize, "monospace font size", validation)

	// Validate line spacing
	if style.LineSpacing != "" {
		if !isValidLineSpacing(style.LineSpacing) {
			validation.AddWarning(fmt.Sprintf("Invalid line spacing '%s', using default 1.5", style.LineSpacing))
		}
	}

	// Validate margins
	validateMargin(style.Margins.Top, "top margin", validation)
	validateMargin(style.Margins.Bottom, "bottom margin", validation)
	validateMargin(style.Margins.Left, "left margin", validation)
	validateMargin(style.Margins.Right, "right margin", validation)

	// Validate template strings (using the template package functions)
	if style.HeaderFooter.HeaderTemplate != "" {
		if warnings := validateTemplateString(style.HeaderFooter.HeaderTemplate); len(warnings) > 0 {
			for _, warning := range warnings {
				validation.AddWarning(fmt.Sprintf("Header template: %s", warning))
			}
		}
	}

	if style.HeaderFooter.FooterTemplate != "" {
		if warnings := validateTemplateString(style.HeaderFooter.FooterTemplate); len(warnings) > 0 {
			for _, warning := range warnings {
				validation.AddWarning(fmt.Sprintf("Footer template: %s", warning))
			}
		}
	}

	return validation
}

// validateColor checks if a color format is valid
func validateColor(color, fieldName string, validation *StyleValidation) {
	if color == "" {
		return
	}

	// Check for hex color format
	hexPattern := regexp.MustCompile(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`)
	if hexPattern.MatchString(color) {
		return
	}

	// Check for RGB format
	rgbPattern := regexp.MustCompile(`^rgb\(\s*\d+\s*,\s*\d+\s*,\s*\d+\s*\)$`)
	if rgbPattern.MatchString(color) {
		return
	}

	validation.AddWarning(fmt.Sprintf("Invalid %s format '%s', using default black", fieldName, color))
}

// validateFontSize checks if a font size is valid
func validateFontSize(fontSize, fieldName string, validation *StyleValidation) {
	if fontSize == "" {
		return
	}

	// Check for valid units: pt, px, em, rem, %
	sizePattern := regexp.MustCompile(`^\d+(\.\d+)?(pt|px|em|rem|%)$`)
	if !sizePattern.MatchString(fontSize) {
		validation.AddWarning(fmt.Sprintf("Invalid %s format '%s', should include unit (pt, px, em, rem, %%)", fieldName, fontSize))
	}
}

// validateMargin checks if a margin value is valid
func validateMargin(margin, fieldName string, validation *StyleValidation) {
	if margin == "" {
		return
	}

	// Check for valid units: in, cm, mm, pt, px
	marginPattern := regexp.MustCompile(`^\d+(\.\d+)?(in|cm|mm|pt|px)$`)
	if !marginPattern.MatchString(margin) {
		validation.AddWarning(fmt.Sprintf("Invalid %s format '%s', should include unit (in, cm, mm, pt, px)", fieldName, margin))
	}
}

// isValidLineSpacing checks if line spacing value is valid
func isValidLineSpacing(spacing string) bool {
	// Allow decimal numbers (1.5, 2.0, etc.)
	spacingPattern := regexp.MustCompile(`^\d+(\.\d+)?$`)
	return spacingPattern.MatchString(spacing)
}

// validateTemplateString validates template variable syntax
func validateTemplateString(template string) []string {
	var warnings []string
	
	validVariables := map[string]bool{
		"{page}":           true,
		"{total_pages}":    true,
		"{chapter_title}":  true,
		"{chapter_number}": true,
		"{document_title}": true,
		"{author}":         true,
		"{date}":          true,
		"{section_title}":  true,
	}

	// Find all {variable} patterns
	for i := 0; i < len(template); i++ {
		if template[i] == '{' {
			// Find closing brace
			end := strings.Index(template[i:], "}")
			if end == -1 {
				warnings = append(warnings, fmt.Sprintf("Unclosed variable at position %d", i))
				continue
			}
			
			variable := template[i : i+end+1]
			if !validVariables[variable] {
				warnings = append(warnings, fmt.Sprintf("Unknown template variable: %s", variable))
			}
		}
	}

	return warnings
}