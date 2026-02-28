package markdown_test

import (
	"bytes"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"

	"github.com/titpetric/vuego/markdown"
)

func TestTemplates(t *testing.T) {
	tplFS := markdown.Templates()
	assert.NotNil(t, tplFS)

	entries, err := fs.ReadDir(tplFS, "markdown")
	assert.NoError(t, err)
	assert.NotEmpty(t, entries)

	expected := []string{
		"autolink.vuego",
		"blockquote.vuego",
		"code_block.vuego",
		"code_span.vuego",
		"emphasis.vuego",
		"hard_break.vuego",
		"heading.vuego",
		"image.vuego",
		"link.vuego",
		"list.vuego",
		"list_item.vuego",
		"paragraph.vuego",
		"raw_html.vuego",
		"strikethrough.vuego",
		"table.vuego",
		"task_checkbox.vuego",
		"thematic_break.vuego",
	}

	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	assert.Equal(t, expected, names)
}

func TestNew(t *testing.T) {
	md := markdown.New(nil)
	assert.NotNil(t, md)
}

func render(t *testing.T, content string) string {
	t.Helper()
	contentFS := fstest.MapFS{"test.md": {Data: []byte(content)}}
	md := markdown.New(contentFS)

	doc, err := md.Load("test.md")
	assert.NoError(t, err)

	var buf bytes.Buffer
	err = doc.Render(&buf)
	assert.NoError(t, err)
	return buf.String()
}

func TestLoad(t *testing.T) {
	contentFS := fstest.MapFS{
		"hello.md": {Data: []byte("# Hello")},
	}
	md := markdown.New(contentFS)
	doc, err := md.Load("hello.md")
	assert.NoError(t, err)
	assert.NotNil(t, doc)
}

func TestLoad_NotFound(t *testing.T) {
	contentFS := fstest.MapFS{}
	md := markdown.New(contentFS)
	_, err := md.Load("missing.md")
	assert.Error(t, err)
}

func TestLoad_FrontMatter(t *testing.T) {
	contentFS := fstest.MapFS{
		"post.md": {Data: []byte("---\ntitle: Hello\nlayout: page\n---\n# Content")},
	}
	md := markdown.New(contentFS)
	doc, err := md.Load("post.md")
	assert.NoError(t, err)
	assert.Equal(t, "Hello", doc.FrontMatter()["title"])
	assert.Equal(t, "page", doc.FrontMatter()["layout"])

	var buf bytes.Buffer
	err = doc.Render(&buf)
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "<h1")
	assert.Contains(t, buf.String(), "Content")
	assert.NotContains(t, buf.String(), "title:")
}

func TestLoad_NoFrontMatter(t *testing.T) {
	contentFS := fstest.MapFS{
		"plain.md": {Data: []byte("# Just markdown")},
	}
	md := markdown.New(contentFS)
	doc, err := md.Load("plain.md")
	assert.NoError(t, err)
	assert.Nil(t, doc.FrontMatter())
}

func TestRender_Heading(t *testing.T) {
	got := render(t, "# Hello World")
	assert.Contains(t, got, "<h1")
	assert.Contains(t, got, `id="hello-world"`)
	assert.Contains(t, got, "Hello World")
}

func TestRender_HeadingLevels(t *testing.T) {
	tests := []struct {
		input string
		tag   string
		id    string
	}{
		{"# One", "h1", "one"},
		{"## Two", "h2", "two"},
		{"### Three", "h3", "three"},
		{"#### Four", "h4", "four"},
		{"##### Five", "h5", "five"},
		{"###### Six", "h6", "six"},
	}

	for _, tt := range tests {
		got := render(t, tt.input)
		assert.Contains(t, got, "<"+tt.tag, tt.input)
		assert.Contains(t, got, `id="`+tt.id+`"`, tt.input)
	}
}

func TestRender_Paragraph(t *testing.T) {
	got := render(t, "Hello world")
	assert.Contains(t, got, "<p>")
	assert.Contains(t, got, "Hello world")
	assert.Contains(t, got, "</p>")
}

func TestRender_CodeBlock(t *testing.T) {
	got := render(t, "```go\nfmt.Println(\"hi\")\n```")
	assert.Contains(t, got, "<pre>")
	assert.Contains(t, got, "<code")
	assert.Contains(t, got, "language-go")
	assert.Contains(t, got, "fmt.Println")
}

func TestRender_CodeBlockNoLanguage(t *testing.T) {
	got := render(t, "```\nplain code\n```")
	assert.Contains(t, got, "<pre>")
	assert.Contains(t, got, "<code>")
	assert.Contains(t, got, "plain code")
}

func TestRender_Blockquote(t *testing.T) {
	got := render(t, "> quoted text")
	assert.Contains(t, got, "<blockquote>")
	assert.Contains(t, got, "quoted text")
}

func TestRender_UnorderedList(t *testing.T) {
	got := render(t, "- alpha\n- beta")
	assert.Contains(t, got, "<ul>")
	assert.Contains(t, got, "<li>")
	assert.Contains(t, got, "alpha")
	assert.Contains(t, got, "beta")
}

func TestRender_OrderedList(t *testing.T) {
	got := render(t, "1. first\n2. second")
	assert.Contains(t, got, "<ol")
	assert.Contains(t, got, "<li>")
	assert.Contains(t, got, "first")
	assert.Contains(t, got, "second")
}

func TestRender_ThematicBreak(t *testing.T) {
	got := render(t, "---")
	assert.Contains(t, got, "<hr>")
}

func TestRender_Emphasis(t *testing.T) {
	got := render(t, "this is *italic* and **bold**")
	assert.Contains(t, got, "<em>")
	assert.Contains(t, got, "italic")
	assert.Contains(t, got, "<strong>")
	assert.Contains(t, got, "bold")
}

func TestRender_Link(t *testing.T) {
	got := render(t, "[click](https://example.com)")
	assert.Contains(t, got, `href="https://example.com"`)
	assert.Contains(t, got, "click")
}

func TestRender_Image(t *testing.T) {
	got := render(t, "![alt text](image.png)")
	assert.Contains(t, got, `src="image.png"`)
	assert.Contains(t, got, `alt="alt text"`)
}

func TestRender_InlineCode(t *testing.T) {
	got := render(t, "use `fmt.Println`")
	assert.Contains(t, got, "<code>")
	assert.Contains(t, got, "fmt.Println")
}

func TestRender_Strikethrough(t *testing.T) {
	got := render(t, "~~deleted~~")
	assert.Contains(t, got, "<del>")
	assert.Contains(t, got, "deleted")
}

func TestRender_Table(t *testing.T) {
	input := "| Name | Age |\n| --- | --- |\n| Alice | 30 |\n| Bob | 25 |"
	got := render(t, input)
	assert.Contains(t, got, "<table>")
	assert.Contains(t, got, "<thead>")
	assert.Contains(t, got, "<th>")
	assert.Contains(t, got, "Name")
	assert.Contains(t, got, "<tbody>")
	assert.Contains(t, got, "<td>")
	assert.Contains(t, got, "Alice")
	assert.Contains(t, got, "Bob")
}

func TestPostProcess(t *testing.T) {
	contentFS := fstest.MapFS{"test.md": {Data: []byte("hello world")}}
	md := markdown.New(contentFS)
	md.PostProcess("paragraph", func(src string) string {
		return strings.ReplaceAll(src, "hello", "goodbye")
	})

	doc, err := md.Load("test.md")
	assert.NoError(t, err)

	var buf bytes.Buffer
	err = doc.Render(&buf)
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "goodbye world")
	assert.NotContains(t, buf.String(), "hello")
}

func TestPostProcess_Heading(t *testing.T) {
	contentFS := fstest.MapFS{"test.md": {Data: []byte("# Title")}}
	md := markdown.New(contentFS)
	md.PostProcess("heading", func(src string) string {
		return strings.ReplaceAll(src, "<h1", `<h1 class="title"`)
	})

	doc, err := md.Load("test.md")
	assert.NoError(t, err)

	var buf bytes.Buffer
	err = doc.Render(&buf)
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), `class="title"`)
}
