package export

import (
	"fmt"
	"strings"
	"time"

	"github.com/gomcpgo/docgen/pkg/types"
)

// TemplateVariables holds all available template variables
type TemplateVariables struct {
	Page          string
	TotalPages    string
	ChapterTitle  string
	ChapterNumber string
	DocumentTitle string
	Author        string
	Date          string
	SectionTitle  string
}

// ProcessTemplate replaces template variables with actual values
func ProcessTemplate(template string, vars TemplateVariables) string {
	if template == "" {
		return ""
	}

	result := template
	replacements := map[string]string{
		"{page}":           vars.Page,
		"{total_pages}":    vars.TotalPages,
		"{chapter_title}":  vars.ChapterTitle,
		"{chapter_number}": vars.ChapterNumber,
		"{document_title}": vars.DocumentTitle,
		"{author}":         vars.Author,
		"{date}":          vars.Date,
		"{section_title}":  vars.SectionTitle,
	}

	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// CreateTemplateVariables creates template variables from document manifest
func CreateTemplateVariables(manifest *types.Manifest) TemplateVariables {
	return TemplateVariables{
		Page:          "{page}",        // Placeholder for page numbers - handled by output format
		TotalPages:    "{total_pages}", // Placeholder for total pages - handled by output format
		ChapterTitle:  "",              // Context-sensitive, filled when processing specific chapters
		ChapterNumber: "",              // Context-sensitive, filled when processing specific chapters
		DocumentTitle: manifest.Document.Title,
		Author:        manifest.Document.Author,
		Date:          time.Now().Format("2006-01-02"),
		SectionTitle:  "",              // Context-sensitive, filled when processing specific sections
	}
}

// ProcessTemplateForPDF converts templates to LaTeX-compatible format
func ProcessTemplateForPDF(template string, vars TemplateVariables) string {
	if template == "" {
		return ""
	}

	// Replace document-level variables
	result := template
	result = strings.ReplaceAll(result, "{document_title}", vars.DocumentTitle)
	result = strings.ReplaceAll(result, "{author}", vars.Author)
	result = strings.ReplaceAll(result, "{date}", vars.Date)

	// Replace page variables with LaTeX commands
	result = strings.ReplaceAll(result, "{page}", "\\thepage")
	result = strings.ReplaceAll(result, "{total_pages}", "\\pageref{LastPage}")

	// Replace context-sensitive variables with LaTeX marks
	result = strings.ReplaceAll(result, "{chapter_title}", "\\leftmark")
	result = strings.ReplaceAll(result, "{chapter_number}", "\\thechapter")
	result = strings.ReplaceAll(result, "{section_title}", "\\rightmark")

	return result
}

// ProcessTemplateForHTML converts templates to HTML format
func ProcessTemplateForHTML(template string, vars TemplateVariables) string {
	if template == "" {
		return ""
	}

	// For HTML, we only process document-level variables
	// Page numbers and context variables are not applicable for web pages
	result := template
	result = strings.ReplaceAll(result, "{document_title}", vars.DocumentTitle)
	result = strings.ReplaceAll(result, "{author}", vars.Author)
	result = strings.ReplaceAll(result, "{date}", vars.Date)

	// Remove page-specific variables for HTML
	result = strings.ReplaceAll(result, "{page}", "")
	result = strings.ReplaceAll(result, "{total_pages}", "")
	result = strings.ReplaceAll(result, "{chapter_title}", "")
	result = strings.ReplaceAll(result, "{chapter_number}", "")
	result = strings.ReplaceAll(result, "{section_title}", "")

	return strings.TrimSpace(result)
}

// ValidateTemplate checks if a template contains valid variables
func ValidateTemplate(template string) []string {
	if template == "" {
		return nil
	}

	validVariables := map[string]bool{
		"{page}":           true,
		"{total_pages}":    true,
		"{chapter_title}":  true,
		"{chapter_number}": true,
		"{document_title}": true,
		"{author}":         true,
		"{date}":          true,
		"{section_title}":  true,
	}

	var warnings []string
	
	// Find all {variable} patterns
	for i := 0; i < len(template); i++ {
		if template[i] == '{' {
			// Find closing brace
			end := strings.Index(template[i:], "}")
			if end == -1 {
				warnings = append(warnings, fmt.Sprintf("Unclosed variable at position %d", i))
				continue
			}
			
			variable := template[i : i+end+1]
			if !validVariables[variable] {
				warnings = append(warnings, fmt.Sprintf("Unknown template variable: %s", variable))
			}
		}
	}

	return warnings
}

// ExtractVariablesFromTemplate returns list of variables used in template
func ExtractVariablesFromTemplate(template string) []string {
	if template == "" {
		return nil
	}

	var variables []string
	used := make(map[string]bool)

	validVariables := []string{
		"{page}", "{total_pages}", "{chapter_title}", "{chapter_number}",
		"{document_title}", "{author}", "{date}", "{section_title}",
	}

	for _, variable := range validVariables {
		if strings.Contains(template, variable) && !used[variable] {
			variables = append(variables, variable)
			used[variable] = true
		}
	}

	return variables
}