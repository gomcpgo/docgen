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

// Style represents document styling configuration
type Style struct {
	FontFamily    string `yaml:"font_family"`
	FontSize      string `yaml:"font_size"`
	Margins       Margins `yaml:"margins"`
	LineSpacing   string `yaml:"line_spacing"`
	HeaderFooter  HeaderFooter `yaml:"header_footer"`
	NumberingStyle NumberingStyle `yaml:"numbering_style"`
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
	Header string `yaml:"header"`
	Footer string `yaml:"footer"`
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
		FontFamily:  "Times New Roman",
		FontSize:    "12pt",
		LineSpacing: "1.5",
		Margins: Margins{
			Top:    "1in",
			Bottom: "1in",
			Left:   "1in",
			Right:  "1in",
		},
		HeaderFooter: HeaderFooter{
			Header: "",
			Footer: "Page \\thepage",
		},
		NumberingStyle: NumberingStyle{
			Chapters: true,
			Sections: true,
			Figures:  true,
			Tables:   true,
		},
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