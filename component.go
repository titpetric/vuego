package vuego

import (
	"bytes"
	"fmt"
	"io/fs"

	"github.com/titpetric/vuego/internal/helpers"
	"golang.org/x/net/html"
)

// Component loads and parses .vuego component files from an fs.FS.
type Component struct {
	FS fs.FS
}

// NewComponent creates a Component backed by fs.
func NewComponent(fs fs.FS) *Component {
	return &Component{
		FS: fs,
	}
}

// Stat checks that filename exists in the component filesystem.
func (c Component) Stat(filename string) error {
	_, err := fs.Stat(c.FS, filename)
	return err
}

// Load parses a full HTML document from the given filename.
func (c Component) Load(filename string) ([]*html.Node, error) {
	template, err := fs.ReadFile(c.FS, filename)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %w", filename, err)
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

// LoadFragment parses a template fragment; if the file is a full document, it falls back to Load.
func (c Component) LoadFragment(filename string) ([]*html.Node, error) {
	template, err := fs.ReadFile(c.FS, filename)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %w", filename, err)
	}

	// Input template contains html/body
	if bytes.Contains(template, []byte("</html>")) {
		return c.Load(filename)
	}

	// Parse the fragment using cached body element
	body := helpers.GetBodyNode()
	return html.ParseFragment(bytes.NewReader(template), body)
}
