package handler

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gomcpgo/mcp/pkg/protocol"
	"github.com/gomcpgo/docgen/pkg/config"
	"github.com/gomcpgo/docgen/pkg/types"
)

// setupTestHandler creates a test handler with temporary directory
func setupTestHandler(t *testing.T) (*DocGenHandler, string) {
	tempDir, err := os.MkdirTemp("", "docgen_handler_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cfg := &config.Config{
		RootDir:       tempDir,
		ExportsDir:    filepath.Join(tempDir, "exports"),
		PandocPath:    "pandoc", // Assume pandoc is available
		ExportTimeout: 30 * time.Second,
		MaxDocuments:  100,
	}

	handler, err := NewDocGenHandler(cfg)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}

	return handler, tempDir
}

// cleanupTestHandler removes temporary directory
func cleanupTestHandler(tempDir string) {
	os.RemoveAll(tempDir)
}

// parseSuccessResponse parses a successful tool response
func parseSuccessResponse(t *testing.T, resp *protocol.CallToolResponse) map[string]interface{} {
	if len(resp.Content) == 0 {
		t.Fatal("Response has no content")
	}

	var result map[string]interface{}
	err := json.Unmarshal([]byte(resp.Content[0].Text), &result)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	return result
}

// expectError checks that the response contains an error
func expectError(t *testing.T, resp *protocol.CallToolResponse, expectedError string) {
	if len(resp.Content) == 0 {
		t.Fatal("Response has no content")
	}

	text := resp.Content[0].Text
	if !contains(text, "Error:") {
		t.Fatalf("Expected error response, got: %s", text)
	}

	if expectedError != "" && !contains(text, expectedError) {
		t.Fatalf("Expected error containing '%s', got: %s", expectedError, text)
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsAt(s, substr, 1)))
}

func containsAt(s, substr string, start int) bool {
	if start >= len(s) {
		return false
	}
	if len(s)-start < len(substr) {
		return false
	}
	if s[start:start+len(substr)] == substr {
		return true
	}
	return containsAt(s, substr, start+1)
}

func TestDocGenHandler_CreateDocument(t *testing.T) {
	handler, tempDir := setupTestHandler(t)
	defer cleanupTestHandler(tempDir)

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid document creation",
			params: map[string]interface{}{
				"title":  "Test Document",
				"author": "Test Author",
				"type":   "book",
			},
			wantError: false,
		},
		{
			name: "missing title",
			params: map[string]interface{}{
				"author": "Test Author",
				"type":   "book",
			},
			wantError: true,
			errorMsg:  "title parameter is required",
		},
		{
			name: "missing author",
			params: map[string]interface{}{
				"title": "Test Document",
				"type":  "book",
			},
			wantError: true,
			errorMsg:  "author parameter is required",
		},
		{
			name: "invalid document type",
			params: map[string]interface{}{
				"title":  "Test Document",
				"author": "Test Author",
				"type":   "invalid",
			},
			wantError: true,
			errorMsg:  "type must be one of",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &protocol.CallToolRequest{
				Name:      "create_document",
				Arguments: tt.params,
			}

			resp, err := handler.CallTool(context.Background(), req)
			if err != nil {
				t.Fatalf("CallTool returned error: %v", err)
			}

			if tt.wantError {
				expectError(t, resp, tt.errorMsg)
			} else {
				result := parseSuccessResponse(t, resp)
				if result["document_id"] == nil {
					t.Error("Expected document_id in response")
				}
				if result["message"] == nil {
					t.Error("Expected message in response")
				}
			}
		})
	}
}

func TestDocGenHandler_AddChapter(t *testing.T) {
	handler, tempDir := setupTestHandler(t)
	defer cleanupTestHandler(tempDir)

	// First create a document
	createReq := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title":  "Test Document",
			"author": "Test Author",
			"type":   "book",
		},
	}

	createResp, err := handler.CallTool(context.Background(), createReq)
	if err != nil {
		t.Fatalf("Failed to create document: %v", err)
	}

	createResult := parseSuccessResponse(t, createResp)
	docID := createResult["document_id"].(string)

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid chapter creation",
			params: map[string]interface{}{
				"document_id": docID,
				"title":       "Chapter 1",
			},
			wantError: false,
		},
		{
			name: "missing document_id",
			params: map[string]interface{}{
				"title": "Chapter 1",
			},
			wantError: true,
			errorMsg:  "document_id parameter is required",
		},
		{
			name: "missing title",
			params: map[string]interface{}{
				"document_id": docID,
			},
			wantError: true,
			errorMsg:  "title parameter is required",
		},
		{
			name: "invalid document_id",
			params: map[string]interface{}{
				"document_id": "invalid-doc-id",
				"title":       "Chapter 1",
			},
			wantError: true,
			errorMsg:  "Failed to add chapter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &protocol.CallToolRequest{
				Name:      "add_chapter",
				Arguments: tt.params,
			}

			resp, err := handler.CallTool(context.Background(), req)
			if err != nil {
				t.Fatalf("CallTool returned error: %v", err)
			}

			if tt.wantError {
				expectError(t, resp, tt.errorMsg)
			} else {
				result := parseSuccessResponse(t, resp)
				if result["chapter_number"] == nil {
					t.Error("Expected chapter_number in response")
				}
				if result["message"] == nil {
					t.Error("Expected message in response")
				}
			}
		})
	}
}

func TestDocGenHandler_AddSection(t *testing.T) {
	handler, tempDir := setupTestHandler(t)
	defer cleanupTestHandler(tempDir)

	// Create document and chapter first
	docID := createTestDocument(t, handler)
	createTestChapter(t, handler, docID)

	tests := []struct {
		name      string
		params    map[string]interface{}
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid section creation",
			params: map[string]interface{}{
				"document_id":    docID,
				"chapter_number": float64(1),
				"title":          "Introduction",
				"content":        "This is the introduction section.",
				"level":          float64(1),
			},
			wantError: false,
		},
		{
			name: "missing content",
			params: map[string]interface{}{
				"document_id":    docID,
				"chapter_number": float64(1),
				"title":          "Introduction",
				"level":          float64(1),
			},
			wantError: true,
			errorMsg:  "content parameter is required",
		},
		{
			name: "invalid level",
			params: map[string]interface{}{
				"document_id":    docID,
				"chapter_number": float64(1),
				"title":          "Introduction",
				"content":        "Content",
				"level":          float64(7), // Invalid level > 6
			},
			wantError: true,
			errorMsg:  "level must be between 1 and 6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &protocol.CallToolRequest{
				Name:      "add_section",
				Arguments: tt.params,
			}

			resp, err := handler.CallTool(context.Background(), req)
			if err != nil {
				t.Fatalf("CallTool returned error: %v", err)
			}

			if tt.wantError {
				expectError(t, resp, tt.errorMsg)
			} else {
				result := parseSuccessResponse(t, resp)
				if result["section_number"] == nil {
					t.Error("Expected section_number in response")
				}
			}
		})
	}
}

func TestDocGenHandler_HelperMethods(t *testing.T) {
	handler, tempDir := setupTestHandler(t)
	defer cleanupTestHandler(tempDir)

	t.Run("getDocumentID", func(t *testing.T) {
		// Valid document ID
		params := map[string]interface{}{
			"document_id": "valid-doc-123",
		}
		docID, err := handler.getDocumentID(params)
		if err != nil {
			t.Errorf("Expected no error for valid document ID, got: %v", err)
		}
		if string(docID) != "valid-doc-123" {
			t.Errorf("Expected 'valid-doc-123', got: %s", string(docID))
		}

		// Missing document ID
		params = map[string]interface{}{}
		_, err = handler.getDocumentID(params)
		if err == nil {
			t.Error("Expected error for missing document_id")
		}
	})

	t.Run("getChapterNumber", func(t *testing.T) {
		// Valid chapter number
		params := map[string]interface{}{
			"chapter_number": float64(5),
		}
		chapterNum, err := handler.getChapterNumber(params)
		if err != nil {
			t.Errorf("Expected no error for valid chapter number, got: %v", err)
		}
		if chapterNum != 5 {
			t.Errorf("Expected 5, got: %d", chapterNum)
		}

		// Invalid chapter number (too small)
		params = map[string]interface{}{
			"chapter_number": float64(0),
		}
		_, err = handler.getChapterNumber(params)
		if err == nil {
			t.Error("Expected error for chapter number 0")
		}
	})

	t.Run("parseSectionNumber", func(t *testing.T) {
		// Valid section number
		sectionNum, err := handler.parseSectionNumber("1.2.3")
		if err != nil {
			t.Errorf("Expected no error for valid section number, got: %v", err)
		}
		expected := types.SectionNumber{1, 2, 3}
		if len(sectionNum) != len(expected) {
			t.Errorf("Expected length %d, got %d", len(expected), len(sectionNum))
		}
		for i, val := range expected {
			if sectionNum[i] != val {
				t.Errorf("Expected %d at index %d, got %d", val, i, sectionNum[i])
			}
		}

		// Invalid section number
		_, err = handler.parseSectionNumber("1")
		if err == nil {
			t.Error("Expected error for invalid section number format")
		}

		_, err = handler.parseSectionNumber("1.abc")
		if err == nil {
			t.Error("Expected error for non-numeric section number")
		}
	})
}

// Helper functions for tests

func createTestDocument(t *testing.T, handler *DocGenHandler) string {
	req := &protocol.CallToolRequest{
		Name: "create_document",
		Arguments: map[string]interface{}{
			"title":  "Test Document",
			"author": "Test Author",
			"type":   "book",
		},
	}

	resp, err := handler.CallTool(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to create test document: %v", err)
	}

	result := parseSuccessResponse(t, resp)
	return result["document_id"].(string)
}

func createTestChapter(t *testing.T, handler *DocGenHandler, docID string) {
	req := &protocol.CallToolRequest{
		Name: "add_chapter",
		Arguments: map[string]interface{}{
			"document_id": docID,
			"title":       "Test Chapter",
		},
	}

	_, err := handler.CallTool(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to create test chapter: %v", err)
	}
}