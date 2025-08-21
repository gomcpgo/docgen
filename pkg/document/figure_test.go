package document

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gomcpgo/docgen/pkg/types"
)

func TestManager_AddImageWithSection(t *testing.T) {
	manager, tempDir := setupTestManager(t)
	defer os.RemoveAll(tempDir)

	// Create a document
	docID, err := manager.CreateDocument("Test Document", "Test Author", types.DocumentTypeBook)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Add a chapter
	chapterNum, err := manager.AddChapter(docID, "Chapter 1", nil)
	if err != nil {
		t.Fatalf("Failed to add chapter: %v", err)
	}

	// Add sections
	sectionNum1, err := manager.AddSection(docID, chapterNum, "Section 1.1", "Content for section 1.1", 1)
	if err != nil {
		t.Fatalf("Failed to add section 1.1: %v", err)
	}

	sectionNum2, err := manager.AddSection(docID, chapterNum, "Section 1.2", "Content for section 1.2", 1)
	if err != nil {
		t.Fatalf("Failed to add section 1.2: %v", err)
	}

	// Create a test image file
	testImagePath := filepath.Join(tempDir, "test-image.png")
	if err := os.WriteFile(testImagePath, []byte("fake png data"), 0644); err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	tests := []struct {
		name          string
		sectionNumber string
		imagePath     string
		caption       string
		position      string
		wantErr       bool
		errContains   string
	}{
		{
			name:          "valid image at end of section",
			sectionNumber: sectionNum1.String(),
			imagePath:     testImagePath,
			caption:       "Test image caption",
			position:      "end",
			wantErr:       false,
		},
		{
			name:          "valid image at beginning of section",
			sectionNumber: sectionNum2.String(),
			imagePath:     testImagePath,
			caption:       "Another test image",
			position:      "beginning",
			wantErr:       false,
		},
		{
			name:          "missing section number",
			sectionNumber: "",
			imagePath:     testImagePath,
			caption:       "Test caption",
			position:      "end",
			wantErr:       true,
			errContains:   "section number is required",
		},
		{
			name:          "invalid section number",
			sectionNumber: "99.99",
			imagePath:     testImagePath,
			caption:       "Test caption",
			position:      "end",
			wantErr:       true,
			errContains:   "does not exist",
		},
		{
			name:          "invalid position",
			sectionNumber: sectionNum1.String(),
			imagePath:     testImagePath,
			caption:       "Test caption",
			position:      "middle",
			wantErr:       true,
			errContains:   "must be 'beginning' or 'end'",
		},
		{
			name:          "missing image path",
			sectionNumber: sectionNum1.String(),
			imagePath:     "",
			caption:       "Test caption",
			position:      "end",
			wantErr:       true,
			errContains:   "image path is required",
		},
		{
			name:          "missing caption",
			sectionNumber: sectionNum1.String(),
			imagePath:     testImagePath,
			caption:       "",
			position:      "end",
			wantErr:       true,
			errContains:   "caption is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			figureID, err := manager.AddImage(docID, chapterNum, tt.sectionNumber, tt.imagePath, tt.caption, tt.position)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.AddImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" {
				if err == nil || !contains(err.Error(), tt.errContains) {
					t.Errorf("Manager.AddImage() error = %v, should contain %v", err, tt.errContains)
				}
				return
			}

			if !tt.wantErr {
				if figureID == "" {
					t.Errorf("Manager.AddImage() should return non-empty figure ID")
				}

				// Verify figure was added to chapter
				chapter, err := manager.storage.LoadChapterMetadata(string(docID), int(chapterNum))
				if err != nil {
					t.Fatalf("Failed to load chapter: %v", err)
				}

				// Find the figure
				found := false
				for _, fig := range chapter.Figures {
					if fig.ID == figureID {
						found = true
						// Verify figure properties
						if fig.SectionNumber != tt.sectionNumber {
							t.Errorf("Figure section number = %v, want %v", fig.SectionNumber, tt.sectionNumber)
						}
						if fig.Caption != tt.caption {
							t.Errorf("Figure caption = %v, want %v", fig.Caption, tt.caption)
						}
						if string(fig.Position) != tt.position {
							t.Errorf("Figure position = %v, want %v", fig.Position, tt.position)
						}
						// Verify image was copied to assets
						if !filepath.IsAbs(fig.ImagePath) {
							// Should be a relative path starting with assets/images/
							if !contains(fig.ImagePath, "assets") || !contains(fig.ImagePath, "images") {
								t.Errorf("Figure image path should be in assets/images, got %v", fig.ImagePath)
							}
						}
						break
					}
				}

				if !found {
					t.Errorf("Figure with ID %v not found in chapter", figureID)
				}
			}
		})
	}
}

func TestManager_UpdateImageCaption(t *testing.T) {
	manager, tempDir := setupTestManager(t)
	defer os.RemoveAll(tempDir)

	// Create a document with a chapter and section
	docID, _ := manager.CreateDocument("Test Document", "Test Author", types.DocumentTypeBook)
	chapterNum, _ := manager.AddChapter(docID, "Chapter 1", nil)
	sectionNum, _ := manager.AddSection(docID, chapterNum, "Section 1.1", "Content", 1)

	// Create test image
	testImagePath := filepath.Join(tempDir, "test-image.png")
	os.WriteFile(testImagePath, []byte("fake png data"), 0644)

	// Add an image
	figureID, err := manager.AddImage(docID, chapterNum, sectionNum.String(), testImagePath, "Original caption", "end")
	if err != nil {
		t.Fatalf("Failed to add image: %v", err)
	}

	// Test updating caption
	newCaption := "Updated caption"
	err = manager.UpdateImageCaption(docID, figureID, newCaption)
	if err != nil {
		t.Errorf("Failed to update image caption: %v", err)
	}

	// Verify caption was updated
	chapter, _ := manager.storage.LoadChapterMetadata(string(docID), int(chapterNum))
	for _, fig := range chapter.Figures {
		if fig.ID == figureID {
			if fig.Caption != newCaption {
				t.Errorf("Caption not updated: got %v, want %v", fig.Caption, newCaption)
			}
			return
		}
	}
	t.Errorf("Figure %v not found after update", figureID)
}

func TestManager_DeleteImage(t *testing.T) {
	manager, tempDir := setupTestManager(t)
	defer os.RemoveAll(tempDir)

	// Create a document with a chapter and section
	docID, _ := manager.CreateDocument("Test Document", "Test Author", types.DocumentTypeBook)
	chapterNum, _ := manager.AddChapter(docID, "Chapter 1", nil)
	sectionNum, _ := manager.AddSection(docID, chapterNum, "Section 1.1", "Content", 1)

	// Create test images
	testImagePath1 := filepath.Join(tempDir, "test-image1.png")
	testImagePath2 := filepath.Join(tempDir, "test-image2.png")
	os.WriteFile(testImagePath1, []byte("fake png data 1"), 0644)
	os.WriteFile(testImagePath2, []byte("fake png data 2"), 0644)

	// Add two images
	figureID1, _ := manager.AddImage(docID, chapterNum, sectionNum.String(), testImagePath1, "First image", "beginning")
	_, _ = manager.AddImage(docID, chapterNum, sectionNum.String(), testImagePath2, "Second image", "end")

	// Delete the first image
	err := manager.DeleteImage(docID, figureID1)
	if err != nil {
		t.Errorf("Failed to delete image: %v", err)
	}

	// Verify image was deleted and second image was renumbered
	chapter, _ := manager.storage.LoadChapterMetadata(string(docID), int(chapterNum))
	
	if len(chapter.Figures) != 1 {
		t.Errorf("Expected 1 figure after deletion, got %d", len(chapter.Figures))
	}

	// Check that the remaining figure is the second one with updated ID
	if len(chapter.Figures) > 0 {
		remainingFig := chapter.Figures[0]
		if remainingFig.Caption != "Second image" {
			t.Errorf("Wrong figure remained: got caption %v", remainingFig.Caption)
		}
		// The second figure should now have sequence 1
		if remainingFig.Sequence != 1 {
			t.Errorf("Figure sequence not updated: got %d, want 1", remainingFig.Sequence)
		}
		expectedID := types.FigureID("fig-1.1")
		if remainingFig.ID != expectedID {
			t.Errorf("Figure ID not updated: got %v, want %v", remainingFig.ID, expectedID)
		}
	}

	// Try to delete non-existent figure - use a figure ID that was never created
	nonExistentID := types.FigureID("fig-1.99")
	err = manager.DeleteImage(docID, nonExistentID)
	if err == nil {
		t.Errorf("Should fail when deleting non-existent figure")
	} else {
		// This is expected - verify the error message
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("Unexpected error when deleting non-existent figure: %v", err)
		}
	}
}

func TestManager_RebuildChapterWithFigures(t *testing.T) {
	manager, tempDir := setupTestManager(t)
	defer os.RemoveAll(tempDir)

	// Create a document with a chapter and sections
	docID, _ := manager.CreateDocument("Test Document", "Test Author", types.DocumentTypeBook)
	chapterNum, _ := manager.AddChapter(docID, "Chapter 1", nil)
	section1, _ := manager.AddSection(docID, chapterNum, "Section 1.1", "Content for section 1.1", 1)
	section2, _ := manager.AddSection(docID, chapterNum, "Section 1.2", "Content for section 1.2", 1)

	// Create test images
	testImagePath1 := filepath.Join(tempDir, "test-image1.png")
	testImagePath2 := filepath.Join(tempDir, "test-image2.png")
	os.WriteFile(testImagePath1, []byte("fake png data 1"), 0644)
	os.WriteFile(testImagePath2, []byte("fake png data 2"), 0644)

	// Add images to different sections with different positions
	figureID1, _ := manager.AddImage(docID, chapterNum, section1.String(), testImagePath1, "Image at beginning", "beginning")
	figureID2, _ := manager.AddImage(docID, chapterNum, section2.String(), testImagePath2, "Image at end", "end")

	// Rebuild chapter markdown
	err := manager.RebuildChapterMarkdown(docID, chapterNum)
	if err != nil {
		t.Errorf("Failed to rebuild chapter markdown: %v", err)
	}

	// Load the chapter content
	chapterContent, err := manager.storage.LoadChapterContent(string(docID), int(chapterNum))
	if err != nil {
		t.Fatalf("Failed to load chapter content: %v", err)
	}

	// Verify that figures are included in the markdown
	if !contains(chapterContent, string(figureID1)) {
		t.Errorf("Chapter content should contain figure ID %v", figureID1)
	}
	if !contains(chapterContent, string(figureID2)) {
		t.Errorf("Chapter content should contain figure ID %v", figureID2)
	}
	if !contains(chapterContent, "Image at beginning") {
		t.Errorf("Chapter content should contain first image caption")
	}
	if !contains(chapterContent, "Image at end") {
		t.Errorf("Chapter content should contain second image caption")
	}

	// Verify markdown image syntax
	if !contains(chapterContent, "![") {
		t.Errorf("Chapter content should contain markdown image syntax")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}