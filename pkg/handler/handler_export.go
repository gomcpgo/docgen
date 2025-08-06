package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gomcpgo/mcp/pkg/protocol"
	"github.com/gomcpgo/docgen/pkg/types"
	"gopkg.in/yaml.v3"
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

	// Get style_name (optional)
	var styleName string
	if styleParam, ok := params["style_name"].(string); ok {
		styleName = strings.TrimSpace(styleParam)
	}

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

	// Ensure default style exists
	if err := h.storage.EnsureDefaultStyle(); err != nil {
		log.Printf("[DOCGEN HANDLER] Warning: Failed to ensure default style: %v", err)
	}

	// Load style using enhanced resolution logic
	style, err := h.resolveStyle(styleName)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to load style: %v", err))
	}
	
	// Load pandoc config (can be nil)
	pandocConfig, _ := h.storage.LoadPandocConfig(string(docID))
	
	// Log the loaded style for debugging
	if style != nil {
		log.Printf("[DOCGEN HANDLER] Loaded style for document %s:", docID)
		log.Printf("  Body: Font=%s, Size=%s, Color=%s", style.Body.FontFamily, style.Body.FontSize, style.Body.Color)
		log.Printf("  Heading: Font=%s, Color=%s", style.Heading.FontFamily, style.Heading.Color)
		log.Printf("  Monospace: Font=%s, Size=%s, Color=%s", style.Monospace.FontFamily, style.Monospace.FontSize, style.Monospace.Color)
		log.Printf("  Link Color: %s", style.LinkColor)
		log.Printf("  Line Spacing: %s", style.LineSpacing)
	} else {
		log.Printf("[DOCGEN HANDLER] No style loaded for document %s, using defaults", docID)
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

// resolveStyle implements the enhanced style resolution logic
func (h *DocGenHandler) resolveStyle(styleName string) (*types.Style, error) {
	// Priority 1: If style_name parameter provided, use it
	if styleName != "" {
		log.Printf("[DOCGEN HANDLER] Using provided style name: %s", styleName)
		style, err := h.storage.LoadStyleByName(styleName)
		if err != nil {
			return nil, fmt.Errorf("style '%s' not found: %w", styleName, err)
		}
		return style, nil
	}
	
	// Priority 2: Check DOCGEN_CURRENT_STYLE environment variable
	currentStyle := os.Getenv("DOCGEN_CURRENT_STYLE")
	if currentStyle != "" {
		log.Printf("[DOCGEN HANDLER] Found DOCGEN_CURRENT_STYLE: %s", currentStyle)
		
		// First try as style name (styles/{currentStyle}.yaml)
		style, err := h.storage.LoadStyleByName(currentStyle)
		if err == nil {
			log.Printf("[DOCGEN HANDLER] Successfully loaded style by name: %s", currentStyle)
			return style, nil
		}
		
		// If that fails, try as file path
		if strings.Contains(currentStyle, "/") || strings.Contains(currentStyle, "\\") {
			log.Printf("[DOCGEN HANDLER] Trying DOCGEN_CURRENT_STYLE as file path: %s", currentStyle)
			style, err := h.loadStyleFromFile(currentStyle)
			if err != nil {
				return nil, fmt.Errorf("failed to load style from path '%s': %w", currentStyle, err)
			}
			log.Printf("[DOCGEN HANDLER] Successfully loaded style from file path: %s", currentStyle)
			return style, nil
		}
		
		// Neither worked
		return nil, fmt.Errorf("style '%s' not found as name or file path", currentStyle)
	}
	
	// Priority 3: Use default style (auto-created if needed)
	log.Printf("[DOCGEN HANDLER] Using default style")
	style, err := h.storage.LoadStyleByName("default")
	if err != nil {
		return nil, fmt.Errorf("default style not found: %w", err)
	}
	return style, nil
}

// loadStyleFromFile loads a style from a file path (supports JSON and YAML)
func (h *DocGenHandler) loadStyleFromFile(filePath string) (*types.Style, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("style file not found: %s", filePath)
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read style file: %w", err)
	}

	var style types.Style
	
	// Determine format by file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".json":
		if err := json.Unmarshal(content, &style); err != nil {
			return nil, fmt.Errorf("failed to parse JSON style file: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(content, &style); err != nil {
			return nil, fmt.Errorf("failed to parse YAML style file: %w", err)
		}
	default:
		// Try JSON first, then YAML
		if err := json.Unmarshal(content, &style); err != nil {
			if err2 := yaml.Unmarshal(content, &style); err2 != nil {
				return nil, fmt.Errorf("failed to parse style file as JSON or YAML: JSON error: %v, YAML error: %v", err, err2)
			}
		}
	}

	return &style, nil
}