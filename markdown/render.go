package markdown

import (
	"bytes"
	"context"
	"io"
	"regexp"
	"strings"

	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"
)

// renderHeading renders a heading (h1-h6) with an auto-generated id.
func (m *Markdown) renderHeading(ctx context.Context, w io.Writer, n *ast.Heading, src []byte) error {
	content := m.inlineContent(n, src)

	if html, handled := m.tryHandlers(ctx, &Node{
		Type:  NodeHeading,
		Raw:   content,
		Level: n.Level,
	}); handled {
		_, err := io.WriteString(w, html)
		return err
	}

	id := headingID(content)
	return m.renderTemplate(w, "heading", map[string]any{
		"level":   n.Level,
		"id":      id,
		"content": content,
	})
}

// renderParagraph renders a paragraph block.
func (m *Markdown) renderParagraph(ctx context.Context, w io.Writer, n *ast.Paragraph, src []byte) error {
	content := m.inlineContent(n, src)

	if html, handled := m.tryHandlers(ctx, &Node{
		Type: NodeParagraph,
		Raw:  content,
	}); handled {
		_, err := io.WriteString(w, html)
		return err
	}

	return m.renderTemplate(w, "paragraph", map[string]any{
		"content": content,
	})
}

// renderFencedCodeBlock renders a fenced code block with optional language.
func (m *Markdown) renderFencedCodeBlock(ctx context.Context, w io.Writer, n *ast.FencedCodeBlock, src []byte) error {
	language := string(n.Language(src))
	code := codeBlockContent(n, src)

	if html, handled := m.tryHandlers(ctx, &Node{
		Type:     NodeCodeBlock,
		Raw:      code,
		Language: language,
	}); handled {
		_, err := io.WriteString(w, html)
		return err
	}

	return m.renderTemplate(w, "code_block", map[string]any{
		"language": language,
		"code":     code,
	})
}

// renderCodeBlock renders an indented code block.
func (m *Markdown) renderCodeBlock(ctx context.Context, w io.Writer, n *ast.CodeBlock, src []byte) error {
	code := codeBlockContent(n, src)

	if html, handled := m.tryHandlers(ctx, &Node{
		Type:     NodeCodeBlock,
		Raw:      code,
		Language: "",
	}); handled {
		_, err := io.WriteString(w, html)
		return err
	}

	return m.renderTemplate(w, "code_block", map[string]any{
		"language": "",
		"code":     code,
	})
}

// renderBlockquote renders a blockquote, recursing into child blocks.
func (m *Markdown) renderBlockquote(ctx context.Context, w io.Writer, n *ast.Blockquote, src []byte) error {
	var buf bytes.Buffer
	if err := m.renderChildren(ctx, &buf, n, src); err != nil {
		return err
	}
	return m.renderTemplate(w, "blockquote", map[string]any{
		"content": buf.String(),
	})
}

// renderList renders an ordered or unordered list.
func (m *Markdown) renderList(ctx context.Context, w io.Writer, n *ast.List, src []byte) error {
	var buf bytes.Buffer
	if err := m.renderChildren(ctx, &buf, n, src); err != nil {
		return err
	}
	return m.renderTemplate(w, "list", map[string]any{
		"ordered": n.IsOrdered(),
		"start":   n.Start,
		"content": buf.String(),
	})
}

// renderListItem renders a single list item.
func (m *Markdown) renderListItem(ctx context.Context, w io.Writer, n *ast.ListItem, src []byte) error {
	var buf bytes.Buffer
	if err := m.renderChildren(ctx, &buf, n, src); err != nil {
		return err
	}
	return m.renderTemplate(w, "list_item", map[string]any{
		"content": buf.String(),
	})
}

// renderHTMLBlock writes raw HTML blocks directly.
func (m *Markdown) renderHTMLBlock(w io.Writer, n *ast.HTMLBlock, src []byte) error {
	for i := 0; i < n.Lines().Len(); i++ {
		line := n.Lines().At(i)
		if _, err := w.Write(line.Value(src)); err != nil {
			return err
		}
	}
	return nil
}

// renderTable renders a GFM table by collecting header and body cells into
// structured data and passing it to the table template.
func (m *Markdown) renderTable(w io.Writer, n *east.Table, src []byte) error {
	var headers []map[string]any
	var rows [][]map[string]any

	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		switch row := child.(type) {
		case *east.TableHeader:
			headers = m.collectCells(row, src)
		case *east.TableRow:
			rows = append(rows, m.collectCells(row, src))
		}
	}

	return m.renderTemplate(w, "table", map[string]any{
		"headers": headers,
		"rows":    rows,
	})
}

// collectCells extracts cell data from a table row or header node.
func (m *Markdown) collectCells(node ast.Node, src []byte) []map[string]any {
	var cells []map[string]any
	for cell := node.FirstChild(); cell != nil; cell = cell.NextSibling() {
		if tc, ok := cell.(*east.TableCell); ok {
			cells = append(cells, map[string]any{
				"align":   alignString(tc.Alignment),
				"content": m.inlineContent(tc, src),
			})
		}
	}
	return cells
}

// renderFootnoteList renders the container for all footnotes.
func (m *Markdown) renderFootnoteList(ctx context.Context, w io.Writer, n *east.FootnoteList, src []byte) error {
	var buf bytes.Buffer
	if err := m.renderChildren(ctx, &buf, n, src); err != nil {
		return err
	}
	return m.renderTemplate(w, "footnote_list", map[string]any{
		"count":   n.Count,
		"content": buf.String(),
	})
}

// renderFootnote renders an individual footnote.
func (m *Markdown) renderFootnote(ctx context.Context, w io.Writer, n *east.Footnote, src []byte) error {
	var buf bytes.Buffer
	if err := m.renderChildren(ctx, &buf, n, src); err != nil {
		return err
	}
	return m.renderTemplate(w, "footnote", map[string]any{
		"index":   n.Index,
		"ref":     string(n.Ref),
		"content": buf.String(),
	})
}

// codeBlockContent extracts the raw text lines from a code block node.
func codeBlockContent(node ast.Node, src []byte) string {
	var buf strings.Builder
	lines := node.Lines()
	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		buf.Write(line.Value(src))
	}
	return buf.String()
}

var (
	stripTagsRe   = regexp.MustCompile(`<[^>]*>`)
	nonAlphanumRe = regexp.MustCompile(`[^a-z0-9 -]`)
	multiHyphenRe = regexp.MustCompile(`-+`)
)

// headingID generates a URL-friendly anchor id from heading content.
func headingID(content string) string {
	s := stripTagsRe.ReplaceAllString(content, "")
	s = strings.ToLower(s)
	s = nonAlphanumRe.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = multiHyphenRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

// alignString returns the alignment string for a table cell.
func alignString(align east.Alignment) string {
	switch align {
	case east.AlignLeft:
		return "left"
	case east.AlignCenter:
		return "center"
	case east.AlignRight:
		return "right"
	default:
		return ""
	}
}

// renderInlineChildren renders inline children of a node, dispatching each through templates.
func (m *Markdown) renderInlineChildren(w io.Writer, node ast.Node, src []byte) error {
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if err := m.renderInlineNode(w, child, src); err != nil {
			return err
		}
	}
	return nil
}

// renderInlineNode renders a single inline AST node through its template.
func (m *Markdown) renderInlineNode(w io.Writer, node ast.Node, src []byte) error {
	switch n := node.(type) {
	case *ast.Text:
		segment := string(n.Segment.Value(src))
		if _, err := io.WriteString(w, segment); err != nil {
			return err
		}
		if n.HardLineBreak() {
			return m.renderTemplate(w, "hard_break", nil)
		} else if n.SoftLineBreak() {
			_, err := io.WriteString(w, "\n")
			return err
		}
		return nil
	case *ast.String:
		_, err := w.Write(n.Value)
		return err
	case *ast.CodeSpan:
		content := codeSpanContent(n, src)
		return m.renderTemplate(w, "code_span", map[string]any{
			"content": content,
		})
	case *ast.Emphasis:
		content := m.inlineContent(n, src)
		return m.renderTemplate(w, "emphasis", map[string]any{
			"level":   n.Level,
			"content": content,
		})
	case *ast.Link:
		content := m.inlineContent(n, src)
		return m.renderTemplate(w, "link", map[string]any{
			"href":    string(n.Destination),
			"title":   string(n.Title),
			"content": content,
		})
	case *ast.Image:
		alt := inlineText(n, src)
		return m.renderTemplate(w, "image", map[string]any{
			"src":   string(n.Destination),
			"alt":   alt,
			"title": string(n.Title),
		})
	case *ast.AutoLink:
		url := string(n.URL(src))
		label := string(n.Label(src))
		href := url
		if n.AutoLinkType == ast.AutoLinkEmail {
			href = "mailto:" + url
		}
		return m.renderTemplate(w, "autolink", map[string]any{
			"href":  href,
			"label": label,
		})
	case *ast.RawHTML:
		var buf bytes.Buffer
		for i := 0; i < n.Segments.Len(); i++ {
			seg := n.Segments.At(i)
			buf.Write(seg.Value(src))
		}
		return m.renderTemplate(w, "raw_html", map[string]any{
			"content": buf.String(),
		})
	case *east.Strikethrough:
		content := m.inlineContent(n, src)
		return m.renderTemplate(w, "strikethrough", map[string]any{
			"content": content,
		})
	case *east.TaskCheckBox:
		return m.renderTemplate(w, "task_checkbox", map[string]any{
			"checked": n.IsChecked,
		})
	case *east.FootnoteLink:
		return m.renderTemplate(w, "footnote_link", map[string]any{
			"index":     n.Index,
			"ref_count": n.RefCount,
			"ref_index": n.RefIndex,
		})
	case *east.FootnoteBacklink:
		return m.renderTemplate(w, "footnote_backlink", map[string]any{
			"index": n.Index,
		})
	default:
		return m.renderInlineChildren(w, node, src)
	}
}

// inlineContent renders all inline children of a node and returns the result as a string.
func (m *Markdown) inlineContent(node ast.Node, src []byte) string {
	var buf bytes.Buffer
	_ = m.renderInlineChildren(&buf, node, src)
	return buf.String()
}

// inlineText extracts plain text from an inline node tree (used for alt text, heading IDs).
func inlineText(node ast.Node, src []byte) string {
	var buf strings.Builder
	for c := node.FirstChild(); c != nil; c = c.NextSibling() {
		if t, ok := c.(*ast.Text); ok {
			buf.Write(t.Segment.Value(src))
		} else if c.HasChildren() {
			buf.WriteString(inlineText(c, src))
		}
	}
	return buf.String()
}

// codeSpanContent extracts text from a code span's children.
func codeSpanContent(n *ast.CodeSpan, src []byte) string {
	var buf strings.Builder
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if t, ok := c.(*ast.Text); ok {
			buf.Write(t.Segment.Value(src))
		}
	}
	return buf.String()
}
