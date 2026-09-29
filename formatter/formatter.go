package formatter

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

// FormatterOptions provides configuration for formatting.
type FormatterOptions struct {
	IndentWidth int  // Number of spaces per indent level (default: 2)
	InsertFinal bool // Insert final newline (default: true)
}

// DefaultFormatterOptions returns the default formatting options.
func DefaultFormatterOptions() FormatterOptions {
	return FormatterOptions{
		IndentWidth: 2,
		InsertFinal: true,
	}
}

// Formatter handles formatting of vuego templates.
type Formatter struct {
	opts FormatterOptions
}

// NewFormatter creates a new formatter with default options.
func NewFormatter() *Formatter {
	return &Formatter{opts: DefaultFormatterOptions()}
}

// NewFormatterWithOptions creates a new formatter with custom options.
func NewFormatterWithOptions(opts FormatterOptions) *Formatter {
	return &Formatter{opts: opts}
}

// Format formats a vuego template string and returns the formatted result.
// It handles both full documents and partial fragments.
func (f *Formatter) Format(content string) (string, error) {
	// Split frontmatter and content
	frontmatter, body := f.splitFrontmatter(content)

	// Trim leading newlines from body
	body = strings.TrimLeft(body, "\n")

	// Check if this looks like a full document (starts with <!DOCTYPE or <html)
	trimmedBody := strings.TrimSpace(body)
	isFullDocument := strings.HasPrefix(trimmedBody, "<!DOCTYPE") || strings.HasPrefix(trimmedBody, "<html")

	if isFullDocument {
		return f.formatFullDocument(frontmatter, body)
	}

	// Handle partial/fragment formatting
	return f.formatFragment(frontmatter, body)
}

// formatFullDocument formats a complete HTML document.
func (f *Formatter) formatFullDocument(frontmatter, body string) (string, error) {
	trimmedBody := strings.TrimSpace(body)

	// Extract DOCTYPE if present
	var doctype string
	var htmlContent string

	if strings.HasPrefix(trimmedBody, "<!DOCTYPE") {
		// Find the end of DOCTYPE declaration
		endIdx := strings.Index(trimmedBody, ">")
		if endIdx != -1 {
			doctype = trimmedBody[:endIdx+1]
			htmlContent = strings.TrimSpace(trimmedBody[endIdx+1:])
		}
	} else {
		htmlContent = trimmedBody
	}

	// Parse HTML
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", fmt.Errorf("parsing HTML: %w", err)
	}

	// Format the parsed HTML
	var result strings.Builder
	f.formatNodeChildren(doc, &result, 0)

	output := strings.TrimRight(result.String(), "\n")

	// Reconstruct with frontmatter and DOCTYPE
	result.Reset()
	result.WriteString(frontmatter)
	if doctype != "" {
		result.WriteString(doctype)
		result.WriteString("\n")
	}
	result.WriteString(output)

	finalResult := result.String()

	if f.opts.InsertFinal {
		finalResult += "\n"
	}

	return finalResult, nil
}

// formatFragment formats a partial HTML fragment using html.ParseFragment
// to avoid injecting html/head/body wrappers.
func (f *Formatter) formatFragment(frontmatter, body string) (string, error) {
	ctx := fragmentContext(body)
	nodes, err := html.ParseFragment(strings.NewReader(body), ctx)
	if err != nil {
		return "", fmt.Errorf("parsing HTML fragment: %w", err)
	}

	var result strings.Builder
	for _, n := range nodes {
		f.formatNode(n, &result, 0)
	}

	output := strings.TrimRight(result.String(), "\n")

	var out strings.Builder
	out.WriteString(frontmatter)
	out.WriteString(output)

	finalResult := out.String()

	if f.opts.InsertFinal {
		finalResult += "\n"
	}

	return finalResult, nil
}

// formatNodeChildren formats only the children of a node (skips the node itself).
func (f *Formatter) formatNodeChildren(n *html.Node, buf *strings.Builder, depth int) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		f.formatNode(c, buf, depth)
	}
}

// splitFrontmatter separates YAML frontmatter from HTML content.
func (f *Formatter) splitFrontmatter(content string) (string, string) {
	lines := strings.Split(content, "\n")

	// Check if content starts with ---
	if len(lines) == 0 || !strings.HasPrefix(lines[0], "---") {
		return "", content
	}

	// Find the closing ---
	for i := 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "---") {
			frontmatter := strings.Join(lines[:i+1], "\n")
			body := strings.Join(lines[i+1:], "\n")
			return frontmatter + "\n", body
		}
	}

	// No closing ---, treat everything as content
	return "", content
}

// FormatString is a convenience function that formats a template string.
func FormatString(content string) (string, error) {
	formatter := NewFormatter()
	return formatter.Format(content)
}

// IndentString indents each line of text by the given level.
func IndentString(text string, level int, width int) string {
	indent := strings.Repeat(" ", level*width)
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			lines[i] = indent + line
		}
	}

	return strings.Join(lines, "\n")
}
