package markdown

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"regexp"
	"strings"

	vuego "github.com/titpetric/vuego"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"

	yaml "gopkg.in/yaml.v3"
)

// PostProcessor transforms rendered HTML for a specific block type.
// It receives the rendered HTML source and returns the transformed result.
type PostProcessor func(src string) string

// Markdown parses GFM markdown and renders block elements through vuego templates.
type Markdown struct {
	contentFS      fs.FS
	tpl            vuego.Template
	parser         parser.Parser
	postProcessors map[string]PostProcessor
}

// Document is a parsed markdown document ready for rendering.
type Document struct {
	md          *Markdown
	doc         ast.Node
	src         []byte
	frontMatter map[string]any
}

// FrontMatter returns the parsed YAML front matter from the document.
// Returns nil if no front matter was present.
func (d *Document) FrontMatter() map[string]any {
	return d.frontMatter
}

// New creates a new Markdown renderer backed by the given filesystem.
// The contentFS is used for loading markdown files via Load.
// Custom templates under "markdown/" in contentFS overlay the embedded defaults.
func New(contentFS fs.FS) *Markdown {
	var tplFS fs.FS
	if contentFS != nil {
		tplFS = vuego.NewOverlayFS(contentFS, Templates())
	} else {
		tplFS = Templates()
	}

	md := goldmark.New(goldmark.WithExtensions(extension.GFM))

	return &Markdown{
		contentFS:      contentFS,
		tpl:            vuego.NewFS(tplFS),
		parser:         md.Parser(),
		postProcessors: make(map[string]PostProcessor),
	}
}

// PostProcess registers a post-processor for a block type (e.g. "paragraph", "heading").
// The processor receives the rendered HTML and returns the transformed result.
func (m *Markdown) PostProcess(blockType string, fn PostProcessor) *Markdown {
	m.postProcessors[blockType] = fn
	return m
}

// Load reads a markdown file from the content filesystem and parses it into a Document.
// YAML front matter delimited by --- is extracted and available via FrontMatter().
func (m *Markdown) Load(filename string) (*Document, error) {
	raw, err := fs.ReadFile(m.contentFS, filename)
	if err != nil {
		return nil, fmt.Errorf("loading %s: %w", filename, err)
	}

	frontMatter, src := splitFrontMatter(raw)

	reader := text.NewReader(src)
	doc := m.parser.Parse(reader)

	return &Document{
		md:          m,
		doc:         doc,
		src:         src,
		frontMatter: frontMatter,
	}, nil
}

// splitFrontMatter extracts YAML front matter from raw markdown content.
// Front matter must be delimited by --- on its own line at the start of the file.
// Returns the parsed front matter map and the remaining markdown body.
func splitFrontMatter(raw []byte) (map[string]any, []byte) {
	content := string(raw)
	if !strings.HasPrefix(content, "---") {
		return nil, raw
	}

	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return nil, raw
	}

	var fm map[string]any
	if err := yaml.Unmarshal([]byte(parts[1]), &fm); err != nil {
		return nil, raw
	}

	body := strings.TrimSpace(parts[2])
	return fm, []byte(body)
}

// Render writes the rendered HTML of the parsed document to w.
func (d *Document) Render(w io.Writer) error {
	return d.md.renderChildren(w, d.doc, d.src)
}

// RenderBytes parses markdown source bytes and writes rendered HTML to w.
// Unlike Load, this does not read from the filesystem or parse front matter.
func (m *Markdown) RenderBytes(w io.Writer, src []byte) error {
	reader := text.NewReader(src)
	doc := m.parser.Parse(reader)
	return m.renderChildren(w, doc, src)
}

// renderChildren renders all direct children of a node.
func (m *Markdown) renderChildren(w io.Writer, node ast.Node, src []byte) error {
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if err := m.renderNode(w, child, src); err != nil {
			return err
		}
	}
	return nil
}

// renderNode dispatches a single AST node to the appropriate template.
func (m *Markdown) renderNode(w io.Writer, node ast.Node, src []byte) error {
	switch n := node.(type) {
	case *ast.Heading:
		return m.renderHeading(w, n, src)
	case *ast.Paragraph:
		return m.renderParagraph(w, n, src)
	case *ast.FencedCodeBlock:
		return m.renderFencedCodeBlock(w, n, src)
	case *ast.CodeBlock:
		return m.renderCodeBlock(w, n, src)
	case *ast.Blockquote:
		return m.renderBlockquote(w, n, src)
	case *ast.List:
		return m.renderList(w, n, src)
	case *ast.ListItem:
		return m.renderListItem(w, n, src)
	case *ast.ThematicBreak:
		return m.renderTemplate(w, "thematic_break", nil)
	case *ast.HTMLBlock:
		return m.renderHTMLBlock(w, n, src)
	case *ast.TextBlock:
		return m.renderInlineChildren(w, n, src)
	case *east.Table:
		return m.renderTable(w, n, src)
	default:
		if node.HasChildren() {
			return m.renderChildren(w, node, src)
		}
	}
	return nil
}

// renderHeading renders a heading (h1-h6) with an auto-generated id.
func (m *Markdown) renderHeading(w io.Writer, n *ast.Heading, src []byte) error {
	content := m.inlineContent(n, src)
	id := headingID(content)
	return m.renderTemplate(w, "heading", map[string]any{
		"level":   n.Level,
		"id":      id,
		"content": content,
	})
}

// renderParagraph renders a paragraph block.
func (m *Markdown) renderParagraph(w io.Writer, n *ast.Paragraph, src []byte) error {
	content := m.inlineContent(n, src)
	return m.renderTemplate(w, "paragraph", map[string]any{
		"content": content,
	})
}

// renderFencedCodeBlock renders a fenced code block with optional language.
func (m *Markdown) renderFencedCodeBlock(w io.Writer, n *ast.FencedCodeBlock, src []byte) error {
	return m.renderTemplate(w, "code_block", map[string]any{
		"language": string(n.Language(src)),
		"code":     codeBlockContent(n, src),
	})
}

// renderCodeBlock renders an indented code block.
func (m *Markdown) renderCodeBlock(w io.Writer, n *ast.CodeBlock, src []byte) error {
	return m.renderTemplate(w, "code_block", map[string]any{
		"language": "",
		"code":     codeBlockContent(n, src),
	})
}

// renderBlockquote renders a blockquote, recursing into child blocks.
func (m *Markdown) renderBlockquote(w io.Writer, n *ast.Blockquote, src []byte) error {
	var buf bytes.Buffer
	if err := m.renderChildren(&buf, n, src); err != nil {
		return err
	}
	return m.renderTemplate(w, "blockquote", map[string]any{
		"content": buf.String(),
	})
}

// renderList renders an ordered or unordered list.
func (m *Markdown) renderList(w io.Writer, n *ast.List, src []byte) error {
	var buf bytes.Buffer
	if err := m.renderChildren(&buf, n, src); err != nil {
		return err
	}
	return m.renderTemplate(w, "list", map[string]any{
		"ordered": n.IsOrdered(),
		"start":   n.Start,
		"content": buf.String(),
	})
}

// renderListItem renders a single list item.
func (m *Markdown) renderListItem(w io.Writer, n *ast.ListItem, src []byte) error {
	var buf bytes.Buffer
	if err := m.renderChildren(&buf, n, src); err != nil {
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

// renderTemplate renders a vuego template with data, applying any registered post-processor.
func (m *Markdown) renderTemplate(w io.Writer, name string, data map[string]any) error {
	var buf bytes.Buffer
	tpl := m.tpl.Load("markdown/" + name + ".vuego").Fill(data)
	if err := tpl.Render(context.Background(), &buf); err != nil {
		return fmt.Errorf("rendering %s template: %w", name, err)
	}

	result := buf.String()
	if pp, ok := m.postProcessors[name]; ok {
		result = pp(result)
	}

	_, err := io.WriteString(w, result)
	return err
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
