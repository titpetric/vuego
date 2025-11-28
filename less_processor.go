package vuego

import (
	"bytes"
	"io/fs"

	"github.com/titpetric/lessgo/dst"
	"github.com/titpetric/lessgo/renderer"
	"golang.org/x/net/html"
)

// LessProcessorError wraps processing errors with context.
type LessProcessorError struct {
	Err    error
	Reason string
}

func (e *LessProcessorError) Error() string {
	return e.Reason + ": " + e.Err.Error()
}

// LessProcessor implements vuego.NodeProcessor to compile LESS to CSS in <script type="text/css+less"> tags.
type LessProcessor struct {
	// fs is optional filesystem for loading @import statements in LESS files
	fs fs.FS
}

// NewLessProcessor creates a new LESS processor.
func NewLessProcessor(fsys ...fs.FS) *LessProcessor {
	var fsVal fs.FS
	if len(fsys) > 0 {
		fsVal = fsys[0]
	}
	return &LessProcessor{fs: fsVal}
}

// Process walks the DOM tree and compiles LESS in <script type="text/css+less"> tags to CSS.
func (lp *LessProcessor) Process(nodes []*html.Node) error {
	for _, node := range nodes {
		if err := lp.processNode(node); err != nil {
			return err
		}
	}
	return nil
}

// processNode recursively processes a single node and its children.
func (lp *LessProcessor) processNode(node *html.Node) error {
	if node == nil {
		return nil
	}

	// Check if this is a script tag with type="text/css+less"
	if node.Type == html.ElementNode && node.Data == "script" {
		if lp.isLessScriptTag(node) {
			if err := lp.compileLessTag(node); err != nil {
				return err
			}
		}
	}

	// Recursively process all children
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if err := lp.processNode(c); err != nil {
			return err
		}
	}

	return nil
}

// isLessScriptTag checks if a script tag has type="text/css+less".
func (lp *LessProcessor) isLessScriptTag(node *html.Node) bool {
	for _, attr := range node.Attr {
		if attr.Key == "type" && attr.Val == "text/css+less" {
			return true
		}
	}
	return false
}

// compileLessTag extracts LESS content from the script tag, compiles it to CSS, and replaces the tag with a style tag.
func (lp *LessProcessor) compileLessTag(scriptNode *html.Node) error {
	// Extract the LESS content from the script tag's text content
	lessContent := ""
	for c := scriptNode.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			lessContent += c.Data
		}
	}

	if lessContent == "" {
		return nil // Empty script tag, nothing to compile
	}

	// Parse and compile LESS to CSS
	parser := dst.NewParser(bytes.NewReader([]byte(lessContent)))
	if lp.fs != nil {
		parser = dst.NewParserWithFS(bytes.NewReader([]byte(lessContent)), lp.fs)
	}

	file, err := parser.Parse()
	if err != nil {
		return &LessProcessorError{Err: err, Reason: "failed to parse LESS"}
	}

	// Render LESS to CSS
	r := renderer.NewRenderer()
	css, err := r.Render(file)
	if err != nil {
		return &LessProcessorError{Err: err, Reason: "failed to render LESS to CSS"}
	}

	// Replace the <script type="text/css+less"> tag with a <style> tag
	lp.replaceWithStyleTag(scriptNode, css)

	return nil
}

// replaceWithStyleTag converts a script tag to a style tag with compiled CSS.
func (lp *LessProcessor) replaceWithStyleTag(scriptNode *html.Node, css string) {
	// Change tag name to "style"
	scriptNode.Data = "style"

	// Keep only relevant attributes (remove type="text/css+less")
	newAttrs := []html.Attribute{}
	for _, attr := range scriptNode.Attr {
		if attr.Key != "type" {
			newAttrs = append(newAttrs, attr)
		}
	}
	scriptNode.Attr = newAttrs

	// Clear children and set CSS content
	scriptNode.FirstChild = nil
	scriptNode.LastChild = nil

	// Create a text node with the compiled CSS
	textNode := &html.Node{
		Type: html.TextNode,
		Data: css,
	}
	scriptNode.FirstChild = textNode
	scriptNode.LastChild = textNode
}
