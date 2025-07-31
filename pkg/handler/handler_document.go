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

		// Parse body text style
		if bodyParams, ok := styleParams["body"].(map[string]interface{}); ok {
			bodyStyle := types.TextStyle{}
			if fontFamily, ok := bodyParams["font_family"].(string); ok {
				bodyStyle.FontFamily = fontFamily
			}
			if fontSize, ok := bodyParams["font_size"].(string); ok {
				bodyStyle.FontSize = fontSize
			}
			if color, ok := bodyParams["color"].(string); ok {
				bodyStyle.Color = color
			}
			style.Body = bodyStyle
		}

		// Parse heading style
		if headingParams, ok := styleParams["heading"].(map[string]interface{}); ok {
			headingStyle := types.TextStyle{}
			if fontFamily, ok := headingParams["font_family"].(string); ok {
				headingStyle.FontFamily = fontFamily
			}
			if fontSize, ok := headingParams["font_size"].(string); ok {
				headingStyle.FontSize = fontSize
			}
			if color, ok := headingParams["color"].(string); ok {
				headingStyle.Color = color
			}
			style.Heading = headingStyle
		}

		// Parse monospace style
		if monospaceParams, ok := styleParams["monospace"].(map[string]interface{}); ok {
			monospaceStyle := types.TextStyle{}
			if fontFamily, ok := monospaceParams["font_family"].(string); ok {
				monospaceStyle.FontFamily = fontFamily
			}
			if fontSize, ok := monospaceParams["font_size"].(string); ok {
				monospaceStyle.FontSize = fontSize
			}
			if color, ok := monospaceParams["color"].(string); ok {
				monospaceStyle.Color = color
			}
			style.Monospace = monospaceStyle
		}

		// Parse global style options
		if linkColor, ok := styleParams["link_color"].(string); ok {
			style.LinkColor = linkColor
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

		// Parse header/footer templates
		if headerFooterParams, ok := styleParams["header_footer"].(map[string]interface{}); ok {
			headerFooter := types.HeaderFooter{}
			if headerTemplate, ok := headerFooterParams["header_template"].(string); ok {
				headerFooter.HeaderTemplate = headerTemplate
			}
			if footerTemplate, ok := headerFooterParams["footer_template"].(string); ok {
				headerFooter.FooterTemplate = footerTemplate
			}
			style.HeaderFooter = headerFooter
		}

		// Parse numbering style
		if numberingParams, ok := styleParams["numbering_style"].(map[string]interface{}); ok {
			numbering := types.NumberingStyle{}
			if chapters, ok := numberingParams["chapters"].(bool); ok {
				numbering.Chapters = chapters
			}
			if sections, ok := numberingParams["sections"].(bool); ok {
				numbering.Sections = sections
			}
			if figures, ok := numberingParams["figures"].(bool); ok {
				numbering.Figures = figures
			}
			if tables, ok := numberingParams["tables"].(bool); ok {
				numbering.Tables = tables
			}
			style.NumberingStyle = numbering
		}

		// Parse output-specific templates
		if referenceDocx, ok := styleParams["reference_docx"].(string); ok {
			style.ReferenceDocx = referenceDocx
		}
		if styleCSS, ok := styleParams["style_css"].(string); ok {
			style.StyleCSS = styleCSS
		}
		if latexHeader, ok := styleParams["latex_header"].(string); ok {
			style.LaTeXHeader = latexHeader
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