package formatter_test

import (
	"strings"
	"testing"

	"github.com/titpetric/vuego/tests/assert"
)

func TestFormatNode_Inline(t *testing.T) {
	// Inline content sits flush against its parent tags.
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no space around child element",
			input: `<button> <i data-lucide="sun"></i> </button>`,
			want:  "<button><i data-lucide=\"sun\"></i></button>\n",
		},
		{
			name:  "no space around text content",
			input: `<span> hello </span>`,
			want:  "<span>hello</span>\n",
		},
		{
			name:  "space between siblings is kept",
			input: `<p> <span>a</span> <span>b</span> </p>`,
			want:  "<p><span>a</span> <span>b</span></p>\n",
		},
		{
			name:  "nested inline elements are trimmed",
			input: `<a href="/"> <i class="icon"></i> <span>Label</span> </a>`,
			want:  "<a href=\"/\"><i class=\"icon\"></i> <span>Label</span></a>\n",
		},
		{
			name:  "text beside an inline element gains no space",
			input: `<p>Hello <strong>world</strong>!</p>`,
			want:  "<p>Hello <strong>world</strong>!</p>\n",
		},
		{
			name:  "span with nested em and strong",
			input: `<span><em><strong>text</strong></em></span>`,
			want:  "<span><em><strong>text</strong></em></span>\n",
		},
		{
			name:  "a with nested i",
			input: `<a href="#"><i class="icon"></i></a>`,
			want:  "<a href=\"#\"><i class=\"icon\"></i></a>\n",
		},
		{
			name:  "label with nested span",
			input: `<label><span>Name:</span></label>`,
			want:  "<label><span>Name:</span></label>\n",
		},
		{
			name:  "code with nested var",
			input: `<code><var>x</var></code>`,
			want:  "<code><var>x</var></code>\n",
		},
		{
			name:  "an empty element stays on one line",
			input: `<div></div>`,
			want:  "<div></div>\n",
		},
		{
			name:  "a comment inside inline content is kept inline",
			input: `<p>text<!-- inline -->more</p>`,
			want:  "<p>text<!-- inline -->more</p>\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, format(t, test.input))
		})
	}
}

func TestFormatNode_PhrasingContainer(t *testing.T) {
	// A block element that holds phrasing content renders on one line when
	// every child is inline-compatible.
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "paragraph", input: `<p><strong>bold</strong> and <em>italic</em></p>`, want: "<p><strong>bold</strong> and <em>italic</em></p>\n"},
		{name: "h1", input: `<h1><span>Title</span></h1>`, want: "<h1><span>Title</span></h1>\n"},
		{name: "h2", input: `<h2>Heading <code>code</code></h2>`, want: "<h2>Heading <code>code</code></h2>\n"},
		{name: "h3", input: `<h3><em>Sub</em></h3>`, want: "<h3><em>Sub</em></h3>\n"},
		{name: "td", input: `<td><a href="#">link</a></td>`, want: "<td><a href=\"#\">link</a></td>\n"},
		{name: "th", input: `<th><strong>Header</strong></th>`, want: "<th><strong>Header</strong></th>\n"},
		{name: "dt", input: `<dt><code>term</code></dt>`, want: "<dt><code>term</code></dt>\n"},
		{name: "dd", input: `<dd><em>definition</em></dd>`, want: "<dd><em>definition</em></dd>\n"},
		{name: "figcaption", input: `<figcaption><em>caption</em></figcaption>`, want: "<figcaption><em>caption</em></figcaption>\n"},
		{name: "summary", input: `<summary><strong>Details</strong></summary>`, want: "<summary><strong>Details</strong></summary>\n"},
		{name: "legend", input: `<legend><span>Group</span></legend>`, want: "<legend><span>Group</span></legend>\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, format(t, test.input))
		})
	}
}

func TestFormatNode_Block(t *testing.T) {
	// A block element with element children always breaks across lines.
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "div with child div", input: `<div><div>inner</div></div>`, want: "<div>\n  <div>inner</div>\n</div>\n"},
		{name: "nav with child ul", input: `<nav><ul><li>item</li></ul></nav>`, want: "<nav>\n  <ul>\n    <li>item</li>\n  </ul>\n</nav>\n"},
		{name: "section with child p", input: `<section><p>text</p></section>`, want: "<section>\n  <p>text</p>\n</section>\n"},
		{name: "article with a header", input: `<article><header>title</header></article>`, want: "<article>\n  <header>title</header>\n</article>\n"},
		// An inline element holding a block child cannot be inlined either.
		{name: "span with a block child", input: `<span><div>block in inline</div></span>`, want: "<span>\n  <div>block in inline</div>\n</span>\n"},
		// A phrasing container holding a block child likewise breaks.
		{name: "p with a block child", input: `<p><div>x</div></p>`, want: "<p></p>\n<div>x</div>\n<p></p>\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, format(t, test.input))
		})
	}
}

func TestFormatNode_VoidElement(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{name: "br", input: `<p>line1<br>line2</p>`, contains: []string{"<br>"}},
		{name: "hr", input: `<div><hr></div>`, contains: []string{"<hr>"}},
		{name: "img", input: `<img src="test.png" alt="test">`, contains: []string{"<img", "src=", "alt="}},
		{name: "input", input: `<input type="text" name="field">`, contains: []string{"<input", "type=", "name="}},
		{name: "meta", input: `<meta charset="UTF-8">`, contains: []string{"<meta", "charset="}},
		{name: "link", input: `<link rel="stylesheet" href="style.css">`, contains: []string{"<link", "rel=", "href="}},
		{name: "source", input: `<video><source src="a.mp4"></video>`, contains: []string{"<source"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			formatted := format(t, test.input)
			for _, want := range test.contains {
				assert.Contains(t, formatted, want)
			}
			// A void element never gets a closing tag.
			assert.NotContains(t, formatted, "</"+test.name+">")
		})
	}
}

func TestFormatNode_Comment(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "top level", input: `<!-- this is a comment -->`, want: "<!-- this is a comment -->\n"},
		{name: "only child of an element", input: `<div><!-- comment --></div>`, want: "<div><!-- comment --></div>\n"},
		{name: "beside block children", input: `<div><!-- c --><p>a</p></div>`, want: "<div>\n  <!-- c -->\n  <p>a</p>\n</div>\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, format(t, test.input))
		})
	}
}

func TestFormatNode_Pre(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "text keeps its whitespace", input: `<pre>  code here  </pre>`, want: "<pre>  code here  </pre>\n"},
		{name: "nested code element", input: `<pre><code>function() { return 1; }</code></pre>`, want: "<pre><code>function() { return 1; }</code></pre>\n"},
		{name: "entities are escaped", input: `<pre>a &lt; b &amp;&amp; c &gt; d</pre>`, want: "<pre>a &lt; b &amp;&amp; c &gt; d</pre>\n"},
		{name: "comment is kept", input: `<pre>text<!-- comment -->more</pre>`, want: "<pre>text<!-- comment -->more</pre>\n"},
		{name: "void element gets no closing tag", input: `<pre>line1<br>line2</pre>`, want: "<pre>line1<br>line2</pre>\n"},
		{name: "newlines are kept", input: "<pre>one\n  two\n</pre>", want: "<pre>one\n  two\n</pre>\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, format(t, test.input))
		})
	}
}

func TestFormatNode_RawTextElement(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "script with content", input: "<script>\nconsole.log('hello');\n</script>", want: "<script>\nconsole.log('hello');\n</script>\n"},
		{name: "empty script with a src", input: `<script src="app.js"></script>`, want: "<script src=\"app.js\"></script>\n"},
		{name: "style with content", input: "<style>\n.foo { color: red; }\n</style>", want: "<style>\n.foo { color: red; }\n</style>\n"},
		{name: "empty style", input: `<style></style>`, want: "<style></style>\n"},
		// Blank lines around the content are dropped, internal indentation is not.
		{name: "leading blank lines are dropped", input: "<script>\n\n\ncode();\n</script>", want: "<script>\ncode();\n</script>\n"},
		{name: "trailing blank lines are dropped", input: "<script>\ncode();\n\n\n</script>", want: "<script>\ncode();\n</script>\n"},
		{name: "whitespace-only content empties the element", input: "<script>   \n\n   </script>", want: "<script></script>\n"},
		{name: "internal indentation is kept", input: "<script>\nif (x) {\n  y();\n}\n</script>", want: "<script>\nif (x) {\n  y();\n}\n</script>\n"},
		// The content is not HTML-escaped.
		{name: "operators are not escaped", input: "<script>\nif (a < b && c > d) {}\n</script>", want: "<script>\nif (a < b && c > d) {}\n</script>\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, format(t, test.input))
		})
	}
}

func TestFormatNode_Indentation(t *testing.T) {
	t.Run("each nesting level adds two spaces", func(t *testing.T) {
		formatted := format(t, "<div><div><div><p>deep</p></div></div></div>")
		assert.Equal(t, "<div>\n  <div>\n    <div>\n      <p>deep</p>\n    </div>\n  </div>\n</div>\n", formatted)
	})

	t.Run("whitespace between block elements is dropped", func(t *testing.T) {
		formatted := format(t, "<div>\n\n   <p>a</p>\n\n   <p>b</p>\n\n</div>")
		assert.Equal(t, "<div>\n  <p>a</p>\n  <p>b</p>\n</div>\n", formatted)
	})

	t.Run("attribute values are collapsed to one line", func(t *testing.T) {
		formatted := format(t, "<div class=\"a\n   b\n   c\">x</div>")
		assert.Equal(t, 1, strings.Count(formatted, "\n"))
		assert.Contains(t, formatted, `class="a b c"`)
	})
}
