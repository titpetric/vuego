package vuego

import (
	"bytes"
	"fmt"
	"io/fs"

	"github.com/titpetric/vuego/internal/helpers"
	"golang.org/x/net/html"
)

// Loader loads and parses .vuego files from an fs.FS.
type Loader struct {
	FS fs.FS
}

// NewLoader creates a Loader backed by fs.
func NewLoader(fs fs.FS) *Loader {
	return &Loader{
		FS: fs,
	}
}

// Stat checks that filename exists in the loader filesystem.
func (l *Loader) Stat(filename string) error {
	_, err := fs.Stat(l.FS, filename)
	return err
}

// Load parses a full HTML document from the given filename.
func (l *Loader) Load(filename string) ([]*html.Node, error) {
	template, err := fs.ReadFile(l.FS, filename)
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
func (l *Loader) LoadFragment(filename string) ([]*html.Node, error) {
	template, err := fs.ReadFile(l.FS, filename)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %w", filename, err)
	}

	// Input template contains html/body
	if bytes.Contains(template, []byte("</html>")) {
		return l.Load(filename)
	}

	// Parse the fragment using cached body element
	body := helpers.GetBodyNode()
	return html.ParseFragment(bytes.NewReader(template), body)
}
