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

func (h *DocGenHandler) handleGetSectionContent(params map[string]interface{}) (*protocol.CallToolResponse, error) {
	// Get document ID
	docID, err := h.getDocumentID(params)
	if err != nil {
		return h.errorResponse(fmt.Sprintf("Invalid document_id: %v", err))
	}

	// Get sections array
	sectionsRaw, ok := params["sections"].([]interface{})
	if !ok || len(sectionsRaw) == 0 {
		return h.errorResponse("sections parameter is required and must be a non-empty array")
	}

	// Result will contain section content
	type SectionContent struct {
		ChapterNumber  int    `json:"chapter_number"`
		SectionNumber  string `json:"section_number"`
		Content        string `json:"content"`
		Title          string `json:"title"`
	}

	var results []SectionContent

	// Process each section request
	for _, sectionRaw := range sectionsRaw {
		section, ok := sectionRaw.(map[string]interface{})
		if !ok {
			continue
		}

		// Get chapter number
		chapterNum, err := h.getChapterNumber(section)
		if err != nil {
			continue
		}

		// Get section number
		sectionNumStr, ok := section["section_number"].(string)
		if !ok || sectionNumStr == "" {
			continue
		}

		// Parse section number
		sectionNum, err := h.parseSectionNumber(sectionNumStr)
		if err != nil {
			continue
		}

		// Load section content
		sectionContent, err := h.manager.GetSectionContent(docID, chapterNum, sectionNum)
		if err != nil {
			// Section might not exist, continue with others
			continue
		}

		// Get chapter metadata to find section title
		structure, err := h.manager.GetDocumentStructure(docID)
		if err != nil {
			continue
		}

		// Find the chapter and section to get the title
		var sectionTitle string
		for _, ch := range structure.Document.Chapters {
			if ch.Number == chapterNum {
				for _, s := range ch.Sections {
					if s.Number.String() == sectionNum.String() {
						sectionTitle = s.Title
						break
					}
				}
				break
			}
		}

		results = append(results, SectionContent{
			ChapterNumber: int(chapterNum),
			SectionNumber: sectionNumStr,
			Content:       sectionContent,
			Title:         sectionTitle,
		})
	}

	if len(results) == 0 {
		return h.errorResponse("No sections found or could not load any section content")
	}

	return h.successResponse(map[string]interface{}{
		"document_id": docID,
		"sections":    results,
	})
}