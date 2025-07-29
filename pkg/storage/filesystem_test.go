package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gomcpgo/docgen/pkg/config"
	"github.com/gomcpgo/docgen/pkg/types"
)

func TestFileSystemStorage_CreateDocumentStructure(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "docgen_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.Config{
		RootDir: tempDir,
	}

	storage := NewFileSystemStorage(cfg)

	docID := types.DocumentID("test-doc")
	doc := &types.Document{
		ID:     docID,
		Title:  "Test Document",
		Author: "Test Author",
		Type:   types.DocumentTypeBook,
	}

	err = storage.CreateDocumentStructure(doc)
	if err != nil {
		t.Fatalf("CreateDocumentStructure() error = %v", err)
	}

	// Check if document directory exists
	docPath := cfg.DocumentPath(string(docID))
	if _, err := os.Stat(docPath); os.IsNotExist(err) {
		t.Errorf("Document directory not created: %s", docPath)
	}

	// Check if chapters directory exists
	chaptersPath := filepath.Join(docPath, "chapters")
	if _, err := os.Stat(chaptersPath); os.IsNotExist(err) {
		t.Errorf("Chapters directory not created: %s", chaptersPath)
	}

	// Check if assets/images directory exists
	assetsPath := cfg.AssetsPath(string(docID))
	if _, err := os.Stat(assetsPath); os.IsNotExist(err) {
		t.Errorf("Assets directory not created: %s", assetsPath)
	}

	// Check if manifest file exists
	manifestPath := cfg.ManifestPath(string(docID))
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Errorf("Manifest file not created: %s", manifestPath)
	}

	// Check if style file exists
	stylePath := cfg.StylePath(string(docID))
	if _, err := os.Stat(stylePath); os.IsNotExist(err) {
		t.Errorf("Style file not created: %s", stylePath)
	}

	// Check if pandoc config file exists
	pandocPath := cfg.PandocConfigPath(string(docID))
	if _, err := os.Stat(pandocPath); os.IsNotExist(err) {
		t.Errorf("Pandoc config file not created: %s", pandocPath)
	}
}

func TestFileSystemStorage_DocumentExists(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "docgen_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.Config{
		RootDir: tempDir,
	}

	storage := NewFileSystemStorage(cfg)

	// Test non-existent document
	exists, err := storage.DocumentExists("non-existent")
	if err != nil {
		t.Errorf("DocumentExists() error = %v", err)
	}
	if exists {
		t.Errorf("DocumentExists() should return false for non-existent document")
	}

	// Create a document and test existence
	docID := "test-doc"
	docPath := cfg.DocumentPath(docID)
	err = os.MkdirAll(docPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create document directory: %v", err)
	}

	exists, err = storage.DocumentExists(docID)
	if err != nil {
		t.Errorf("DocumentExists() error = %v", err)
	}
	if !exists {
		t.Errorf("DocumentExists() should return true for existing document")
	}
}

func TestFileSystemStorage_DeleteDocument(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "docgen_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.Config{
		RootDir: tempDir,
	}

	storage := NewFileSystemStorage(cfg)

	// Create a document first
	docID := "test-doc"
	doc := &types.Document{
		ID:     types.DocumentID(docID),
		Title:  "Test Document",
		Author: "Test Author",
		Type:   types.DocumentTypeBook,
	}

	err = storage.CreateDocumentStructure(doc)
	if err != nil {
		t.Fatalf("CreateDocumentStructure() error = %v", err)
	}

	// Verify it exists
	exists, err := storage.DocumentExists(docID)
	if err != nil {
		t.Fatalf("DocumentExists() error = %v", err)
	}
	if !exists {
		t.Fatalf("Document should exist before deletion")
	}

	// Delete the document
	err = storage.DeleteDocument(docID)
	if err != nil {
		t.Errorf("DeleteDocument() error = %v", err)
	}

	// Verify it's gone
	exists, err = storage.DocumentExists(docID)
	if err != nil {
		t.Errorf("DocumentExists() error = %v", err)
	}
	if exists {
		t.Errorf("Document should not exist after deletion")
	}
}

func TestFileSystemStorage_CreateChapterStructure(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "docgen_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.Config{
		RootDir: tempDir,
	}

	storage := NewFileSystemStorage(cfg)

	// Create document first
	docID := "test-doc"
	doc := &types.Document{
		ID:     types.DocumentID(docID),
		Title:  "Test Document",
		Author: "Test Author",
		Type:   types.DocumentTypeBook,
	}

	err = storage.CreateDocumentStructure(doc)
	if err != nil {
		t.Fatalf("CreateDocumentStructure() error = %v", err)
	}

	// Create chapter
	chapter := &types.Chapter{
		Number:  1,
		Title:   "Introduction",
		Content: "This is the introduction chapter.",
	}

	err = storage.CreateChapterStructure(docID, chapter)
	if err != nil {
		t.Fatalf("CreateChapterStructure() error = %v", err)
	}

	// Check if chapter directory exists
	chapterPath := cfg.ChapterPath(docID, int(chapter.Number))
	if _, err := os.Stat(chapterPath); os.IsNotExist(err) {
		t.Errorf("Chapter directory not created: %s", chapterPath)
	}

	// Check if chapter.md file exists
	contentPath := cfg.ChapterContentPath(docID, int(chapter.Number))
	if _, err := os.Stat(contentPath); os.IsNotExist(err) {
		t.Errorf("Chapter content file not created: %s", contentPath)
	}

	// Check if metadata.yaml file exists
	metadataPath := cfg.ChapterMetadataPath(docID, int(chapter.Number))
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		t.Errorf("Chapter metadata file not created: %s", metadataPath)
	}
}