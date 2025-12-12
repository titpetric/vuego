package vuego

import (
	"bytes"
	"io/fs"
	"strings"

	"github.com/titpetric/lessgo/dst"
	"github.com/titpetric/lessgo/renderer"
	"golang.org/x/net/html"
)

// LessProcessorError wraps processing errors with context.
type LessProcessorError struct {
	// Err is the underlying error from the LESS processor.
	Err error
	// Reason is a descriptive message about the LESS processing failure.
	Reason string
}

func (e *LessProcessorError) Error() string {
	return e.Reason + ": " + e.Err.Error()
}

// LessProcessor implements vuego.NodeProcessor.
//
// It implements the functionality to compile Less CSS to standard css.
// It processes `script` tags with a `type="text/css+less"` attribute.
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

// New creates a new allocation of *LessProcessor.
func (lp *LessProcessor) New() NodeProcessor {
	return &LessProcessor{
		fs: lp.fs,
	}
}

// PreProcess.
func (lp *LessProcessor) PreProcess(nodes []*html.Node) error {
	return nil
}

// PostProcess walks the DOM tree and performs compilation.
func (lp *LessProcessor) PostProcess(nodes []*html.Node) error {
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

	// Check if this is a style tag with type="text/css+less"
	if node.Type == html.ElementNode && node.Data == "style" {
		if lp.isLessStyleTag(node) {
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

// isLessStyleTag checks if a style tag has type="text/css+less".
func (lp *LessProcessor) isLessStyleTag(node *html.Node) bool {
	for _, attr := range node.Attr {
		if attr.Key == "type" && attr.Val == "text/css+less" {
			return true
		}
	}
	return false
}

// compileLessTag extracts LESS content from the style tag, compiles it to CSS, and replaces the tag with a style tag.
func (lp *LessProcessor) compileLessTag(styleNode *html.Node) error {
	// Extract the LESS content from the style tag's text content
	lessContent := ""
	for c := styleNode.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			lessContent += c.Data
		}
	}

	if lessContent == "" {
		return nil // Empty style tag, nothing to compile
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

	// Replace the <style type="text/css+less"> tag with a <style> tag
	lp.replaceWithStyleTag(styleNode, css)

	return nil
}

// replaceWithStyleTag converts a style tag to a style tag with compiled CSS.
func (lp *LessProcessor) replaceWithStyleTag(styleNode *html.Node, css string) {
	// Change tag name to "style"
	styleNode.Data = "style"

	// Keep only relevant attributes (remove type="text/css+less")
	newAttrs := []html.Attribute{}
	for _, attr := range styleNode.Attr {
		if attr.Key != "type" {
			newAttrs = append(newAttrs, attr)
		}
	}
	styleNode.Attr = newAttrs

	// Clear children and set CSS content
	styleNode.FirstChild = nil
	styleNode.LastChild = nil

	// Create a text node with the compiled CSS
	textNode := &html.Node{
		Type: html.TextNode,
		Data: "\n" + strings.TrimSpace(css) + "\n",
	}
	styleNode.FirstChild = textNode
	styleNode.LastChild = textNode
}
