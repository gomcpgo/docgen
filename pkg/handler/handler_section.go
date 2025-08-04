package handler

import (
	"fmt"

	"github.com/gomcpgo/mcp/pkg/protocol"
)

// Section operations

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
		"document_id":    docID,
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
		"document_id":    docID,
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
		"document_id":    docID,
		"section_number": sectionNumStr,
		"message":        fmt.Sprintf("Section %s deleted successfully", sectionNumStr),
	})
}