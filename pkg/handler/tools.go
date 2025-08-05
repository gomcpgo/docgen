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
			Name:        "list_documents",
			Description: "List all available documents with their metadata. Returns document IDs, titles, authors, types, creation dates, and basic statistics. Use this to see what documents exist before performing operations.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"sort_by": {
						"type": "string",
						"enum": ["created_at", "updated_at", "title"],
						"default": "updated_at",
						"description": "Field to sort by"
					},
					"sort_order": {
						"type": "string",
						"enum": ["asc", "desc"],
						"default": "desc",
						"description": "Sort order"
					},
					"limit": {
						"type": "integer",
						"default": 50,
						"minimum": 1,
						"maximum": 100,
						"description": "Maximum number of documents to return"
					}
				}
			}`),
		},
		{
			Name:        "create_document",
			Description: "Create a new document - this is the first step in document creation. Creates the document structure and returns a document_id (e.g., 'technical-manual-1234567890') that you'll need for all subsequent operations on this document. Choose document type carefully as it affects styling and structure.",
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
						"description": "Document type: 'book' for multi-chapter works, 'report' for structured reports, 'article' for papers, 'letter' for formal letters"
					}
				},
				"required": ["title", "author", "type"]
			}`),
		},
		{
			Name:        "get_document_structure",
			Description: "Get the complete overview of a document's current structure - shows all chapters, sections, figures, and tables with their numbers and titles. Use this to check the current state of the document before making changes or to understand the document organization. Essential for knowing what chapters exist before adding content.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID returned from create_document (e.g., 'technical-manual-1234567890')"
					}
				},
				"required": ["document_id"]
			}`),
		},
		{
			Name:        "delete_document",
			Description: "Permanently delete a document and all its contents including chapters, sections, figures, and exported files. This action cannot be undone. Use only when the user explicitly requests document deletion.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID of the document to permanently delete"
					}
				},
				"required": ["document_id"]
			}`),
		},
		{
			Name:        "configure_document",
			Description: "Update document styling (fonts, margins, spacing) and export settings (PDF engine, table of contents). Use this to customize the appearance and formatting of the final exported document. Changes apply to future exports, not existing ones.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID returned from create_document (e.g., 'technical-manual-1234567890')"
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
						"description": "Style updates: font_family, font_size, line_spacing, margins with top/bottom/left/right"
					},
					"pandoc_options": {
						"type": "object",
						"properties": {
							"pdf_engine": {"type": "string"},
							"toc": {"type": "boolean"},
							"toc_depth": {"type": "integer"},
							"citation_style": {"type": "string"}
						},
						"description": "Export settings: pdf_engine, toc (table of contents), toc_depth, citation_style"
					}
				},
				"required": ["document_id"]
			}`),
		},
		{
			Name:        "add_chapter",
			Description: "Add a new chapter to a document. Creates chapter structure but not content - use add_section to add actual content. Chapters are automatically numbered sequentially (1, 2, 3...). Returns the assigned chapter number. Use this before adding any content to a chapter.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID returned from create_document (e.g., 'technical-manual-1234567890')"
					},
					"title": {
						"type": "string",
						"description": "Chapter title"
					},
					"position": {
						"type": "integer",
						"description": "Position to insert chapter (optional). If specified, existing chapters are renumbered. If omitted, chapter is added at end.",
						"minimum": 1
					}
				},
				"required": ["document_id", "title"]
			}`),
		},
		{
			Name:        "get_chapter",
			Description: "Get detailed information about a specific chapter including its title, content, sections, figures, and tables. Use this to view or edit existing chapter content. Returns the full chapter structure with all sections and assets.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID returned from create_document (e.g., 'technical-manual-1234567890')"
					},
					"chapter_number": {
						"type": "integer",
						"description": "Chapter number (1-based, sequential). Use get_document_structure to see available chapter numbers.",
						"minimum": 1
					}
				},
				"required": ["document_id", "chapter_number"]
			}`),
		},
		{
			Name:        "update_chapter_metadata",
			Description: "Update chapter information like title. Use this to rename chapters or update chapter-level information. Does not affect chapter content - use update_section for content changes.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID returned from create_document (e.g., 'technical-manual-1234567890')"
					},
					"chapter_number": {
						"type": "integer",
						"description": "Chapter number (1-based, sequential). Use get_document_structure to see available chapter numbers.",
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
			Description: "Delete a chapter and all its content permanently. Automatically renumbers subsequent chapters (chapter 3 becomes 2, chapter 4 becomes 3, etc.). All sections, figures, and tables in the chapter are also deleted. Use only when user explicitly requests chapter deletion.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID returned from create_document (e.g., 'technical-manual-1234567890')"
					},
					"chapter_number": {
						"type": "integer",
						"description": "Chapter number to permanently delete (warning: this removes all chapter content)",
						"minimum": 1
					}
				},
				"required": ["document_id", "chapter_number"]
			}`),
		},
		{
			Name:        "move_chapter",
			Description: "Reorder chapters by moving a chapter from one position to another. Automatically renumbers all chapters and updates cross-references. For example, moving chapter 3 to position 1 makes it chapter 1, and chapters 1-2 become 2-3. Use get_document_structure first to see current chapter order.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID returned from create_document (e.g., 'technical-manual-1234567890')"
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
			Name:        "get_chapter_content",
			Description: "Get the full markdown content of a chapter. Returns the raw markdown text including all sections concatenated together. Use this to get the complete chapter content for editing or display.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID returned from create_document"
					},
					"chapter_number": {
						"type": "integer",
						"description": "Chapter number to get content for",
						"minimum": 1
					}
				},
				"required": ["document_id", "chapter_number"]
			}`),
		},
		{
			Name:        "update_chapter_content",
			Description: "Update the entire markdown content of a chapter. Replaces all existing sections with the provided markdown text. The tool will parse the markdown and create appropriate sections based on headings.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID returned from create_document"
					},
					"chapter_number": {
						"type": "integer",
						"description": "Chapter number to update",
						"minimum": 1
					},
					"content": {
						"type": "string",
						"description": "Complete markdown content for the chapter"
					}
				},
				"required": ["document_id", "chapter_number", "content"]
			}`),
		},
		{
			Name:        "add_section",
			Description: "Add actual content to a chapter by creating a section. This is where you put the real text, paragraphs, lists, and formatting. Sections are automatically numbered (1.1, 1.2, 2.1, etc.). The chapter must exist first - use add_chapter if needed. Supports full markdown formatting.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID returned from create_document (e.g., 'technical-manual-1234567890')"
					},
					"chapter_number": {
						"type": "integer",
						"description": "Chapter number (1-based, sequential). Use get_document_structure to see available chapter numbers.",
						"minimum": 1
					},
					"title": {
						"type": "string",
						"description": "Section title"
					},
					"content": {
						"type": "string",
						"description": "Section content in markdown format. Supports headings, paragraphs, lists, code blocks, emphasis, links, etc."
					},
					"level": {
						"type": "integer",
						"description": "Section hierarchy level: 1=main section (1.1), 2=subsection (1.1.1), 3=sub-subsection (1.1.1.1), etc.",
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
			Description: "Modify the content of an existing section within a chapter. Use this to edit, revise, or replace section text while preserving the document structure. Find the section number using get_document_structure or get_chapter first.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID returned from create_document (e.g., 'technical-manual-1234567890')"
					},
					"chapter_number": {
						"type": "integer",
						"description": "Chapter number (1-based, sequential). Use get_document_structure to see available chapter numbers.",
						"minimum": 1
					},
					"section_number": {
						"type": "string",
						"description": "Section number (e.g., '1.1', '1.2.1')"
					},
					"content": {
						"type": "string",
						"description": "New section content"
					}
				},
				"required": ["document_id", "chapter_number", "section_number", "content"]
			}`),
		},
		{
			Name:        "delete_section",
			Description: "Permanently remove a section from a chapter and automatically renumber subsequent sections. This removes the section content and adjusts section numbering (1.2 becomes 1.1, 1.3 becomes 1.2, etc.). Use only when user explicitly requests section deletion.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID returned from create_document (e.g., 'technical-manual-1234567890')"
					},
					"chapter_number": {
						"type": "integer",
						"description": "Chapter number (1-based, sequential). Use get_document_structure to see available chapter numbers.",
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
			Description: "Add an image/figure to a chapter with automatic numbering (fig-1.1, fig-1.2, etc.). Images are automatically numbered within each chapter and include captions. Supports positioning, sizing, and alignment options. The image file must exist at the specified path.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID returned from create_document (e.g., 'technical-manual-1234567890')"
					},
					"chapter_number": {
						"type": "integer",
						"description": "Chapter number (1-based, sequential). Use get_document_structure to see available chapter numbers.",
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
			Description: "Change the caption text of an existing figure while preserving the image and its position. Use the figure_id (like 'fig-1.1') to identify which image to update. Find figure IDs using get_document_structure or get_chapter.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID returned from create_document (e.g., 'technical-manual-1234567890')"
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
			Description: "Permanently remove an image/figure from a chapter and automatically renumber remaining figures (fig-1.2 becomes fig-1.1, fig-1.3 becomes fig-1.2, etc.). This removes both the image reference and its caption. Use only when user explicitly requests image deletion.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID returned from create_document (e.g., 'technical-manual-1234567890')"
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
			Description: "Export a document to PDF, DOCX, or HTML format when the user explicitly requests it and the document is ready. Do NOT export automatically - only when the user specifically asks for export. Returns the full file path where the exported document was saved (in the exports/ directory). Use validate_document first to check for issues.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID returned from create_document (e.g., 'technical-manual-1234567890')"
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
			Description: "Generate a quick preview/export of a single chapter for review before full document export. Useful for checking formatting, content, and layout of individual chapters. Supports PDF and HTML formats. Much faster than full document export.",
			InputSchema: json.RawMessage(`{
				"type": "object",
				"properties": {
					"document_id": {
						"type": "string",
						"description": "Document ID returned from create_document (e.g., 'technical-manual-1234567890')"
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
			Description: "Check document integrity and identify potential issues before export. Validates document structure, verifies all referenced files exist, checks for missing content, and ensures proper numbering. Run this before export_document to catch problems early.",
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