package markdown

import (
	"context"
	"io"
)

// FrontMatter returns the parsed YAML front matter from the document.
// Returns nil if no front matter was present.
func (d *Document) FrontMatter() map[string]any {
	return d.frontMatter
}

// Render writes the rendered HTML of the parsed document to w.
func (d *Document) Render(w io.Writer) error {
	return d.RenderContext(context.Background(), w)
}

// RenderContext writes the rendered HTML of the parsed document to w with context.
func (d *Document) RenderContext(ctx context.Context, w io.Writer) error {
	return d.md.renderChildren(ctx, w, d.doc, d.src)
}
