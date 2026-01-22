package formatter

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/titpetric/vuego/internal/helpers"
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

	// Format the parsed HTML - skip auto-inserted html/head/body elements
	var result strings.Builder
	f.formatNodeChildren(doc, 0, &result)

	output := result.String()

	// Remove trailing newline if present (formatNode adds one)
	output = strings.TrimRight(output, "\n")

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

// formatFragment formats a partial HTML fragment (e.g., single element or list of elements).
func (f *Formatter) formatFragment(frontmatter, body string) (string, error) {
	// Wrap fragment in a root div to parse it
	wrapped := "<div>" + body + "</div>"
	doc, err := html.Parse(strings.NewReader(wrapped))
	if err != nil {
		return "", fmt.Errorf("parsing HTML fragment: %w", err)
	}

	// Find the wrapping div and format its children
	var result strings.Builder
	f.formatFragmentChildren(doc, 0, &result)

	output := result.String()

	// Remove trailing newline if present
	output = strings.TrimRight(output, "\n")

	// Reconstruct with frontmatter
	result.Reset()
	result.WriteString(frontmatter)
	result.WriteString(output)

	finalResult := result.String()

	if f.opts.InsertFinal {
		finalResult += "\n"
	}

	return finalResult, nil
}

// formatFragmentChildren formats children of a fragment's wrapper element,
// skipping auto-inserted wrappers (html, head, body, and the wrapper div itself).
func (f *Formatter) formatFragmentChildren(n *html.Node, depth int, buf *strings.Builder) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		f.formatNode(c, depth, buf)
	}
}

// formatNodeChildren formats only the children of a node (skips the node itself).
func (f *Formatter) formatNodeChildren(n *html.Node, depth int, buf *strings.Builder) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		f.formatNode(c, depth, buf)
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

// formatNode recursively formats an HTML node tree.
func (f *Formatter) formatNode(n *html.Node, depth int, buf *strings.Builder) {
	indent := strings.Repeat(" ", depth*f.opts.IndentWidth)

	switch n.Type {
	case html.DocumentNode:
		// Process children of document node
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f.formatNode(c, depth, buf)
		}

	case html.ElementNode:
		// Check if this is a style or script block - preserve as-is
		if n.Data == "style" || n.Data == "script" {
			buf.WriteString(indent)
			buf.WriteString(f.renderOpenTag(n))

			// Collect content
			var content strings.Builder
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.TextNode {
					content.WriteString(c.Data)
				}
			}

			contentStr := content.String()
			trimmedContent := strings.TrimSpace(contentStr)

			// If content is empty or only whitespace, output inline
			if trimmedContent == "" {
				buf.WriteString(f.renderCloseTag(n))
				buf.WriteString("\n")
			} else {
				// Content exists: newline + trimmed content + newline + closing tag
				buf.WriteString("\n")
				buf.WriteString(trimmedContent)
				buf.WriteString("\n")
				buf.WriteString(indent)
				buf.WriteString(f.renderCloseTag(n))
				buf.WriteString("\n")
			}
			return
		}

		// Write opening tag
		buf.WriteString(indent)
		buf.WriteString(f.renderOpenTag(n))
		// Check if this is a void element (self-closing)
		if isVoidElement(n.DataAtom) {
			buf.WriteString("\n")
			return
		}

		// Check if element should keep content inline
		if f.shouldKeepInline(n) {
			// Process children inline
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				switch c.Type {
				case html.TextNode:
					text := strings.TrimSpace(c.Data)
					if text != "" {
						buf.WriteString(text)
					}
				case html.ElementNode:
					// Inline child elements
					buf.WriteString(f.renderOpenTag(c))
					if !isVoidElement(c.DataAtom) {
						for cc := c.FirstChild; cc != nil; cc = cc.NextSibling {
							if cc.Type == html.TextNode {
								text := strings.TrimSpace(cc.Data)
								if text != "" {
									buf.WriteString(text)
								}
							}
						}
						buf.WriteString(f.renderCloseTag(c))
					}
				}
			}
			buf.WriteString(f.renderCloseTag(n))
			buf.WriteString("\n")
			return
		}

		// Process children normally
		isFirst := true
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			// Skip ignorable whitespace nodes
			if c.Type == html.TextNode && f.isIgnorableWhitespace(c) {
				continue
			}

			// Add newline before each child except the first
			if !isFirst {
				buf.WriteString("\n")
			}
			isFirst = false

			f.formatNode(c, depth+1, buf)
		}

		// Write closing tag
		buf.WriteString(indent)
		buf.WriteString(f.renderCloseTag(n))
		buf.WriteString("\n")

	case html.TextNode:
		text := strings.TrimSpace(n.Data)
		if text != "" && !f.isIgnorableWhitespace(n) {
			buf.WriteString(indent)
			buf.WriteString(text)
			buf.WriteString("\n")
		}

	case html.CommentNode:
		buf.WriteString(indent)
		buf.WriteString("<!--")
		buf.WriteString(n.Data)
		buf.WriteString("-->\n")
	}
}

// shouldKeepInline determines if element content should be kept inline.
func (f *Formatter) shouldKeepInline(n *html.Node) bool {
	// First, check if element has only text children (no nested elements)
	hasOnlyText := true

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			hasOnlyText = false
			break
		}
	}

	// If it only has text content, keep it inline
	if hasOnlyText {
		return true
	}

	// Also check against list of inline elements
	inlineElements := []atom.Atom{
		atom.A, atom.Abbr, atom.B, atom.Bdi, atom.Bdo, atom.Br, atom.Button,
		atom.Cite, atom.Code, atom.Data, atom.Dfn, atom.Em, atom.I, atom.Kbd,
		atom.Label, atom.Mark, atom.Q, atom.S, atom.Samp, atom.Small, atom.Span,
		atom.Strong, atom.Sub, atom.Sup, atom.Time, atom.U, atom.Var, atom.Wbr,
	}

	for _, elem := range inlineElements {
		if n.DataAtom == elem {
			// Check if it only contains text children (no nested elements)
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode {
					return false // Has nested elements, use block formatting
				}
			}
			return true // Only text, keep inline
		}
	}
	return false
}

// isIgnorableWhitespace checks if a text node is only whitespace between block elements.
func (f *Formatter) isIgnorableWhitespace(n *html.Node) bool {
	if n.Type != html.TextNode {
		return false
	}
	// Check if text is only whitespace
	return strings.TrimSpace(n.Data) == ""
}

// renderOpenTag renders an opening tag with attributes.
func (f *Formatter) renderOpenTag(n *html.Node) string {
	var buf strings.Builder
	buf.WriteString("<")
	buf.WriteString(n.Data)

	for _, attr := range n.Attr {
		buf.WriteString(" ")
		buf.WriteString(attr.Key)
		if attr.Val != "" {
			buf.WriteString("=\"")
			buf.WriteString(helpers.FormatAttr(attr.Val))
			buf.WriteString("\"")
		}
	}

	buf.WriteString(">")
	return buf.String()
}

// renderCloseTag renders a closing tag.
func (f *Formatter) renderCloseTag(n *html.Node) string {
	return "</" + n.Data + ">"
}

// isVoidElement checks if an element is void (self-closing).
func isVoidElement(a atom.Atom) bool {
	voidElements := []atom.Atom{
		atom.Area, atom.Base, atom.Br, atom.Col, atom.Embed, atom.Hr,
		atom.Img, atom.Input, atom.Link, atom.Meta, atom.Param, atom.Source,
		atom.Track, atom.Wbr,
	}

	for _, elem := range voidElements {
		if a == elem {
			return true
		}
	}
	return false
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
