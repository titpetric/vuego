package parser

import (
	"bytes"

	"golang.org/x/net/html"

	"github.com/titpetric/vuego/internal/helpers"
)

// ParseTemplateBytes parses template bytes into HTML nodes, handling both full documents and fragments.
// If the content contains a full HTML document (</html> tag), it uses html.Parse.
// Otherwise, it parses as a fragment using a cached body element.
func ParseTemplateBytes(templateBytes []byte) ([]*html.Node, error) {
	// Check if input template contains html/body
	if bytes.Contains(templateBytes, []byte("</html>")) {
		doc, err := html.Parse(bytes.NewReader(templateBytes))
		if err != nil {
			return nil, err
		}
		var nodes []*html.Node
		for node := range doc.ChildNodes() {
			nodes = append(nodes, node)
		}
		return nodes, nil
	}

	// Parse the fragment using cached body element
	body := helpers.GetBodyNode()
	nodes, err := html.ParseFragment(bytes.NewReader(templateBytes), body)
	if err != nil {
		return nil, err
	}
	return nodes, nil
}
