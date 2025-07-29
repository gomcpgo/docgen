package document

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gomcpgo/docgen/pkg/types"
)

// renumberChapters renumbers chapters starting from a given position with an offset
// offset: +1 to shift up (insert), -1 to shift down (delete)
func (m *Manager) renumberChapters(docID string, manifest *types.Manifest, startFrom int, offset int) error {
	if offset == 0 {
		return nil // No change needed
	}

	// Get all chapters that need renumbering
	var toRename []types.ChapterNumber
	for _, chapter := range manifest.Document.Chapters {
		if int(chapter.Number) >= startFrom {
			toRename = append(toRename, chapter.Number)
		}
	}

	// Rename chapter directories in the correct order to avoid conflicts
	if offset > 0 {
		// Renaming upward (insert): start from highest number
		for i := len(toRename) - 1; i >= 0; i-- {
			oldNum := int(toRename[i])
			newNum := oldNum + offset
			
			if err := m.renameChapterDirectory(docID, oldNum, newNum); err != nil {
				return fmt.Errorf("failed to rename chapter %d to %d: %w", oldNum, newNum, err)
			}
		}
	} else {
		// Renaming downward (delete): start from lowest number
		for _, chapterNum := range toRename {
			oldNum := int(chapterNum)
			newNum := oldNum + offset
			
			if err := m.renameChapterDirectory(docID, oldNum, newNum); err != nil {
				return fmt.Errorf("failed to rename chapter %d to %d: %w", oldNum, newNum, err)
			}
		}
	}

	return nil
}

// renameChapterDirectory renames a chapter directory from old number to new number
func (m *Manager) renameChapterDirectory(docID string, oldNum, newNum int) error {
	oldPath := m.config.ChapterPath(docID, oldNum)
	newPath := m.config.ChapterPath(docID, newNum)

	// Check if old directory exists
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return nil // Directory doesn't exist, nothing to rename
	}

	// Create parent directory for new path if needed
	if err := os.MkdirAll(filepath.Dir(newPath), 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Rename the directory
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to rename directory: %w", err)
	}

	return nil
}

// generateFigureSequence generates the next figure sequence number for a chapter
func generateFigureSequence(existing []types.Figure) int {
	maxSequence := 0
	for _, figure := range existing {
		if figure.Sequence > maxSequence {
			maxSequence = figure.Sequence
		}
	}
	return maxSequence + 1
}

// generateTableSequence generates the next table sequence number for a chapter
func generateTableSequence(existing []types.Table) int {
	maxSequence := 0
	for _, table := range existing {
		if table.Sequence > maxSequence {
			maxSequence = table.Sequence
		}
	}
	return maxSequence + 1
}

// renumberFigures updates figure IDs when chapters are renumbered
func renumberFigures(figures []types.Figure, chapterNum types.ChapterNumber, offset int) []types.Figure {
	var updated []types.Figure
	
	for _, fig := range figures {
		if fig.Chapter == chapterNum {
			// Update the figure's chapter number
			newChapter := types.ChapterNumber(int(fig.Chapter) + offset)
			fig.Chapter = newChapter
			fig.ID = types.GenerateFigureID(newChapter, fig.Sequence)
		}
		updated = append(updated, fig)
	}
	
	return updated
}

// renumberTables updates table IDs when chapters are renumbered
func renumberTables(tables []types.Table, chapterNum types.ChapterNumber, offset int) []types.Table {
	var updated []types.Table
	
	for _, table := range tables {
		if table.Chapter == chapterNum {
			// Update the table's chapter number
			newChapter := types.ChapterNumber(int(table.Chapter) + offset)
			table.Chapter = newChapter
			table.ID = types.GenerateTableID(newChapter, table.Sequence)
		}
		updated = append(updated, table)
	}
	
	return updated
}

// generateSectionNumber creates a section number within a chapter
func generateSectionNumber(chapter types.ChapterNumber, level int, existing []types.Section) types.SectionNumber {
	// Find the highest section number at the current level
	maxAtLevel := 0
	
	for _, section := range existing {
		if len(section.Number) >= level && section.Number[0] == int(chapter) {
			if len(section.Number) == level {
				if section.Number[level-1] > maxAtLevel {
					maxAtLevel = section.Number[level-1]
				}
			}
		}
	}
	
	// Create new section number
	sectionParts := []int{int(chapter)}
	for i := 1; i < level; i++ {
		if i == level-1 {
			sectionParts = append(sectionParts, maxAtLevel+1)
		} else {
			sectionParts = append(sectionParts, 1)
		}
	}
	
	return types.NewSectionNumber(sectionParts...)
}

// parseFigureID parses a figure ID to extract chapter and sequence
func parseFigureID(id types.FigureID) (types.ChapterNumber, int, error) {
	// Expected format: fig-{chapter}.{sequence}
	idStr := string(id)
	if len(idStr) < 6 || idStr[:4] != "fig-" {
		return 0, 0, fmt.Errorf("invalid figure ID format: %s", id)
	}
	
	parts := idStr[4:] // Remove "fig-" prefix
	dotIndex := -1
	for i, c := range parts {
		if c == '.' {
			dotIndex = i
			break
		}
	}
	
	if dotIndex == -1 {
		return 0, 0, fmt.Errorf("invalid figure ID format: %s", id)
	}
	
	chapterStr := parts[:dotIndex]
	sequenceStr := parts[dotIndex+1:]
	
	chapter, err := strconv.Atoi(chapterStr)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid chapter number in figure ID: %s", id)
	}
	
	sequence, err := strconv.Atoi(sequenceStr)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid sequence number in figure ID: %s", id)
	}
	
	return types.ChapterNumber(chapter), sequence, nil
}

// parseTableID parses a table ID to extract chapter and sequence
func parseTableID(id types.TableID) (types.ChapterNumber, int, error) {
	// Expected format: table-{chapter}.{sequence}
	idStr := string(id)
	if len(idStr) < 8 || idStr[:6] != "table-" {
		return 0, 0, fmt.Errorf("invalid table ID format: %s", id)
	}
	
	parts := idStr[6:] // Remove "table-" prefix
	dotIndex := -1
	for i, c := range parts {
		if c == '.' {
			dotIndex = i
			break
		}
	}
	
	if dotIndex == -1 {
		return 0, 0, fmt.Errorf("invalid table ID format: %s", id)
	}
	
	chapterStr := parts[:dotIndex]
	sequenceStr := parts[dotIndex+1:]
	
	chapter, err := strconv.Atoi(chapterStr)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid chapter number in table ID: %s", id)
	}
	
	sequence, err := strconv.Atoi(sequenceStr)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid sequence number in table ID: %s", id)
	}
	
	return types.ChapterNumber(chapter), sequence, nil
}