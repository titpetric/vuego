package vuego

import (
	"io"
	"strings"

	"golang.org/x/net/html"
)

// Engine is an abstracted interface to a template evaluator.
type Engine interface {
	Evaluate(w io.Writer, root *html.Node, data map[string]any) error
}

// Render will use the default template evaluator to render the template.
func Render(w io.Writer, template string, data map[string]any) error {
	return RenderEngine(w, New(), template, data)
}

// RenderEngine renders a template snippet to w using innerHTML semantics.
// It parses the snippet as a fragment of a minimal document, avoiding html/head/body wrappers.
// The engine can be overriden with extended implementations implementing the Engine interface.
func RenderEngine(w io.Writer, engine Engine, template string, data map[string]any) error {
	// Step 1: parse a minimal document
	doc, err := html.Parse(strings.NewReader("<html><body></body></html>"))
	if err != nil {
		return err
	}

	// Step 2: find <body> node
	var body *html.Node
	var findBody func(*html.Node)
	findBody = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "body" {
			body = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if body == nil {
				findBody(c)
			}
		}
	}
	findBody(doc)
	if body == nil {
		return err
	}

	// Step 3: parse the fragment using <body> as context
	nodes, err := html.ParseFragment(strings.NewReader(template), body)
	if err != nil {
		return err
	}

	// Step 4: evaluate template nodes individually
	for _, n := range nodes {
		if err := engine.Evaluate(w, n, data); err != nil {
			return err
		}
	}

	return nil
}
