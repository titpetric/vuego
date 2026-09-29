package markdown_test

import (
	"strings"
	"testing"

	"github.com/titpetric/vuego/tests/assert"
)

func TestRenderHeading(t *testing.T) {
	t.Run("renders the tag, the id and the content", func(t *testing.T) {
		got := render(t, "# Hello World")
		assert.Contains(t, got, "<h1")
		assert.Contains(t, got, `id="hello-world"`)
		assert.Contains(t, got, "Hello World")
	})

	levels := []struct {
		input string
		tag   string
		id    string
	}{
		{input: "# One", tag: "h1", id: "one"},
		{input: "## Two", tag: "h2", id: "two"},
		{input: "### Three", tag: "h3", id: "three"},
		{input: "#### Four", tag: "h4", id: "four"},
		{input: "##### Five", tag: "h5", id: "five"},
		{input: "###### Six", tag: "h6", id: "six"},
	}

	for _, level := range levels {
		t.Run(level.tag, func(t *testing.T) {
			got := render(t, level.input)
			assert.Contains(t, got, "<"+level.tag)
			assert.Contains(t, got, `id="`+level.id+`"`)
		})
	}

	ids := []struct {
		name  string
		input string
		want  string
	}{
		{name: "punctuation is dropped", input: "# Hello, World!", want: "hello-world"},
		{name: "case is folded", input: "# MiXeD CaSe", want: "mixed-case"},
		{name: "runs of hyphens collapse", input: "# a -- b", want: "a-b"},
		{name: "inline markup is stripped", input: "# A `code` span", want: "a-code-span"},
		{name: "hyphens are kept", input: "# well-known name", want: "well-known-name"},
	}

	for _, id := range ids {
		t.Run(id.name, func(t *testing.T) {
			assert.Contains(t, render(t, id.input), `id="`+id.want+`"`)
		})
	}
}

func TestRenderParagraph(t *testing.T) {
	got := render(t, "Hello world")
	assert.Contains(t, got, "<p>")
	assert.Contains(t, got, "Hello world")
	assert.Contains(t, got, "</p>")
}

func TestRenderCodeBlock(t *testing.T) {
	t.Run("fenced block carries the language class", func(t *testing.T) {
		got := render(t, "```go\nfmt.Println(\"hi\")\n```")
		assert.Contains(t, got, "<pre")
		assert.Contains(t, got, "<code")
		assert.Contains(t, got, "language-go")
		assert.Contains(t, got, "fmt.Println")
	})

	t.Run("fenced block without a language", func(t *testing.T) {
		got := render(t, "```\nplain code\n```")
		assert.Contains(t, got, "<pre")
		assert.Contains(t, got, "<code")
		assert.Contains(t, got, "plain code")
		assert.NotContains(t, got, "language-")
	})

	t.Run("indented block", func(t *testing.T) {
		got := render(t, "text\n\n    indented code\n")
		assert.Contains(t, got, "<pre")
		assert.Contains(t, got, "indented code")
	})
}

func TestRenderBlockquote(t *testing.T) {
	t.Run("wraps the quoted text", func(t *testing.T) {
		got := render(t, "> quoted text")
		assert.Contains(t, got, "<blockquote>")
		assert.Contains(t, got, "quoted text")
	})

	t.Run("recurses into child blocks", func(t *testing.T) {
		got := render(t, "> # Heading\n>\n> paragraph")
		assert.Contains(t, got, "<blockquote>")
		assert.Contains(t, got, "<h1")
		assert.Contains(t, got, "<p>")
	})
}

func TestRenderList(t *testing.T) {
	t.Run("unordered", func(t *testing.T) {
		got := render(t, "- alpha\n- beta")
		assert.Contains(t, got, "<ul>")
		assert.Contains(t, got, "<li>")
		assert.Contains(t, got, "alpha")
		assert.Contains(t, got, "beta")
	})

	t.Run("ordered", func(t *testing.T) {
		got := render(t, "1. first\n2. second")
		assert.Contains(t, got, "<ol")
		assert.Contains(t, got, "<li>")
		assert.Contains(t, got, "first")
		assert.Contains(t, got, "second")
	})

	t.Run("ordered with a start other than one", func(t *testing.T) {
		assert.Contains(t, render(t, "3. third\n4. fourth"), `start="3"`)
	})

	t.Run("nested", func(t *testing.T) {
		got := render(t, "- outer\n  - inner")
		assert.Contains(t, got, "outer")
		assert.Contains(t, got, "inner")
		assert.Equal(t, 2, strings.Count(got, "<ul>"))
	})

	t.Run("task list checkbox", func(t *testing.T) {
		got := render(t, "- [x] done\n- [ ] todo")
		assert.Contains(t, got, `type="checkbox"`)
		assert.Contains(t, got, "checked")
	})
}

func TestRenderThematicBreak(t *testing.T) {
	assert.Contains(t, render(t, "text\n\n---\n\nmore"), "<hr>")
}

func TestRenderHTMLBlock(t *testing.T) {
	got := render(t, "<div class=\"raw\">\n  <span>kept</span>\n</div>")
	assert.Contains(t, got, `<div class="raw">`)
	assert.Contains(t, got, "<span>kept</span>")
}

func TestRenderTable(t *testing.T) {
	t.Run("header and body cells", func(t *testing.T) {
		got := render(t, "| Name | Age |\n| --- | --- |\n| Alice | 30 |\n| Bob | 25 |")
		assert.Contains(t, got, "<table")
		assert.Contains(t, got, "<thead>")
		assert.Contains(t, got, "<th>")
		assert.Contains(t, got, "Name")
		assert.Contains(t, got, "<tbody>")
		assert.Contains(t, got, "<td>")
		assert.Contains(t, got, "Alice")
		assert.Contains(t, got, "Bob")
	})

	alignments := []struct {
		name  string
		rule  string
		align string
	}{
		{name: "left", rule: ":---", align: "left"},
		{name: "center", rule: ":---:", align: "center"},
		{name: "right", rule: "---:", align: "right"},
	}

	for _, alignment := range alignments {
		t.Run("alignment "+alignment.name, func(t *testing.T) {
			got := render(t, "| Name |\n| "+alignment.rule+" |\n| Alice |")
			assert.Contains(t, got, alignment.align)
		})
	}

	t.Run("no alignment rule leaves the cell unaligned", func(t *testing.T) {
		got := render(t, "| Name |\n| --- |\n| Alice |")
		assert.NotContains(t, got, "left")
		assert.NotContains(t, got, "center")
		assert.NotContains(t, got, "right")
	})
}

func TestRenderFootnote(t *testing.T) {
	t.Run("links the reference to the definition and back", func(t *testing.T) {
		got := render(t, "This is a sentence with a footnote[^1].\n\n[^1]: This is the footnote content.")

		assert.Contains(t, got, `<sup id="fnref:1"`)
		assert.Contains(t, got, `href="#fn:1"`)
		assert.Contains(t, got, `<section class="footnotes"`)
		assert.Contains(t, got, `<li id="fn:1"`)
		assert.Contains(t, got, "This is the footnote content")
		assert.Contains(t, got, `href="#fnref:1"`)
	})

	t.Run("numbers several footnotes", func(t *testing.T) {
		got := render(t, "First[^1] and second[^2].\n\n[^1]: First note.\n[^2]: Second note with [link](https://example.com).")

		assert.Contains(t, got, `id="fnref:1"`)
		assert.Contains(t, got, `id="fnref:2"`)
		assert.Contains(t, got, `id="fn:1"`)
		assert.Contains(t, got, `id="fn:2"`)
		assert.Contains(t, got, "First note")
		assert.Contains(t, got, "Second note")
		assert.Contains(t, got, "example.com")
	})
}

func TestRenderEmphasis(t *testing.T) {
	got := render(t, "this is *italic* and **bold**")
	assert.Contains(t, got, "<em>")
	assert.Contains(t, got, "italic")
	assert.Contains(t, got, "<strong>")
	assert.Contains(t, got, "bold")
}

func TestRenderLink(t *testing.T) {
	t.Run("href and content", func(t *testing.T) {
		got := render(t, "[click](https://example.com)")
		assert.Contains(t, got, `href="https://example.com"`)
		assert.Contains(t, got, "click")
	})

	t.Run("title", func(t *testing.T) {
		assert.Contains(t, render(t, `[click](https://example.com "tooltip")`), `title="tooltip"`)
	})

	t.Run("inline markup in the label", func(t *testing.T) {
		got := render(t, "[*click*](https://example.com)")
		assert.Contains(t, got, "<em>")
		assert.Contains(t, got, "click")
	})
}

func TestRenderAutoLink(t *testing.T) {
	t.Run("url", func(t *testing.T) {
		got := render(t, "<https://example.com>")
		assert.Contains(t, got, `href="https://example.com"`)
	})

	t.Run("email gets a mailto scheme", func(t *testing.T) {
		assert.Contains(t, render(t, "<user@example.com>"), `href="mailto:user@example.com"`)
	})
}

func TestRenderImage(t *testing.T) {
	t.Run("src and alt", func(t *testing.T) {
		got := render(t, "![alt text](image.png)")
		assert.Contains(t, got, `src="image.png"`)
		assert.Contains(t, got, `alt="alt text"`)
	})

	t.Run("title", func(t *testing.T) {
		assert.Contains(t, render(t, `![alt](image.png "caption")`), `title="caption"`)
	})

	t.Run("alt text is taken as plain text", func(t *testing.T) {
		got := render(t, "![*emphasised* alt](image.png)")
		assert.Contains(t, got, `alt="emphasised alt"`)
	})
}

func TestRenderCodeSpan(t *testing.T) {
	got := render(t, "use `fmt.Println`")
	assert.Contains(t, got, "<code>")
	assert.Contains(t, got, "fmt.Println")
}

func TestRenderStrikethrough(t *testing.T) {
	got := render(t, "~~deleted~~")
	assert.Contains(t, got, "<del>")
	assert.Contains(t, got, "deleted")
}

func TestRenderRawHTML(t *testing.T) {
	got := render(t, "text with <span>inline html</span> in it")
	assert.Contains(t, got, "<span>")
	assert.Contains(t, got, "inline html")
}

func TestRenderLineBreaks(t *testing.T) {
	t.Run("a soft break stays a newline", func(t *testing.T) {
		got := render(t, "first\nsecond")
		assert.Contains(t, got, "first\nsecond")
		assert.NotContains(t, got, "<br")
	})

	t.Run("a hard break becomes a br tag", func(t *testing.T) {
		assert.Contains(t, render(t, "first  \nsecond"), "<br")
	})
}
