package document

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gomcpgo/docgen/pkg/config"
	"github.com/gomcpgo/docgen/pkg/types"
)

// MockStorage implements the Storage interface for testing
type MockStorage struct {
	chapters       map[string]*types.Chapter
	sectionContent map[string]string // key: docID-chapterNum-sectionNum
	chapterContent map[string]string // key: docID-chapterNum
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		chapters:       make(map[string]*types.Chapter),
		sectionContent: make(map[string]string),
		chapterContent: make(map[string]string),
	}
}

func (m *MockStorage) LoadChapterMetadata(documentID string, chapterNumber int) (*types.Chapter, error) {
	key := documentID + "-" + string(rune(chapterNumber))
	chapter, exists := m.chapters[key]
	if !exists {
		return nil, fmt.Errorf("chapter not found")
	}
	return chapter, nil
}

func (m *MockStorage) LoadSectionContent(documentID string, chapterNumber int, sectionNumber types.SectionNumber) (string, error) {
	key := documentID + "-" + string(rune(chapterNumber)) + "-" + sectionNumber.String()
	content, exists := m.sectionContent[key]
	if !exists {
		return "", fmt.Errorf("section content not found")
	}
	return content, nil
}

func (m *MockStorage) SaveChapterContent(documentID string, chapterNumber int, content string) error {
	key := documentID + "-" + string(rune(chapterNumber))
	m.chapterContent[key] = content
	return nil
}

func (m *MockStorage) SaveSectionContent(documentID string, chapterNumber int, sectionNumber types.SectionNumber, content string) error {
	key := documentID + "-" + string(rune(chapterNumber)) + "-" + sectionNumber.String()
	m.sectionContent[key] = content
	return nil
}

func (m *MockStorage) DeleteSectionFile(documentID string, chapterNumber int, sectionNumber types.SectionNumber) error {
	key := documentID + "-" + string(rune(chapterNumber)) + "-" + sectionNumber.String()
	delete(m.sectionContent, key)
	return nil
}

func (m *MockStorage) CreateSectionsDirectory(documentID string, chapterNumber int) error {
	return nil // No-op for mock
}

func (m *MockStorage) SaveChapterMetadata(documentID string, chapter *types.Chapter) error {
	key := documentID + "-" + string(rune(int(chapter.Number)))
	m.chapters[key] = chapter
	return nil
}

// Implement other required Storage interface methods as no-ops for testing
func (m *MockStorage) CreateDocumentStructure(doc *types.Document) error                            { return nil }
func (m *MockStorage) DocumentExists(documentID string) (bool, error)                              { return true, nil }
func (m *MockStorage) DeleteDocument(documentID string) error                                       { return nil }
func (m *MockStorage) ListDocuments() ([]string, error)                                            { return []string{}, nil }
func (m *MockStorage) CreateChapterStructure(documentID string, chapter *types.Chapter) error     { return nil }
func (m *MockStorage) ChapterExists(documentID string, chapterNumber int) (bool, error)           { return true, nil }
func (m *MockStorage) DeleteChapter(documentID string, chapterNumber int) error                    { return nil }
func (m *MockStorage) SaveManifest(documentID string, manifest *types.Manifest) error             { return nil }
func (m *MockStorage) LoadManifest(documentID string) (*types.Manifest, error)                    { return nil, nil }
func (m *MockStorage) SaveStyle(documentID string, style *types.Style) error                      { return nil }
func (m *MockStorage) LoadStyle(documentID string) (*types.Style, error)                          { return nil, nil }
func (m *MockStorage) SaveStyleByName(styleName string, style *types.Style) error                { return nil }
func (m *MockStorage) LoadStyleByName(styleName string) (*types.Style, error)                    { return nil, nil }
func (m *MockStorage) EnsureDefaultStyle() error                                                  { return nil }
func (m *MockStorage) SavePandocConfig(documentID string, config *types.PandocConfig) error       { return nil }
func (m *MockStorage) LoadPandocConfig(documentID string) (*types.PandocConfig, error)            { return nil, nil }
func (m *MockStorage) LoadChapterContent(documentID string, chapterNumber int) (string, error)    { return "", nil }

func TestRebuildChapterMarkdown_SimpleStructure(t *testing.T) {
	// Create mock storage and manager
	mockStorage := NewMockStorage()
	cfg := &config.Config{}
	manager := NewManager(cfg, mockStorage)

	docID := types.DocumentID("test-doc")
	chapterNum := types.ChapterNumber(1)

	// Set up test data
	chapter := &types.Chapter{
		Number: chapterNum,
		Title:  "Introduction to Self-Hosted LLMs",
		Sections: []types.Section{
			{
				Number:    types.SectionNumber([]int{1, 1}),
				Title:     "Definition and Overview",
				Level:     1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				Number:    types.SectionNumber([]int{1, 2}),
				Title:     "Advantages of Self-Hosting",
				Level:     1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				Number:    types.SectionNumber([]int{1, 2, 1}),
				Title:     "Data Privacy Benefits",
				Level:     2,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Store chapter metadata
	mockStorage.chapters["test-doc-\x01"] = chapter

	// Store section content
	mockStorage.sectionContent["test-doc-\x01-1.1"] = "Self-hosted Large Language Models (LLMs) represent a paradigm shift in how organizations deploy AI."
	mockStorage.sectionContent["test-doc-\x01-1.2"] = "Self-hosting LLMs provides several key advantages over cloud-based solutions."
	mockStorage.sectionContent["test-doc-\x01-1.2.1"] = "When you self-host an LLM, your data never leaves your infrastructure."

	// Rebuild chapter markdown
	err := manager.RebuildChapterMarkdown(docID, chapterNum)
	if err != nil {
		t.Fatalf("Failed to rebuild chapter markdown: %v", err)
	}

	// Check the generated content
	generatedContent := mockStorage.chapterContent["test-doc-\x01"]
	
	// Verify chapter title
	if !strings.Contains(generatedContent, "# Chapter 1: Introduction to Self-Hosted LLMs") {
		t.Errorf("Generated content missing chapter title. Content: %s", generatedContent)
	}

	// Verify level-1 sections
	if !strings.Contains(generatedContent, "## 1.1 Definition and Overview") {
		t.Errorf("Generated content missing section 1.1 header. Content: %s", generatedContent)
	}
	if !strings.Contains(generatedContent, "## 1.2 Advantages of Self-Hosting") {
		t.Errorf("Generated content missing section 1.2 header. Content: %s", generatedContent)
	}

	// Verify level-2 section
	if !strings.Contains(generatedContent, "### 1.2.1 Data Privacy Benefits") {
		t.Errorf("Generated content missing section 1.2.1 header. Content: %s", generatedContent)
	}

	// Verify section content
	if !strings.Contains(generatedContent, "Self-hosted Large Language Models (LLMs) represent a paradigm shift") {
		t.Errorf("Generated content missing section 1.1 content. Content: %s", generatedContent)
	}
	if !strings.Contains(generatedContent, "When you self-host an LLM, your data never leaves your infrastructure") {
		t.Errorf("Generated content missing section 1.2.1 content. Content: %s", generatedContent)
	}
}

func TestRebuildChapterMarkdown_EmptyChapter(t *testing.T) {
	// Test rebuilding a chapter with no sections
	mockStorage := NewMockStorage()
	cfg := &config.Config{}
	manager := NewManager(cfg, mockStorage)

	docID := types.DocumentID("empty-doc")
	chapterNum := types.ChapterNumber(1)

	// Set up empty chapter
	chapter := &types.Chapter{
		Number:    chapterNum,
		Title:     "Empty Chapter",
		Sections:  []types.Section{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockStorage.chapters["empty-doc-\x01"] = chapter

	// Rebuild chapter markdown
	err := manager.RebuildChapterMarkdown(docID, chapterNum)
	if err != nil {
		t.Fatalf("Failed to rebuild empty chapter markdown: %v", err)
	}

	// Check the generated content
	generatedContent := mockStorage.chapterContent["empty-doc-\x01"]
	
	// Should only contain chapter title
	expectedContent := "# Chapter 1: Empty Chapter\n\n"
	if generatedContent != expectedContent {
		t.Errorf("Expected empty chapter content '%s', got '%s'", expectedContent, generatedContent)
	}
}

func TestRebuildChapterMarkdown_DeepNesting(t *testing.T) {
	// Test rebuilding with deep section nesting
	mockStorage := NewMockStorage()
	cfg := &config.Config{}
	manager := NewManager(cfg, mockStorage)

	docID := types.DocumentID("deep-doc")
	chapterNum := types.ChapterNumber(2)

	// Set up deeply nested sections
	chapter := &types.Chapter{
		Number: chapterNum,
		Title:  "Deep Nesting Chapter",
		Sections: []types.Section{
			{Number: types.SectionNumber([]int{2, 1}), Title: "Level 1", Level: 1},
			{Number: types.SectionNumber([]int{2, 1, 1}), Title: "Level 2", Level: 2},
			{Number: types.SectionNumber([]int{2, 1, 1, 1}), Title: "Level 3", Level: 3},
			{Number: types.SectionNumber([]int{2, 1, 1, 1, 1}), Title: "Level 4", Level: 4},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockStorage.chapters["deep-doc-\x02"] = chapter

	// Store section content
	mockStorage.sectionContent["deep-doc-\x02-2.1"] = "Level 1 content"
	mockStorage.sectionContent["deep-doc-\x02-2.1.1"] = "Level 2 content"
	mockStorage.sectionContent["deep-doc-\x02-2.1.1.1"] = "Level 3 content"
	mockStorage.sectionContent["deep-doc-\x02-2.1.1.1.1"] = "Level 4 content"

	// Rebuild chapter markdown
	err := manager.RebuildChapterMarkdown(docID, chapterNum)
	if err != nil {
		t.Fatalf("Failed to rebuild deep nested chapter markdown: %v", err)
	}

	// Check the generated content
	generatedContent := mockStorage.chapterContent["deep-doc-\x02"]
	
	// Verify correct header levels
	if !strings.Contains(generatedContent, "## 2.1 Level 1") {
		t.Errorf("Generated content missing level 1 header (##). Content: %s", generatedContent)
	}
	if !strings.Contains(generatedContent, "### 2.1.1 Level 2") {
		t.Errorf("Generated content missing level 2 header (###). Content: %s", generatedContent)
	}
	if !strings.Contains(generatedContent, "#### 2.1.1.1 Level 3") {
		t.Errorf("Generated content missing level 3 header (####). Content: %s", generatedContent)
	}
	if !strings.Contains(generatedContent, "##### 2.1.1.1.1 Level 4") {
		t.Errorf("Generated content missing level 4 header (#####). Content: %s", generatedContent)
	}

	// Verify all content is present
	if !strings.Contains(generatedContent, "Level 1 content") {
		t.Errorf("Generated content missing level 1 content")
	}
	if !strings.Contains(generatedContent, "Level 4 content") {
		t.Errorf("Generated content missing level 4 content")
	}
}