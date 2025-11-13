package vuego

import (
	"bytes"
	"fmt"
	"io/fs"
	"strings"

	"golang.org/x/net/html"
)

type Component struct {
	FS fs.FS
}

func NewComponent(fs fs.FS) *Component {
	return &Component{
		FS: fs,
	}
}

func (c Component) Stat(filename string) error {
	_, err := fs.Stat(c.FS, filename)
	return err
}

func (c Component) Load(filename string) ([]*html.Node, error) {
	template, err := fs.ReadFile(c.FS, filename)
	if err != nil {
		return nil, fmt.Errorf("File not found: %s", filename)
	}

	doc, err := html.Parse(bytes.NewReader(template))
	if err != nil {
		return nil, err
	}

	var result []*html.Node
	for node := range doc.ChildNodes() {
		result = append(result, node)
	}
	return result, nil
}

func (c Component) LoadFragment(filename string) ([]*html.Node, error) {
	template, err := fs.ReadFile(c.FS, filename)
	if err != nil {
		return nil, fmt.Errorf("File not found: %s", filename)
	}

	// Input template contains html/body
	if bytes.Contains(template, []byte("</html>")) {
		return c.Load(filename)
	}

	// Step 1: parse a minimal document
	doc, err := html.Parse(strings.NewReader("<html><body></body></html>"))
	if err != nil {
		return nil, err
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
		return nil, err
	}

	// Step 3: parse the fragment using <body> as context
	return html.ParseFragment(bytes.NewReader(template), body)
}
