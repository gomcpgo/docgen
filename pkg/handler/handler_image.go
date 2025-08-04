package handler

import (
	"fmt"

	"github.com/gomcpgo/mcp/pkg/protocol"
	"github.com/gomcpgo/docgen/pkg/types"
)

// Image operations

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
		"document_id": docID,
		"figure_id":   figureID,
		"message":     fmt.Sprintf("Image added successfully with ID %s", figureID),
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
		"document_id": docID,
		"figure_id":   figureID,
		"message":     fmt.Sprintf("Image caption updated successfully for %s", figureID),
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
		"document_id": docID,
		"figure_id":   figureID,
		"message":     fmt.Sprintf("Image %s deleted successfully", figureID),
	})
}