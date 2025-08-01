package handler

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/gomcpgo/mcp/pkg/protocol"
	"github.com/gomcpgo/docgen/pkg/types"
)

// Export operations

func (h *DocGenHandler) handleExportDocument(params map[string]interface{}) (*protocol.CallToolResponse, error) {
	docID, err := h.getDocumentID(params)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Invalid document_id: %v", err))
	}

	// Get format
	format, ok := params["format"].(string)
	if !ok || format == "" {
		return h.errorResponse("format parameter is required")
	}

	// Validate format
	validFormats := map[string]bool{
		"pdf": true, "docx": true, "html": true,
	}
	if !validFormats[format] {
		return h.errorResponse("format must be one of: pdf, docx, html")
	}

	exportFormat := types.ExportFormat(format)

	// Get chapters (optional)
	var chapters []types.ChapterNumber
	if chaptersParam, ok := params["chapters"].([]interface{}); ok {
		for _, ch := range chaptersParam {
			if chNum, ok := ch.(float64); ok {
				chapters = append(chapters, types.ChapterNumber(chNum))
			}
		}
	}

	// Load document manifest
	manifest, err := h.manager.GetDocumentStructure(docID)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to load document: %v", err))
	}

	// Load document styles and pandoc config (can be nil)
	style, _ := h.storage.LoadStyle(string(docID))
	pandocConfig, _ := h.storage.LoadPandocConfig(string(docID))
	
	// Check for global default style from environment variable
	defaultStylePath := os.Getenv("DOCGEN_DEFAULT_STYLE")
	if defaultStylePath != "" {
		fmt.Printf("[DOCGEN HANDLER] Found DOCGEN_DEFAULT_STYLE: %s\n", defaultStylePath)
		
		// Load global style from file
		globalStyle, err := h.loadGlobalStyleFromFile(defaultStylePath)
		if err != nil {
			fmt.Printf("[DOCGEN HANDLER] Failed to load global style: %v\n", err)
		} else {
			fmt.Printf("[DOCGEN HANDLER] Successfully loaded global style from %s\n", defaultStylePath)
			// Use global style instead of document-specific style
			style = globalStyle
		}
	}
	
	// Log the loaded style for debugging
	if style != nil {
		fmt.Printf("[DOCGEN HANDLER] Loaded style for document %s:\n", docID)
		fmt.Printf("  Body: Font=%s, Size=%s, Color=%s\n", style.Body.FontFamily, style.Body.FontSize, style.Body.Color)
		fmt.Printf("  Heading: Font=%s, Color=%s\n", style.Heading.FontFamily, style.Heading.Color)
		fmt.Printf("  Monospace: Font=%s, Size=%s, Color=%s\n", style.Monospace.FontFamily, style.Monospace.FontSize, style.Monospace.Color)
		fmt.Printf("  Link Color: %s\n", style.LinkColor)
		fmt.Printf("  Line Spacing: %s\n", style.LineSpacing)
	} else {
		fmt.Printf("[DOCGEN HANDLER] No style loaded for document %s, using defaults\n", docID)
	}

	// Create export options
	options := &types.ExportOptions{
		Format:   exportFormat,
		Chapters: chapters,
	}

	// Export the document
	outputPath, err := h.exporter.ExportDocument(string(docID), manifest, style, pandocConfig, options, h.manager.RebuildChapterMarkdown)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to export document: %v", err))
	}

	return h.successResponse(map[string]interface{}{
		"output_path": outputPath,
		"format":      format,
		"message":     fmt.Sprintf("Document exported successfully to %s", outputPath),
	})
}

func (h *DocGenHandler) handlePreviewChapter(params map[string]interface{}) (*protocol.CallToolResponse, error) {
	// Get document ID
	docID, err := h.getDocumentID(params)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Invalid document_id: %v", err))
	}

	// Get chapter number
	chapterNum, err := h.getChapterNumber(params)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Invalid chapter_number: %v", err))
	}

	// Get format (optional, defaults to "html")
	format := "html"
	if fmt, ok := params["format"].(string); ok && fmt != "" {
		format = fmt
	}

	// Validate format
	validFormats := map[string]bool{
		"html": true, "pdf": true,
	}
	if !validFormats[format] {
		return h.errorResponse(fmt.Sprintf("invalid format: %s (must be html or pdf)", format))
	}

	// Preview the chapter
	previewPath, err := h.exporter.PreviewChapter(string(docID), chapterNum, types.ExportFormat(format), h.manager.RebuildChapterMarkdown)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to preview chapter: %v", err))
	}

	return h.successResponse(map[string]interface{}{
		"preview_path": previewPath,
		"format":       format,
		"chapter":      int(chapterNum),
		"message":      fmt.Sprintf("Chapter %d preview generated: %s", chapterNum, previewPath),
	})
}

func (h *DocGenHandler) handleValidateDocument(params map[string]interface{}) (*protocol.CallToolResponse, error) {
	docID, err := h.getDocumentID(params)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Invalid document_id: %v", err))
	}

	// Load document manifest
	manifest, err := h.manager.GetDocumentStructure(docID)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to load document: %v", err))
	}

	// Validate the document
	report := h.exporter.ValidateDocument(string(docID), manifest)

	return h.successResponse(map[string]interface{}{
		"validation_report": report,
		"message":           fmt.Sprintf("Document %s validation completed", docID),
	})
}

// loadGlobalStyleFromFile loads a style from a JSON file path
func (h *DocGenHandler) loadGlobalStyleFromFile(filePath string) (*types.Style, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("style file not found: %s", filePath)
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read style file: %w", err)
	}

	// Parse JSON
	var style types.Style
	if err := json.Unmarshal(content, &style); err != nil {
		return nil, fmt.Errorf("failed to parse style JSON: %w", err)
	}

	return &style, nil
}