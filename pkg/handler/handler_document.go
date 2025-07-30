package handler

import (
	"fmt"

	"github.com/gomcpgo/mcp/pkg/protocol"
	"github.com/gomcpgo/docgen/pkg/types"
)

// Document operations

func (h *DocGenHandler) handleCreateDocument(params map[string]interface{}) (*protocol.CallToolResponse, error) {
	// Get title
	title, ok := params["title"].(string)
	if !ok || title == "" {
		return h.errorResponse("title parameter is required")
	}

	// Get author
	author, ok := params["author"].(string)
	if !ok || author == "" {
		return h.errorResponse("author parameter is required")
	}

	// Get document type
	docTypeStr, ok := params["type"].(string)
	if !ok || docTypeStr == "" {
		return h.errorResponse("type parameter is required")
	}

	// Validate document type
	validTypes := map[string]bool{
		"book": true, "report": true, "article": true, "letter": true,
	}
	if !validTypes[docTypeStr] {
		return h.errorResponse("type must be one of: book, report, article, letter")
	}

	docType := types.DocumentType(docTypeStr)

	// Create the document
	docID, err := h.manager.CreateDocument(title, author, docType)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to create document: %v", err))
	}

	return h.successResponse(map[string]interface{}{
		"document_id": docID,
		"message":     fmt.Sprintf("Document '%s' created successfully", title),
	})
}

func (h *DocGenHandler) handleGetDocumentStructure(params map[string]interface{}) (*protocol.CallToolResponse, error) {
	docID, err := h.getDocumentID(params)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Invalid document_id: %v", err))
	}

	manifest, err := h.manager.GetDocumentStructure(docID)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to get document: %v", err))
	}

	return h.successResponse(manifest)
}

func (h *DocGenHandler) handleDeleteDocument(params map[string]interface{}) (*protocol.CallToolResponse, error) {
	docID, err := h.getDocumentID(params)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Invalid document_id: %v", err))
	}

	err = h.manager.DeleteDocument(docID)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to delete document: %v", err))
	}

	return h.successResponse(map[string]interface{}{
		"document_id": docID,
		"message":     fmt.Sprintf("Document %s deleted successfully", docID),
	})
}

func (h *DocGenHandler) handleConfigureDocument(params map[string]interface{}) (*protocol.CallToolResponse, error) {
	docID, err := h.getDocumentID(params)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Invalid document_id: %v", err))
	}

	// Parse style options
	var styleOptions *types.Style
	if styleParams, ok := params["style"].(map[string]interface{}); ok {
		style := &types.Style{}

		if fontFamily, ok := styleParams["font_family"].(string); ok {
			style.FontFamily = fontFamily
		}
		if fontSize, ok := styleParams["font_size"].(string); ok {
			style.FontSize = fontSize
		}
		if lineSpacing, ok := styleParams["line_spacing"].(string); ok {
			style.LineSpacing = lineSpacing
		}

		// Parse margins
		if marginParams, ok := styleParams["margins"].(map[string]interface{}); ok {
			margins := types.Margins{}
			if top, ok := marginParams["top"].(string); ok {
				margins.Top = top
			}
			if bottom, ok := marginParams["bottom"].(string); ok {
				margins.Bottom = bottom
			}
			if left, ok := marginParams["left"].(string); ok {
				margins.Left = left
			}
			if right, ok := marginParams["right"].(string); ok {
				margins.Right = right
			}
			style.Margins = margins
		}

		styleOptions = style
	}

	// Parse pandoc options
	var pandocOptions *types.PandocConfig
	if pandocParams, ok := params["export_settings"].(map[string]interface{}); ok {
		pandoc := &types.PandocConfig{}

		if pdfEngine, ok := pandocParams["pdf_engine"].(string); ok {
			pandoc.PDFEngine = pdfEngine
		}
		if toc, ok := pandocParams["toc"].(bool); ok {
			pandoc.TOC = toc
		}
		if tocDepth, ok := pandocParams["toc_depth"].(float64); ok {
			pandoc.TOCDepth = int(tocDepth)
		}
		if citationStyle, ok := pandocParams["citation_style"].(string); ok {
			pandoc.CitationStyle = citationStyle
		}

		pandocOptions = pandoc
	}

	// Update document configuration
	err = h.manager.ConfigureDocument(docID, styleOptions, pandocOptions)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to configure document: %v", err))
	}

	return h.successResponse(map[string]interface{}{
		"document_id": docID,
		"message":     "Document configuration updated successfully",
	})
}