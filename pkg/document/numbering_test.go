package document

import (
	"fmt"
	"os"
	"testing"

	"github.com/gomcpgo/docgen/pkg/types"
)

func TestRenumberChapters(t *testing.T) {
	manager, tempDir := setupTestManager(t)
	defer os.RemoveAll(tempDir)

	// Create a document with several chapters
	docID, err := manager.CreateDocument("Test Document", "Test Author", types.DocumentTypeBook)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Add chapters 1, 2, 3
	for i := 1; i <= 3; i++ {
		_, err := manager.AddChapter(docID, fmt.Sprintf("Chapter %d", i), nil)
		if err != nil {
			t.Fatalf("Failed to add chapter %d: %v", i, err)
		}
	}

	// Load the manifest to use for renumbering
	manifest, err := manager.storage.LoadManifest(string(docID))
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	// Test renumbering: shift chapters 2 and 3 up by 1 (insert at position 2)
	err = manager.renumberChapters(string(docID), manifest, 2, 1)
	if err != nil {
		t.Errorf("renumberChapters() error = %v", err)
	}

	// Verify that chapter directories were renamed correctly
	// Original chapter 2 should now be chapter 3
	exists, err := manager.storage.ChapterExists(string(docID), 3)
	if err != nil {
		t.Errorf("Failed to check chapter 3 existence: %v", err)
	}
	if !exists {
		t.Errorf("Chapter 3 should exist after renumbering")
	}

	// Original chapter 3 should now be chapter 4
	exists, err = manager.storage.ChapterExists(string(docID), 4)
	if err != nil {
		t.Errorf("Failed to check chapter 4 existence: %v", err)
	}
	if !exists {
		t.Errorf("Chapter 4 should exist after renumbering")
	}
}

func TestGenerateFigureSequence(t *testing.T) {
	tests := []struct {
		name     string
		existing []types.Figure
		want     int
	}{
		{
			name:     "no existing figures",
			existing: []types.Figure{},
			want:     1,
		},
		{
			name: "existing figures",
			existing: []types.Figure{
				{ID: "fig-1.1", Sequence: 1},
				{ID: "fig-1.2", Sequence: 2},
			},
			want: 3,
		},
		{
			name: "gaps in sequence",
			existing: []types.Figure{
				{ID: "fig-1.1", Sequence: 1},
				{ID: "fig-1.3", Sequence: 3},
			},
			want: 4, // Should continue from highest, not fill gaps
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateFigureSequence(tt.existing)
			if got != tt.want {
				t.Errorf("generateFigureSequence() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateTableSequence(t *testing.T) {
	tests := []struct {
		name     string
		existing []types.Table
		want     int
	}{
		{
			name:     "no existing tables",
			existing: []types.Table{},
			want:     1,
		},
		{
			name: "existing tables",
			existing: []types.Table{
				{ID: "table-1.1", Sequence: 1},
				{ID: "table-1.2", Sequence: 2},
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateTableSequence(tt.existing)
			if got != tt.want {
				t.Errorf("generateTableSequence() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRenumberFigures(t *testing.T) {
	figures := []types.Figure{
		{ID: "fig-2.1", Chapter: 2, Sequence: 1},
		{ID: "fig-2.2", Chapter: 2, Sequence: 2},
		{ID: "fig-3.1", Chapter: 3, Sequence: 1},
	}

	// Renumber chapter 2 to chapter 1 (shift down)
	updated := renumberFigures(figures, 2, -1)

	// Check that chapter 2 figures are now chapter 1
	found1_1 := false
	found1_2 := false
	found3_1 := false

	for _, fig := range updated {
		switch fig.ID {
		case "fig-1.1":
			if fig.Chapter == 1 && fig.Sequence == 1 {
				found1_1 = true
			}
		case "fig-1.2":
			if fig.Chapter == 1 && fig.Sequence == 2 {
				found1_2 = true
			}
		case "fig-3.1":
			if fig.Chapter == 3 && fig.Sequence == 1 {
				found3_1 = true
			}
		}
	}

	if !found1_1 {
		t.Errorf("Expected to find fig-1.1 after renumbering")
	}
	if !found1_2 {
		t.Errorf("Expected to find fig-1.2 after renumbering")
	}
	if !found3_1 {
		t.Errorf("Expected to find fig-3.1 unchanged")
	}
}