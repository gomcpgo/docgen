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
	// Document operations
	case "create_document":
		return h.handleCreateDocument(req.Arguments)
	case "get_document_structure":
		return h.handleGetDocumentStructure(req.Arguments)
	case "delete_document":
		return h.handleDeleteDocument(req.Arguments)
	case "configure_document":
		return h.handleConfigureDocument(req.Arguments)

	// Chapter operations
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

	// Section operations
	case "add_section":
		return h.handleAddSection(req.Arguments)
	case "update_section":
		return h.handleUpdateSection(req.Arguments)
	case "delete_section":
		return h.handleDeleteSection(req.Arguments)

	// Image operations
	case "add_image":
		return h.handleAddImage(req.Arguments)
	case "update_image_caption":
		return h.handleUpdateImageCaption(req.Arguments)
	case "delete_image":
		return h.handleDeleteImage(req.Arguments)

	// Export operations
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
		}, nil
	}
}

// Helper methods

func (h *DocGenHandler) getDocumentID(params map[string]interface{}) (types.DocumentID, error) {
	docIDStr, ok := params["document_id"].(string)
	if !ok || docIDStr == "" {
		return "", fmt.Errorf("document_id parameter is required")
	}

	docID := types.DocumentID(docIDStr)
	if err := docID.Validate(); err != nil {
		return "", err
	}

	return docID, nil
}

func (h *DocGenHandler) getChapterNumber(params map[string]interface{}) (types.ChapterNumber, error) {
	chapterNumFloat, ok := params["chapter_number"].(float64)
	if !ok {
		return 0, fmt.Errorf("chapter_number parameter is required")
	}

	chapterNum := types.ChapterNumber(chapterNumFloat)
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
	}, nil
}