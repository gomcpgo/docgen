package handler

import (
	"fmt"
	"strings"

	"github.com/gomcpgo/mcp/pkg/protocol"
	"github.com/gomcpgo/docgen/pkg/types"
)

// Chapter operations

func (h *DocGenHandler) handleAddChapter(params map[string]interface{}) (*protocol.CallToolResponse, error) {
	// Get document ID
	docID, err := h.getDocumentID(params)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Invalid document_id: %v", err))
	}

	// Get chapter title
	title, ok := params["title"].(string)
	if !ok || title == "" {
		return h.errorResponse("title parameter is required")
	}

	// Get position (optional)
	var position *int
	if pos, ok := params["position"].(float64); ok {
		posInt := int(pos)
		position = &posInt
	}

	// Add the chapter
	chapterNum, err := h.manager.AddChapter(docID, title, position)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to add chapter: %v", err))
	}

	return h.successResponse(map[string]interface{}{
		"document_id":    docID,
		"chapter_number": int(chapterNum),
		"message":        fmt.Sprintf("Chapter '%s' added as chapter %d", title, chapterNum),
	})
}

func (h *DocGenHandler) handleGetChapter(params map[string]interface{}) (*protocol.CallToolResponse, error) {
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

	// Get the chapter
	chapter, err := h.manager.GetChapter(docID, chapterNum)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to get chapter: %v", err))
	}

	return h.successResponse(chapter)
}

func (h *DocGenHandler) handleUpdateChapterMetadata(params map[string]interface{}) (*protocol.CallToolResponse, error) {
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

	// Get new title
	title, ok := params["title"].(string)
	if !ok || title == "" {
		return h.errorResponse("title parameter is required")
	}

	// Update chapter metadata
	err = h.manager.UpdateChapterMetadata(docID, chapterNum, title)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to update chapter: %v", err))
	}

	return h.successResponse(map[string]interface{}{
		"document_id":    docID,
		"chapter_number": int(chapterNum),
		"message":        fmt.Sprintf("Chapter %d title updated to '%s'", chapterNum, title),
	})
}

func (h *DocGenHandler) handleDeleteChapter(params map[string]interface{}) (*protocol.CallToolResponse, error) {
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

	// Delete the chapter
	err = h.manager.DeleteChapter(docID, chapterNum)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to delete chapter: %v", err))
	}

	return h.successResponse(map[string]interface{}{
		"document_id":    docID,
		"chapter_number": int(chapterNum),
		"message":        fmt.Sprintf("Chapter %d deleted successfully", chapterNum),
	})
}

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
		"document_id":    docID,
		"from_position": int(fromPos),
		"to_position":   int(toPos),
		"message":       fmt.Sprintf("Chapter moved from position %d to %d successfully", fromPos, toPos),
	})
}

// handleGetChapterContent returns the full markdown content of a chapter
func (h *DocGenHandler) handleGetChapterContent(params map[string]interface{}) (*protocol.CallToolResponse, error) {
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

	// Load chapter metadata
	chapter, err := h.storage.LoadChapterMetadata(string(docID), int(chapterNum))
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to load chapter metadata: %v", err))
	}

	// Load chapter content
	content, err := h.storage.LoadChapterContent(string(docID), int(chapterNum))
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to load chapter content: %v", err))
	}

	// Get the full chapter with sections
	fullChapter, err := h.manager.GetChapter(types.DocumentID(docID), chapterNum)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to get chapter details: %v", err))
	}

	// Build full markdown content
	var fullContent strings.Builder
	
	// Start with chapter content if any
	if content != "" {
		fullContent.WriteString(content)
		fullContent.WriteString("\n\n")
	}

	// Add all sections
	for _, section := range fullChapter.Sections {
		// Add appropriate heading level
		for i := 0; i < section.Level; i++ {
			fullContent.WriteString("#")
		}
		fullContent.WriteString(" ")
		fullContent.WriteString(section.Title)
		fullContent.WriteString("\n\n")
		fullContent.WriteString(section.Content)
		fullContent.WriteString("\n\n")
	}

	return h.successResponse(map[string]interface{}{
		"document_id":    docID,
		"chapter_number": int(chapterNum),
		"title":         chapter.Title,
		"content":       strings.TrimSpace(fullContent.String()),
	})
}

// handleUpdateChapterContent updates the entire content of a chapter
func (h *DocGenHandler) handleUpdateChapterContent(params map[string]interface{}) (*protocol.CallToolResponse, error) {
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

	// Get content
	content, ok := params["content"].(string)
	if !ok {
		return h.errorResponse("content parameter is required")
	}

	// Parse markdown to identify sections
	// For now, we'll save the content as-is and clear existing sections
	// In a full implementation, we'd parse the markdown and create sections based on headings

	// Get the chapter to clear existing sections
	chapter, err := h.manager.GetChapter(types.DocumentID(docID), chapterNum)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to get chapter: %v", err))
	}

	// Clear existing sections
	for _, section := range chapter.Sections {
		if err := h.manager.DeleteSection(types.DocumentID(docID), chapterNum, section.Number); err != nil {
			// Log error but continue
			fmt.Printf("Warning: Failed to delete section %s: %v\n", section.Number, err)
		}
	}

	// Save the new content
	if err := h.storage.SaveChapterContent(string(docID), int(chapterNum), content); err != nil {
		return h.errorResponse(fmt.Sprintf("Failed to save chapter content: %v", err))
	}

	// Note: In a full implementation, we would update the manifest here
	// to reflect any changes in word count, section structure, etc.

	return h.successResponse(map[string]interface{}{
		"document_id":    docID,
		"chapter_number": int(chapterNum),
		"message":       "Chapter content updated successfully",
	})
}