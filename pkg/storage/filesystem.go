package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gomcpgo/docgen/pkg/config"
	"github.com/gomcpgo/docgen/pkg/types"
	"gopkg.in/yaml.v3"
)

// Storage defines the interface for document storage operations
type Storage interface {
	// Document operations
	CreateDocumentStructure(doc *types.Document) error
	DocumentExists(documentID string) (bool, error)
	DeleteDocument(documentID string) error
	ListDocuments() ([]string, error)

	// Chapter operations
	CreateChapterStructure(documentID string, chapter *types.Chapter) error
	ChapterExists(documentID string, chapterNumber int) (bool, error)
	DeleteChapter(documentID string, chapterNumber int) error

	// Manifest operations
	SaveManifest(documentID string, manifest *types.Manifest) error
	LoadManifest(documentID string) (*types.Manifest, error)

	// Style operations (document-specific - deprecated)
	SaveStyle(documentID string, style *types.Style) error
	LoadStyle(documentID string) (*types.Style, error)
	
	// Style operations (by name in styles folder)
	SaveStyleByName(styleName string, style *types.Style) error
	LoadStyleByName(styleName string) (*types.Style, error)
	EnsureDefaultStyle() error

	// Pandoc config operations
	SavePandocConfig(documentID string, config *types.PandocConfig) error
	LoadPandocConfig(documentID string) (*types.PandocConfig, error)

	// Chapter content operations
	SaveChapterContent(documentID string, chapterNumber int, content string) error
	LoadChapterContent(documentID string, chapterNumber int) (string, error)

	// Chapter metadata operations
	SaveChapterMetadata(documentID string, chapter *types.Chapter) error
	LoadChapterMetadata(documentID string, chapterNumber int) (*types.Chapter, error)
	
	// Section file operations
	SaveSectionContent(documentID string, chapterNumber int, sectionNumber types.SectionNumber, content string) error
	LoadSectionContent(documentID string, chapterNumber int, sectionNumber types.SectionNumber) (string, error)
	DeleteSectionFile(documentID string, chapterNumber int, sectionNumber types.SectionNumber) error
	CreateSectionsDirectory(documentID string, chapterNumber int) error
}

// FileSystemStorage implements Storage using the local filesystem
type FileSystemStorage struct {
	config *config.Config
}

// NewFileSystemStorage creates a new filesystem storage instance
func NewFileSystemStorage(cfg *config.Config) Storage {
	return &FileSystemStorage{
		config: cfg,
	}
}

// CreateDocumentStructure creates the directory structure for a new document
func (fs *FileSystemStorage) CreateDocumentStructure(doc *types.Document) error {
	docID := string(doc.ID)
	docPath := fs.config.DocumentPath(docID)

	// Create document root directory
	if err := os.MkdirAll(docPath, 0755); err != nil {
		return fmt.Errorf("failed to create document directory: %w", err)
	}

	// Create chapters directory
	chaptersPath := filepath.Join(docPath, "chapters")
	if err := os.MkdirAll(chaptersPath, 0755); err != nil {
		return fmt.Errorf("failed to create chapters directory: %w", err)
	}

	// Create assets/images directory
	assetsPath := fs.config.AssetsPath(docID)
	if err := os.MkdirAll(assetsPath, 0755); err != nil {
		return fmt.Errorf("failed to create assets directory: %w", err)
	}

	// Create initial manifest
	manifest := &types.Manifest{
		Document:      *doc,
		ChapterCounts: make(map[types.ChapterNumber]types.ChapterCount),
		CreatedAt:     doc.CreatedAt,
		UpdatedAt:     doc.UpdatedAt,
	}
	if err := fs.SaveManifest(docID, manifest); err != nil {
		return fmt.Errorf("failed to create manifest: %w", err)
	}

	// Note: Document-specific styles are no longer created
	// Styles are now managed globally in the styles/ folder

	// Create default pandoc config
	pandocConfig := types.DefaultPandocConfig()
	if err := fs.SavePandocConfig(docID, &pandocConfig); err != nil {
		return fmt.Errorf("failed to create pandoc config: %w", err)
	}

	return nil
}

// DocumentExists checks if a document exists
func (fs *FileSystemStorage) DocumentExists(documentID string) (bool, error) {
	docPath := fs.config.DocumentPath(documentID)
	_, err := os.Stat(docPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check document existence: %w", err)
	}
	return true, nil
}

// DeleteDocument removes a document and all its contents
func (fs *FileSystemStorage) DeleteDocument(documentID string) error {
	docPath := fs.config.DocumentPath(documentID)
	if err := os.RemoveAll(docPath); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	return nil
}

// ListDocuments returns a list of all document IDs
func (fs *FileSystemStorage) ListDocuments() ([]string, error) {
	// Create root directory if it doesn't exist
	if err := os.MkdirAll(fs.config.RootDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create root directory: %w", err)
	}

	entries, err := os.ReadDir(fs.config.RootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	var documents []string
	for _, entry := range entries {
		if entry.IsDir() {
			// Verify it's a valid document by checking for manifest
			manifestPath := fs.config.ManifestPath(entry.Name())
			if _, err := os.Stat(manifestPath); err == nil {
				documents = append(documents, entry.Name())
			}
		}
	}

	return documents, nil
}

// CreateChapterStructure creates the directory structure for a new chapter
func (fs *FileSystemStorage) CreateChapterStructure(documentID string, chapter *types.Chapter) error {
	chapterPath := fs.config.ChapterPath(documentID, int(chapter.Number))

	// Create chapter directory
	if err := os.MkdirAll(chapterPath, 0755); err != nil {
		return fmt.Errorf("failed to create chapter directory: %w", err)
	}

	// Create sections subdirectory
	if err := fs.CreateSectionsDirectory(documentID, int(chapter.Number)); err != nil {
		return fmt.Errorf("failed to create sections directory: %w", err)
	}

	// Save chapter content
	if err := fs.SaveChapterContent(documentID, int(chapter.Number), chapter.Content); err != nil {
		return fmt.Errorf("failed to save chapter content: %w", err)
	}

	// Save chapter metadata
	if err := fs.SaveChapterMetadata(documentID, chapter); err != nil {
		return fmt.Errorf("failed to save chapter metadata: %w", err)
	}

	return nil
}

// ChapterExists checks if a chapter exists
func (fs *FileSystemStorage) ChapterExists(documentID string, chapterNumber int) (bool, error) {
	chapterPath := fs.config.ChapterPath(documentID, chapterNumber)
	_, err := os.Stat(chapterPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check chapter existence: %w", err)
	}
	return true, nil
}

// DeleteChapter removes a chapter directory and all its contents
func (fs *FileSystemStorage) DeleteChapter(documentID string, chapterNumber int) error {
	chapterPath := fs.config.ChapterPath(documentID, chapterNumber)
	if err := os.RemoveAll(chapterPath); err != nil {
		return fmt.Errorf("failed to delete chapter: %w", err)
	}
	return nil
}

// SaveManifest saves the document manifest
func (fs *FileSystemStorage) SaveManifest(documentID string, manifest *types.Manifest) error {
	manifestPath := fs.config.ManifestPath(documentID)
	return fs.saveYAMLFile(manifestPath, manifest)
}

// LoadManifest loads the document manifest
func (fs *FileSystemStorage) LoadManifest(documentID string) (*types.Manifest, error) {
	manifestPath := fs.config.ManifestPath(documentID)
	var manifest types.Manifest
	if err := fs.loadYAMLFile(manifestPath, &manifest); err != nil {
		return nil, err
	}
	return &manifest, nil
}

// SaveStyle saves the document style
func (fs *FileSystemStorage) SaveStyle(documentID string, style *types.Style) error {
	stylePath := fs.config.StylePath(documentID)
	return fs.saveYAMLFile(stylePath, style)
}

// LoadStyle loads the document style
func (fs *FileSystemStorage) LoadStyle(documentID string) (*types.Style, error) {
	stylePath := fs.config.StylePath(documentID)
	var style types.Style
	if err := fs.loadYAMLFile(stylePath, &style); err != nil {
		return nil, err
	}
	return &style, nil
}

// SavePandocConfig saves the pandoc configuration
func (fs *FileSystemStorage) SavePandocConfig(documentID string, config *types.PandocConfig) error {
	configPath := fs.config.PandocConfigPath(documentID)
	return fs.saveYAMLFile(configPath, config)
}

// LoadPandocConfig loads the pandoc configuration
func (fs *FileSystemStorage) LoadPandocConfig(documentID string) (*types.PandocConfig, error) {
	configPath := fs.config.PandocConfigPath(documentID)
	var pandocConfig types.PandocConfig
	if err := fs.loadYAMLFile(configPath, &pandocConfig); err != nil {
		return nil, err
	}
	return &pandocConfig, nil
}

// SaveChapterContent saves chapter content to the chapter.md file
func (fs *FileSystemStorage) SaveChapterContent(documentID string, chapterNumber int, content string) error {
	contentPath := fs.config.ChapterContentPath(documentID, chapterNumber)
	if err := os.WriteFile(contentPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to save chapter content: %w", err)
	}
	return nil
}

// LoadChapterContent loads chapter content from the chapter.md file
func (fs *FileSystemStorage) LoadChapterContent(documentID string, chapterNumber int) (string, error) {
	contentPath := fs.config.ChapterContentPath(documentID, chapterNumber)
	content, err := os.ReadFile(contentPath)
	if err != nil {
		return "", fmt.Errorf("failed to load chapter content: %w", err)
	}
	return string(content), nil
}

// SaveChapterMetadata saves chapter metadata to the metadata.yaml file
func (fs *FileSystemStorage) SaveChapterMetadata(documentID string, chapter *types.Chapter) error {
	metadataPath := fs.config.ChapterMetadataPath(documentID, int(chapter.Number))
	return fs.saveYAMLFile(metadataPath, chapter)
}

// LoadChapterMetadata loads chapter metadata from the metadata.yaml file
func (fs *FileSystemStorage) LoadChapterMetadata(documentID string, chapterNumber int) (*types.Chapter, error) {
	metadataPath := fs.config.ChapterMetadataPath(documentID, chapterNumber)
	var chapter types.Chapter
	if err := fs.loadYAMLFile(metadataPath, &chapter); err != nil {
		return nil, err
	}
	return &chapter, nil
}

// saveYAMLFile saves data to a YAML file
func (fs *FileSystemStorage) saveYAMLFile(filePath string, data interface{}) error {
	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	defer encoder.Close()

	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode YAML: %w", err)
	}

	return nil
}

// loadYAMLFile loads data from a YAML file
func (fs *FileSystemStorage) loadYAMLFile(filePath string, data interface{}) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(data); err != nil {
		return fmt.Errorf("failed to decode YAML: %w", err)
	}

	return nil
}

// SaveSectionContent saves section content to individual section file
func (fs *FileSystemStorage) SaveSectionContent(documentID string, chapterNumber int, sectionNumber types.SectionNumber, content string) error {
	sectionPath := fs.config.SectionPath(documentID, chapterNumber, sectionNumber.String())
	
	// Ensure sections directory exists
	if err := fs.CreateSectionsDirectory(documentID, chapterNumber); err != nil {
		return fmt.Errorf("failed to create sections directory: %w", err)
	}
	
	if err := os.WriteFile(sectionPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to save section content: %w", err)
	}
	return nil
}

// LoadSectionContent loads section content from individual section file
func (fs *FileSystemStorage) LoadSectionContent(documentID string, chapterNumber int, sectionNumber types.SectionNumber) (string, error) {
	sectionPath := fs.config.SectionPath(documentID, chapterNumber, sectionNumber.String())
	content, err := os.ReadFile(sectionPath)
	if err != nil {
		return "", fmt.Errorf("failed to load section content: %w", err)
	}
	return string(content), nil
}

// DeleteSectionFile removes a section file
func (fs *FileSystemStorage) DeleteSectionFile(documentID string, chapterNumber int, sectionNumber types.SectionNumber) error {
	sectionPath := fs.config.SectionPath(documentID, chapterNumber, sectionNumber.String())
	if err := os.Remove(sectionPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete section file: %w", err)
	}
	return nil
}

// CreateSectionsDirectory creates the sections subdirectory for a chapter
func (fs *FileSystemStorage) CreateSectionsDirectory(documentID string, chapterNumber int) error {
	sectionsDir := fs.config.SectionsPath(documentID, chapterNumber)
	if err := os.MkdirAll(sectionsDir, 0755); err != nil {
		return fmt.Errorf("failed to create sections directory: %w", err)
	}
	return nil
}

// SaveStyleByName saves a style by name to the styles folder
func (fs *FileSystemStorage) SaveStyleByName(styleName string, style *types.Style) error {
	stylePath := fs.config.StyleByNamePath(styleName)
	return fs.saveYAMLFile(stylePath, style)
}

// LoadStyleByName loads a style by name from the styles folder
func (fs *FileSystemStorage) LoadStyleByName(styleName string) (*types.Style, error) {
	stylePath := fs.config.StyleByNamePath(styleName)
	var style types.Style
	if err := fs.loadYAMLFile(stylePath, &style); err != nil {
		return nil, err
	}
	return &style, nil
}

// EnsureDefaultStyle creates the default style if it doesn't exist
func (fs *FileSystemStorage) EnsureDefaultStyle() error {
	defaultStylePath := fs.config.StyleByNamePath("default")
	
	// Check if default style already exists
	if _, err := os.Stat(defaultStylePath); err == nil {
		return nil // Already exists
	}
	
	// Create styles directory if it doesn't exist
	stylesDir := fs.config.StylesPath()
	if err := os.MkdirAll(stylesDir, 0755); err != nil {
		return fmt.Errorf("failed to create styles directory: %w", err)
	}
	
	// Create default style
	defaultStyle := types.DefaultStyle()
	if err := fs.SaveStyleByName("default", &defaultStyle); err != nil {
		return fmt.Errorf("failed to create default style: %w", err)
	}
	
	return nil
}