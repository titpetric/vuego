package vuego

import (
	"context"
	"io"
	"strings"

	"golang.org/x/net/html"
)

// Component is a renderable .vuego template component.
type Component interface {
	// Render renders the component template to the given writer.
	Render(ctx context.Context, w io.Writer, filename string, data any) error
}

// Renderer renders parsed HTML nodes to output.
type Renderer interface {
	// Render renders HTML nodes to the given writer.
	Render(ctx context.Context, w io.Writer, nodes []*html.Node) error
}

// htmlRenderer is the default Renderer implementation.
type htmlRenderer struct{}

// Render renders HTML nodes to the given writer.
func (r *htmlRenderer) Render(ctx context.Context, w io.Writer, nodes []*html.Node) error {
	for _, node := range nodes {
		if err := renderNode(w, node, 0); err != nil {
			return err
		}
	}
	return nil
}

// NewRenderer creates a new Renderer.
func NewRenderer() Renderer {
	return &htmlRenderer{}
}

// shouldIgnoreAttr returns true if an attribute should be skipped in HTML output.
// This includes Vue directives and binding attributes that are only used during evaluation.
func shouldIgnoreAttr(key string) bool {
	// Vue directives that should not appear in final HTML
	switch key {
	case "v-if", "v-else-if", "v-else", "v-for", "v-pre", "v-html", "v-text", "v-show", "v-once", "v-once-id", "data-v-html-content", "data-v-text-content":
		return true
	}
	return false
}

func renderAttrs(attrs []html.Attribute) string {
	if len(attrs) == 0 {
		return ""
	}
	var sb strings.Builder
	for _, a := range attrs {
		// Skip Vue directives and internal binding attributes
		if shouldIgnoreAttr(a.Key) {
			continue
		}
		sb.WriteByte(' ')
		sb.WriteString(a.Key)
		sb.WriteByte('=')
		sb.WriteByte('"')
		sb.WriteString(a.Val)
		sb.WriteByte('"')
	}
	return sb.String()
}

var indentCache = [256]string{}

func init() {
	for i := 0; i < len(indentCache); i++ {
		indentCache[i] = strings.Repeat(" ", i)
	}
}

func getIndent(indent int) string {
	if indent < len(indentCache) {
		return indentCache[indent]
	}
	return strings.Repeat(" ", indent)
}

// shouldEscapeTextNode checks if a text node needs HTML escaping.
// Returns false if the text appears to be already HTML-escaped (from interpolation),
// true if it contains raw HTML special characters that need escaping.
func shouldEscapeTextNode(data string) bool {
	// If the text contains HTML entity references like &lt; &amp; &#39; etc,
	// it's likely from interpolation and already escaped
	if strings.Contains(data, "&") && strings.Contains(data, ";") {
		return false
	}
	// Check if text contains unescaped HTML special characters
	return strings.ContainsAny(data, "<>&\"'")
}

func renderNode(w io.Writer, node *html.Node, indent int) error {
	ctx := VueContext{}
	return renderNodeWithContext(ctx, w, node, indent)
}

func renderNodeWithContext(ctx VueContext, w io.Writer, node *html.Node, indent int) error {
	switch node.Type {
	case html.TextNode:
		if strings.TrimSpace(node.Data) == "" {
			return nil
		}
		spaces := getIndent(indent)
		parentTag := ctx.CurrentTag()
		// Skip HTML escaping inside script and style tags
		if parentTag == "script" || parentTag == "style" {
			_, _ = w.Write([]byte(spaces + node.Data))
		} else if shouldEscapeTextNode(node.Data) {
			_, _ = w.Write([]byte(spaces + html.EscapeString(node.Data)))
		} else {
			_, _ = w.Write([]byte(spaces + node.Data))
		}

	case html.ElementNode:
		// Count children without allocating slice
		childCount := 0
		firstChild := node.FirstChild
		for c := firstChild; c != nil; c = c.NextSibling {
			childCount++
		}

		spaces := getIndent(indent)
		tagName := node.Data

		// Check if this element has evaluated v-html or v-text content (stored in internal attributes)
		var vhtmlContent string
		var vtextContent string
		for _, attr := range node.Attr {
			if attr.Key == "data-v-html-content" {
				vhtmlContent = attr.Val
				break
			}
			if attr.Key == "data-v-text-content" {
				vtextContent = attr.Val
				break
			}
		}

		// If this element has v-html or v-text content, output it directly without indentation
		if vhtmlContent != "" || vtextContent != "" {
			content := vhtmlContent
			if vtextContent != "" {
				content = vtextContent
			}
			// Special case for <template>: output content only, not the template tags
			if tagName == "template" {
				_, _ = w.Write([]byte(content))
			} else {
				_, _ = w.Write([]byte(spaces + "<" + tagName + renderAttrs(node.Attr) + ">"))
				_, _ = w.Write([]byte(content))
				_, _ = w.Write([]byte("</" + tagName + ">\n"))
			}
			return nil
		}

		// Special handling for <template> elements without v-html: output children without template tag
		if tagName == "template" {
			for c := firstChild; c != nil; c = c.NextSibling {
				if err := renderNodeWithContext(ctx, w, c, indent); err != nil {
					return err
				}
			}
			return nil
		}

		// compact single-entry text nodes
		if childCount == 0 {
			_, _ = w.Write([]byte(spaces + "<" + tagName + renderAttrs(node.Attr) + "></" + tagName + ">\n"))
		} else if childCount == 1 && firstChild.Type == html.TextNode {
			_, _ = w.Write([]byte(spaces + "<" + tagName + renderAttrs(node.Attr) + ">"))
			// Skip HTML escaping inside script and style tags
			if tagName == "script" || tagName == "style" {
				_, _ = w.Write([]byte(firstChild.Data))
			} else if shouldEscapeTextNode(firstChild.Data) {
				_, _ = w.Write([]byte(html.EscapeString(firstChild.Data)))
			} else {
				_, _ = w.Write([]byte(firstChild.Data))
			}
			_, _ = w.Write([]byte("</" + tagName + ">\n"))
		} else {
			_, _ = w.Write([]byte(spaces + "<" + tagName + renderAttrs(node.Attr) + ">\n"))
			ctx.PushTag(tagName)
			childIndent := indent + 2
			for c := firstChild; c != nil; c = c.NextSibling {
				if err := renderNodeWithContext(ctx, w, c, childIndent); err != nil {
					return err
				}
			}
			ctx.PopTag()
			_, _ = w.Write([]byte(spaces + "</" + tagName + ">\n"))
		}
	}

	return nil
}
