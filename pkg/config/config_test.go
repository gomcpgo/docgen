package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// Save original env vars
	originalRootDir := os.Getenv("DOCGEN_ROOT_DIR")
	originalPandocPath := os.Getenv("PANDOC_PATH")
	originalMaxDocs := os.Getenv("DOCGEN_MAX_DOCUMENTS")

	defer func() {
		// Restore original env vars
		os.Setenv("DOCGEN_ROOT_DIR", originalRootDir)
		os.Setenv("PANDOC_PATH", originalPandocPath)
		os.Setenv("DOCGEN_MAX_DOCUMENTS", originalMaxDocs)
	}()

	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
		check   func(*Config) bool
	}{
		{
			name: "default config",
			envVars: map[string]string{
				"DOCGEN_ROOT_DIR": "/tmp/docgen",
			},
			wantErr: false,
			check: func(c *Config) bool {
				return c.RootDir == "/tmp/docgen" &&
					c.PandocPath == "pandoc" &&
					c.MaxDocuments == 100 &&
					c.MaxFileSize == 10*1024*1024
			},
		},
		{
			name: "custom config",
			envVars: map[string]string{
				"DOCGEN_ROOT_DIR":      "/custom/path",
				"PANDOC_PATH":          "/usr/bin/pandoc",
				"DOCGEN_MAX_DOCUMENTS": "50",
			},
			wantErr: false,
			check: func(c *Config) bool {
				return c.RootDir == "/custom/path" &&
					c.PandocPath == "/usr/bin/pandoc" &&
					c.MaxDocuments == 50 &&
					c.ExportsDir == "/custom/path/exports"
			},
		},
		{
			name:    "missing root dir",
			envVars: map[string]string{},
			wantErr: true,
		},
		{
			name: "invalid max documents",
			envVars: map[string]string{
				"DOCGEN_ROOT_DIR":      "/tmp/docgen",
				"DOCGEN_MAX_DOCUMENTS": "invalid",
			},
			wantErr: true,
		},
		{
			name: "negative max documents",
			envVars: map[string]string{
				"DOCGEN_ROOT_DIR":      "/tmp/docgen",
				"DOCGEN_MAX_DOCUMENTS": "-1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all env vars first
			os.Unsetenv("DOCGEN_ROOT_DIR")
			os.Unsetenv("PANDOC_PATH")
			os.Unsetenv("DOCGEN_MAX_DOCUMENTS")

			// Set test env vars
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			cfg, err := LoadConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil {
				if !tt.check(cfg) {
					t.Errorf("LoadConfig() configuration validation failed")
				}
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				RootDir:      "/tmp/docgen",
				PandocPath:   "pandoc",
				MaxDocuments: 100,
				MaxFileSize:  10 * 1024 * 1024,
				ExportTimeout: 5 * time.Minute,
			},
			wantErr: false,
		},
		{
			name: "empty root dir",
			config: &Config{
				RootDir:      "",
				PandocPath:   "pandoc",
				MaxDocuments: 100,
			},
			wantErr: true,
		},
		{
			name: "empty pandoc path",
			config: &Config{
				RootDir:      "/tmp/docgen",
				PandocPath:   "",
				MaxDocuments: 100,
			},
			wantErr: true,
		},
		{
			name: "zero max documents",
			config: &Config{
				RootDir:      "/tmp/docgen",
				PandocPath:   "pandoc",
				MaxDocuments: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_DocumentPath(t *testing.T) {
	cfg := &Config{
		RootDir: "/tmp/docgen",
	}

	tests := []struct {
		name   string
		docID  string
		want   string
	}{
		{"simple id", "doc1", "/tmp/docgen/doc1"},
		{"complex id", "my-book-2023", "/tmp/docgen/my-book-2023"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cfg.DocumentPath(tt.docID); got != tt.want {
				t.Errorf("Config.DocumentPath() = %v, want %v", got, tt.want)
			}
		})
	}
}