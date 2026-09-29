package markdown

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	yaml "gopkg.in/yaml.v3"

	"github.com/titpetric/vuego"
)

// PostProcessor transforms rendered HTML for a specific block type.
// It receives the rendered HTML source and returns the transformed result.
type PostProcessor func(src string) string

// NodeType identifies the type of markdown node.
type NodeType string

// Limited functionality of node identifiers desired
// to be rendered with a handler.
const (
	NodeParagraph NodeType = "paragraph"
	NodeHeading   NodeType = "heading"
	NodeCodeBlock NodeType = "code_block"
)

// Node provides information about a markdown node for custom handling.
type Node struct {
	Type     NodeType
	Raw      string // raw source content
	Language string // for code blocks
	Level    int    // for headings (1-6)
}

// Handler processes a node, returning HTML if handled.
// Multiple handlers are called in registration order until one returns handled=true.
type Handler func(ctx context.Context, n *Node) (html string, handled bool)

// Document is a parsed markdown document ready for rendering.
type Document struct {
	md          *Markdown
	doc         ast.Node
	src         []byte
	frontMatter map[string]any
}

// Markdown parses GFM markdown and renders block elements through vuego templates.
type Markdown struct {
	contentFS      fs.FS
	tpl            vuego.Template
	parser         parser.Parser
	postProcessors map[string]PostProcessor
	handlers       map[NodeType][]Handler
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

	md := goldmark.New(goldmark.WithExtensions(
		extension.GFM,
		extension.Footnote,
	))

	return &Markdown{
		contentFS:      contentFS,
		tpl:            vuego.NewFS(tplFS),
		parser:         md.Parser(),
		postProcessors: make(map[string]PostProcessor),
		handlers:       make(map[NodeType][]Handler),
	}
}

// PostProcess registers a post-processor for a block type (e.g. "paragraph", "heading").
// The processor receives the rendered HTML and returns the transformed result.
func (m *Markdown) PostProcess(blockType string, fn PostProcessor) *Markdown {
	m.postProcessors[blockType] = fn
	return m
}

// Register adds a handler for a node type.
// Multiple handlers are called in registration order until one returns handled=true.
func (m *Markdown) Register(h Handler, nodeType NodeType) *Markdown {
	m.handlers[nodeType] = append(m.handlers[nodeType], h)
	return m
}

// tryHandlers calls registered handlers for a node type until one handles it.
func (m *Markdown) tryHandlers(ctx context.Context, n *Node) (html string, handled bool) {
	for _, h := range m.handlers[n.Type] {
		if html, handled = h(ctx, n); handled {
			return html, true
		}
	}
	return "", false
}

// Load reads a markdown file from the content filesystem and parses it into a Document.
// YAML front matter delimited by --- is extracted and available via FrontMatter().
// A renderer built by New(nil) has no content filesystem and returns an error;
// use RenderBytes to render source that is already in memory.
func (m *Markdown) Load(filename string) (*Document, error) {
	if m.contentFS == nil {
		return nil, fmt.Errorf("loading %s: no content filesystem", filename)
	}

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

// RenderBytes parses markdown source bytes and writes rendered HTML to w.
// Unlike Load, this does not read from the filesystem or parse front matter.
func (m *Markdown) RenderBytes(w io.Writer, src []byte) error {
	return m.RenderBytesContext(context.Background(), w, src)
}

// RenderBytesContext parses markdown source bytes and writes rendered HTML to w with context.
func (m *Markdown) RenderBytesContext(ctx context.Context, w io.Writer, src []byte) error {
	reader := text.NewReader(src)
	doc := m.parser.Parse(reader)
	return m.renderChildren(ctx, w, doc, src)
}

// renderChildren renders all direct children of a node.
func (m *Markdown) renderChildren(ctx context.Context, w io.Writer, node ast.Node, src []byte) error {
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if err := m.renderNode(ctx, w, child, src); err != nil {
			return err
		}
	}
	return nil
}

// renderNode dispatches a single AST node to the appropriate template.
func (m *Markdown) renderNode(ctx context.Context, w io.Writer, node ast.Node, src []byte) error {
	switch n := node.(type) {
	case *ast.Heading:
		return m.renderHeading(ctx, w, n, src)
	case *ast.Paragraph:
		return m.renderParagraph(ctx, w, n, src)
	case *ast.FencedCodeBlock:
		return m.renderFencedCodeBlock(ctx, w, n, src)
	case *ast.CodeBlock:
		return m.renderCodeBlock(ctx, w, n, src)
	case *ast.Blockquote:
		return m.renderBlockquote(ctx, w, n, src)
	case *ast.List:
		return m.renderList(ctx, w, n, src)
	case *ast.ListItem:
		return m.renderListItem(ctx, w, n, src)
	case *ast.ThematicBreak:
		return m.renderTemplate(w, "thematic_break", nil)
	case *ast.HTMLBlock:
		return m.renderHTMLBlock(w, n, src)
	case *ast.TextBlock:
		return m.renderInlineChildren(w, n, src)
	case *east.Table:
		return m.renderTable(w, n, src)
	case *east.FootnoteList:
		return m.renderFootnoteList(ctx, w, n, src)
	case *east.Footnote:
		return m.renderFootnote(ctx, w, n, src)
	default:
		if node.HasChildren() {
			return m.renderChildren(ctx, w, node, src)
		}
	}
	return nil
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
