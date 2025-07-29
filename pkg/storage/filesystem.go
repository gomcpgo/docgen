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

	// Style operations
	SaveStyle(documentID string, style *types.Style) error
	LoadStyle(documentID string) (*types.Style, error)

	// Pandoc config operations
	SavePandocConfig(documentID string, config *types.PandocConfig) error
	LoadPandocConfig(documentID string) (*types.PandocConfig, error)

	// Chapter content operations
	SaveChapterContent(documentID string, chapterNumber int, content string) error
	LoadChapterContent(documentID string, chapterNumber int) (string, error)

	// Chapter metadata operations
	SaveChapterMetadata(documentID string, chapter *types.Chapter) error
	LoadChapterMetadata(documentID string, chapterNumber int) (*types.Chapter, error)
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

	// Create default style
	style := types.DefaultStyle()
	if err := fs.SaveStyle(docID, &style); err != nil {
		return fmt.Errorf("failed to create style file: %w", err)
	}

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