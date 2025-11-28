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

	htmlStr := fmt.Sprint(val)
	if !ok {
		return nil
	}

	n.FirstChild = nil
	n.LastChild = nil

	// parse fragment using cached body element
	body := helpers.GetBodyNode()
	nodes, err := html.ParseFragment(strings.NewReader(htmlStr), body)
	if err != nil {
		return fmt.Errorf("v-html parse fragment error: %w", err)
	}
	if len(nodes) == 0 {
		return nil
	}

	// append parsed nodes as children
	var prev *html.Node
	for _, c := range nodes {
		c.Parent = n
		c.PrevSibling = prev
		if prev != nil {
			prev.NextSibling = c
		} else {
			n.FirstChild = c
		}
		prev = c
	}
	n.LastChild = prev

	return nil
}
