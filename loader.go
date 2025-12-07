package vuego

import (
	"bytes"
	"fmt"
	"io/fs"

	"golang.org/x/net/html"
	yaml "gopkg.in/yaml.v3"

	"github.com/titpetric/vuego/internal/helpers"
)

// Loader loads and parses .vuego files from an fs.FS.
type Loader struct {
	// FS is the file system used to load templates.
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
	if l.FS == nil {
		return nil, fmt.Errorf("error reading %s: no filesystem configured", filename)
	}
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

// extractFrontMatter extracts YAML front-matter from a template if it starts with ---.
// Returns the front-matter data as map[string]any and the remaining template content.
// If no front-matter is present, returns nil map and the original content.
func extractFrontMatter(content []byte) (map[string]any, []byte, error) {
	// Check if content starts with ---
	if !bytes.HasPrefix(content, []byte("---")) {
		return nil, content, nil
	}

	// Find the closing --- delimiter
	rest := content[3:]
	delimIdx := bytes.Index(rest, []byte("\n---"))
	if delimIdx == -1 {
		// No closing delimiter found
		return nil, content, nil
	}

	// Extract front-matter YAML (between the delimiters)
	frontMatterBytes := rest[:delimIdx]
	remainingContent := rest[delimIdx+4:] // Skip "\n---"
	// Skip optional newline after closing ---
	if len(remainingContent) > 0 && remainingContent[0] == '\n' {
		remainingContent = remainingContent[1:]
	}

	// Parse YAML
	data := make(map[string]any)
	if err := yaml.Unmarshal(frontMatterBytes, &data); err != nil {
		return nil, nil, fmt.Errorf("error parsing front-matter YAML: %w", err)
	}

	return data, remainingContent, nil
}

// LoadFragment parses a template fragment; if the file is a full document, it falls back to Load.
// Front-matter is extracted and discarded; use loadFragmentInternal to access it.
func (l *Loader) LoadFragment(filename string) ([]*html.Node, error) {
	_, nodes, err := l.loadFragmentInternal(filename)
	return nodes, err
}

// loadFragmentInternal parses a template fragment and returns both front-matter data and DOM nodes.
func (l *Loader) loadFragmentInternal(filename string) (map[string]any, []*html.Node, error) {
	if l.FS == nil {
		return nil, nil, fmt.Errorf("error reading %s: no filesystem configured", filename)
	}
	template, err := fs.ReadFile(l.FS, filename)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading %s: %w", filename, err)
	}

	// Extract front-matter
	frontMatter, template, err := extractFrontMatter(template)
	if err != nil {
		return nil, nil, err
	}

	// Input template contains html/body
	var nodes []*html.Node
	if bytes.Contains(template, []byte("</html>")) {
		doc, err := html.Parse(bytes.NewReader(template))
		if err != nil {
			return nil, nil, err
		}
		for node := range doc.ChildNodes() {
			nodes = append(nodes, node)
		}
	} else {
		// Parse the fragment using cached body element
		body := helpers.GetBodyNode()
		nodes, err = html.ParseFragment(bytes.NewReader(template), body)
		if err != nil {
			return nil, nil, err
		}
	}

	return frontMatter, nodes, nil
}
