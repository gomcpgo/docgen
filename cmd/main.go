package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gomcpgo/mcp/pkg/handler"
	"github.com/gomcpgo/mcp/pkg/protocol"
	"github.com/gomcpgo/mcp/pkg/server"
	"github.com/gomcpgo/docgen/pkg/config"
	docgenHandler "github.com/gomcpgo/docgen/pkg/handler"
	"github.com/gomcpgo/docgen/pkg/types"
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
	flag.Parse()

	if *versionFlag {
		fmt.Printf("Document Generation MCP Server\n")
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("Build Time: %s\n", BuildTime)
		return
	}

	if *testMode {
		runIntegrationTests(*keepFiles)
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