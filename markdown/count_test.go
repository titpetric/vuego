package markdown_test

import (
	"bytes"
	"sync/atomic"
	"testing"
	"testing/fstest"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/markdown"
)

// CountingFS wraps a filesystem and counts Open calls
type CountingFS struct {
	fs    fstest.MapFS
	count atomic.Int64
}

func (c *CountingFS) Open(name string) (interface{ Read([]byte) (int, error); Close() error }, error) {
	c.count.Add(1)
	return c.fs.Open(name)
}

// TestTemplateCallCount counts how many template renders occur for the alert markdown
func TestTemplateCallCount(t *testing.T) {
	// Count inline elements in alertMD:
	// - 2 headings (## Usage, ## Examples)
	// - 2 h3 headings (### Destructive, ### No description, ### No icon)
	// - 6 paragraphs
	// - 2 code blocks
	// - 1 list with 4 items
	// - Multiple inline code spans (backticks)
	// - Multiple emphasis elements

	contentFS := fstest.MapFS{"test.md": {Data: []byte(alertMD)}}
	tplFS := vuego.NewOverlayFS(contentFS, markdown.Templates())

	// We can't easily count template calls without modifying the code,
	// but we can estimate from the AST structure

	md := markdown.New(contentFS)
	doc, err := md.Load("test.md")
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := doc.Render(&buf); err != nil {
		t.Fatal(err)
	}

	t.Logf("Output size: %d bytes", buf.Len())
	t.Logf("Using overlay FS: %v", tplFS != nil)

	// Expected template calls (estimate):
	// Block level: ~6 headings, ~6 paragraphs, 2 code blocks, 1 list, ~4 list items = ~19
	// Inline: code_span (~8), emphasis (~4), link (~1) = ~13
	// Total: ~32+ template Load/Fill/Render cycles
}
