# Document Generation MCP Server

A comprehensive Model Context Protocol (MCP) server for generating professional documents (PDF, DOCX, HTML) using Pandoc. This server enables iterative document building with support for books, reports, articles, and letters.

## Features

- **Document Types**: Support for books, reports, articles, and letters
- **Iterative Building**: Create and refine documents over multiple interactions
- **Automatic Numbering**: Sequential numbering for chapters, sections, figures, and tables
- **Export Formats**: PDF, DOCX, and HTML output via Pandoc
- **File-based Storage**: Transparent storage using markdown and YAML files
- **Comprehensive Toolset**: 18 tools for complete document management

## Prerequisites

- Go 1.21 or later
- Pandoc (for document export)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/gomcpgo/docgen
cd docgen
```

2. Build the server:
```bash
./run.sh build
```

3. Set up environment variables:
```bash
export DOCGEN_ROOT_DIR="/path/to/your/documents"
export PANDOC_PATH="pandoc"  # optional, defaults to "pandoc"
```

## Configuration

The server is configured via environment variables:

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DOCGEN_ROOT_DIR` | Yes | - | Root directory for document storage |
| `PANDOC_PATH` | No | `pandoc` | Path to pandoc executable |
| `DOCGEN_DEFAULT_STYLE` | No | - | Path to default style template |
| `DOCGEN_MAX_DOCUMENTS` | No | `100` | Maximum number of documents |
| `DOCGEN_MAX_FILE_SIZE` | No | `10MB` | Maximum file size for uploads |
| `DOCGEN_EXPORT_TIMEOUT` | No | `300s` | Export operation timeout |

## Usage

### Running the Server

```bash
# Run the MCP server
./bin/docgen

# Show version information
./bin/docgen -version

# Run integration tests (temporary files deleted after)
./bin/docgen -test

# Run integration tests and keep generated files for inspection
./bin/docgen -test -keep-files
```

### Document Structure

Each document is stored in the following structure:

```
DOCGEN_ROOT_DIR/
├── exports/                # Exported documents (PDF, DOCX, HTML)
│   ├── document1.pdf
│   └── document2.docx
├── DocumentID/
│   ├── manifest.yaml       # Document metadata and structure
│   ├── style.yaml         # Document-specific styling
│   ├── pandoc-config.yaml # Pandoc settings
│   ├── chapters/
│   │   ├── 01/
│   │   │   ├── chapter.md    # Chapter content
│   │   │   └── metadata.yaml # Chapter metadata
│   │   └── 02/
│   │       ├── chapter.md
│   │       └── metadata.yaml
│   └── assets/
│       └── images/
│           ├── fig-1.1.png   # Numbered figures
│           └── fig-2.3.png
└── AnotherDocumentID/
    └── ...
```

## Available Tools

### Document Management
- `create_document` - Create a new document
- `get_document_structure` - Get complete document structure
- `delete_document` - Remove a document
- `configure_document` - Update document styling and settings

### Chapter Operations
- `add_chapter` - Add a new chapter
- `get_chapter` - Retrieve chapter content
- `update_chapter_metadata` - Update chapter title/metadata
- `delete_chapter` - Remove a chapter (with automatic renumbering)
- `move_chapter` - Reorder chapters

### Content Operations
- `add_section` - Add sections to chapters
- `update_section` - Modify section content
- `delete_section` - Remove sections

### Asset Management
- `add_image` - Add figures with captions
- `update_image_caption` - Modify figure captions
- `delete_image` - Remove figures (with automatic renumbering)

### Export Operations
- `export_document` - Export to PDF/DOCX/HTML
- `preview_chapter` - Generate single chapter previews
- `validate_document` - Check document integrity

## Examples

### Creating a Book

```json
{
  "name": "create_document",
  "arguments": {
    "title": "My Technical Book",
    "author": "John Doe",
    "type": "book"
  }
}
```

### Adding a Chapter

```json
{
  "name": "add_chapter",
  "arguments": {
    "document_id": "my-technical-book-123456",
    "title": "Introduction to the Topic"
  }
}
```

### Exporting to PDF

```json
{
  "name": "export_document",
  "arguments": {
    "document_id": "my-technical-book-123456",
    "format": "pdf"
  }
}
```

## Integration Tests

The server includes comprehensive integration tests that create sample documents and test all major functionality:

```bash
# Run tests (files cleaned up automatically)
./bin/docgen -test

# Run tests and keep generated files for inspection
./bin/docgen -test -keep-files
```

This will:
1. Create sample book and report documents
2. Add chapters with realistic content
3. Test document configuration
4. Export to PDF (if Pandoc is available)
5. Validate document integrity
6. Test chapter management operations

When using `-keep-files`, the generated documents and PDF exports are preserved in a temporary directory for manual inspection.

## Development

### Building

```bash
# Build the server
./run.sh build

# Run tests
./run.sh test

# Run integration tests
./run.sh integration-test

# Clean build artifacts
./run.sh clean
```

### Project Structure

```
docgen/
├── cmd/
│   └── main.go              # MCP server entry point
├── pkg/
│   ├── types/               # Core data structures
│   ├── config/              # Configuration management
│   ├── storage/             # File system operations
│   ├── document/            # Document management logic
│   ├── export/              # Pandoc export functionality
│   └── handler/             # MCP tool handlers
├── test/
│   └── integration_test.go  # End-to-end tests
└── README.md
```

## Architecture

The server follows clean architecture principles:

- **Types**: Core data structures and domain models
- **Storage**: File system abstraction layer
- **Document Manager**: High-level document operations
- **Export**: Pandoc integration and document generation
- **Handler**: MCP protocol implementation

## Pandoc Integration

The server uses Pandoc for professional document generation with support for:

- Multiple output formats (PDF, DOCX, HTML)
- Table of contents generation
- Cross-references and citations
- Custom styling and templates
- Professional typography

## Security

- All operations are restricted to the configured root directory
- Input validation prevents directory traversal attacks
- File size limits prevent resource exhaustion
- No arbitrary code execution

## Limitations (MVP)

- Single user/agent access only
- No version control or change tracking
- Basic error recovery
- Images only for assets (no data files)
- Markdown tables only
- No collaborative editing

## Building

```bash
# Build the server
go build -o bin/docgen cmd/main.go

# Or use the run script
./run.sh build
```

## License

MIT License

## Contributing

Pull requests welcome. Please ensure:
- Tests pass (`./run.sh test`)
- Integration tests pass (`./run.sh integration-test`)
- New features include documentation
- Code follows project style