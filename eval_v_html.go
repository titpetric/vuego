package vuego

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/titpetric/vuego/internal/helpers"
	"golang.org/x/net/html"
)

func (v *Vue) evalVHtml(n *html.Node) error {
	vHtmlExpr := helpers.GetAttr(n, "v-html")
	if vHtmlExpr == "" {
		return nil
	}
	helpers.RemoveAttr(n, "v-html")

	val, ok := v.stack.Resolve(vHtmlExpr)
	if !ok {
		return nil
	}

	htmlStr := fmt.Sprintf("%v", val)
	if htmlStr == "" {
		n.FirstChild = nil
		n.LastChild = nil
		return nil
	}

	// use a neutral <div> as context for parsing
	wrapperDoc, err := html.Parse(bytes.NewReader([]byte("<html><body></body></html>")))
	if err != nil {
		return fmt.Errorf("v-html parse wrapper error: %w", err)
	}

	// find the body element
	var body *html.Node
	var findBody func(*html.Node)
	findBody = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "body" {
			body = node
			return
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			findBody(c)
		}
	}
	findBody(wrapperDoc)

	// parse fragment into wrapper's body
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
