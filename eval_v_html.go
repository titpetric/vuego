package vuego

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"

	"github.com/titpetric/vuego/internal/helpers"
)

func (v *Vue) evalVHtml(ctx VueContext, n *html.Node) error {
	expr := helpers.GetAttr(n, "v-html")
	if expr == "" {
		return nil
	}

	val, ok := ctx.stack.Resolve(expr)
	if !ok {
		// v-html may be a function call like "file(src)"
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

	// Evaluate v-html expression to its string value and store in internal attribute
	htmlStr := fmt.Sprint(val)
	n.Attr = append(n.Attr, html.Attribute{Key: "data-v-html-content", Val: htmlStr})

	// Clear children - v-html content will be output directly during rendering
	n.FirstChild = nil
	n.LastChild = nil

	return nil
}
