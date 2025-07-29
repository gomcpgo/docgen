package handler

import (
	"context"
	"encoding/json"

	"github.com/gomcpgo/mcp/pkg/protocol"
)

// ListTools provides a list of all available tools in the docgen handler
func (h *DocGenHandler) ListTools(ctx context.Context) (*protocol.ListToolsResponse, error) {
	tools := []protocol.Tool{
		{
			Name:        "create_document",
			Description: "Create a new document with title, author, and type. Returns the document ID for future operations.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"title": {
						"type": "string",
						"description": "Document title"
					},
					"author": {
						"type": "string",
						"description": "Document author"
					},
					"type": {
						"type": "string",
						"enum": ["book", "report", "article", "letter"],
						"description": "Document type"
					}
				},
				"required": ["title", "author", "type"]
			}`),
		},
		{
			Name:        "get_document_structure",
			Description: "Get the complete structure of a document including all chapters, sections, figures, and tables.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID"
					}
				},
				"required": ["document_id"]
			}`),
		},
		{
			Name:        "delete_document",
			Description: "Delete a document and all its contents permanently.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID to delete"
					}
				},
				"required": ["document_id"]
			}`),
		},
		{
			Name:        "configure_document",
			Description: "Update document styling and pandoc configuration options.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID"
					},
					"style_updates": {
						"type": "object",
						"properties": {
							"font_family": {"type": "string"},
							"font_size": {"type": "string"},
							"line_spacing": {"type": "string"},
							"margins": {
								"type": "object",
								"properties": {
									"top": {"type": "string"},
									"bottom": {"type": "string"},
									"left": {"type": "string"},
									"right": {"type": "string"}
								}
							}
						},
						"description": "Style configuration updates"
					},
					"pandoc_options": {
						"type": "object",
						"properties": {
							"pdf_engine": {"type": "string"},
							"toc": {"type": "boolean"},
							"toc_depth": {"type": "integer"},
							"citation_style": {"type": "string"}
						},
						"description": "Pandoc configuration updates"
					}
				},
				"required": ["document_id"]
			}`),
		},
		{
			Name:        "add_chapter",
			Description: "Add a new chapter to a document. Chapters are automatically numbered sequentially.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID"
					},
					"title": {
						"type": "string",
						"description": "Chapter title"
					},
					"position": {
						"type": "integer",
						"description": "Position to insert chapter (optional, defaults to end)",
						"minimum": 1
					}
				},
				"required": ["document_id", "title"]
			}`),
		},
		{
			Name:        "get_chapter",
			Description: "Get the content and metadata of a specific chapter.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID"
					},
					"chapter_number": {
						"type": "integer",
						"description": "Chapter number",
						"minimum": 1
					}
				},
				"required": ["document_id", "chapter_number"]
			}`),
		},
		{
			Name:        "update_chapter_metadata",
			Description: "Update chapter metadata such as title.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID"
					},
					"chapter_number": {
						"type": "integer",
						"description": "Chapter number",
						"minimum": 1
					},
					"title": {
						"type": "string",
						"description": "New chapter title"
					}
				},
				"required": ["document_id", "chapter_number", "title"]
			}`),
		},
		{
			Name:        "delete_chapter",
			Description: "Delete a chapter and automatically renumber subsequent chapters.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID"
					},
					"chapter_number": {
						"type": "integer",
						"description": "Chapter number to delete",
						"minimum": 1
					}
				},
				"required": ["document_id", "chapter_number"]
			}`),
		},
		{
			Name:        "move_chapter",
			Description: "Move a chapter from one position to another and renumber chapters accordingly.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID"
					},
					"from_number": {
						"type": "integer",
						"description": "Current chapter number",
						"minimum": 1
					},
					"to_number": {
						"type": "integer",
						"description": "Target chapter number",
						"minimum": 1
					}
				},
				"required": ["document_id", "from_number", "to_number"]
			}`),
		},
		{
			Name:        "add_section",
			Description: "Add a new section to a chapter with markdown content.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID"
					},
					"chapter_number": {
						"type": "integer",
						"description": "Chapter number",
						"minimum": 1
					},
					"title": {
						"type": "string",
						"description": "Section title"
					},
					"content": {
						"type": "string",
						"description": "Section content in markdown format"
					},
					"level": {
						"type": "integer",
						"description": "Section level (1=section, 2=subsection, etc.)",
						"minimum": 1,
						"maximum": 6,
						"default": 1
					}
				},
				"required": ["document_id", "chapter_number", "title", "content"]
			}`),
		},
		{
			Name:        "update_section",
			Description: "Update the content of an existing section.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID"
					},
					"chapter_number": {
						"type": "integer",
						"description": "Chapter number",
						"minimum": 1
					},
					"section_number": {
						"type": "string",
						"description": "Section number (e.g., '1.1', '1.2.1')"
					},
					"content": {
						"type": "string",
						"description": "New section content in markdown format"
					}
				},
				"required": ["document_id", "chapter_number", "section_number", "content"]
			}`),
		},
		{
			Name:        "delete_section",
			Description: "Delete a section from a chapter.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID"
					},
					"chapter_number": {
						"type": "integer",
						"description": "Chapter number",
						"minimum": 1
					},
					"section_number": {
						"type": "string",
						"description": "Section number to delete (e.g., '1.1', '1.2.1')"
					}
				},
				"required": ["document_id", "chapter_number", "section_number"]
			}`),
		},
		{
			Name:        "add_image",
			Description: "Add an image figure to a chapter with caption and positioning options.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID"
					},
					"chapter_number": {
						"type": "integer",
						"description": "Chapter number",
						"minimum": 1
					},
					"image_path": {
						"type": "string",
						"description": "Path to the image file"
					},
					"caption": {
						"type": "string",
						"description": "Image caption"
					},
					"position": {
						"type": "string",
						"enum": ["here", "top", "bottom", "page", "float"],
						"description": "Image position (default: here)",
						"default": "here"
					},
					"width": {
						"type": "string",
						"description": "Image width (e.g., '50%', '10cm')"
					},
					"alignment": {
						"type": "string",
						"enum": ["left", "center", "right"],
						"description": "Image alignment (default: center)",
						"default": "center"
					}
				},
				"required": ["document_id", "chapter_number", "image_path", "caption"]
			}`),
		},
		{
			Name:        "update_image_caption",
			Description: "Update the caption of an existing image figure.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID"
					},
					"figure_id": {
						"type": "string",
						"description": "Figure ID (e.g., 'fig-1.1')"
					},
					"new_caption": {
						"type": "string",
						"description": "New caption text"
					}
				},
				"required": ["document_id", "figure_id", "new_caption"]
			}`),
		},
		{
			Name:        "delete_image",
			Description: "Delete an image figure and automatically renumber subsequent figures in the chapter.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID"
					},
					"figure_id": {
						"type": "string",
						"description": "Figure ID to delete (e.g., 'fig-1.1')"
					}
				},
				"required": ["document_id", "figure_id"]
			}`),
		},
		{
			Name:        "export_document",
			Description: "Export a document to PDF, DOCX, or HTML format using Pandoc.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID"
					},
					"format": {
						"type": "string",
						"enum": ["pdf", "docx", "html"],
						"description": "Export format"
					},
					"chapters": {
						"type": "array",
						"items": {
							"type": "integer",
							"minimum": 1
						},
						"description": "Specific chapters to export (optional, defaults to all)"
					}
				},
				"required": ["document_id", "format"]
			}`),
		},
		{
			Name:        "preview_chapter",
			Description: "Generate a preview of a single chapter in the specified format.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID"
					},
					"chapter_number": {
						"type": "integer",
						"description": "Chapter number to preview",
						"minimum": 1
					},
					"format": {
						"type": "string",
						"enum": ["pdf", "html"],
						"description": "Preview format (default: html)",
						"default": "html"
					}
				},
				"required": ["document_id", "chapter_number"]
			}`),
		},
		{
			Name:        "validate_document",
			Description: "Validate a document structure and check for any issues before export.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID to validate"
					}
				},
				"required": ["document_id"]
			}`),
		},
	}

	return &protocol.ListToolsResponse{Tools: tools}, nil
}