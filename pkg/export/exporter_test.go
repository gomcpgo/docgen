package export

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gomcpgo/docgen/pkg/config"
	"github.com/gomcpgo/docgen/pkg/types"
)

func setupTestExporter(t *testing.T) (*Exporter, string) {
	tempDir, err := os.MkdirTemp("", "docgen_export_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cfg := &config.Config{
		RootDir:       tempDir,
		PandocPath:    "pandoc", // Assume pandoc is available in test environment
		TempDir:       tempDir,
		ExportTimeout: 30,
	}

	exporter := NewExporter(cfg)
	return exporter, tempDir
}

func createTestDocument(t *testing.T, tempDir string) (*types.Document, *types.Manifest, *types.Style, *types.PandocConfig) {
	doc := &types.Document{
		ID:     "test-doc",
		Title:  "Test Document",
		Author: "Test Author",
		Type:   types.DocumentTypeBook,
		Chapters: []types.Chapter{
			{
				Number:  1,
				Title:   "Introduction",
				Content: "# Introduction\n\nThis is the introduction chapter.",
			},
			{
				Number:  2,
				Title:   "Methods",
				Content: "# Methods\n\nThis chapter describes the methods used.",
			},
		},
	}

	manifest := &types.Manifest{
		Document: *doc,
		ChapterCounts: map[types.ChapterNumber]types.ChapterCount{
			1: {Sections: 1, Figures: 0, Tables: 0},
			2: {Sections: 1, Figures: 0, Tables: 0},
		},
	}

	style := &types.Style{
		FontFamily:  "Times New Roman",
		FontSize:    "12pt",
		LineSpacing: "1.5",
	}

	pandocConfig := &types.PandocConfig{
		PDFEngine: "pdflatex",
		TOC:       true,
		TOCDepth:  3,
	}

	return doc, manifest, style, pandocConfig
}

func TestExporter_GenerateMarkdown(t *testing.T) {
	exporter, tempDir := setupTestExporter(t)
	defer os.RemoveAll(tempDir)

	doc, manifest, _, _ := createTestDocument(t, tempDir)

	// Create chapter content files
	docPath := filepath.Join(tempDir, "test-doc")
	chapter1Path := filepath.Join(docPath, "chapters", "01")
	chapter2Path := filepath.Join(docPath, "chapters", "02")

	os.MkdirAll(chapter1Path, 0755)
	os.MkdirAll(chapter2Path, 0755)

	os.WriteFile(filepath.Join(chapter1Path, "chapter.md"), []byte(doc.Chapters[0].Content), 0644)
	os.WriteFile(filepath.Join(chapter2Path, "chapter.md"), []byte(doc.Chapters[1].Content), 0644)

	options := &types.ExportOptions{
		Format: types.ExportFormatPDF,
	}

	markdown, err := exporter.GenerateMarkdown("test-doc", manifest, options)
	if err != nil {
		t.Fatalf("GenerateMarkdown() error = %v", err)
	}

	// Check that the markdown contains both chapters
	if !strings.Contains(markdown, "# Introduction") {
		t.Errorf("Generated markdown should contain Introduction chapter")
	}

	if !strings.Contains(markdown, "# Methods") {
		t.Errorf("Generated markdown should contain Methods chapter")
	}

	if !strings.Contains(markdown, "This is the introduction chapter") {
		t.Errorf("Generated markdown should contain chapter content")
	}
}

func TestExporter_GeneratePandocCommand(t *testing.T) {
	exporter, tempDir := setupTestExporter(t)
	defer os.RemoveAll(tempDir)

	_, manifest, style, pandocConfig := createTestDocument(t, tempDir)

	options := &types.ExportOptions{
		Format: types.ExportFormatPDF,
	}

	inputFile := filepath.Join(tempDir, "input.md")
	outputFile := filepath.Join(tempDir, "output.pdf")

	cmd := exporter.GeneratePandocCommand("test-doc", inputFile, outputFile, manifest, style, pandocConfig, options)

	// Check basic command structure (path might be full path to pandoc)
	if !strings.Contains(cmd.Path, "pandoc") {
		t.Errorf("Expected pandoc command, got %s", cmd.Path)
	}

	// Check that input and output files are included
	args := strings.Join(cmd.Args, " ")
	if !strings.Contains(args, inputFile) {
		t.Errorf("Command should contain input file: %s", inputFile)
	}

	if !strings.Contains(args, outputFile) {
		t.Errorf("Command should contain output file: %s", outputFile)
	}

	// Check for TOC option
	if !strings.Contains(args, "--toc") {
		t.Errorf("Command should contain --toc option")
	}
}

func TestExporter_ValidateDocument(t *testing.T) {
	exporter, tempDir := setupTestExporter(t)
	defer os.RemoveAll(tempDir)

	_, manifest, _, _ := createTestDocument(t, tempDir)

	tests := []struct {
		name    string
		setup   func()
		wantErr bool
	}{
		{
			name: "valid document",
			setup: func() {
				// Create chapter directories and files
				docPath := filepath.Join(tempDir, "test-doc")
				chapter1Path := filepath.Join(docPath, "chapters", "01")
				chapter2Path := filepath.Join(docPath, "chapters", "02")

				os.MkdirAll(chapter1Path, 0755)
				os.MkdirAll(chapter2Path, 0755)

				os.WriteFile(filepath.Join(chapter1Path, "chapter.md"), []byte("# Chapter 1"), 0644)
				os.WriteFile(filepath.Join(chapter2Path, "chapter.md"), []byte("# Chapter 2"), 0644)
				
				// Create manifest file
				os.WriteFile(filepath.Join(docPath, "manifest.yaml"), []byte("document:\n  title: Test"), 0644)
			},
			wantErr: false,
		},
		{
			name: "missing chapter file",
			setup: func() {
				// Create directory but not the chapter.md file
				docPath := filepath.Join(tempDir, "test-doc")
				chapter1Path := filepath.Join(docPath, "chapters", "01")
				os.MkdirAll(chapter1Path, 0755)
				// Don't create chapter.md
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up
			os.RemoveAll(filepath.Join(tempDir, "test-doc"))
			
			// Setup test case
			tt.setup()

			report := exporter.ValidateDocument("test-doc", manifest)
			if (len(report.Errors) > 0) != tt.wantErr {
				t.Errorf("ValidateDocument() errors = %v, wantErr %v", report.Errors, tt.wantErr)
			}
		})
	}
}

func TestGenerateYAMLMetadata(t *testing.T) {
	doc := &types.Document{
		Title:  "Test Document",
		Author: "Test Author",
		Type:   types.DocumentTypeBook,
	}

	style := &types.Style{
		FontFamily: "Times New Roman",
		FontSize:   "12pt",
	}

	yaml := generateYAMLMetadata(doc, style)

	if !strings.Contains(yaml, "title: Test Document") {
		t.Errorf("YAML should contain document title")
	}

	if !strings.Contains(yaml, "author: Test Author") {
		t.Errorf("YAML should contain document author")
	}

	if !strings.Contains(yaml, "documentclass: book") {
		t.Errorf("YAML should contain document class")
	}
}