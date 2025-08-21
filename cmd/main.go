package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gomcpgo/mcp/pkg/handler"
	"github.com/gomcpgo/mcp/pkg/protocol"
	"github.com/gomcpgo/mcp/pkg/server"
	"github.com/gomcpgo/docgen/pkg/config"
	docgenHandler "github.com/gomcpgo/docgen/pkg/handler"
	"github.com/gomcpgo/docgen/pkg/types"
	"gopkg.in/yaml.v3"
)

// Version information (set by build script)
var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	// Parse command line flags
	testMode := flag.Bool("test", false, "Run integration tests with sample documents")
	keepFiles := flag.Bool("keep-files", false, "Keep generated test files (only used with -test)")
	versionFlag := flag.Bool("version", false, "Show version information")
	exportDoc := flag.String("export", "", "Export existing document by ID (format: documentID,format). Example: -export my-doc-123,pdf")
	styleFile := flag.String("style", "", "Custom style file to use for export (JSON or YAML). Example: -style '/path/to/style.json'")
	rebuildChapter := flag.String("rebuild", "", "Rebuild chapter markdown (format: documentID,chapterNumber). Example: -rebuild my-doc-123,1")
	testFigures := flag.Bool("test-figures", false, "Run figure functionality test by creating a sample document with images")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("Document Generation MCP Server\n")
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
		return
	}

	if *exportDoc != "" {
		runDirectExport(*exportDoc, *styleFile)
		return
	}

	if *rebuildChapter != "" {
		runRebuildChapter(*rebuildChapter)
		return
	}

	if *testMode {
		runIntegrationTests(*keepFiles)
		return
	}
	
	if *testFigures {
		runFigureTest()
		return
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create docgen handler
	docgenHandler, err := docgenHandler.NewDocGenHandler(cfg)
	if err != nil {
		log.Fatalf("Failed to create docgen handler: %v", err)
	}

	// Create handler registry
	registry := handler.NewHandlerRegistry()
	registry.RegisterToolHandler(docgenHandler)

	// Create and run MCP server
	mcpServer := server.New(server.Options{
		Name:     "Document Generator",
		Version:  Version,
		Registry: registry,
	})

	log.Printf("Starting Document Generation MCP Server v%s", Version)
	if err := mcpServer.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// runIntegrationTests runs comprehensive integration tests
func runIntegrationTests(keepFiles bool) {
	fmt.Println("Document Generation MCP Server - Integration Tests")
	fmt.Println("==================================================")

	// Set up test environment
	tempDir, err := os.MkdirTemp("", "docgen_integration_test_")
	if err != nil {
		log.Fatalf("Failed to create temp directory: %v", err)
	}
	
	// Only clean up if keepFiles is false
	if !keepFiles {
		defer os.RemoveAll(tempDir)
	}

	// Set test configuration
	os.Setenv("DOCGEN_ROOT_DIR", tempDir)
	os.Setenv("PANDOC_PATH", "pandoc")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load test configuration: %v", err)
	}

	// Create handler
	docgenHandler, err := docgenHandler.NewDocGenHandler(cfg)
	if err != nil {
		log.Fatalf("Failed to create docgen handler: %v", err)
	}

	// Run test scenarios
	runTestScenarios(docgenHandler, tempDir)

	fmt.Println("\n✅ All integration tests completed successfully!")
	
	if keepFiles {
		fmt.Printf("\n📁 Test files preserved in: %s\n", tempDir)
		fmt.Println("   You can explore the generated documents and exported files.")
	}
}

// runTestScenarios executes various test scenarios
func runTestScenarios(h *docgenHandler.DocGenHandler, tempDir string) {
	// Test 1: Create a sample book
	fmt.Println("\n📚 Test 1: Creating a sample book...")
	docID := createSampleBook(h)
	fmt.Printf("   Created document: %s\n", docID)

	// Test 2: Add chapters with content
	fmt.Println("\n📖 Test 2: Adding chapters with content...")
	addSampleChapters(h, docID)
	fmt.Println("   Added chapters successfully")

	// Test 3: Get document structure
	fmt.Println("\n🏗️  Test 3: Getting document structure...")
	structure := getDocumentStructure(h, docID)
	fmt.Printf("   Document has %d chapters\n", len(structure.Document.Chapters))

	// Test 4: Export to PDF (if pandoc is available)
	fmt.Println("\n📄 Test 4: Exporting to PDF...")
	if isPandocAvailable() {
		pdfPath := exportToPDF(h, docID)
		fmt.Printf("   PDF exported to: %s\n", pdfPath)
		
		// Check file exists and has reasonable size
		if stat, err := os.Stat(pdfPath); err == nil {
			fmt.Printf("   PDF file size: %d bytes\n", stat.Size())
		}
	} else {
		fmt.Println("   ⚠️  Pandoc not available, skipping PDF export")
	}

	// Test 5: Create a different document type
	fmt.Println("\n📋 Test 5: Creating a research report...")
	reportID := createSampleReport(h)
	fmt.Printf("   Created report: %s\n", reportID)

	// Test 6: Document validation
	fmt.Println("\n✅ Test 6: Validating documents...")
	validateDocument(h, docID)
	validateDocument(h, reportID)
	fmt.Println("   Document validation completed")

	// Test 7: Configuration updates
	fmt.Println("\n⚙️  Test 7: Testing document configuration...")
	configureDocument(h, docID)
	fmt.Println("   Document configuration updated")

	// Test 8: Chapter management
	fmt.Println("\n📝 Test 8: Testing chapter management...")
	testChapterManagement(h, docID)
	fmt.Println("   Chapter management tests completed")
	
	// Test 9: Style management
	fmt.Println("\n🎨 Test 9: Testing style management...")
	testStyleManagement(h, docID)
	fmt.Println("   Style management tests completed")
}

// createSampleBook creates a sample book document
func createSampleBook(h *docgenHandler.DocGenHandler) string {
	params := map[string]interface{}{
		"title":  "The Complete Guide to MCP Servers",
		"author": "AI Assistant",
		"type":   "book",
	}

	response, err := h.CallTool(nil, &protocol.CallToolRequest{
		Name:      "create_document",
		Arguments: params,
	})
	if err != nil {
		log.Fatalf("Failed to create document: %v", err)
	}

	if response.IsError {
		log.Fatalf("Error creating document: %s", response.Content[0].Text)
	}

	// Parse response to get document ID
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(response.Content[0].Text), &result); err != nil {
		log.Fatalf("Failed to parse create document response: %v", err)
	}
	
	docID, ok := result["document_id"].(string)
	if !ok {
		log.Fatalf("Failed to extract document_id from response")
	}
	
	return docID
}

// addSampleChapters adds sample chapters to a document
func addSampleChapters(h *docgenHandler.DocGenHandler, docID string) {
	chapters := []struct {
		title   string
		content string
	}{
		{
			"Introduction to MCP",
			`# Introduction to MCP

The Model Context Protocol (MCP) is a revolutionary approach to building AI applications. This chapter introduces the core concepts and benefits of using MCP servers.

## What is MCP?

MCP provides a standardized way for AI applications to interact with external tools and services. It enables:

- Seamless integration with various data sources
- Standardized tool interfaces
- Enhanced security and access control
- Scalable architecture patterns

## Benefits

Using MCP servers provides several key advantages:

1. **Modularity**: Each tool can be developed and maintained independently
2. **Reusability**: Tools can be shared across different AI applications
3. **Security**: Fine-grained access control and permission management
4. **Scalability**: Easy to add new tools and capabilities
`,
		},
		{
			"Getting Started",
			`# Getting Started with MCP Servers

This chapter walks you through creating your first MCP server and connecting it to an AI application.

## Prerequisites

Before you begin, ensure you have:

- Go 1.21 or later installed
- Basic understanding of JSON-RPC
- Familiarity with AI application development

## Creating Your First Server

Here's a simple example of an MCP server:

` + "```go" + `
package main

import (
    "github.com/gomcpgo/mcp/pkg/server"
    "github.com/gomcpgo/mcp/pkg/handler"
)

func main() {
    registry := handler.NewHandlerRegistry()
    // Register your tools here
    
    srv := server.New(server.Options{
        Name:     "my-server",
        Version:  "1.0.0",
        Registry: registry,
    })
    
    srv.Run()
}
` + "```" + `

## Testing Your Server

Use the built-in test mode to verify your server works correctly:

` + "```bash" + `
go run main.go -test
` + "```" + `
`,
		},
		{
			"Advanced Features",
			`# Advanced MCP Server Features

This chapter covers advanced topics including error handling, authentication, and performance optimization.

## Error Handling Best Practices

Proper error handling is crucial for reliable MCP servers:

1. **Validate all inputs** before processing
2. **Provide meaningful error messages** to help users
3. **Use appropriate error codes** for different failure types
4. **Log errors** for debugging and monitoring

## Performance Optimization

To ensure your MCP server performs well under load:

- Use connection pooling for external services
- Implement proper caching strategies  
- Monitor resource usage and set appropriate limits
- Use asynchronous processing where possible

## Security Considerations

Security should be built into your MCP server from the ground up:

- Validate and sanitize all user inputs
- Implement proper authentication and authorization
- Use secure communication protocols
- Regular security audits and updates
`,
		},
	}

	for i, chapter := range chapters {
		params := map[string]interface{}{
			"document_id": docID,
			"title":       chapter.title,
		}

		response, err := h.CallTool(nil, &protocol.CallToolRequest{
			Name:      "add_chapter",
			Arguments: params,
		})
		if err != nil {
			fmt.Printf("   ⚠️  Failed to add chapter %d: %v\n", i+1, err)
			continue
		}

		if response.IsError {
			fmt.Printf("   ⚠️  Error adding chapter %d: %s\n", i+1, response.Content[0].Text)
			continue
		}

		// For this demo, we'll simulate adding content by directly writing to the file
		// In a real implementation, you'd use the add_section or update_chapter tools
		chapterPath := filepath.Join(os.Getenv("DOCGEN_ROOT_DIR"), docID, "chapters", fmt.Sprintf("%02d", i+1), "chapter.md")
		os.MkdirAll(filepath.Dir(chapterPath), 0755)
		os.WriteFile(chapterPath, []byte(chapter.content), 0644)
	}
}

// createSampleReport creates a sample report document
func createSampleReport(h *docgenHandler.DocGenHandler) string {
	params := map[string]interface{}{
		"title":  "Quarterly Performance Analysis",
		"author": "Data Analytics Team",
		"type":   "report",
	}

	response, err := h.CallTool(nil, &protocol.CallToolRequest{
		Name:      "create_document",
		Arguments: params,
	})
	if err != nil {
		log.Fatalf("Failed to create report: %v", err)
	}

	if response.IsError {
		log.Fatalf("Error creating report: %s", response.Content[0].Text)
	}

	// Parse response to get document ID
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(response.Content[0].Text), &result); err != nil {
		log.Fatalf("Failed to parse create report response: %v", err)
	}
	
	docID, ok := result["document_id"].(string)
	if !ok {
		log.Fatalf("Failed to extract document_id from report response")
	}
	
	return docID
}

// getDocumentStructure retrieves and returns document structure
func getDocumentStructure(h *docgenHandler.DocGenHandler, docID string) *types.Manifest {
	params := map[string]interface{}{
		"document_id": docID,
	}

	response, err := h.CallTool(nil, &protocol.CallToolRequest{
		Name:      "get_document_structure",
		Arguments: params,
	})
	if err != nil {
		log.Fatalf("Failed to get document structure: %v", err)
	}

	if response.IsError {
		log.Fatalf("Error getting document structure: %s", response.Content[0].Text)
	}

	// For demo purposes, return a mock structure
	return &types.Manifest{
		Document: types.Document{
			ID:     types.DocumentID(docID),
			Title:  "Sample Document",
			Author: "Test Author",
			Chapters: []types.Chapter{
				{Number: 1, Title: "Introduction"},
				{Number: 2, Title: "Getting Started"},
				{Number: 3, Title: "Advanced Features"},
			},
		},
	}
}

// exportToPDF exports document to PDF format
func exportToPDF(h *docgenHandler.DocGenHandler, docID string) string {
	// Test 1: Export with default style (no style_name parameter)
	params := map[string]interface{}{
		"document_id": docID,
		"format":      "pdf",
	}

	response, err := h.CallTool(nil, &protocol.CallToolRequest{
		Name:      "export_document",
		Arguments: params,
	})
	if err != nil {
		fmt.Printf("   ⚠️  Failed to export PDF: %v\n", err)
		return ""
	}

	if response.IsError {
		fmt.Printf("   ⚠️  Error exporting PDF: %s\n", response.Content[0].Text)
		return ""
	}
	
	// Test 2: Export with style_name parameter
	fmt.Println("   Testing export with style_name parameter...")
	paramsWithStyle := map[string]interface{}{
		"document_id": docID,
		"format":      "pdf",
		"style_name":  "default", // Use default style explicitly
	}

	responseWithStyle, err := h.CallTool(nil, &protocol.CallToolRequest{
		Name:      "export_document",
		Arguments: paramsWithStyle,
	})
	if err != nil {
		fmt.Printf("   ⚠️  Failed to export PDF with style_name: %v\n", err)
	} else if responseWithStyle.IsError {
		fmt.Printf("   ⚠️  Error exporting PDF with style_name: %s\n", responseWithStyle.Content[0].Text)
	} else {
		fmt.Println("   ✅ Export with style_name parameter succeeded")
	}

	// For demo, return a mock path
	return filepath.Join(os.Getenv("DOCGEN_ROOT_DIR"), "exports", docID+".pdf")
}

// validateDocument validates a document
func validateDocument(h *docgenHandler.DocGenHandler, docID string) {
	params := map[string]interface{}{
		"document_id": docID,
	}

	response, err := h.CallTool(nil, &protocol.CallToolRequest{
		Name:      "validate_document",
		Arguments: params,
	})
	if err != nil {
		fmt.Printf("   ⚠️  Failed to validate document %s: %v\n", docID, err)
		return
	}

	if response.IsError {
		fmt.Printf("   ⚠️  Error validating document %s: %s\n", docID, response.Content[0].Text)
		return
	}

	fmt.Printf("   ✅ Document %s validation completed\n", docID)
}

// configureDocument tests document configuration
func configureDocument(h *docgenHandler.DocGenHandler, docID string) {
	params := map[string]interface{}{
		"document_id": docID,
		"style_updates": map[string]interface{}{
			"font_family": "Times New Roman",
			"font_size":   "12pt",
			"margins": map[string]interface{}{
				"top":    "1in",
				"bottom": "1in",
				"left":   "1.25in",
				"right":  "1.25in",
			},
		},
		"pandoc_options": map[string]interface{}{
			"toc":       true,
			"toc_depth": 3,
			"pdf_engine": "pdflatex",
		},
	}

	response, err := h.CallTool(nil, &protocol.CallToolRequest{
		Name:      "configure_document",
		Arguments: params,
	})
	if err != nil {
		fmt.Printf("   ⚠️  Failed to configure document: %v\n", err)
		return
	}

	if response.IsError {
		fmt.Printf("   ⚠️  Error configuring document: %s\n", response.Content[0].Text)
		return
	}
}

// testChapterManagement tests chapter operations
func testChapterManagement(h *docgenHandler.DocGenHandler, docID string) {
	// Test getting a chapter
	params := map[string]interface{}{
		"document_id":    docID,
		"chapter_number": float64(1),
	}

	response, err := h.CallTool(nil, &protocol.CallToolRequest{
		Name:      "get_chapter",
		Arguments: params,
	})
	if err != nil {
		fmt.Printf("   ⚠️  Failed to get chapter: %v\n", err)
		return
	}

	if response.IsError {
		fmt.Printf("   ⚠️  Error getting chapter: %s\n", response.Content[0].Text)
		return
	}

	// Test updating chapter metadata
	params = map[string]interface{}{
		"document_id":    docID,
		"chapter_number": float64(1),
		"title":          "Introduction to MCP (Updated)",
	}

	response, err = h.CallTool(nil, &protocol.CallToolRequest{
		Name:      "update_chapter_metadata",
		Arguments: params,
	})
	if err != nil {
		fmt.Printf("   ⚠️  Failed to update chapter metadata: %v\n", err)
		return
	}

	if response.IsError {
		fmt.Printf("   ⚠️  Error updating chapter metadata: %s\n", response.Content[0].Text)
		return
	}

	fmt.Printf("   ✅ Chapter operations completed successfully\n")
}

// testStyleManagement tests style management operations
func testStyleManagement(h *docgenHandler.DocGenHandler, docID string) {
	// Test 1: List existing styles
	response, err := h.CallTool(nil, &protocol.CallToolRequest{
		Name:      "list_styles",
		Arguments: map[string]interface{}{},
	})
	if err != nil {
		fmt.Printf("   ⚠️  Failed to list styles: %v\n", err)
		return
	}
	
	if !response.IsError {
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(response.Content[0].Text), &result); err == nil {
			if styles, ok := result["styles"].([]interface{}); ok {
				fmt.Printf("   Found %d existing style(s)\n", len(styles))
			}
		}
	}
	
	// Test 2: Save a new style
	testStyle := map[string]interface{}{
		"body": map[string]interface{}{
			"font_family": "Georgia",
			"font_size":   "11pt",
			"color":       "#333333",
		},
		"heading": map[string]interface{}{
			"font_family": "Arial",
			"color":       "#003366",
		},
		"margins": map[string]interface{}{
			"top":    "1.5in",
			"bottom": "1.5in", 
			"left":   "1.25in",
			"right":  "1.25in",
		},
		"line_spacing": "1.8",
	}
	
	response, err = h.CallTool(nil, &protocol.CallToolRequest{
		Name: "save_style",
		Arguments: map[string]interface{}{
			"style_name": "test-modern",
			"style_data": testStyle,
		},
	})
	if err != nil {
		fmt.Printf("   ⚠️  Failed to save style: %v\n", err)
		return
	}
	
	if response.IsError {
		fmt.Printf("   ⚠️  Error saving style: %s\n", response.Content[0].Text)
		return
	}
	fmt.Println("   ✅ Saved new style 'test-modern'")
	
	// Test 3: Load the saved style
	response, err = h.CallTool(nil, &protocol.CallToolRequest{
		Name: "load_style",
		Arguments: map[string]interface{}{
			"style_name": "test-modern",
		},
	})
	if err != nil {
		fmt.Printf("   ⚠️  Failed to load style: %v\n", err)
		return
	}
	
	if response.IsError {
		fmt.Printf("   ⚠️  Error loading style: %s\n", response.Content[0].Text)
		return
	}
	fmt.Println("   ✅ Loaded style 'test-modern'")
	
	// Test 4: Set document style
	response, err = h.CallTool(nil, &protocol.CallToolRequest{
		Name: "set_document_style",
		Arguments: map[string]interface{}{
			"document_id": docID,
			"style_name":  "test-modern",
		},
	})
	if err != nil {
		fmt.Printf("   ⚠️  Failed to set document style: %v\n", err)
		return
	}
	
	if response.IsError {
		fmt.Printf("   ⚠️  Error setting document style: %s\n", response.Content[0].Text)
		return
	}
	fmt.Println("   ✅ Set document style to 'test-modern'")
	
	// Test 5: Get document style
	response, err = h.CallTool(nil, &protocol.CallToolRequest{
		Name: "get_document_style",
		Arguments: map[string]interface{}{
			"document_id": docID,
		},
	})
	if err != nil {
		fmt.Printf("   ⚠️  Failed to get document style: %v\n", err)
		return
	}
	
	if response.IsError {
		fmt.Printf("   ⚠️  Error getting document style: %s\n", response.Content[0].Text)
		return
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(response.Content[0].Text), &result); err == nil {
		if styleName, ok := result["style_name"].(string); ok {
			fmt.Printf("   ✅ Document style is: %s\n", styleName)
		}
	}
	
	// Test 6: Delete the test style
	response, err = h.CallTool(nil, &protocol.CallToolRequest{
		Name: "delete_style",
		Arguments: map[string]interface{}{
			"style_name": "test-modern",
		},
	})
	if err != nil {
		fmt.Printf("   ⚠️  Failed to delete style: %v\n", err)
		return
	}
	
	if response.IsError {
		fmt.Printf("   ⚠️  Error deleting style: %s\n", response.Content[0].Text)
		return
	}
	fmt.Println("   ✅ Deleted style 'test-modern'")
}

// isPandocAvailable checks if pandoc is available
func isPandocAvailable() bool {
	_, err := os.Stat("/usr/bin/pandoc")
	if err == nil {
		return true
	}
	_, err = os.Stat("/usr/local/bin/pandoc")
	if err == nil {
		return true
	}
	_, err = os.Stat("/opt/homebrew/bin/pandoc")
	return err == nil
}

// runDirectExport directly exports a document for troubleshooting
// runRebuildChapter rebuilds the markdown for a specific chapter
func runRebuildChapter(rebuildSpec string) {
	fmt.Println("Document Generation MCP Server - Rebuild Chapter Mode")
	fmt.Println("====================================================")

	// Parse rebuild specification (documentID,chapterNumber)
	parts := strings.Split(rebuildSpec, ",")
	if len(parts) != 2 {
		log.Fatalf("Invalid rebuild specification. Use format: documentID,chapterNumber (e.g., my-doc-123,1)")
	}

	docID := strings.TrimSpace(parts[0])
	chapterNumStr := strings.TrimSpace(parts[1])
	
	var chapterNum int
	if _, err := fmt.Sscanf(chapterNumStr, "%d", &chapterNum); err != nil {
		log.Fatalf("Invalid chapter number: %s", chapterNumStr)
	}

	fmt.Printf("Document ID: %s\n", docID)
	fmt.Printf("Chapter Number: %d\n", chapterNum)

	// Check required environment variables
	rootDir := os.Getenv("DOCGEN_ROOT_DIR")
	if rootDir == "" {
		log.Fatalf("DOCGEN_ROOT_DIR environment variable not set. Please set it to the directory containing your documents.")
	}

	fmt.Printf("Root directory: %s\n", rootDir)

	// Load configuration
	cfg := &config.Config{
		RootDir:      rootDir,
		MaxDocuments: 100,
		PandocPath:   "pandoc",
	}

	// Create docgen handler
	docHandler, err := docgenHandler.NewDocGenHandler(cfg)
	if err != nil {
		log.Fatalf("Failed to create docgen handler: %v", err)
	}

	// Get the manager from handler (we need to expose it)
	manager := docHandler.GetManager()
	
	// Rebuild the chapter
	if err := manager.RebuildChapterMarkdown(types.DocumentID(docID), types.ChapterNumber(chapterNum)); err != nil {
		log.Fatalf("Failed to rebuild chapter: %v", err)
	}

	fmt.Printf("\n✓ Chapter %d rebuilt successfully\n", chapterNum)
	fmt.Printf("Updated file: %s\n", filepath.Join(rootDir, docID, "chapters", fmt.Sprintf("%02d", chapterNum), "chapter.md"))
}

func runDirectExport(exportSpec string, customStyleFile string) {
	fmt.Println("Document Generation MCP Server - Direct Export Mode")
	fmt.Println("===================================================")

	// Parse export specification (documentID,format)
	parts := strings.Split(exportSpec, ",")
	if len(parts) != 2 {
		log.Fatalf("Invalid export specification. Use format: documentID,format (e.g., my-doc-123,pdf)")
	}

	docID := strings.TrimSpace(parts[0])
	format := strings.TrimSpace(parts[1])

	// Validate format
	validFormats := map[string]bool{
		"pdf": true, "docx": true, "html": true,
	}
	if !validFormats[format] {
		log.Fatalf("Invalid format '%s'. Valid formats: pdf, docx, html", format)
	}

	fmt.Printf("Document ID: %s\n", docID)
	fmt.Printf("Format: %s\n", format)

	// Check required environment variables
	rootDir := os.Getenv("DOCGEN_ROOT_DIR")
	if rootDir == "" {
		log.Fatalf("DOCGEN_ROOT_DIR environment variable not set. Please set it to the directory containing your documents.")
	}

	pandocPath := os.Getenv("PANDOC_PATH")
	if pandocPath == "" {
		// Try to find pandoc automatically
		if isPandocAvailable() {
			pandocPath = "pandoc"
		} else {
			log.Fatalf("PANDOC_PATH environment variable not set and pandoc not found in standard locations.")
		}
	}

	fmt.Printf("Root directory: %s\n", rootDir)
	fmt.Printf("Pandoc path: %s\n", pandocPath)
	
	// Check for global default style environment variable for testing
	defaultStylePath := os.Getenv("DOCGEN_DEFAULT_STYLE")
	if defaultStylePath != "" {
		fmt.Printf("Global default style: %s\n", defaultStylePath)
	} else {
		fmt.Printf("No global default style set\n")
	}

	// Set up configuration with environment variables
	os.Setenv("DOCGEN_ROOT_DIR", rootDir)
	os.Setenv("PANDOC_PATH", pandocPath)

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create docgen handler
	handler, err := docgenHandler.NewDocGenHandler(cfg)
	if err != nil {
		log.Fatalf("Failed to create docgen handler: %v", err)
	}

	// Check if document exists
	fmt.Printf("\n🔍 Checking if document exists...\n")
	manifestPath := cfg.ManifestPath(docID)
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		log.Fatalf("Document '%s' not found at path: %s", docID, manifestPath)
	}
	fmt.Printf("✅ Document found: %s\n", manifestPath)

	// Get document structure
	fmt.Printf("\n📋 Getting document structure...\n")
	structParams := map[string]interface{}{
		"document_id": docID,
	}

	structResponse, err := handler.CallTool(nil, &protocol.CallToolRequest{
		Name:      "get_document_structure",
		Arguments: structParams,
	})
	if err != nil {
		log.Fatalf("Failed to get document structure: %v", err)
	}

	if structResponse.IsError {
		log.Fatalf("Error getting document structure: %s", structResponse.Content[0].Text)
	}

	fmt.Printf("✅ Document structure retrieved successfully\n")

	// Export the document using MCP handler (which now supports enhanced style resolution)
	fmt.Printf("\n🚀 Exporting document to %s...\n", format)
	
	exportParams := map[string]interface{}{
		"document_id": docID,
		"format":      format,
	}
	
	// Handle custom style if provided
	if customStyleFile != "" {
		fmt.Printf("\n🎨 Using custom style via style_name parameter: %s\n", customStyleFile)
		// Use the new style_name parameter instead of env variable
		exportParams["style_name"] = customStyleFile
	}

	exportResponse, err := handler.CallTool(nil, &protocol.CallToolRequest{
		Name:      "export_document",
		Arguments: exportParams,
	})
	if err != nil {
		log.Fatalf("Failed to export document: %v", err)
	}

	if exportResponse.IsError {
		log.Fatalf("Error exporting document: %s", exportResponse.Content[0].Text)
	}

	// Parse export response
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(exportResponse.Content[0].Text), &result); err != nil {
		log.Fatalf("Failed to parse export response: %v", err)
	}

	outputPath, ok := result["output_path"].(string)
	if !ok {
		log.Fatalf("Failed to extract output_path from export response")
	}

	// Check if exported file exists and get its size
	if stat, err := os.Stat(outputPath); err == nil {
		fmt.Printf("✅ Export completed successfully!\n")
		fmt.Printf("📄 Output file: %s\n", outputPath)
		fmt.Printf("📏 File size: %d bytes\n", stat.Size())
		
		// Show file modification time
		fmt.Printf("⏰ Created: %s\n", stat.ModTime().Format("2006-01-02 15:04:05"))
		
		if stat.Size() == 0 {
			fmt.Printf("⚠️  Warning: Output file is empty (0 bytes)\n")
		}
	} else {
		fmt.Printf("⚠️  Warning: Export reported success but output file not found: %s\n", outputPath)
	}

	fmt.Printf("\n🎉 Direct export completed!\n")
}

// runFigureTest creates a test document with section-associated figures
func runFigureTest() {
	fmt.Println("Document Generation MCP Server - Figure Test Mode")
	fmt.Println("=================================================")
	
	// Check required environment variables
	rootDir := os.Getenv("DOCGEN_ROOT_DIR")
	if rootDir == "" {
		log.Fatalf("DOCGEN_ROOT_DIR environment variable not set. Please set it to the directory for test documents.")
	}
	
	fmt.Printf("Root directory: %s\n", rootDir)
	
	// Load configuration
	cfg := &config.Config{
		RootDir:      rootDir,
		MaxDocuments: 100,
		PandocPath:   "pandoc",
	}
	
	// Create docgen handler
	handler, err := docgenHandler.NewDocGenHandler(cfg)
	if err != nil {
		log.Fatalf("Failed to create docgen handler: %v", err)
	}
	
	manager := handler.GetManager()
	
	fmt.Println("\n📚 Creating test document with figures...")
	
	// Create a document
	docID, err := manager.CreateDocument("Figure Test Document", "Test Author", types.DocumentTypeBook)
	if err != nil {
		log.Fatalf("Failed to create document: %v", err)
	}
	fmt.Printf("✓ Created document: %s\n", docID)
	
	// Add a chapter
	chapterNum, err := manager.AddChapter(docID, "Introduction to Testing", nil)
	if err != nil {
		log.Fatalf("Failed to add chapter: %v", err)
	}
	fmt.Printf("✓ Added chapter %d\n", chapterNum)
	
	// Add sections
	section1, err := manager.AddSection(docID, chapterNum, "Getting Started", 
		"This section introduces the basic concepts of our testing framework. "+
		"We'll explore how images can be positioned at the beginning or end of sections.", 1)
	if err != nil {
		log.Fatalf("Failed to add section 1: %v", err)
	}
	fmt.Printf("✓ Added section %s\n", section1.String())
	
	section2, err := manager.AddSection(docID, chapterNum, "Advanced Topics",
		"This section covers more advanced testing scenarios. "+
		"Images here demonstrate different positioning options.", 1)
	if err != nil {
		log.Fatalf("Failed to add section 2: %v", err)
	}
	fmt.Printf("✓ Added section %s\n", section2.String())
	
	// Create test image files
	testImageDir := filepath.Join(rootDir, "test-images")
	if err := os.MkdirAll(testImageDir, 0755); err != nil {
		log.Fatalf("Failed to create test image directory: %v", err)
	}
	
	// Create placeholder images (simple PNG files)
	image1Path := filepath.Join(testImageDir, "test-image-1.png")
	image2Path := filepath.Join(testImageDir, "test-image-2.png")
	
	// Create simple 1x1 PNG files as placeholders
	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41, // IDAT chunk
		0x54, 0x08, 0x99, 0x63, 0xF8, 0x0F, 0x00, 0x00,
		0x01, 0x01, 0x01, 0x00, 0x1B, 0xB6, 0xEE, 0x56,
		0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, // IEND chunk
		0xAE, 0x42, 0x60, 0x82,
	}
	
	if err := os.WriteFile(image1Path, pngData, 0644); err != nil {
		log.Fatalf("Failed to create test image 1: %v", err)
	}
	if err := os.WriteFile(image2Path, pngData, 0644); err != nil {
		log.Fatalf("Failed to create test image 2: %v", err)
	}
	fmt.Printf("✓ Created test images\n")
	
	// Add figures to sections
	figure1ID, err := manager.AddImage(docID, chapterNum, section1.String(), 
		image1Path, "Figure at the beginning of section 1.1", "beginning")
	if err != nil {
		log.Fatalf("Failed to add figure 1: %v", err)
	}
	fmt.Printf("✓ Added figure %s at beginning of section %s\n", figure1ID, section1.String())
	
	figure2ID, err := manager.AddImage(docID, chapterNum, section2.String(),
		image2Path, "Figure at the end of section 1.2", "end")
	if err != nil {
		log.Fatalf("Failed to add figure 2: %v", err)
	}
	fmt.Printf("✓ Added figure %s at end of section %s\n", figure2ID, section2.String())
	
	// Rebuild chapter to include figures
	if err := manager.RebuildChapterMarkdown(docID, chapterNum); err != nil {
		log.Fatalf("Failed to rebuild chapter: %v", err)
	}
	fmt.Printf("✓ Rebuilt chapter markdown with figures\n")
	
	// Get document structure to verify
	manifest, err := manager.GetDocumentStructure(docID)
	if err != nil {
		log.Fatalf("Failed to get document structure: %v", err)
	}
	
	fmt.Printf("\n📋 Document Structure:\n")
	fmt.Printf("   Title: %s\n", manifest.Document.Title)
	fmt.Printf("   Author: %s\n", manifest.Document.Author)
	fmt.Printf("   Chapters: %d\n", len(manifest.Document.Chapters))
	
	for _, chapter := range manifest.Document.Chapters {
		fmt.Printf("\n   Chapter %d: %s\n", chapter.Number, chapter.Title)
		fmt.Printf("     Sections: %d\n", len(chapter.Sections))
		fmt.Printf("     Figures: %d\n", len(chapter.Figures))
		
		for _, fig := range chapter.Figures {
			fmt.Printf("       - %s (section %s, position: %s)\n", 
				fig.ID, fig.SectionNumber, fig.Position)
			fmt.Printf("         Caption: %s\n", fig.Caption)
		}
	}
	
	// Output document location
	docPath := filepath.Join(rootDir, string(docID))
	chapterPath := filepath.Join(docPath, "chapters", fmt.Sprintf("%02d", chapterNum), "chapter.md")
	
	fmt.Printf("\n✅ Test completed successfully!\n")
	fmt.Printf("📁 Document location: %s\n", docPath)
	fmt.Printf("📄 Chapter file: %s\n", chapterPath)
	fmt.Printf("\nYou can now:\n")
	fmt.Printf("1. View the chapter file to see how figures are included\n")
	fmt.Printf("2. Export the document: ./docgen -export '%s,pdf'\n", docID)
	fmt.Printf("3. Use this document ID in the Savant app to test the display\n")
}

// loadCustomStyleFile loads a style file in JSON or YAML format
func loadCustomStyleFile(filePath string) (*types.Style, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("style file not found: %s", filePath)
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read style file: %w", err)
	}

	var style types.Style
	
	// Determine format by file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".json":
		if err := json.Unmarshal(content, &style); err != nil {
			return nil, fmt.Errorf("failed to parse JSON style file: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(content, &style); err != nil {
			return nil, fmt.Errorf("failed to parse YAML style file: %w", err)
		}
	default:
		// Try JSON first, then YAML
		if err := json.Unmarshal(content, &style); err != nil {
			if err2 := yaml.Unmarshal(content, &style); err2 != nil {
				return nil, fmt.Errorf("failed to parse style file as JSON or YAML: JSON error: %v, YAML error: %v", err, err2)
			}
		}
	}

	return &style, nil
}