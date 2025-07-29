package document

import (
	"os"
	"testing"

	"github.com/gomcpgo/docgen/pkg/config"
	"github.com/gomcpgo/docgen/pkg/storage"
	"github.com/gomcpgo/docgen/pkg/types"
)

func setupTestManager(t *testing.T) (*Manager, string) {
	tempDir, err := os.MkdirTemp("", "docgen_manager_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cfg := &config.Config{
		RootDir:      tempDir,
		MaxDocuments: 10,
	}

	stor := storage.NewFileSystemStorage(cfg)
	manager := NewManager(cfg, stor)

	return manager, tempDir
}

func TestManager_CreateDocument(t *testing.T) {
	manager, tempDir := setupTestManager(t)
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name    string
		title   string
		author  string
		docType types.DocumentType
		wantErr bool
	}{
		{
			name:    "valid document",
			title:   "Test Document",
			author:  "Test Author",
			docType: types.DocumentTypeBook,
			wantErr: false,
		},
		{
			name:    "empty title",
			title:   "",
			author:  "Test Author",
			docType: types.DocumentTypeBook,
			wantErr: true,
		},
		{
			name:    "empty author",
			title:   "Test Document",
			author:  "",
			docType: types.DocumentTypeBook,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			docID, err := manager.CreateDocument(tt.title, tt.author, tt.docType)
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.CreateDocument() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if docID == "" {
					t.Errorf("Manager.CreateDocument() should return non-empty document ID")
				}

				// Verify document exists
				exists, err := manager.storage.DocumentExists(string(docID))
				if err != nil {
					t.Errorf("Failed to check document existence: %v", err)
				}
				if !exists {
					t.Errorf("Document should exist after creation")
				}
			}
		})
	}
}

func TestManager_GetDocumentStructure(t *testing.T) {
	manager, tempDir := setupTestManager(t)
	defer os.RemoveAll(tempDir)

	// Create a document first
	docID, err := manager.CreateDocument("Test Document", "Test Author", types.DocumentTypeBook)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Add some chapters
	_, err = manager.AddChapter(docID, "Introduction", nil)
	if err != nil {
		t.Fatalf("Failed to add chapter: %v", err)
	}

	_, err = manager.AddChapter(docID, "Methods", nil)
	if err != nil {
		t.Fatalf("Failed to add chapter: %v", err)
	}

	// Get document structure
	structure, err := manager.GetDocumentStructure(docID)
	if err != nil {
		t.Fatalf("GetDocumentStructure() error = %v", err)
	}

	if structure.Document.Title != "Test Document" {
		t.Errorf("Expected title 'Test Document', got %s", structure.Document.Title)
	}

	if len(structure.Document.Chapters) != 2 {
		t.Errorf("Expected 2 chapters, got %d", len(structure.Document.Chapters))
	}

	if structure.Document.Chapters[0].Title != "Introduction" {
		t.Errorf("Expected first chapter title 'Introduction', got %s", structure.Document.Chapters[0].Title)
	}
}

func TestManager_AddChapter(t *testing.T) {
	manager, tempDir := setupTestManager(t)
	defer os.RemoveAll(tempDir)

	// Create a document first
	docID, err := manager.CreateDocument("Test Document", "Test Author", types.DocumentTypeBook)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	tests := []struct {
		name     string
		title    string
		position *int
		wantErr  bool
		wantNum  types.ChapterNumber
	}{
		{
			name:    "first chapter",
			title:   "Introduction",
			wantErr: false,
			wantNum: 1,
		},
		{
			name:    "second chapter",
			title:   "Methods",
			wantErr: false,
			wantNum: 2,
		},
		{
			name:    "empty title",
			title:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chapterNum, err := manager.AddChapter(docID, tt.title, tt.position)
			if (err != nil) != tt.wantErr {
				t.Errorf("Manager.AddChapter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if chapterNum != tt.wantNum {
					t.Errorf("Expected chapter number %d, got %d", tt.wantNum, chapterNum)
				}

				// Verify chapter exists
				exists, err := manager.storage.ChapterExists(string(docID), int(chapterNum))
				if err != nil {
					t.Errorf("Failed to check chapter existence: %v", err)
				}
				if !exists {
					t.Errorf("Chapter should exist after creation")
				}
			}
		})
	}
}

func TestManager_GetChapter(t *testing.T) {
	manager, tempDir := setupTestManager(t)
	defer os.RemoveAll(tempDir)

	// Create a document and chapter
	docID, err := manager.CreateDocument("Test Document", "Test Author", types.DocumentTypeBook)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	chapterNum, err := manager.AddChapter(docID, "Introduction", nil)
	if err != nil {
		t.Fatalf("Failed to add chapter: %v", err)
	}

	// Get the chapter
	chapter, err := manager.GetChapter(docID, chapterNum)
	if err != nil {
		t.Fatalf("GetChapter() error = %v", err)
	}

	if chapter.Title != "Introduction" {
		t.Errorf("Expected chapter title 'Introduction', got %s", chapter.Title)
	}

	if chapter.Number != chapterNum {
		t.Errorf("Expected chapter number %d, got %d", chapterNum, chapter.Number)
	}

	// Test non-existent chapter
	_, err = manager.GetChapter(docID, 999)
	if err == nil {
		t.Errorf("Expected error for non-existent chapter")
	}
}

func TestManager_DeleteDocument(t *testing.T) {
	manager, tempDir := setupTestManager(t)
	defer os.RemoveAll(tempDir)

	// Create a document
	docID, err := manager.CreateDocument("Test Document", "Test Author", types.DocumentTypeBook)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	// Verify it exists
	exists, err := manager.storage.DocumentExists(string(docID))
	if err != nil {
		t.Fatalf("Failed to check document existence: %v", err)
	}
	if !exists {
		t.Fatalf("Document should exist before deletion")
	}

	// Delete the document
	err = manager.DeleteDocument(docID)
	if err != nil {
		t.Errorf("DeleteDocument() error = %v", err)
	}

	// Verify it's gone
	exists, err = manager.storage.DocumentExists(string(docID))
	if err != nil {
		t.Errorf("Failed to check document existence: %v", err)
	}
	if exists {
		t.Errorf("Document should not exist after deletion")
	}
}

func TestManager_UpdateChapterMetadata(t *testing.T) {
	manager, tempDir := setupTestManager(t)
	defer os.RemoveAll(tempDir)

	// Create a document and chapter
	docID, err := manager.CreateDocument("Test Document", "Test Author", types.DocumentTypeBook)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	chapterNum, err := manager.AddChapter(docID, "Introduction", nil)
	if err != nil {
		t.Fatalf("Failed to add chapter: %v", err)
	}

	// Update chapter title
	err = manager.UpdateChapterMetadata(docID, chapterNum, "New Introduction Title")
	if err != nil {
		t.Errorf("UpdateChapterMetadata() error = %v", err)
	}

	// Verify the update
	chapter, err := manager.GetChapter(docID, chapterNum)
	if err != nil {
		t.Fatalf("Failed to get chapter: %v", err)
	}

	if chapter.Title != "New Introduction Title" {
		t.Errorf("Expected updated title 'New Introduction Title', got %s", chapter.Title)
	}
}