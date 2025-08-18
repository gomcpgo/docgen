package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gomcpgo/mcp/pkg/protocol"
	"github.com/gomcpgo/docgen/pkg/types"
	"gopkg.in/yaml.v3"
)

// Style management operations

// handleSaveStyle saves or updates a style
func (h *DocGenHandler) handleSaveStyle(params map[string]interface{}) (*protocol.CallToolResponse, error) {
	// Get style name
	styleName, ok := params["style_name"].(string)
	if !ok || styleName == "" {
		return h.errorResponse("style_name parameter is required")
	}

	// Validate style name (must be valid filename)
	if !isValidStyleName(styleName) {
		return h.errorResponse(fmt.Sprintf("Invalid style name '%s'. Style names must be valid filenames without special characters", styleName))
	}

	// Get style data
	styleData, ok := params["style_data"]
	if !ok {
		return h.errorResponse("style_data parameter is required")
	}

	// Convert style data to Style struct
	var style types.Style
	
	// Handle both direct struct and JSON string
	switch v := styleData.(type) {
	case string:
		// If it's a JSON string, unmarshal it
		if err := json.Unmarshal([]byte(v), &style); err != nil {
			return h.errorResponse(fmt.Sprintf("Failed to parse style_data as JSON: %v", err))
		}
	case map[string]interface{}:
		// If it's already a map, convert it
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return h.errorResponse(fmt.Sprintf("Failed to process style_data: %v", err))
		}
		if err := json.Unmarshal(jsonBytes, &style); err != nil {
			return h.errorResponse(fmt.Sprintf("Failed to parse style_data: %v", err))
		}
	default:
		return h.errorResponse("style_data must be a JSON object or string")
	}

	// Validate style
	validation := types.ValidateStyle(&style)
	if !validation.IsValid() {
		return h.errorResponse(fmt.Sprintf("Style validation failed: %v", validation.Errors))
	}

	// Save style
	if err := h.storage.SaveStyleByName(styleName, &style); err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to save style: %v", err))
	}

	log.Printf("[DOCGEN HANDLER] Saved style '%s'", styleName)

	return h.successResponse(map[string]interface{}{
		"style_name": styleName,
		"message":    fmt.Sprintf("Style '%s' saved successfully", styleName),
		"warnings":   validation.Warnings,
	})
}

// handleLoadStyle loads a specific style
func (h *DocGenHandler) handleLoadStyle(params map[string]interface{}) (*protocol.CallToolResponse, error) {
	// Get style name
	styleName, ok := params["style_name"].(string)
	if !ok || styleName == "" {
		return h.errorResponse("style_name parameter is required")
	}

	// Load style
	style, err := h.storage.LoadStyleByName(styleName)
	if err != nil {
		if os.IsNotExist(err) {
			return h.errorResponse(fmt.Sprintf("Style '%s' not found", styleName))
		}
		return h.errorResponse(fmt.Sprintf("Failed to load style: %v", err))
	}

	log.Printf("[DOCGEN HANDLER] Loaded style '%s'", styleName)

	// Convert style to JSON for response
	styleJSON, err := json.Marshal(style)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to serialize style: %v", err))
	}

	var styleMap map[string]interface{}
	if err := json.Unmarshal(styleJSON, &styleMap); err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to process style: %v", err))
	}

	return h.successResponse(map[string]interface{}{
		"style_name": styleName,
		"style_data": styleMap,
		"message":    fmt.Sprintf("Style '%s' loaded successfully", styleName),
	})
}

// handleListStyles lists all available styles
func (h *DocGenHandler) handleListStyles(params map[string]interface{}) (*protocol.CallToolResponse, error) {
	stylesDir := filepath.Join(h.config.RootDir, "styles")

	// Create styles directory if it doesn't exist
	if err := os.MkdirAll(stylesDir, 0755); err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to create styles directory: %v", err))
	}

	// Ensure default style exists
	if err := h.storage.EnsureDefaultStyle(); err != nil {
		log.Printf("[DOCGEN HANDLER] Warning: Failed to ensure default style: %v", err)
	}

	// Read all .yaml files from styles directory
	entries, err := os.ReadDir(stylesDir)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to read styles directory: %v", err))
	}

	var styles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") {
			styleName := strings.TrimSuffix(entry.Name(), ".yaml")
			styles = append(styles, styleName)
		}
	}

	log.Printf("[DOCGEN HANDLER] Found %d styles", len(styles))

	return h.successResponse(map[string]interface{}{
		"styles":  styles,
		"count":   len(styles),
		"message": fmt.Sprintf("Found %d style(s)", len(styles)),
	})
}

// handleDeleteStyle deletes a style
func (h *DocGenHandler) handleDeleteStyle(params map[string]interface{}) (*protocol.CallToolResponse, error) {
	// Get style name
	styleName, ok := params["style_name"].(string)
	if !ok || styleName == "" {
		return h.errorResponse("style_name parameter is required")
	}

	// Check if style exists
	stylePath := filepath.Join(h.config.RootDir, "styles", fmt.Sprintf("%s.yaml", styleName))
	if _, err := os.Stat(stylePath); os.IsNotExist(err) {
		return h.errorResponse(fmt.Sprintf("Style '%s' not found", styleName))
	}

	// Delete the style file
	if err := os.Remove(stylePath); err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to delete style: %v", err))
	}

	log.Printf("[DOCGEN HANDLER] Deleted style '%s'", styleName)

	return h.successResponse(map[string]interface{}{
		"style_name": styleName,
		"message":    fmt.Sprintf("Style '%s' deleted successfully", styleName),
	})
}

// handleSetDocumentStyle sets the style for a document
func (h *DocGenHandler) handleSetDocumentStyle(params map[string]interface{}) (*protocol.CallToolResponse, error) {
	// Get document ID
	docID, err := h.getDocumentID(params)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Invalid document_id: %v", err))
	}

	// Get style name
	styleName, ok := params["style_name"].(string)
	if !ok || styleName == "" {
		return h.errorResponse("style_name parameter is required")
	}

	// Verify style exists
	if _, err := h.storage.LoadStyleByName(styleName); err != nil {
		if os.IsNotExist(err) {
			return h.errorResponse(fmt.Sprintf("Style '%s' not found", styleName))
		}
		return h.errorResponse(fmt.Sprintf("Failed to verify style: %v", err))
	}

	// Load document manifest
	manifest, err := h.manager.GetDocumentStructure(docID)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to load document: %v", err))
	}

	// Update style name
	manifest.Document.StyleName = styleName
	manifest.UpdatedAt = time.Now()

	// Save updated manifest
	if err := h.storage.SaveManifest(string(docID), manifest); err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to update document: %v", err))
	}

	log.Printf("[DOCGEN HANDLER] Set style '%s' for document %s", styleName, docID)

	return h.successResponse(map[string]interface{}{
		"document_id": string(docID),
		"style_name":  styleName,
		"message":     fmt.Sprintf("Document style set to '%s'", styleName),
	})
}

// handleGetDocumentStyle gets the style for a document
func (h *DocGenHandler) handleGetDocumentStyle(params map[string]interface{}) (*protocol.CallToolResponse, error) {
	// Get document ID
	docID, err := h.getDocumentID(params)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Invalid document_id: %v", err))
	}

	// Load document manifest
	manifest, err := h.manager.GetDocumentStructure(docID)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to load document: %v", err))
	}

	// Get style name (may be empty)
	styleName := manifest.Document.StyleName

	response := map[string]interface{}{
		"document_id": string(docID),
		"style_name":  styleName,
	}

	if styleName == "" {
		response["message"] = "Document has no style set"
	} else {
		response["message"] = fmt.Sprintf("Document uses style '%s'", styleName)
		
		// Also load and return the style data if it exists
		if style, err := h.storage.LoadStyleByName(styleName); err == nil {
			// Convert style to JSON
			styleJSON, _ := json.Marshal(style)
			var styleMap map[string]interface{}
			if json.Unmarshal(styleJSON, &styleMap) == nil {
				response["style_data"] = styleMap
			}
		}
	}

	log.Printf("[DOCGEN HANDLER] Retrieved style for document %s: %s", docID, styleName)

	return h.successResponse(response)
}

// Helper function to validate style names
func isValidStyleName(name string) bool {
	// Style name must be a valid filename
	// No: / \ : * ? " < > |
	// Allow: letters, numbers, spaces, hyphens, underscores, dots
	invalidChars := regexp.MustCompile(`[/\\:*?"<>|]`)
	if invalidChars.MatchString(name) {
		return false
	}
	
	// Must not be empty or just whitespace
	if strings.TrimSpace(name) == "" {
		return false
	}
	
	// Must not start or end with dot
	if strings.HasPrefix(name, ".") || strings.HasSuffix(name, ".") {
		return false
	}
	
	return true
}

// Helper function to validate font size
func isValidFontSize(size string) bool {
	if size == "" {
		return true // Optional
	}
	// Valid patterns: 12pt, 14px, 1.2em, 100%, 1.5rem
	validSize := regexp.MustCompile(`^\d+(\.\d+)?(pt|px|em|rem|%|ex|ch|vw|vh|vmin|vmax)$`)
	return validSize.MatchString(size)
}

// Helper function to convert YAML to JSON
func yamlToJSON(yamlData []byte) ([]byte, error) {
	var data interface{}
	if err := yaml.Unmarshal(yamlData, &data); err != nil {
		return nil, err
	}
	return json.Marshal(data)
}

// Helper function to convert JSON to YAML
func jsonToYAML(jsonData []byte) ([]byte, error) {
	var data interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, err
	}
	return yaml.Marshal(data)
}