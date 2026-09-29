package formatter

import (
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/titpetric/vuego/internal/helpers"
)

// collectChildren returns non-ignorable children of a node.
func (f *Formatter) collectChildren(n *html.Node) []*html.Node {
	var children []*html.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode && f.isIgnorableWhitespace(c) {
			continue
		}
		children = append(children, c)
	}
	return children
}

// formatNode recursively formats an HTML node tree.
func (f *Formatter) formatNode(n *html.Node, buf *strings.Builder, depth int) {
	indent := strings.Repeat(" ", depth*f.opts.IndentWidth)

	switch n.Type {
	case html.DocumentNode:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f.formatNode(c, buf, depth)
		}

	case html.ElementNode:
		// Style/script blocks - preserve content as-is
		if n.Data == "style" || n.Data == "script" {
			f.formatRawTextElement(n, buf, indent)
			return
		}

		// Pre blocks - preserve content whitespace and escape entities
		if n.Data == "pre" {
			buf.WriteString(indent)
			buf.WriteString(f.renderOpenTag(n))
			f.renderPreContent(n, buf)
			buf.WriteString(f.renderCloseTag(n))
			buf.WriteString("\n")
			return
		}

		openTag := f.renderOpenTag(n)
		buf.WriteString(indent)
		buf.WriteString(openTag)

		// Void elements (self-closing)
		if isVoidElement(n.DataAtom) {
			buf.WriteString("\n")
			return
		}

		children := f.collectChildren(n)

		// Empty element - keep on one line
		if len(children) == 0 {
			buf.WriteString(f.renderCloseTag(n))
			buf.WriteString("\n")
			return
		}

		// Inline content - render on one line
		if f.shouldKeepInline(n) {
			buf.WriteString(f.renderInlineChildren(n))
			buf.WriteString(f.renderCloseTag(n))
			buf.WriteString("\n")
			return
		}

		// Block mode - newline after opening tag, indent children
		buf.WriteString("\n")
		for _, c := range children {
			f.formatNode(c, buf, depth+1)
		}
		buf.WriteString(indent)
		buf.WriteString(f.renderCloseTag(n))
		buf.WriteString("\n")

	case html.TextNode:
		text := strings.TrimSpace(n.Data)
		if text != "" {
			buf.WriteString(indent)
			buf.WriteString(escapeText(text))
			buf.WriteString("\n")
		}

	case html.CommentNode:
		buf.WriteString(indent)
		buf.WriteString("<!--")
		buf.WriteString(n.Data)
		buf.WriteString("-->\n")
	}
}

// formatRawTextElement formats script/style elements preserving their content.
func (f *Formatter) formatRawTextElement(n *html.Node, buf *strings.Builder, indent string) {
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
	trimmedContent := trimRawContent(contentStr)

	if trimmedContent == "" {
		buf.WriteString(f.renderCloseTag(n))
		buf.WriteString("\n")
	} else {
		buf.WriteString("\n")
		buf.WriteString(trimmedContent)
		buf.WriteString("\n")
		buf.WriteString(indent)
		buf.WriteString(f.renderCloseTag(n))
		buf.WriteString("\n")
	}
}

// renderPreContent recursively renders children of a <pre> element,
// preserving whitespace and escaping text content.
func (f *Formatter) renderPreContent(n *html.Node, buf *strings.Builder) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.TextNode:
			buf.WriteString(escapeText(c.Data))
		case html.ElementNode:
			buf.WriteString(f.renderOpenTag(c))
			if !isVoidElement(c.DataAtom) {
				f.renderPreContent(c, buf)
				buf.WriteString(f.renderCloseTag(c))
			}
		case html.CommentNode:
			buf.WriteString("<!--")
			buf.WriteString(c.Data)
			buf.WriteString("-->")
		}
	}
}

// shouldKeepInline determines if element content should be kept on one line.
//
// Rules:
//  1. Elements with only text children (no element children) are always inline.
//  2. Inline elements (span, strong, em, etc.) stay inline if all their
//     children are also inline-compatible (text, inline elements, void elements).
//  3. Phrasing containers (p, h1-h6, li, td, th, dt, dd, label, button, caption)
//     stay inline if all their children are inline-compatible.
//  4. All other block elements (div, nav, section, etc.) are never inlined
//     when they contain element children.
func (f *Formatter) shouldKeepInline(n *html.Node) bool {
	// Check if there are any element children at all
	hasElementChildren := false
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			hasElementChildren = true
			break
		}
	}

	// Text-only content: always inline
	if !hasElementChildren {
		return true
	}

	// For elements with child elements, only allow inline rendering
	// for inline elements and phrasing containers
	if isInlineAtom(n.DataAtom) || isPhrasingContainer(n.DataAtom) {
		return f.allChildrenAreInline(n)
	}

	return false
}

// allChildrenAreInline checks if all children of a node are inline-compatible
// (text nodes, comment nodes, or inline element nodes).
func (f *Formatter) allChildrenAreInline(n *html.Node) bool {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.TextNode:
			// text is always inline
		case html.CommentNode:
			// comments are inline-compatible
		case html.ElementNode:
			if isVoidElement(c.DataAtom) {
				continue
			}
			if !isInlineAtom(c.DataAtom) {
				return false
			}
			// Recursively check that inline children also only have inline content
			if !f.allChildrenAreInline(c) {
				return false
			}
		default:
			return false
		}
	}
	return true
}

// isPhrasingContainer checks if an atom is a block element that typically
// contains phrasing (inline) content and should be rendered inline when
// all its children are inline-compatible.
func isPhrasingContainer(a atom.Atom) bool {
	phrasingContainers := []atom.Atom{
		atom.P, atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6,
		atom.Td, atom.Th, atom.Dt, atom.Dd, atom.Caption,
		atom.Figcaption, atom.Summary, atom.Legend,
	}
	for _, elem := range phrasingContainers {
		if a == elem {
			return true
		}
	}
	return false
}

// renderInlineChildren recursively renders children as inline content on one line.
// Leading and trailing whitespace is trimmed so that inline content sits flush
// against its parent tags (e.g. <span><i>x</i></span> not <span> <i>x</i> </span>).
func (f *Formatter) renderInlineChildren(n *html.Node) string {
	var b strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		switch c.Type {
		case html.TextNode:
			b.WriteString(escapeText(normalizeInlineText(c.Data)))
		case html.CommentNode:
			b.WriteString("<!--")
			b.WriteString(c.Data)
			b.WriteString("-->")
		case html.ElementNode:
			b.WriteString(f.renderOpenTag(c))
			if !isVoidElement(c.DataAtom) {
				b.WriteString(f.renderInlineChildren(c))
				b.WriteString(f.renderCloseTag(c))
			}
		}
	}
	return strings.TrimSpace(b.String())
}

// isIgnorableWhitespace checks if a text node is only whitespace between block elements.
func (f *Formatter) isIgnorableWhitespace(n *html.Node) bool {
	if n.Type != html.TextNode {
		return false
	}
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

// isInlineAtom checks if an atom represents an inline HTML element.
func isInlineAtom(a atom.Atom) bool {
	inlineElements := []atom.Atom{
		atom.A, atom.Abbr, atom.B, atom.Bdi, atom.Bdo, atom.Br, atom.Button,
		atom.Cite, atom.Code, atom.Data, atom.Dfn, atom.Em, atom.I, atom.Kbd,
		atom.Label, atom.Mark, atom.Q, atom.S, atom.Samp, atom.Small, atom.Span,
		atom.Strong, atom.Sub, atom.Sup, atom.Time, atom.U, atom.Var, atom.Wbr,
	}

	for _, elem := range inlineElements {
		if a == elem {
			return true
		}
	}
	return false
}

// fragmentContext returns an appropriate context node for html.ParseFragment
// based on the leading tag in the body. Table-scoped elements like <td>, <th>,
// <tr>, <thead>, <tbody>, <tfoot>, <caption>, and <colgroup> require ancestor
// context nodes to avoid being stripped by the HTML5 parser.
func fragmentContext(body string) *html.Node {
	trimmed := strings.TrimSpace(body)

	type contextMapping struct {
		prefix   string
		dataAtom atom.Atom
		data     string
	}

	// Order matters: check more specific elements first.
	mappings := []contextMapping{
		{"<td", atom.Tr, "tr"},
		{"<th", atom.Tr, "tr"},
		{"<tr", atom.Tbody, "tbody"},
		{"<thead", atom.Table, "table"},
		{"<tbody", atom.Table, "table"},
		{"<tfoot", atom.Table, "table"},
		{"<caption", atom.Table, "table"},
		{"<colgroup", atom.Table, "table"},
		{"<col", atom.Colgroup, "colgroup"},
	}

	lower := strings.ToLower(trimmed)
	for _, m := range mappings {
		if strings.HasPrefix(lower, m.prefix) {
			// Ensure the prefix is followed by a space, >, or end-of-string
			// to avoid false matches (e.g., "<the" matching "<th").
			rest := lower[len(m.prefix):]
			if len(rest) == 0 || rest[0] == ' ' || rest[0] == '>' || rest[0] == '\n' || rest[0] == '\t' || rest[0] == '/' {
				return &html.Node{Type: html.ElementNode, DataAtom: m.dataAtom, Data: m.data}
			}
		}
	}

	return &html.Node{Type: html.ElementNode, DataAtom: atom.Body, Data: "body"}
}
