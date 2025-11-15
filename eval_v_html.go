package vuego

import (
	"fmt"
	"strings"

	"github.com/titpetric/vuego/internal/helpers"
	"golang.org/x/net/html"
)

func (v *Vue) evalVHtml(ctx VueContext, n *html.Node) error {
	vHtmlExpr := helpers.GetAttr(n, "v-html")
	if vHtmlExpr == "" {
		return nil
	}
	helpers.RemoveAttr(n, "v-html")

	val, ok := ctx.stack.Resolve(vHtmlExpr)
	if !ok {
		return nil
	}

	htmlStr := fmt.Sprint(val)
	if htmlStr == "" {
		n.FirstChild = nil
		n.LastChild = nil
		return nil
	}

	// parse fragment using cached body element
	body := helpers.GetBodyNode()
	nodes, err := html.ParseFragment(strings.NewReader(htmlStr), body)
	if err != nil {
		return fmt.Errorf("v-html parse fragment error: %w", err)
	}

	// clear existing children
	n.FirstChild = nil
	n.LastChild = nil

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
