package handler

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gomcpgo/docgen/pkg/config"
	"github.com/gomcpgo/docgen/pkg/types"
)

// TestIsValidStyleName tests the style name validation
func TestIsValidStyleName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid simple name", "my-style", true},
		{"Valid with underscore", "my_style", true},
		{"Valid with numbers", "style123", true},
		{"Valid with dots", "style.v2", true},
		{"Invalid with slash", "my/style", false},
		{"Invalid with backslash", "my\\style", false},
		{"Invalid with colon", "my:style", false},
		{"Invalid with asterisk", "my*style", false},
		{"Invalid with question", "my?style", false},
		{"Invalid with quotes", "my\"style", false},
		{"Invalid with less than", "my<style", false},
		{"Invalid with greater than", "my>style", false},
		{"Invalid with pipe", "my|style", false},
		{"Invalid empty", "", false},
		{"Invalid spaces only", "   ", false},
		{"Invalid starts with dot", ".style", false},
		{"Invalid ends with dot", "style.", false},
		{"Valid with spaces", "my style", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidStyleName(tt.input)
			if result != tt.expected {
				t.Errorf("isValidStyleName(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestIsValidFontSize tests font size validation
func TestIsValidFontSize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Valid pt", "12pt", true},
		{"Valid px", "14px", true},
		{"Valid em", "1.2em", true},
		{"Valid rem", "1.5rem", true},
		{"Valid percent", "120%", true},
		{"Valid decimal pt", "10.5pt", true},
		{"Valid ex", "2ex", true},
		{"Valid ch", "80ch", true},
		{"Valid vw", "100vw", true},
		{"Valid vh", "50vh", true},
		{"Valid vmin", "10vmin", true},
		{"Valid vmax", "20vmax", true},
		{"Empty is valid", "", true},
		{"Invalid no unit", "12", false},
		{"Invalid unit", "12inches", false},
		{"Invalid format", "pt12", false},
		{"Invalid multiple dots", "12.5.5pt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidFontSize(tt.input)
			if result != tt.expected {
				t.Errorf("isValidFontSize(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestHandleSaveStyle tests the save style handler
func TestHandleSaveStyle(t *testing.T) {
	// Create temporary test directory
	tempDir := t.TempDir()
	
	// Create handler with test config
	cfg := &config.Config{
		RootDir: tempDir,
	}
	h, err := NewDocGenHandler(cfg)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Test valid style save
	params := map[string]interface{}{
		"style_name": "test-style",
		"style_data": map[string]interface{}{
			"body": map[string]interface{}{
				"font_family": "Arial",
				"font_size":   "12pt",
				"color":       "#000000",
			},
			"heading": map[string]interface{}{
				"font_family": "Helvetica",
				"color":       "#333333",
			},
			"margins": map[string]interface{}{
				"top":    "1in",
				"bottom": "1in",
				"left":   "1in",
				"right":  "1in",
			},
		},
	}

	response, err := h.handleSaveStyle(params)
	if err != nil {
		t.Fatalf("handleSaveStyle failed: %v", err)
	}

	if response.IsError {
		t.Errorf("Expected success but got error: %s", response.Content[0].Text)
	}

	// Verify file was created
	stylePath := filepath.Join(tempDir, "styles", "test-style.yaml")
	if _, err := os.Stat(stylePath); os.IsNotExist(err) {
		t.Errorf("Style file was not created at %s", stylePath)
	}

	// Test invalid style name
	params = map[string]interface{}{
		"style_name": "test/style",
		"style_data": map[string]interface{}{
			"body": map[string]interface{}{
				"font_family": "Arial",
			},
		},
	}

	response, err = h.handleSaveStyle(params)
	if err != nil {
		t.Fatalf("handleSaveStyle failed: %v", err)
	}

	if !response.IsError {
		t.Errorf("Expected error for invalid style name but got success")
	}
}

// TestHandleLoadStyle tests the load style handler
func TestHandleLoadStyle(t *testing.T) {
	// Create temporary test directory
	tempDir := t.TempDir()
	
	// Create handler with test config
	cfg := &config.Config{
		RootDir: tempDir,
	}
	h, err := NewDocGenHandler(cfg)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// First save a style
	testStyle := &types.Style{
		Body: types.TextStyle{
			FontFamily: "Georgia",
			FontSize:   "11pt",
			Color:      "#111111",
		},
		Heading: types.TextStyle{
			FontFamily: "Arial",
			Color:      "#222222",
		},
	}

	if err := h.storage.SaveStyleByName("load-test", testStyle); err != nil {
		t.Fatalf("Failed to save test style: %v", err)
	}

	// Test loading the style
	params := map[string]interface{}{
		"style_name": "load-test",
	}

	response, err := h.handleLoadStyle(params)
	if err != nil {
		t.Fatalf("handleLoadStyle failed: %v", err)
	}

	if response.IsError {
		t.Errorf("Expected success but got error: %s", response.Content[0].Text)
	}

	// Test loading non-existent style
	params = map[string]interface{}{
		"style_name": "non-existent",
	}

	response, err = h.handleLoadStyle(params)
	if err != nil {
		t.Fatalf("handleLoadStyle failed: %v", err)
	}

	if !response.IsError {
		t.Errorf("Expected error for non-existent style but got success")
	}
}

// TestHandleListStyles tests the list styles handler
func TestHandleListStyles(t *testing.T) {
	// Create temporary test directory
	tempDir := t.TempDir()
	
	// Create handler with test config
	cfg := &config.Config{
		RootDir: tempDir,
	}
	h, err := NewDocGenHandler(cfg)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Save some test styles
	testStyle := &types.Style{
		Body: types.TextStyle{
			FontFamily: "Arial",
		},
	}

	h.storage.SaveStyleByName("style1", testStyle)
	h.storage.SaveStyleByName("style2", testStyle)

	// Test listing styles
	params := map[string]interface{}{}

	response, err := h.handleListStyles(params)
	if err != nil {
		t.Fatalf("handleListStyles failed: %v", err)
	}

	if response.IsError {
		t.Errorf("Expected success but got error: %s", response.Content[0].Text)
	}
}

// TestHandleDeleteStyle tests the delete style handler
func TestHandleDeleteStyle(t *testing.T) {
	// Create temporary test directory
	tempDir := t.TempDir()
	
	// Create handler with test config
	cfg := &config.Config{
		RootDir: tempDir,
	}
	h, err := NewDocGenHandler(cfg)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Save a test style
	testStyle := &types.Style{
		Body: types.TextStyle{
			FontFamily: "Arial",
		},
	}

	if err := h.storage.SaveStyleByName("delete-test", testStyle); err != nil {
		t.Fatalf("Failed to save test style: %v", err)
	}

	// Test deleting the style
	params := map[string]interface{}{
		"style_name": "delete-test",
	}

	response, err := h.handleDeleteStyle(params)
	if err != nil {
		t.Fatalf("handleDeleteStyle failed: %v", err)
	}

	if response.IsError {
		t.Errorf("Expected success but got error: %s", response.Content[0].Text)
	}

	// Verify file was deleted
	stylePath := filepath.Join(tempDir, "styles", "delete-test.yaml")
	if _, err := os.Stat(stylePath); !os.IsNotExist(err) {
		t.Errorf("Style file was not deleted at %s", stylePath)
	}

	// Test deleting non-existent style
	params = map[string]interface{}{
		"style_name": "non-existent",
	}

	response, err = h.handleDeleteStyle(params)
	if err != nil {
		t.Fatalf("handleDeleteStyle failed: %v", err)
	}

	if !response.IsError {
		t.Errorf("Expected error for non-existent style but got success")
	}
}

// TestDocumentStyleOperations tests set and get document style
func TestDocumentStyleOperations(t *testing.T) {
	// Create temporary test directory
	tempDir := t.TempDir()
	
	// Create handler with test config
	cfg := &config.Config{
		RootDir: tempDir,
	}
	h, err := NewDocGenHandler(cfg)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	// Create a test document
	docID := types.DocumentID("test-doc-123")
	manifest := &types.Manifest{
		Document: types.Document{
			ID:     docID,
			Title:  "Test Document",
			Author: "Test Author",
			Type:   types.DocumentType("report"),
		},
	}

	if err := h.storage.SaveManifest(string(docID), manifest); err != nil {
		t.Fatalf("Failed to save test document: %v", err)
	}

	// Save a test style
	testStyle := &types.Style{
		Body: types.TextStyle{
			FontFamily: "Arial",
		},
	}

	if err := h.storage.SaveStyleByName("doc-style", testStyle); err != nil {
		t.Fatalf("Failed to save test style: %v", err)
	}

	// Test setting document style
	params := map[string]interface{}{
		"document_id": string(docID),
		"style_name":  "doc-style",
	}

	response, err := h.handleSetDocumentStyle(params)
	if err != nil {
		t.Fatalf("handleSetDocumentStyle failed: %v", err)
	}

	if response.IsError {
		t.Errorf("Expected success but got error: %s", response.Content[0].Text)
	}

	// Test getting document style
	params = map[string]interface{}{
		"document_id": string(docID),
	}

	response, err = h.handleGetDocumentStyle(params)
	if err != nil {
		t.Fatalf("handleGetDocumentStyle failed: %v", err)
	}

	if response.IsError {
		t.Errorf("Expected success but got error: %s", response.Content[0].Text)
	}

	// Test setting non-existent style
	params = map[string]interface{}{
		"document_id": string(docID),
		"style_name":  "non-existent",
	}

	response, err = h.handleSetDocumentStyle(params)
	if err != nil {
		t.Fatalf("handleSetDocumentStyle failed: %v", err)
	}

	if !response.IsError {
		t.Errorf("Expected error for non-existent style but got success")
	}
}