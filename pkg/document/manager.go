package document

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gomcpgo/docgen/pkg/config"
	"github.com/gomcpgo/docgen/pkg/storage"
	"github.com/gomcpgo/docgen/pkg/types"
)

// Manager handles high-level document operations
type Manager struct {
	config  *config.Config
	storage storage.Storage
}

// NewManager creates a new document manager
func NewManager(cfg *config.Config, stor storage.Storage) *Manager {
	return &Manager{
		config:  cfg,
		storage: stor,
	}
}

// CreateDocument creates a new document with the given parameters
func (m *Manager) CreateDocument(title, author string, docType types.DocumentType) (types.DocumentID, error) {
	if title == "" {
		return "", fmt.Errorf("document title is required")
	}
	if author == "" {
		return "", fmt.Errorf("document author is required")
	}

	// Generate a unique document ID
	docID := types.DocumentID(generateDocumentID(title))

	// Validate the document ID
	if err := docID.Validate(); err != nil {
		return "", fmt.Errorf("invalid document ID: %w", err)
	}

	// Check if document already exists
	exists, err := m.storage.DocumentExists(string(docID))
	if err != nil {
		return "", fmt.Errorf("failed to check document existence: %w", err)
	}
	if exists {
		return "", fmt.Errorf("document with ID %s already exists", docID)
	}

	// Check document count limit
	if err := m.checkDocumentLimit(); err != nil {
		return "", err
	}

	// Create document structure
	now := time.Now()
	doc := &types.Document{
		ID:        docID,
		Title:     title,
		Author:    author,
		Type:      docType,
		CreatedAt: now,
		UpdatedAt: now,
		Chapters:  []types.Chapter{},
	}

	if err := m.storage.CreateDocumentStructure(doc); err != nil {
		return "", fmt.Errorf("failed to create document structure: %w", err)
	}

	return docID, nil
}

// GetDocumentStructure returns the complete document structure
func (m *Manager) GetDocumentStructure(docID types.DocumentID) (*types.Manifest, error) {
	if err := docID.Validate(); err != nil {
		return nil, fmt.Errorf("invalid document ID: %w", err)
	}

	exists, err := m.storage.DocumentExists(string(docID))
	if err != nil {
		return nil, fmt.Errorf("failed to check document existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("document %s not found", docID)
	}

	manifest, err := m.storage.LoadManifest(string(docID))
	if err != nil {
		return nil, fmt.Errorf("failed to load document manifest: %w", err)
	}

	return manifest, nil
}

// DeleteDocument removes a document and all its contents
func (m *Manager) DeleteDocument(docID types.DocumentID) error {
	if err := docID.Validate(); err != nil {
		return fmt.Errorf("invalid document ID: %w", err)
	}

	exists, err := m.storage.DocumentExists(string(docID))
	if err != nil {
		return fmt.Errorf("failed to check document existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("document %s not found", docID)
	}

	if err := m.storage.DeleteDocument(string(docID)); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}

// AddChapter adds a new chapter to the document
func (m *Manager) AddChapter(docID types.DocumentID, title string, position *int) (types.ChapterNumber, error) {
	if err := docID.Validate(); err != nil {
		return 0, fmt.Errorf("invalid document ID: %w", err)
	}
	if title == "" {
		return 0, fmt.Errorf("chapter title is required")
	}

	// Load current manifest
	manifest, err := m.storage.LoadManifest(string(docID))
	if err != nil {
		return 0, fmt.Errorf("failed to load document manifest: %w", err)
	}

	// Determine chapter number
	var chapterNum types.ChapterNumber
	if position != nil {
		// Insert at specific position (renumber subsequent chapters)
		chapterNum = types.ChapterNumber(*position)
		if err := m.renumberChapters(string(docID), manifest, int(chapterNum), 1); err != nil {
			return 0, fmt.Errorf("failed to renumber chapters: %w", err)
		}
	} else {
		// Add at the end
		chapterNum = types.ChapterNumber(len(manifest.Document.Chapters) + 1)
	}

	// Create chapter
	now := time.Now()
	chapter := &types.Chapter{
		Number:    chapterNum,
		Title:     title,
		Content:   "",
		Sections:  []types.Section{},
		Figures:   []types.Figure{},
		Tables:    []types.Table{},
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Create chapter structure
	if err := m.storage.CreateChapterStructure(string(docID), chapter); err != nil {
		return 0, fmt.Errorf("failed to create chapter structure: %w", err)
	}

	// Update manifest
	manifest.Document.Chapters = append(manifest.Document.Chapters, *chapter)
	manifest.ChapterCounts[chapterNum] = types.ChapterCount{
		Sections: 0,
		Figures:  0,
		Tables:   0,
	}
	manifest.UpdatedAt = now

	if err := m.storage.SaveManifest(string(docID), manifest); err != nil {
		return 0, fmt.Errorf("failed to update manifest: %w", err)
	}

	return chapterNum, nil
}

// GetChapter returns a specific chapter
func (m *Manager) GetChapter(docID types.DocumentID, chapterNum types.ChapterNumber) (*types.Chapter, error) {
	if err := docID.Validate(); err != nil {
		return nil, fmt.Errorf("invalid document ID: %w", err)
	}

	exists, err := m.storage.ChapterExists(string(docID), int(chapterNum))
	if err != nil {
		return nil, fmt.Errorf("failed to check chapter existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("chapter %d not found in document %s", chapterNum, docID)
	}

	chapter, err := m.storage.LoadChapterMetadata(string(docID), int(chapterNum))
	if err != nil {
		return nil, fmt.Errorf("failed to load chapter metadata: %w", err)
	}

	// Load chapter content
	content, err := m.storage.LoadChapterContent(string(docID), int(chapterNum))
	if err != nil {
		return nil, fmt.Errorf("failed to load chapter content: %w", err)
	}
	chapter.Content = content

	return chapter, nil
}

// UpdateChapterMetadata updates chapter metadata (title, etc.)
func (m *Manager) UpdateChapterMetadata(docID types.DocumentID, chapterNum types.ChapterNumber, title string) error {
	if err := docID.Validate(); err != nil {
		return fmt.Errorf("invalid document ID: %w", err)
	}
	if title == "" {
		return fmt.Errorf("chapter title is required")
	}

	// Load current chapter
	chapter, err := m.storage.LoadChapterMetadata(string(docID), int(chapterNum))
	if err != nil {
		return fmt.Errorf("failed to load chapter metadata: %w", err)
	}

	// Update fields
	chapter.Title = title
	chapter.UpdatedAt = time.Now()

	// Save updated metadata
	if err := m.storage.SaveChapterMetadata(string(docID), chapter); err != nil {
		return fmt.Errorf("failed to save chapter metadata: %w", err)
	}

	// Update manifest
	manifest, err := m.storage.LoadManifest(string(docID))
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	for i := range manifest.Document.Chapters {
		if manifest.Document.Chapters[i].Number == chapterNum {
			manifest.Document.Chapters[i].Title = title
			manifest.Document.Chapters[i].UpdatedAt = time.Now()
			break
		}
	}
	manifest.UpdatedAt = time.Now()

	if err := m.storage.SaveManifest(string(docID), manifest); err != nil {
		return fmt.Errorf("failed to update manifest: %w", err)
	}

	return nil
}

// DeleteChapter removes a chapter and renumbers subsequent chapters
func (m *Manager) DeleteChapter(docID types.DocumentID, chapterNum types.ChapterNumber) error {
	if err := docID.Validate(); err != nil {
		return fmt.Errorf("invalid document ID: %w", err)
	}

	exists, err := m.storage.ChapterExists(string(docID), int(chapterNum))
	if err != nil {
		return fmt.Errorf("failed to check chapter existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("chapter %d not found in document %s", chapterNum, docID)
	}

	// Load manifest
	manifest, err := m.storage.LoadManifest(string(docID))
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	// Delete chapter directory
	if err := m.storage.DeleteChapter(string(docID), int(chapterNum)); err != nil {
		return fmt.Errorf("failed to delete chapter: %w", err)
	}

	// Renumber subsequent chapters
	if err := m.renumberChapters(string(docID), manifest, int(chapterNum)+1, -1); err != nil {
		return fmt.Errorf("failed to renumber chapters: %w", err)
	}

	// Update manifest by removing the chapter
	var updatedChapters []types.Chapter
	for _, ch := range manifest.Document.Chapters {
		if ch.Number != chapterNum {
			if ch.Number > chapterNum {
				ch.Number--
			}
			updatedChapters = append(updatedChapters, ch)
		}
	}
	manifest.Document.Chapters = updatedChapters

	// Update chapter counts
	updatedCounts := make(map[types.ChapterNumber]types.ChapterCount)
	for chNum, count := range manifest.ChapterCounts {
		if chNum != chapterNum {
			if chNum > chapterNum {
				updatedCounts[chNum-1] = count
			} else {
				updatedCounts[chNum] = count
			}
		}
	}
	manifest.ChapterCounts = updatedCounts
	manifest.UpdatedAt = time.Now()

	if err := m.storage.SaveManifest(string(docID), manifest); err != nil {
		return fmt.Errorf("failed to update manifest: %w", err)
	}

	return nil
}

// ConfigureDocument updates document configuration (style, pandoc options)
func (m *Manager) ConfigureDocument(docID types.DocumentID, styleUpdates *types.Style, pandocOptions *types.PandocConfig) error {
	if err := docID.Validate(); err != nil {
		return fmt.Errorf("invalid document ID: %w", err)
	}

	exists, err := m.storage.DocumentExists(string(docID))
	if err != nil {
		return fmt.Errorf("failed to check document existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("document %s not found", docID)
	}

	// Update style if provided
	if styleUpdates != nil {
		if err := m.storage.SaveStyle(string(docID), styleUpdates); err != nil {
			return fmt.Errorf("failed to save style: %w", err)
		}
	}

	// Update pandoc config if provided
	if pandocOptions != nil {
		if err := m.storage.SavePandocConfig(string(docID), pandocOptions); err != nil {
			return fmt.Errorf("failed to save pandoc config: %w", err)
		}
	}

	return nil
}

// AddSection adds a new section to a chapter
func (m *Manager) AddSection(docID types.DocumentID, chapterNum types.ChapterNumber, title, content string, level int) (types.SectionNumber, error) {
	if err := docID.Validate(); err != nil {
		return nil, fmt.Errorf("invalid document ID: %w", err)
	}
	if title == "" {
		return nil, fmt.Errorf("section title is required")
	}
	if content == "" {
		return nil, fmt.Errorf("section content is required")
	}
	if level < 1 || level > 6 {
		return nil, fmt.Errorf("section level must be between 1 and 6")
	}

	// Load current chapter
	chapter, err := m.storage.LoadChapterMetadata(string(docID), int(chapterNum))
	if err != nil {
		return nil, fmt.Errorf("failed to load chapter: %w", err)
	}

	// Generate next section number
	sectionNum, err := m.generateNextSectionNumber(chapter, level)
	if err != nil {
		return nil, fmt.Errorf("failed to generate section number: %w", err)
	}

	// Create new section
	now := time.Now()
	section := types.Section{
		Number:    sectionNum,
		Title:     title,
		Content:   content,
		Level:     level,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Add section to chapter
	chapter.Sections = append(chapter.Sections, section)
	chapter.UpdatedAt = now

	// Save updated chapter metadata
	if err := m.storage.SaveChapterMetadata(string(docID), chapter); err != nil {
		return nil, fmt.Errorf("failed to save chapter metadata: %w", err)
	}

	return sectionNum, nil
}

// generateNextSectionNumber generates the next section number for the given level
func (m *Manager) generateNextSectionNumber(chapter *types.Chapter, level int) (types.SectionNumber, error) {
	// Find the highest section number at the specified level
	var maxNumbers = make([]int, level)
	chapterNum := int(chapter.Number)

	for _, section := range chapter.Sections {
		parts := []int(section.Number)
		if len(parts) >= level {
			// Only consider sections at the same depth or deeper
			for i := 0; i < level && i < len(parts); i++ {
				if parts[i] > maxNumbers[i] {
					maxNumbers[i] = parts[i]
				}
			}
		}
	}

	// Increment the last level
	maxNumbers[level-1]++

	// Build section number: [chapter, level1, level2, ...]
	sectionNum := make([]int, level+1)
	sectionNum[0] = chapterNum
	for i := 0; i < level; i++ {
		sectionNum[i+1] = maxNumbers[i]
	}

	return types.SectionNumber(sectionNum), nil
}

// UpdateSection updates the content of an existing section
func (m *Manager) UpdateSection(docID types.DocumentID, chapterNum types.ChapterNumber, sectionNum types.SectionNumber, content string) error {
	if err := docID.Validate(); err != nil {
		return fmt.Errorf("invalid document ID: %w", err)
	}
	if content == "" {
		return fmt.Errorf("section content is required")
	}

	// Load current chapter
	chapter, err := m.storage.LoadChapterMetadata(string(docID), int(chapterNum))
	if err != nil {
		return fmt.Errorf("failed to load chapter: %w", err)
	}

	// Find and update the section
	found := false
	for i, section := range chapter.Sections {
		if m.sectionNumbersEqual(section.Number, sectionNum) {
			chapter.Sections[i].Content = content
			chapter.Sections[i].UpdatedAt = time.Now()
			chapter.UpdatedAt = time.Now()
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("section %s not found in chapter %d", sectionNum.String(), chapterNum)
	}

	// Save updated chapter metadata
	if err := m.storage.SaveChapterMetadata(string(docID), chapter); err != nil {
		return fmt.Errorf("failed to save chapter metadata: %w", err)
	}

	return nil
}

// sectionNumbersEqual compares two section numbers for equality
func (m *Manager) sectionNumbersEqual(a, b types.SectionNumber) bool {
	if len(a) != len(b) {
		return false
	}
	for i, val := range a {
		if val != b[i] {
			return false
		}
	}
	return true
}

// DeleteSection removes a section from a chapter
func (m *Manager) DeleteSection(docID types.DocumentID, chapterNum types.ChapterNumber, sectionNum types.SectionNumber) error {
	if err := docID.Validate(); err != nil {
		return fmt.Errorf("invalid document ID: %w", err)
	}

	// Load current chapter
	chapter, err := m.storage.LoadChapterMetadata(string(docID), int(chapterNum))
	if err != nil {
		return fmt.Errorf("failed to load chapter: %w", err)
	}

	// Find and remove the section
	found := false
	for i, section := range chapter.Sections {
		if m.sectionNumbersEqual(section.Number, sectionNum) {
			// Remove the section
			chapter.Sections = append(chapter.Sections[:i], chapter.Sections[i+1:]...)
			chapter.UpdatedAt = time.Now()
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("section %s not found in chapter %d", sectionNum.String(), chapterNum)
	}

	// Renumber sections if needed (sections with same level and deeper)
	m.renumberSectionsAfterDeletion(chapter, sectionNum)

	// Save updated chapter metadata
	if err := m.storage.SaveChapterMetadata(string(docID), chapter); err != nil {
		return fmt.Errorf("failed to save chapter metadata: %w", err)
	}

	return nil
}

// renumberSectionsAfterDeletion renumbers sections after a deletion
func (m *Manager) renumberSectionsAfterDeletion(chapter *types.Chapter, deletedSectionNum types.SectionNumber) {
	if len(deletedSectionNum) < 2 {
		return // Invalid section number
	}

	deletedLevel := len(deletedSectionNum) - 1
	chapterNum := deletedSectionNum[0]

	for i := range chapter.Sections {
		section := &chapter.Sections[i]
		if len(section.Number) > deletedLevel && section.Number[0] == chapterNum {
			// Check if this section needs renumbering
			shouldRenumber := true
			for j := 1; j < deletedLevel; j++ {
				if j >= len(section.Number) || section.Number[j] != deletedSectionNum[j] {
					shouldRenumber = false
					break
				}
			}

			if shouldRenumber && len(section.Number) > deletedLevel {
				if section.Number[deletedLevel] > deletedSectionNum[deletedLevel] {
					// Decrement this level
					section.Number[deletedLevel]--
				}
			}
		}
	}
}

// MoveChapter moves a chapter from one position to another and renumbers accordingly
func (m *Manager) MoveChapter(docID types.DocumentID, fromPos, toPos types.ChapterNumber) error {
	if err := docID.Validate(); err != nil {
		return fmt.Errorf("invalid document ID: %w", err)
	}

	// Load the document manifest
	manifest, err := m.storage.LoadManifest(string(docID))
	if err != nil {
		return fmt.Errorf("failed to load document manifest: %w", err)
	}

	// Validate positions
	if int(fromPos) > len(manifest.Document.Chapters) || int(toPos) > len(manifest.Document.Chapters) {
		return fmt.Errorf("invalid position: document has %d chapters", len(manifest.Document.Chapters))
	}

	// Convert to 0-based indexing
	fromIdx := int(fromPos) - 1
	toIdx := int(toPos) - 1

	if fromIdx == toIdx {
		return nil // No movement needed
	}

	// Store the chapter to move
	chapterToMove := manifest.Document.Chapters[fromIdx]

	// Remove the chapter from its current position
	manifest.Document.Chapters = append(manifest.Document.Chapters[:fromIdx], manifest.Document.Chapters[fromIdx+1:]...)

	// Insert the chapter at the new position
	if toIdx >= len(manifest.Document.Chapters) {
		// Insert at the end
		manifest.Document.Chapters = append(manifest.Document.Chapters, chapterToMove)
	} else {
		// Insert at the specified position
		manifest.Document.Chapters = append(manifest.Document.Chapters[:toIdx], append([]types.Chapter{chapterToMove}, manifest.Document.Chapters[toIdx:]...)...)
	}

	// Renumber all chapters and update filesystem
	// For move operation, we need to renumber all chapters from position 1
	if err := m.renumberChapters(string(docID), manifest, 0, 0); err != nil {
		return fmt.Errorf("failed to renumber chapters: %w", err)
	}

	// Update manifest
	manifest.UpdatedAt = time.Now()
	if err := m.storage.SaveManifest(string(docID), manifest); err != nil {
		return fmt.Errorf("failed to save manifest: %w", err)
	}

	return nil
}

// AddImage adds a new image figure to a chapter
func (m *Manager) AddImage(docID types.DocumentID, chapterNum types.ChapterNumber, imagePath, caption, position string) (types.FigureID, error) {
	if err := docID.Validate(); err != nil {
		return "", fmt.Errorf("invalid document ID: %w", err)
	}
	if imagePath == "" {
		return "", fmt.Errorf("image path is required")
	}
	if caption == "" {
		return "", fmt.Errorf("caption is required")
	}

	// Validate position
	validPositions := map[string]bool{
		"here": true, "top": true, "bottom": true, "page": true, "float": true,
	}
	if !validPositions[position] {
		return "", fmt.Errorf("invalid position: %s (must be one of: here, top, bottom, page, float)", position)
	}

	// Load current chapter
	chapter, err := m.storage.LoadChapterMetadata(string(docID), int(chapterNum))
	if err != nil {
		return "", fmt.Errorf("failed to load chapter: %w", err)
	}

	// Generate next figure sequence number for this chapter
	sequence := len(chapter.Figures) + 1

	// Generate figure ID
	figureID := types.FigureID(fmt.Sprintf("fig-%d.%d", chapterNum, sequence))

	// Create new figure
	now := time.Now()
	figure := types.Figure{
		ID:        figureID,
		Chapter:   chapterNum,
		Sequence:  sequence,
		Caption:   caption,
		ImagePath: imagePath,
		Position:  types.ImagePosition(position),
		Alignment: types.AlignCenter, // Default alignment
		Width:     "",                // Will be determined automatically
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Add figure to chapter
	chapter.Figures = append(chapter.Figures, figure)
	chapter.UpdatedAt = now

	// Save updated chapter metadata
	if err := m.storage.SaveChapterMetadata(string(docID), chapter); err != nil {
		return "", fmt.Errorf("failed to save chapter metadata: %w", err)
	}

	return figureID, nil
}

// UpdateImageCaption updates the caption of an existing image figure
func (m *Manager) UpdateImageCaption(docID types.DocumentID, figureID types.FigureID, newCaption string) error {
	if err := docID.Validate(); err != nil {
		return fmt.Errorf("invalid document ID: %w", err)
	}
	if newCaption == "" {
		return fmt.Errorf("caption is required")
	}

	// Parse figure ID to get chapter number
	chapterNum, err := m.parseFigureIDChapter(figureID)
	if err != nil {
		return fmt.Errorf("invalid figure ID: %w", err)
	}

	// Load current chapter
	chapter, err := m.storage.LoadChapterMetadata(string(docID), int(chapterNum))
	if err != nil {
		return fmt.Errorf("failed to load chapter: %w", err)
	}

	// Find and update the figure
	found := false
	for i, figure := range chapter.Figures {
		if figure.ID == figureID {
			chapter.Figures[i].Caption = newCaption
			chapter.Figures[i].UpdatedAt = time.Now()
			chapter.UpdatedAt = time.Now()
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("figure %s not found", figureID)
	}

	// Save updated chapter metadata
	if err := m.storage.SaveChapterMetadata(string(docID), chapter); err != nil {
		return fmt.Errorf("failed to save chapter metadata: %w", err)
	}

	return nil
}

// DeleteImage removes an image figure and renumbers subsequent figures in the chapter
func (m *Manager) DeleteImage(docID types.DocumentID, figureID types.FigureID) error {
	if err := docID.Validate(); err != nil {
		return fmt.Errorf("invalid document ID: %w", err)
	}

	// Parse figure ID to get chapter number
	chapterNum, err := m.parseFigureIDChapter(figureID)
	if err != nil {
		return fmt.Errorf("invalid figure ID: %w", err)
	}

	// Load current chapter
	chapter, err := m.storage.LoadChapterMetadata(string(docID), int(chapterNum))
	if err != nil {
		return fmt.Errorf("failed to load chapter: %w", err)
	}

	// Find and remove the figure
	found := false
	deletedSequence := 0
	for i, figure := range chapter.Figures {
		if figure.ID == figureID {
			deletedSequence = figure.Sequence
			// Remove the figure
			chapter.Figures = append(chapter.Figures[:i], chapter.Figures[i+1:]...)
			chapter.UpdatedAt = time.Now()
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("figure %s not found", figureID)
	}

	// Renumber subsequent figures
	for i := range chapter.Figures {
		if chapter.Figures[i].Sequence > deletedSequence {
			chapter.Figures[i].Sequence--
			// Update the figure ID to reflect new sequence
			chapter.Figures[i].ID = types.FigureID(fmt.Sprintf("fig-%d.%d", chapterNum, chapter.Figures[i].Sequence))
		}
	}

	// Save updated chapter metadata
	if err := m.storage.SaveChapterMetadata(string(docID), chapter); err != nil {
		return fmt.Errorf("failed to save chapter metadata: %w", err)
	}

	return nil
}

// parseFigureIDChapter extracts the chapter number from a figure ID (e.g., "fig-1.2" -> 1)
func (m *Manager) parseFigureIDChapter(figureID types.FigureID) (types.ChapterNumber, error) {
	idStr := string(figureID)
	if !strings.HasPrefix(idStr, "fig-") {
		return 0, fmt.Errorf("figure ID must start with 'fig-'")
	}

	parts := strings.Split(idStr[4:], ".")
	if len(parts) != 2 {
		return 0, fmt.Errorf("figure ID must be in format 'fig-X.Y'")
	}

	chapterNum, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("invalid chapter number in figure ID: %w", err)
	}

	return types.ChapterNumber(chapterNum), nil
}

// checkDocumentLimit checks if the document count is within limits
func (m *Manager) checkDocumentLimit() error {
	docs, err := m.storage.ListDocuments()
	if err != nil {
		return fmt.Errorf("failed to list documents: %w", err)
	}

	if len(docs) >= m.config.MaxDocuments {
		return fmt.Errorf("maximum number of documents (%d) reached", m.config.MaxDocuments)
	}

	return nil
}

// renumberChapters is implemented in numbering.go

// generateDocumentID generates a document ID from the title
func generateDocumentID(title string) string {
	// Simple implementation: convert to lowercase and replace spaces with hyphens
	// In a real implementation, you might want to use a more sophisticated approach
	id := title
	id = strings.ToLower(id)
	id = strings.ReplaceAll(id, " ", "-")
	id = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(id, "")
	
	// Limit length
	if len(id) > 30 {
		id = id[:30]
	}
	
	// Add timestamp to ensure uniqueness
	timestamp := time.Now().Unix()
	id = fmt.Sprintf("%s-%d", id, timestamp)
	
	return id
}