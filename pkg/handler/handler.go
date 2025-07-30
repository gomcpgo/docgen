package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gomcpgo/mcp/pkg/protocol"
	"github.com/gomcpgo/docgen/pkg/config"
	"github.com/gomcpgo/docgen/pkg/document"
	"github.com/gomcpgo/docgen/pkg/export"
	"github.com/gomcpgo/docgen/pkg/storage"
	"github.com/gomcpgo/docgen/pkg/types"
)

// DocGenHandler implements the MCP handler for document generation
type DocGenHandler struct {
	config   *config.Config
	manager  *document.Manager
	exporter *export.Exporter
	storage  storage.Storage
}

// NewDocGenHandler creates a new document generation handler
func NewDocGenHandler(cfg *config.Config) (*DocGenHandler, error) {
	stor := storage.NewFileSystemStorage(cfg)
	manager := document.NewManager(cfg, stor)
	exporter := export.NewExporter(cfg)

	return &DocGenHandler{
		config:   cfg,
		manager:  manager,
		exporter: exporter,
		storage:  stor,
	}, nil
}

// CallTool executes a tool
func (h *DocGenHandler) CallTool(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResponse, error) {
	switch req.Name {
	case "create_document":
		return h.handleCreateDocument(req.Arguments)
	case "get_document_structure":
		return h.handleGetDocumentStructure(req.Arguments)
	case "delete_document":
		return h.handleDeleteDocument(req.Arguments)
	case "configure_document":
		return h.handleConfigureDocument(req.Arguments)
	case "add_chapter":
		return h.handleAddChapter(req.Arguments)
	case "get_chapter":
		return h.handleGetChapter(req.Arguments)
	case "update_chapter_metadata":
		return h.handleUpdateChapterMetadata(req.Arguments)
	case "delete_chapter":
		return h.handleDeleteChapter(req.Arguments)
	case "move_chapter":
		return h.handleMoveChapter(req.Arguments)
	case "add_section":
		return h.handleAddSection(req.Arguments)
	case "update_section":
		return h.handleUpdateSection(req.Arguments)
	case "delete_section":
		return h.handleDeleteSection(req.Arguments)
	case "add_image":
		return h.handleAddImage(req.Arguments)
	case "update_image_caption":
		return h.handleUpdateImageCaption(req.Arguments)
	case "delete_image":
		return h.handleDeleteImage(req.Arguments)
	case "export_document":
		return h.handleExportDocument(req.Arguments)
	case "preview_chapter":
		return h.handlePreviewChapter(req.Arguments)
	case "validate_document":
		return h.handleValidateDocument(req.Arguments)
	default:
		return &protocol.CallToolResponse{
			Content: []protocol.ToolContent{
				{
					Type: "text",
					Text: fmt.Sprintf("Unknown tool: %s", req.Name),
				},
			},
			IsError: true,
		}, nil
	}
}

// handleCreateDocument creates a new document
func (h *DocGenHandler) handleCreateDocument(params map[string]interface{}) (*protocol.CallToolResponse, error) {
	title, ok := params["title"].(string)
	if !ok || title == "" {
		return h.errorResponse("title parameter is required and must be a non-empty string")
	}

	author, ok := params["author"].(string)
	if !ok || author == "" {
		return h.errorResponse("author parameter is required and must be a non-empty string")
	}

	docTypeStr, ok := params["type"].(string)
	if !ok || docTypeStr == "" {
		return h.errorResponse("type parameter is required and must be a non-empty string")
	}

	docType := types.DocumentType(docTypeStr)
	if docType != types.DocumentTypeBook && docType != types.DocumentTypeReport && 
	   docType != types.DocumentTypeArticle && docType != types.DocumentTypeLetter {
		return h.errorResponse("type must be one of: book, report, article, letter")
	}

	docID, err := h.manager.CreateDocument(title, author, docType)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to create document: %v", err))
	}

	result := map[string]interface{}{
		"document_id": string(docID),
		"title":       title,
		"author":      author,
		"type":        docTypeStr,
		"message":     "Document created successfully",
	}

	return h.successResponse(result)
}

// handleGetDocumentStructure gets the document structure
func (h *DocGenHandler) handleGetDocumentStructure(params map[string]interface{}) (*protocol.CallToolResponse, error) {
	docID, err := h.getDocumentID(params)
	if err != nil {
		return h.errorResponse(err.Error())
	}

	manifest, err := h.manager.GetDocumentStructure(docID)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to get document structure: %v", err))
	}

	return h.successResponse(manifest)
}

// handleDeleteDocument deletes a document
func (h *DocGenHandler) handleDeleteDocument(params map[string]interface{}) (*protocol.CallToolResponse, error) {
	docID, err := h.getDocumentID(params)
	if err != nil {
		return h.errorResponse(err.Error())
	}

	if err := h.manager.DeleteDocument(docID); err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to delete document: %v", err))
	}

	result := map[string]interface{}{
		"document_id": string(docID),
		"message":     "Document deleted successfully",
	}

	return h.successResponse(result)
}

// handleConfigureDocument configures document settings
func (h *DocGenHandler) handleConfigureDocument(params map[string]interface{}) (*protocol.CallToolResponse, error) {
	docID, err := h.getDocumentID(params)
	if err != nil {
		return h.errorResponse(err.Error())
	}

	var styleUpdates *types.Style
	var pandocOptions *types.PandocConfig

	// Parse style updates if provided
	if styleData, ok := params["style_updates"].(map[string]interface{}); ok {
		styleUpdates = &types.Style{}
		if fontFamily, ok := styleData["font_family"].(string); ok {
			styleUpdates.FontFamily = fontFamily
		}
		if fontSize, ok := styleData["font_size"].(string); ok {
			styleUpdates.FontSize = fontSize
		}
		if lineSpacing, ok := styleData["line_spacing"].(string); ok {
			styleUpdates.LineSpacing = lineSpacing
		}
		if marginsData, ok := styleData["margins"].(map[string]interface{}); ok {
			if top, ok := marginsData["top"].(string); ok {
				styleUpdates.Margins.Top = top
			}
			if bottom, ok := marginsData["bottom"].(string); ok {
				styleUpdates.Margins.Bottom = bottom
			}
			if left, ok := marginsData["left"].(string); ok {
				styleUpdates.Margins.Left = left
			}
			if right, ok := marginsData["right"].(string); ok {
				styleUpdates.Margins.Right = right
			}
		}
	}

	// Parse pandoc options if provided
	if pandocData, ok := params["pandoc_options"].(map[string]interface{}); ok {
		pandocOptions = &types.PandocConfig{}
		if pdfEngine, ok := pandocData["pdf_engine"].(string); ok {
			pandocOptions.PDFEngine = pdfEngine
		}
		if toc, ok := pandocData["toc"].(bool); ok {
			pandocOptions.TOC = toc
		}
		if tocDepth, ok := pandocData["toc_depth"].(float64); ok {
			pandocOptions.TOCDepth = int(tocDepth)
		}
		if citationStyle, ok := pandocData["citation_style"].(string); ok {
			pandocOptions.CitationStyle = citationStyle
		}
	}

	if err := h.manager.ConfigureDocument(docID, styleUpdates, pandocOptions); err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to configure document: %v", err))
	}

	result := map[string]interface{}{
		"document_id": string(docID),
		"message":     "Document configuration updated successfully",
	}

	return h.successResponse(result)
}

// handleAddChapter adds a new chapter
func (h *DocGenHandler) handleAddChapter(params map[string]interface{}) (*protocol.CallToolResponse, error) {
	docID, err := h.getDocumentID(params)
	if err != nil {
		return h.errorResponse(err.Error())
	}

	title, ok := params["title"].(string)
	if !ok || title == "" {
		return h.errorResponse("title parameter is required and must be a non-empty string")
	}

	var position *int
	if posFloat, ok := params["position"].(float64); ok {
		pos := int(posFloat)
		position = &pos
	}

	chapterNum, err := h.manager.AddChapter(docID, title, position)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to add chapter: %v", err))
	}

	result := map[string]interface{}{
		"document_id":    string(docID),
		"chapter_number": int(chapterNum),
		"title":          title,
		"message":        "Chapter added successfully",
	}

	return h.successResponse(result)
}

// handleGetChapter gets a specific chapter
func (h *DocGenHandler) handleGetChapter(params map[string]interface{}) (*protocol.CallToolResponse, error) {
	docID, err := h.getDocumentID(params)
	if err != nil {
		return h.errorResponse(err.Error())
	}

	chapterNum, err := h.getChapterNumber(params)
	if err != nil {
		return h.errorResponse(err.Error())
	}

	chapter, err := h.manager.GetChapter(docID, chapterNum)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to get chapter: %v", err))
	}

	return h.successResponse(chapter)
}

// handleUpdateChapterMetadata updates chapter metadata
func (h *DocGenHandler) handleUpdateChapterMetadata(params map[string]interface{}) (*protocol.CallToolResponse, error) {
	docID, err := h.getDocumentID(params)
	if err != nil {
		return h.errorResponse(err.Error())
	}

	chapterNum, err := h.getChapterNumber(params)
	if err != nil {
		return h.errorResponse(err.Error())
	}

	title, ok := params["title"].(string)
	if !ok || title == "" {
		return h.errorResponse("title parameter is required and must be a non-empty string")
	}

	if err := h.manager.UpdateChapterMetadata(docID, chapterNum, title); err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to update chapter metadata: %v", err))
	}

	result := map[string]interface{}{
		"document_id":    string(docID),
		"chapter_number": int(chapterNum),
		"title":          title,
		"message":        "Chapter metadata updated successfully",
	}

	return h.successResponse(result)
}

// handleDeleteChapter deletes a chapter
func (h *DocGenHandler) handleDeleteChapter(params map[string]interface{}) (*protocol.CallToolResponse, error) {
	docID, err := h.getDocumentID(params)
	if err != nil {
		return h.errorResponse(err.Error())
	}

	chapterNum, err := h.getChapterNumber(params)
	if err != nil {
		return h.errorResponse(err.Error())
	}

	if err := h.manager.DeleteChapter(docID, chapterNum); err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to delete chapter: %v", err))
	}

	result := map[string]interface{}{
		"document_id":    string(docID),
		"chapter_number": int(chapterNum),
		"message":        "Chapter deleted successfully",
	}

	return h.successResponse(result)
}

// handleExportDocument exports a document
func (h *DocGenHandler) handleExportDocument(params map[string]interface{}) (*protocol.CallToolResponse, error) {
	docID, err := h.getDocumentID(params)
	if err != nil {
		return h.errorResponse(err.Error())
	}

	formatStr, ok := params["format"].(string)
	if !ok || formatStr == "" {
		return h.errorResponse("format parameter is required")
	}

	format := types.ExportFormat(formatStr)
	if format != types.ExportFormatPDF && format != types.ExportFormatDOCX && format != types.ExportFormatHTML {
		return h.errorResponse("format must be one of: pdf, docx, html")
	}

	// Parse chapters if provided
	var chapters []types.ChapterNumber
	if chaptersData, ok := params["chapters"].([]interface{}); ok {
		for _, chapterData := range chaptersData {
			if chapterFloat, ok := chapterData.(float64); ok {
				chapters = append(chapters, types.ChapterNumber(int(chapterFloat)))
			}
		}
	}

	// Load document data
	manifest, err := h.storage.LoadManifest(string(docID))
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to load document manifest: %v", err))
	}

	style, err := h.storage.LoadStyle(string(docID))
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to load document style: %v", err))
	}

	pandocConfig, err := h.storage.LoadPandocConfig(string(docID))
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to load pandoc config: %v", err))
	}

	options := &types.ExportOptions{
		Format:   format,
		Chapters: chapters,
	}

	outputPath, err := h.exporter.ExportDocument(string(docID), manifest, style, pandocConfig, options)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to export document: %v", err))
	}

	result := map[string]interface{}{
		"document_id":   string(docID),
		"format":        formatStr,
		"output_path":   outputPath,
		"message":       "Document exported successfully",
	}

	return h.successResponse(result)
}

// handleValidateDocument validates a document
func (h *DocGenHandler) handleValidateDocument(params map[string]interface{}) (*protocol.CallToolResponse, error) {
	docID, err := h.getDocumentID(params)
	if err != nil {
		return h.errorResponse(err.Error())
	}

	manifest, err := h.storage.LoadManifest(string(docID))
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to load document manifest: %v", err))
	}

	report := h.exporter.ValidateDocument(string(docID), manifest)
	return h.successResponse(report)
}

// Helper methods for the remaining tools (sections, images) will be implemented
// For now, let's implement basic placeholders

func (h *DocGenHandler) handleMoveChapter(params map[string]interface{}) (*protocol.CallToolResponse, error) {
	// Get document ID
	docID, err := h.getDocumentID(params)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Invalid document_id: %v", err))
	}

	// Get from position
	fromPosFloat, ok := params["from_position"].(float64)
	if !ok {
		return h.errorResponse("from_position parameter is required")
	}
	fromPos := types.ChapterNumber(fromPosFloat)
	if fromPos < 1 {
		return h.errorResponse("from_position must be at least 1")
	}

	// Get to position
	toPosFloat, ok := params["to_position"].(float64)
	if !ok {
		return h.errorResponse("to_position parameter is required")
	}
	toPos := types.ChapterNumber(toPosFloat)
	if toPos < 1 {
		return h.errorResponse("to_position must be at least 1")
	}

	if fromPos == toPos {
		return h.errorResponse("from_position and to_position cannot be the same")
	}

	// Move the chapter
	err = h.manager.MoveChapter(docID, fromPos, toPos)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to move chapter: %v", err))
	}

	return h.successResponse(map[string]interface{}{
		"from_position": int(fromPos),
		"to_position":   int(toPos),
		"message":       fmt.Sprintf("Chapter moved from position %d to %d successfully", fromPos, toPos),
	})
}

func (h *DocGenHandler) handleAddSection(params map[string]interface{}) (*protocol.CallToolResponse, error) {
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

	// Get section title
	title, ok := params["title"].(string)
	if !ok || title == "" {
		return h.errorResponse("title parameter is required")
	}

	// Get section content
	content, ok := params["content"].(string)
	if !ok || content == "" {
		return h.errorResponse("content parameter is required")
	}

	// Get level (optional, defaults to 1)
	level := 1
	if levelFloat, ok := params["level"].(float64); ok {
		level = int(levelFloat)
		if level < 1 || level > 6 {
			return h.errorResponse("level must be between 1 and 6")
		}
	}

	// Add the section
	sectionNum, err := h.manager.AddSection(docID, chapterNum, title, content, level)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to add section: %v", err))
	}

	return h.successResponse(map[string]interface{}{
		"section_number": sectionNum.String(),
		"message":        fmt.Sprintf("Section '%s' added successfully to chapter %d", title, chapterNum),
	})
}

func (h *DocGenHandler) handleUpdateSection(params map[string]interface{}) (*protocol.CallToolResponse, error) {
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

	// Get section number
	sectionNumStr, ok := params["section_number"].(string)
	if !ok || sectionNumStr == "" {
		return h.errorResponse("section_number parameter is required")
	}

	// Parse section number
	sectionNum, err := h.parseSectionNumber(sectionNumStr)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Invalid section_number format: %v", err))
	}

	// Get new content
	content, ok := params["content"].(string)
	if !ok || content == "" {
		return h.errorResponse("content parameter is required")
	}

	// Update the section
	err = h.manager.UpdateSection(docID, chapterNum, sectionNum, content)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to update section: %v", err))
	}

	return h.successResponse(map[string]interface{}{
		"section_number": sectionNumStr,
		"message":        fmt.Sprintf("Section %s updated successfully", sectionNumStr),
	})
}

func (h *DocGenHandler) handleDeleteSection(params map[string]interface{}) (*protocol.CallToolResponse, error) {
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

	// Get section number
	sectionNumStr, ok := params["section_number"].(string)
	if !ok || sectionNumStr == "" {
		return h.errorResponse("section_number parameter is required")
	}

	// Parse section number
	sectionNum, err := h.parseSectionNumber(sectionNumStr)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Invalid section_number format: %v", err))
	}

	// Delete the section
	err = h.manager.DeleteSection(docID, chapterNum, sectionNum)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to delete section: %v", err))
	}

	return h.successResponse(map[string]interface{}{
		"section_number": sectionNumStr,
		"message":        fmt.Sprintf("Section %s deleted successfully", sectionNumStr),
	})
}

func (h *DocGenHandler) handleAddImage(params map[string]interface{}) (*protocol.CallToolResponse, error) {
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

	// Get image path
	imagePath, ok := params["image_path"].(string)
	if !ok || imagePath == "" {
		return h.errorResponse("image_path parameter is required")
	}

	// Get caption
	caption, ok := params["caption"].(string)
	if !ok || caption == "" {
		return h.errorResponse("caption parameter is required")
	}

	// Get position (optional, defaults to "here")
	position := "here"
	if pos, ok := params["position"].(string); ok && pos != "" {
		position = pos
	}

	// Add the image
	figureID, err := h.manager.AddImage(docID, chapterNum, imagePath, caption, position)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to add image: %v", err))
	}

	return h.successResponse(map[string]interface{}{
		"figure_id": figureID,
		"message":   fmt.Sprintf("Image added successfully with ID %s", figureID),
	})
}

func (h *DocGenHandler) handleUpdateImageCaption(params map[string]interface{}) (*protocol.CallToolResponse, error) {
	// Get document ID
	docID, err := h.getDocumentID(params)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Invalid document_id: %v", err))
	}

	// Get figure ID
	figureID, ok := params["figure_id"].(string)
	if !ok || figureID == "" {
		return h.errorResponse("figure_id parameter is required")
	}

	// Get new caption
	newCaption, ok := params["new_caption"].(string)
	if !ok || newCaption == "" {
		return h.errorResponse("new_caption parameter is required")
	}

	// Update the image caption
	err = h.manager.UpdateImageCaption(docID, types.FigureID(figureID), newCaption)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to update image caption: %v", err))
	}

	return h.successResponse(map[string]interface{}{
		"figure_id": figureID,
		"message":   fmt.Sprintf("Image caption updated successfully for %s", figureID),
	})
}

func (h *DocGenHandler) handleDeleteImage(params map[string]interface{}) (*protocol.CallToolResponse, error) {
	// Get document ID
	docID, err := h.getDocumentID(params)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Invalid document_id: %v", err))
	}

	// Get figure ID
	figureID, ok := params["figure_id"].(string)
	if !ok || figureID == "" {
		return h.errorResponse("figure_id parameter is required")
	}

	// Delete the image
	err = h.manager.DeleteImage(docID, types.FigureID(figureID))
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to delete image: %v", err))
	}

	return h.successResponse(map[string]interface{}{
		"figure_id": figureID,
		"message":   fmt.Sprintf("Image %s deleted successfully", figureID),
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
	previewPath, err := h.exporter.PreviewChapter(string(docID), chapterNum, types.ExportFormat(format))
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

// Helper methods

func (h *DocGenHandler) getDocumentID(params map[string]interface{}) (types.DocumentID, error) {
	docIDStr, ok := params["document_id"].(string)
	if !ok || docIDStr == "" {
		return "", fmt.Errorf("document_id parameter is required and must be a non-empty string")
	}

	docID := types.DocumentID(docIDStr)
	if err := docID.Validate(); err != nil {
		return "", fmt.Errorf("invalid document_id: %w", err)
	}

	return docID, nil
}

func (h *DocGenHandler) getChapterNumber(params map[string]interface{}) (types.ChapterNumber, error) {
	chapterFloat, ok := params["chapter_number"].(float64)
	if !ok {
		return 0, fmt.Errorf("chapter_number parameter is required and must be a number")
	}

	chapterNum := types.ChapterNumber(int(chapterFloat))
	if chapterNum < 1 {
		return 0, fmt.Errorf("chapter_number must be at least 1")
	}

	return chapterNum, nil
}

func (h *DocGenHandler) parseSectionNumber(sectionNumStr string) (types.SectionNumber, error) {
	parts := strings.Split(sectionNumStr, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("section number must have at least 2 parts (e.g., '1.1')")
	}

	sectionNum := make(types.SectionNumber, len(parts))
	for i, part := range parts {
		num, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid number in section: %s", part)
		}
		if num <= 0 {
			return nil, fmt.Errorf("section numbers must be positive")
		}
		sectionNum[i] = num
	}

	return sectionNum, nil
}

func (h *DocGenHandler) successResponse(data interface{}) (*protocol.CallToolResponse, error) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to marshal response: %v", err))
	}

	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: string(jsonBytes),
			},
		},
	}, nil
}

func (h *DocGenHandler) errorResponse(message string) (*protocol.CallToolResponse, error) {
	return &protocol.CallToolResponse{
		Content: []protocol.ToolContent{
			{
				Type: "text",
				Text: fmt.Sprintf("Error: %s", message),
			},
		},
		IsError: true,
	}, nil
}