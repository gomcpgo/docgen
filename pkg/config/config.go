package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Config holds the configuration for the docgen MCP server
type Config struct {
	// RootDir is the root directory where documents are stored
	RootDir string
	
	// PandocPath is the path to the pandoc executable
	PandocPath string
	
	// DefaultStylePath is the path to the default style template
	DefaultStylePath string
	
	// MaxDocuments is the maximum number of documents allowed
	MaxDocuments int
	
	// ExportsDir is the directory for exported documents (within RootDir)
	ExportsDir string
	
	// MaxFileSize is the maximum file size for uploads in bytes
	MaxFileSize int64
	
	// ExportTimeout is the timeout for export operations
	ExportTimeout time.Duration
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() (*Config, error) {
	cfg := &Config{
		PandocPath:    "pandoc",
		MaxDocuments:  100,
		MaxFileSize:   10 * 1024 * 1024, // 10MB
		ExportTimeout: 5 * time.Minute,
	}
	
	// DOCGEN_ROOT_DIR (required)
	cfg.RootDir = os.Getenv("DOCGEN_ROOT_DIR")
	if cfg.RootDir == "" {
		return nil, fmt.Errorf("DOCGEN_ROOT_DIR environment variable is required")
	}
	
	// Set exports directory within root dir
	cfg.ExportsDir = filepath.Join(cfg.RootDir, "exports")
	
	// PANDOC_PATH (optional)
	if val := os.Getenv("PANDOC_PATH"); val != "" {
		cfg.PandocPath = val
	}
	
	// DOCGEN_DEFAULT_STYLE (optional)
	if val := os.Getenv("DOCGEN_DEFAULT_STYLE"); val != "" {
		cfg.DefaultStylePath = val
	}
	
	// DOCGEN_MAX_DOCUMENTS (optional)
	if val := os.Getenv("DOCGEN_MAX_DOCUMENTS"); val != "" {
		maxDocs, err := strconv.Atoi(val)
		if err != nil {
			return nil, fmt.Errorf("invalid DOCGEN_MAX_DOCUMENTS value: %s", val)
		}
		if maxDocs <= 0 {
			return nil, fmt.Errorf("DOCGEN_MAX_DOCUMENTS must be positive")
		}
		cfg.MaxDocuments = maxDocs
	}
	
	
	// DOCGEN_MAX_FILE_SIZE (optional)
	if val := os.Getenv("DOCGEN_MAX_FILE_SIZE"); val != "" {
		maxSize, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid DOCGEN_MAX_FILE_SIZE value: %s", val)
		}
		if maxSize <= 0 {
			return nil, fmt.Errorf("DOCGEN_MAX_FILE_SIZE must be positive")
		}
		cfg.MaxFileSize = maxSize
	}
	
	// DOCGEN_EXPORT_TIMEOUT (optional)
	if val := os.Getenv("DOCGEN_EXPORT_TIMEOUT"); val != "" {
		timeoutSecs, err := strconv.Atoi(val)
		if err != nil {
			return nil, fmt.Errorf("invalid DOCGEN_EXPORT_TIMEOUT value: %s", val)
		}
		if timeoutSecs <= 0 {
			return nil, fmt.Errorf("DOCGEN_EXPORT_TIMEOUT must be positive")
		}
		cfg.ExportTimeout = time.Duration(timeoutSecs) * time.Second
	}
	
	return cfg, cfg.Validate()
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.RootDir == "" {
		return fmt.Errorf("root directory is required")
	}
	
	if c.PandocPath == "" {
		return fmt.Errorf("pandoc path is required")
	}
	
	if c.MaxDocuments <= 0 {
		return fmt.Errorf("max documents must be positive")
	}
	
	if c.MaxFileSize <= 0 {
		return fmt.Errorf("max file size must be positive")
	}
	
	if c.ExportTimeout <= 0 {
		return fmt.Errorf("export timeout must be positive")
	}
	
	return nil
}

// DocumentPath returns the full path to a document directory
func (c *Config) DocumentPath(documentID string) string {
	return filepath.Join(c.RootDir, documentID)
}

// ChapterPath returns the full path to a chapter directory
func (c *Config) ChapterPath(documentID string, chapterNumber int) string {
	return filepath.Join(c.DocumentPath(documentID), "chapters", fmt.Sprintf("%02d", chapterNumber))
}

// AssetsPath returns the full path to the assets directory
func (c *Config) AssetsPath(documentID string) string {
	return filepath.Join(c.DocumentPath(documentID), "assets", "images")
}

// ManifestPath returns the full path to the manifest file
func (c *Config) ManifestPath(documentID string) string {
	return filepath.Join(c.DocumentPath(documentID), "manifest.yaml")
}

// StylePath returns the full path to the style file
func (c *Config) StylePath(documentID string) string {
	return filepath.Join(c.DocumentPath(documentID), "style.yaml")
}

// PandocConfigPath returns the full path to the pandoc config file
func (c *Config) PandocConfigPath(documentID string) string {
	return filepath.Join(c.DocumentPath(documentID), "pandoc-config.yaml")
}

// ChapterContentPath returns the full path to a chapter's content file
func (c *Config) ChapterContentPath(documentID string, chapterNumber int) string {
	return filepath.Join(c.ChapterPath(documentID, chapterNumber), "chapter.md")
}

// ChapterMetadataPath returns the full path to a chapter's metadata file
func (c *Config) ChapterMetadataPath(documentID string, chapterNumber int) string {
	return filepath.Join(c.ChapterPath(documentID, chapterNumber), "metadata.yaml")
}

// ExportPath returns the path for export files
func (c *Config) ExportPath(documentID, format string) string {
	filename := fmt.Sprintf("%s.%s", documentID, format)
	return filepath.Join(c.ExportsDir, filename)
}