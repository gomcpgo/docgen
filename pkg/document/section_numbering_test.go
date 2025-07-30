package document

import (
	"testing"
	"time"

	"github.com/gomcpgo/docgen/pkg/types"
)

func TestGenerateNextSectionNumber_Sequential(t *testing.T) {
	// Test sequential numbering at same level
	chapter := &types.Chapter{
		Number:    1,
		Title:     "Test Chapter",
		Sections:  []types.Section{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	manager := &Manager{}

	// Add first level-1 section
	sectionNum1, err := manager.generateNextSectionNumber(chapter, 1)
	if err != nil {
		t.Fatalf("Failed to generate first section number: %v", err)
	}
	expected1 := types.SectionNumber([]int{1, 1})
	if !sectionsEqual(sectionNum1, expected1) {
		t.Errorf("Expected first section to be [1,1], got %v", sectionNum1)
	}

	// Add section to chapter
	chapter.Sections = append(chapter.Sections, types.Section{
		Number: sectionNum1,
		Title:  "Section 1.1",
		Level:  1,
	})

	// Add second level-1 section
	sectionNum2, err := manager.generateNextSectionNumber(chapter, 1)
	if err != nil {
		t.Fatalf("Failed to generate second section number: %v", err)
	}
	expected2 := types.SectionNumber([]int{1, 2})
	if !sectionsEqual(sectionNum2, expected2) {
		t.Errorf("Expected second section to be [1,2], got %v", sectionNum2)
	}

	// Add section to chapter
	chapter.Sections = append(chapter.Sections, types.Section{
		Number: sectionNum2,
		Title:  "Section 1.2",
		Level:  1,
	})

	// Add third level-1 section
	sectionNum3, err := manager.generateNextSectionNumber(chapter, 1)
	if err != nil {
		t.Fatalf("Failed to generate third section number: %v", err)
	}
	expected3 := types.SectionNumber([]int{1, 3})
	if !sectionsEqual(sectionNum3, expected3) {
		t.Errorf("Expected third section to be [1,3], got %v", sectionNum3)
	}
}

func TestGenerateNextSectionNumber_Hierarchical(t *testing.T) {
	// Test hierarchical numbering
	chapter := &types.Chapter{
		Number:   1,
		Title:    "Test Chapter",
		Sections: []types.Section{
			{Number: types.SectionNumber([]int{1, 1}), Title: "Section 1.1", Level: 1},
			{Number: types.SectionNumber([]int{1, 2}), Title: "Section 1.2", Level: 1},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	manager := &Manager{}

	// Add first level-2 subsection (should be 1.2.1 since 1.2 is the latest level-1)
	sectionNum1, err := manager.generateNextSectionNumber(chapter, 2)
	if err != nil {
		t.Fatalf("Failed to generate first level-2 section number: %v", err)
	}
	expected1 := types.SectionNumber([]int{1, 2, 1})
	if !sectionsEqual(sectionNum1, expected1) {
		t.Errorf("Expected first level-2 section to be [1,2,1], got %v", sectionNum1)
	}

	// Add section to chapter
	chapter.Sections = append(chapter.Sections, types.Section{
		Number: sectionNum1,
		Title:  "Section 1.2.1",
		Level:  2,
	})

	// Add second level-2 subsection
	sectionNum2, err := manager.generateNextSectionNumber(chapter, 2)
	if err != nil {
		t.Fatalf("Failed to generate second level-2 section number: %v", err)
	}
	expected2 := types.SectionNumber([]int{1, 2, 2})
	if !sectionsEqual(sectionNum2, expected2) {
		t.Errorf("Expected second level-2 section to be [1,2,2], got %v", sectionNum2)
	}

	// Add section to chapter
	chapter.Sections = append(chapter.Sections, types.Section{
		Number: sectionNum2,
		Title:  "Section 1.2.2",
		Level:  2,
	})

	// Add level-1 section after subsections
	sectionNum3, err := manager.generateNextSectionNumber(chapter, 1)
	if err != nil {
		t.Fatalf("Failed to generate level-1 section after subsections: %v", err)
	}
	expected3 := types.SectionNumber([]int{1, 3})
	if !sectionsEqual(sectionNum3, expected3) {
		t.Errorf("Expected level-1 section after subsections to be [1,3], got %v", sectionNum3)
	}

	// Add section to chapter
	chapter.Sections = append(chapter.Sections, types.Section{
		Number: sectionNum3,
		Title:  "Section 1.3",
		Level:  1,
	})

	// Add level-2 subsection under new level-1 section
	sectionNum4, err := manager.generateNextSectionNumber(chapter, 2)
	if err != nil {
		t.Fatalf("Failed to generate level-2 section under new parent: %v", err)
	}
	expected4 := types.SectionNumber([]int{1, 3, 1})
	if !sectionsEqual(sectionNum4, expected4) {
		t.Errorf("Expected level-2 section under new parent to be [1,3,1], got %v", sectionNum4)
	}
}

func TestGenerateNextSectionNumber_DeepNesting(t *testing.T) {
	// Test deep nesting levels
	chapter := &types.Chapter{
		Number: 2, // Chapter 2
		Title:  "Test Chapter 2",
		Sections: []types.Section{
			{Number: types.SectionNumber([]int{2, 1}), Title: "Section 2.1", Level: 1},
			{Number: types.SectionNumber([]int{2, 1, 1}), Title: "Section 2.1.1", Level: 2},
			{Number: types.SectionNumber([]int{2, 1, 1, 1}), Title: "Section 2.1.1.1", Level: 3},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	manager := &Manager{}

	// Add level-4 section
	sectionNum, err := manager.generateNextSectionNumber(chapter, 4)
	if err != nil {
		t.Fatalf("Failed to generate level-4 section number: %v", err)
	}
	expected := types.SectionNumber([]int{2, 1, 1, 1, 1})
	if !sectionsEqual(sectionNum, expected) {
		t.Errorf("Expected level-4 section to be [2,1,1,1,1], got %v", sectionNum)
	}

	// Add section to chapter
	chapter.Sections = append(chapter.Sections, types.Section{
		Number: sectionNum,
		Title:  "Section 2.1.1.1.1",
		Level:  4,
	})

	// Add another level-4 section
	sectionNum2, err := manager.generateNextSectionNumber(chapter, 4)
	if err != nil {
		t.Fatalf("Failed to generate second level-4 section number: %v", err)
	}
	expected2 := types.SectionNumber([]int{2, 1, 1, 1, 2})
	if !sectionsEqual(sectionNum2, expected2) {
		t.Errorf("Expected second level-4 section to be [2,1,1,1,2], got %v", sectionNum2)
	}
}

func TestGenerateNextSectionNumber_MissingParent(t *testing.T) {
	// Test behavior when parent levels are missing
	chapter := &types.Chapter{
		Number:    1,
		Title:     "Test Chapter",
		Sections:  []types.Section{}, // No existing sections
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	manager := &Manager{}

	// Try to add level-3 section when no level-1 or level-2 exist
	sectionNum, err := manager.generateNextSectionNumber(chapter, 3)
	if err != nil {
		t.Fatalf("Failed to generate level-3 section with missing parents: %v", err)
	}
	expected := types.SectionNumber([]int{1, 1, 1, 1})
	if !sectionsEqual(sectionNum, expected) {
		t.Errorf("Expected level-3 section with missing parents to be [1,1,1,1], got %v", sectionNum)
	}
}

func TestGenerateNextSectionNumber_ComplexScenario(t *testing.T) {
	// Test the exact scenario from the design document
	chapter := &types.Chapter{
		Number: 1,
		Title:  "Introduction to Self-Hosted LLMs",
		Sections: []types.Section{
			{Number: types.SectionNumber([]int{1, 1}), Title: "Definition and Overview", Level: 1},
			{Number: types.SectionNumber([]int{1, 2}), Title: "Advantages of Self-Hosting", Level: 1},
			{Number: types.SectionNumber([]int{1, 2, 1}), Title: "Data Privacy Benefits", Level: 2},
			{Number: types.SectionNumber([]int{1, 2, 2}), Title: "Cost Benefits", Level: 2},
			{Number: types.SectionNumber([]int{1, 3}), Title: "Challenges and Considerations", Level: 1},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	manager := &Manager{}

	// Add level-1 section - should be 1.4
	sectionNum1, err := manager.generateNextSectionNumber(chapter, 1)
	if err != nil {
		t.Fatalf("Failed to generate level-1 section: %v", err)
	}
	expected1 := types.SectionNumber([]int{1, 4})
	if !sectionsEqual(sectionNum1, expected1) {
		t.Errorf("Expected level-1 section to be [1,4], got %v", sectionNum1)
	}

	// Add level-2 section - should be 1.3.1 (under the latest level-1 section)
	sectionNum2, err := manager.generateNextSectionNumber(chapter, 2)
	if err != nil {
		t.Fatalf("Failed to generate level-2 section: %v", err)
	}
	expected2 := types.SectionNumber([]int{1, 3, 1})
	if !sectionsEqual(sectionNum2, expected2) {
		t.Errorf("Expected level-2 section to be [1,3,1], got %v", sectionNum2)
	}
}

// Helper function to compare section numbers
func sectionsEqual(a, b types.SectionNumber) bool {
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