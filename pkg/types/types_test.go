package types

import (
	"testing"
)

func TestDocumentID_Validate(t *testing.T) {
	tests := []struct {
		name    string
		id      DocumentID
		wantErr bool
	}{
		{"valid short id", "doc123", false},
		{"valid with dashes", "my-doc-1", false},
		{"empty id", "", true},
		{"too long", "this-is-a-very-long-document-id-that-exceeds-limits", true},
		{"invalid chars", "doc/123", true},
		{"spaces", "doc 123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.id.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("DocumentID.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestChapterNumber_String(t *testing.T) {
	tests := []struct {
		name   string
		number ChapterNumber
		want   string
	}{
		{"chapter 1", 1, "1"},
		{"chapter 10", 10, "10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.number.String(); got != tt.want {
				t.Errorf("ChapterNumber.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSectionNumber_String(t *testing.T) {
	tests := []struct {
		name    string
		section SectionNumber
		want    string
	}{
		{"section 1.1", SectionNumber{1, 1}, "1.1"},
		{"section 2.3", SectionNumber{2, 3}, "2.3"},
		{"subsection 1.2.1", SectionNumber{1, 2, 1}, "1.2.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.section.String(); got != tt.want {
				t.Errorf("SectionNumber.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFigureID_Validate(t *testing.T) {
	tests := []struct {
		name    string
		id      FigureID
		wantErr bool
	}{
		{"valid figure", "fig-1.1", false},
		{"valid double digit", "fig-10.5", false},
		{"invalid format", "figure-1.1", true},
		{"missing dash", "fig1.1", true},
		{"invalid number", "fig-a.1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.id.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("FigureID.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDocument_AddChapter(t *testing.T) {
	doc := &Document{
		ID:       "test-doc",
		Title:    "Test Document",
		Chapters: []Chapter{},
	}

	chapter := Chapter{
		Number: 1,
		Title:  "Introduction",
	}

	doc.AddChapter(chapter)

	if len(doc.Chapters) != 1 {
		t.Errorf("Expected 1 chapter, got %d", len(doc.Chapters))
	}

	if doc.Chapters[0].Title != "Introduction" {
		t.Errorf("Expected chapter title 'Introduction', got %s", doc.Chapters[0].Title)
	}
}

func TestManifest_TotalSections(t *testing.T) {
	manifest := &Manifest{
		ChapterCounts: map[ChapterNumber]ChapterCount{
			1: {Sections: 3, Figures: 2, Tables: 1},
			2: {Sections: 5, Figures: 1, Tables: 0},
		},
	}

	total := manifest.TotalSections()
	expected := 8 // 3 + 5

	if total != expected {
		t.Errorf("Expected total sections %d, got %d", expected, total)
	}
}