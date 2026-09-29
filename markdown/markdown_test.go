package markdown_test

import (
	"bytes"
	"context"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/titpetric/vuego/markdown"
	"github.com/titpetric/vuego/tests/assert"
)

// render parses content as a standalone document and returns the rendered HTML.
// It is shared by every test file in this package.
func render(t *testing.T, content string) string {
	t.Helper()

	contentFS := fstest.MapFS{"test.md": {Data: []byte(content)}}
	md := markdown.New(contentFS)

	doc, err := md.Load("test.md")
	assert.NoError(t, err)

	var buf bytes.Buffer
	assert.NoError(t, doc.Render(&buf))
	return buf.String()
}

func TestNew(t *testing.T) {
	t.Run("without a content filesystem", func(t *testing.T) {
		md := markdown.New(nil)
		assert.NotNil(t, md)

		// The embedded templates are still reachable.
		var buf bytes.Buffer
		assert.NoError(t, md.RenderBytes(&buf, []byte("# Hi")))
		assert.Contains(t, buf.String(), "<h1")
	})

	t.Run("content filesystem overrides an embedded template", func(t *testing.T) {
		contentFS := fstest.MapFS{
			"markdown/paragraph.vuego": {Data: []byte(`<div class="para">{{ content }}</div>`)},
		}
		md := markdown.New(contentFS)

		var buf bytes.Buffer
		assert.NoError(t, md.RenderBytes(&buf, []byte("hello")))
		assert.Contains(t, buf.String(), `class="para"`)
		assert.NotContains(t, buf.String(), "<p>")
	})
}

func TestNodeType(t *testing.T) {
	// The constants are the template base names, so a handler registered for
	// a node type lines up with the template it replaces.
	tests := []struct {
		nodeType markdown.NodeType
		want     string
	}{
		{nodeType: markdown.NodeParagraph, want: "paragraph"},
		{nodeType: markdown.NodeHeading, want: "heading"},
		{nodeType: markdown.NodeCodeBlock, want: "code_block"},
	}

	for _, test := range tests {
		t.Run(test.want, func(t *testing.T) {
			assert.Equal(t, test.want, string(test.nodeType))

			_, err := fs.Stat(markdown.Templates(), "markdown/"+test.want+".vuego")
			assert.NoError(t, err)
		})
	}
}

func TestPostProcessor(t *testing.T) {
	var fn markdown.PostProcessor = strings.ToUpper

	md := markdown.New(nil)
	md.PostProcess("paragraph", fn)

	var buf bytes.Buffer
	assert.NoError(t, md.RenderBytes(&buf, []byte("text")))
	assert.Contains(t, buf.String(), "TEXT")
}

func TestMarkdown_PostProcess(t *testing.T) {
	tests := []struct {
		name        string
		blockType   string
		source      string
		replace     [2]string
		wantContain string
		wantAbsent  string
	}{
		{
			name:        "paragraph",
			blockType:   "paragraph",
			source:      "hello world",
			replace:     [2]string{"hello", "goodbye"},
			wantContain: "goodbye world",
			wantAbsent:  "hello",
		},
		{
			name:        "heading",
			blockType:   "heading",
			source:      "# Title",
			replace:     [2]string{"<h1", `<h1 class="title"`},
			wantContain: `class="title"`,
		},
		{
			name:        "code block",
			blockType:   "code_block",
			source:      "```go\nx := 1\n```",
			replace:     [2]string{"<pre", `<pre data-copy="1"`},
			wantContain: `data-copy="1"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			md := markdown.New(nil)
			md.PostProcess(test.blockType, func(src string) string {
				return strings.ReplaceAll(src, test.replace[0], test.replace[1])
			})

			var buf bytes.Buffer
			assert.NoError(t, md.RenderBytes(&buf, []byte(test.source)))
			assert.Contains(t, buf.String(), test.wantContain)
			if test.wantAbsent != "" {
				assert.NotContains(t, buf.String(), test.wantAbsent)
			}
		})
	}

	t.Run("returns the receiver for chaining", func(t *testing.T) {
		md := markdown.New(nil)
		assert.Equal(t, md, md.PostProcess("paragraph", func(src string) string { return src }))
	})
}

func TestMarkdown_Register(t *testing.T) {
	handler := func(html string) markdown.Handler {
		return func(_ context.Context, _ *markdown.Node) (string, bool) {
			return html, true
		}
	}

	t.Run("a handler replaces the template output", func(t *testing.T) {
		md := markdown.New(nil)
		md.Register(handler("<h1>handled</h1>"), markdown.NodeHeading)

		var buf bytes.Buffer
		assert.NoError(t, md.RenderBytes(&buf, []byte("# Title")))
		assert.Equal(t, "<h1>handled</h1>", buf.String())
	})

	t.Run("a handler sees the node it was registered for", func(t *testing.T) {
		var seen markdown.Node
		md := markdown.New(nil)
		md.Register(func(_ context.Context, n *markdown.Node) (string, bool) {
			seen = *n
			return "", true
		}, markdown.NodeCodeBlock)

		var buf bytes.Buffer
		assert.NoError(t, md.RenderBytes(&buf, []byte("```go\nx := 1\n```")))
		assert.Equal(t, markdown.NodeCodeBlock, seen.Type)
		assert.Equal(t, "go", seen.Language)
		assert.Equal(t, "x := 1\n", seen.Raw)
	})

	t.Run("handlers run in registration order until one handles", func(t *testing.T) {
		var called []string
		md := markdown.New(nil)
		md.Register(func(_ context.Context, _ *markdown.Node) (string, bool) {
			called = append(called, "first")
			return "", false
		}, markdown.NodeParagraph)
		md.Register(func(_ context.Context, _ *markdown.Node) (string, bool) {
			called = append(called, "second")
			return "<div>second</div>", true
		}, markdown.NodeParagraph)
		md.Register(func(_ context.Context, _ *markdown.Node) (string, bool) {
			called = append(called, "third")
			return "<div>third</div>", true
		}, markdown.NodeParagraph)

		var buf bytes.Buffer
		assert.NoError(t, md.RenderBytes(&buf, []byte("text")))
		assert.Equal(t, []string{"first", "second"}, called)
		assert.Equal(t, "<div>second</div>", buf.String())
	})

	t.Run("a handler for another node type is not called", func(t *testing.T) {
		md := markdown.New(nil)
		md.Register(handler("<h1>handled</h1>"), markdown.NodeHeading)

		var buf bytes.Buffer
		assert.NoError(t, md.RenderBytes(&buf, []byte("text")))
		assert.Contains(t, buf.String(), "<p>")
	})

	t.Run("returns the receiver for chaining", func(t *testing.T) {
		md := markdown.New(nil)
		assert.Equal(t, md, md.Register(handler(""), markdown.NodeHeading))
	})
}

func TestMarkdown_Load(t *testing.T) {
	t.Run("parses a file", func(t *testing.T) {
		md := markdown.New(fstest.MapFS{"hello.md": {Data: []byte("# Hello")}})
		doc, err := md.Load("hello.md")
		assert.NoError(t, err)
		assert.NotNil(t, doc)
	})

	t.Run("reports a missing file", func(t *testing.T) {
		md := markdown.New(fstest.MapFS{})
		_, err := md.Load("missing.md")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "loading missing.md")
	})

	t.Run("reports a nil content filesystem", func(t *testing.T) {
		_, err := markdown.New(nil).Load("any.md")
		assert.Error(t, err)
	})

	t.Run("extracts front matter and strips it from the body", func(t *testing.T) {
		md := markdown.New(fstest.MapFS{
			"post.md": {Data: []byte("---\ntitle: Hello\nlayout: page\n---\n# Content")},
		})
		doc, err := md.Load("post.md")
		assert.NoError(t, err)
		assert.Equal(t, "Hello", doc.FrontMatter()["title"])
		assert.Equal(t, "page", doc.FrontMatter()["layout"])

		var buf bytes.Buffer
		assert.NoError(t, doc.Render(&buf))
		assert.Contains(t, buf.String(), "<h1")
		assert.Contains(t, buf.String(), "Content")
		assert.NotContains(t, buf.String(), "title:")
	})

	t.Run("a file without front matter reports none", func(t *testing.T) {
		md := markdown.New(fstest.MapFS{"plain.md": {Data: []byte("# Just markdown")}})
		doc, err := md.Load("plain.md")
		assert.NoError(t, err)
		assert.Nil(t, doc.FrontMatter())
	})

	t.Run("an unterminated front matter block is body text", func(t *testing.T) {
		md := markdown.New(fstest.MapFS{"bad.md": {Data: []byte("---\ntitle: Hello\n")}})
		doc, err := md.Load("bad.md")
		assert.NoError(t, err)
		assert.Nil(t, doc.FrontMatter())
	})

	t.Run("front matter that is not YAML is body text", func(t *testing.T) {
		md := markdown.New(fstest.MapFS{"bad.md": {Data: []byte("---\n\ttitle: [\n---\n# Body")}})
		doc, err := md.Load("bad.md")
		assert.NoError(t, err)
		assert.Nil(t, doc.FrontMatter())
	})
}

func TestMarkdown_RenderBytes(t *testing.T) {
	t.Run("renders without touching the filesystem", func(t *testing.T) {
		var buf bytes.Buffer
		assert.NoError(t, markdown.New(nil).RenderBytes(&buf, []byte("# Hello")))
		assert.Contains(t, buf.String(), "<h1")
		assert.Contains(t, buf.String(), "Hello")
	})

	t.Run("does not strip front matter", func(t *testing.T) {
		var buf bytes.Buffer
		assert.NoError(t, markdown.New(nil).RenderBytes(&buf, []byte("---\ntitle: Hello\n---\n# Body")))
		assert.Contains(t, buf.String(), "<hr>")
	})

	t.Run("empty source writes nothing", func(t *testing.T) {
		var buf bytes.Buffer
		assert.NoError(t, markdown.New(nil).RenderBytes(&buf, nil))
		assert.Equal(t, "", buf.String())
	})
}

func TestMarkdown_RenderBytesContext(t *testing.T) {
	t.Run("passes the context to handlers", func(t *testing.T) {
		type key struct{}

		var got any
		md := markdown.New(nil)
		md.Register(func(ctx context.Context, _ *markdown.Node) (string, bool) {
			got = ctx.Value(key{})
			return "", true
		}, markdown.NodeHeading)

		ctx := context.WithValue(context.Background(), key{}, "value")
		var buf bytes.Buffer
		assert.NoError(t, md.RenderBytesContext(ctx, &buf, []byte("# Title")))
		assert.Equal(t, "value", got)
	})

	t.Run("renders the same HTML as RenderBytes", func(t *testing.T) {
		src := []byte("# Title\n\ntext")

		var withCtx, withoutCtx bytes.Buffer
		assert.NoError(t, markdown.New(nil).RenderBytesContext(context.Background(), &withCtx, src))
		assert.NoError(t, markdown.New(nil).RenderBytes(&withoutCtx, src))
		assert.Equal(t, withoutCtx.String(), withCtx.String())
	})
}

func BenchmarkRenderBytes(b *testing.B) {
	md := markdown.New(nil)
	src := []byte(alertMD)

	for b.Loop() {
		var buf bytes.Buffer
		_ = md.RenderBytes(&buf, src)
	}
}

func BenchmarkRenderBytesSimple(b *testing.B) {
	md := markdown.New(nil)
	src := []byte(simpleMD)

	for b.Loop() {
		var buf bytes.Buffer
		_ = md.RenderBytes(&buf, src)
	}
}
