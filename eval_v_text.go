package vuego

import (
	"fmt"
	"html"
	"strings"

	htmlnode "golang.org/x/net/html"

	"github.com/titpetric/vuego/internal/helpers"
)

// evalVText evaluates the v-text directive, setting escaped text content on the node.
// Unlike v-html which outputs raw HTML, v-text escapes the content for HTML safety.
func (v *Vue) evalVText(ctx VueContext, n *htmlnode.Node) error {
	expr := helpers.GetAttr(n, "v-text")
	if expr == "" {
		return nil
	}

	val, ok := ctx.stack.Resolve(expr)
	if !ok {
		// v-text may be a function call like "file(src)"
		var err error
		if strings.Contains(expr, "|") || helpers.IsFunctionCall(expr) || helpers.IsComplexExpr(expr) {
			pipe := parsePipeExpr(expr)
			val, err = v.evalPipe(ctx, pipe)
			if err != nil {
				return fmt.Errorf("in expression '{{ %s }}': %w", expr, err)
			}
			ok = true
		}
	}

	if !ok {
		return nil
	}

	// Evaluate v-text expression to its string value and escape for HTML
	textStr := fmt.Sprint(val)
	escapedStr := html.EscapeString(textStr)
	n.Attr = append(n.Attr, htmlnode.Attribute{Key: "data-v-text-content", Val: escapedStr})

	// Clear children - v-text content will be output directly during rendering
	n.FirstChild = nil
	n.LastChild = nil

	return nil
}
