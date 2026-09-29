package markdown_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing/fstest"

	"github.com/titpetric/vuego/markdown"
)

// The examples below are the runnable form of the package documentation in
// doc.go.

func Example() {
	contentFS := fstest.MapFS{
		"post.md": {Data: []byte("---\ntitle: Hello\n---\n# Heading\n\nA paragraph.")},
	}

	md := markdown.New(contentFS)

	doc, err := md.Load("post.md")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(doc.FrontMatter()["title"])
	if err := doc.Render(os.Stdout); err != nil {
		fmt.Println(err)
	}
	// Output:
	// Hello
	// <h1 id="heading">Heading</h1>
	// <p>A paragraph.</p>
}

func ExampleMarkdown_PostProcess() {
	md := markdown.New(nil)
	md.PostProcess("paragraph", func(src string) string {
		return strings.ReplaceAll(src, "foo", "bar")
	})

	if err := md.RenderBytes(os.Stdout, []byte("a foo walks in")); err != nil {
		fmt.Println(err)
	}
	// Output: <p>a bar walks in</p>
}

func ExampleMarkdown_Register() {
	md := markdown.New(nil)
	md.Register(func(_ context.Context, n *markdown.Node) (string, bool) {
		if n.Language != "mermaid" {
			return "", false
		}
		return `<div class="mermaid">` + n.Raw + "</div>", true
	}, markdown.NodeCodeBlock)

	if err := md.RenderBytes(os.Stdout, []byte("```mermaid\ngraph TD;\n```")); err != nil {
		fmt.Println(err)
	}
	// Output: <div class="mermaid">graph TD;
	// </div>
}

func ExampleNew_overlay() {
	// A .vuego file under markdown/ in the content filesystem replaces the
	// embedded default of the same name.
	contentFS := fstest.MapFS{
		"markdown/paragraph.vuego": {Data: []byte(`<div class="para">{{ content }}</div>`)},
	}

	if err := markdown.New(contentFS).RenderBytes(os.Stdout, []byte("text")); err != nil {
		fmt.Println(err)
	}
	// Output: <div class="para">text</div>
}
