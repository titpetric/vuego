package formatter_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego/diff"
	"github.com/titpetric/vuego/formatter"
)

const example = `<!DOCTYPE html>
<html lang="en">
  <head>
    <link rel="stylesheet" href="/assets/css/styles.css">
    <template include="partials/basecoat.vuego"></template>
    <template include="partials/hljs.vuego"></template>
    <!-- comment -->
    <script src="https://unpkg.com/htmx.org@2.0.4"></script>
  </head>
  <body>
    <template v-html="content"></template>
    <script src="https://unpkg.com/lucide@latest"></script>
    <script>
lucide.createIcons();
    </script>
  </body>
</html>`

func TestNewFormatter(t *testing.T) {
	f := formatter.NewFormatter()
	require.NotNil(t, f)
}

func TestFormatter_Format_FrontmatterPreserved(t *testing.T) {
	f := formatter.NewFormatter()

	formatted, err := f.Format(example)
	require.NoError(t, err)

	require.Contains(t, formatted, "></template>", "Tags without content need to be one-lined (template tag).")
	require.Contains(t, formatted, "></script>", "Tags without content need to be one-lined (script tag).")

	diff.EqualHTML(t, []byte(example), []byte(formatted), []byte(example), nil)
}

func TestFormatter_Format_SimpleHTML(t *testing.T) {
	f := formatter.NewFormatter()

	content := `<div><h1>Title</h1></div>`

	formatted, err := f.Format(content)
	require.NoError(t, err)
	require.NotEmpty(t, formatted)
}

func TestFormatter_Format_VuegoTemplate(t *testing.T) {
	f := formatter.NewFormatter()

	content := `<template v-for="item in items">
  <div>{{ item.name }}</div>
</template>`

	formatted, err := f.Format(content)
	require.NoError(t, err)
	require.NotEmpty(t, formatted)
}

func TestFormatter_Format_InsertsFinalNewline(t *testing.T) {
	f := formatter.NewFormatter()

	content := `<div>Content</div>`

	formatted, err := f.Format(content)
	require.NoError(t, err)
	require.True(t, len(formatted) > 0 && formatted[len(formatted)-1] == '\n',
		"formatted content should end with newline")
}

func TestFormatter_Format_NoFinalNewline(t *testing.T) {
	opts := formatter.FormatterOptions{
		IndentWidth: 2,
		InsertFinal: false,
	}
	f := formatter.NewFormatterWithOptions(opts)

	content := `<div>Content</div>`

	formatted, err := f.Format(content)
	require.NoError(t, err)
	require.False(t, len(formatted) > 0 && formatted[len(formatted)-1] == '\n',
		"formatted content should not end with newline when InsertFinal is false")
}

func TestFormatString(t *testing.T) {
	content := `<div><span>Test</span></div>`

	formatted, err := formatter.FormatString(content)
	require.NoError(t, err)
	require.NotEmpty(t, formatted)
}

func TestDefaultFormatterOptions(t *testing.T) {
	opts := formatter.DefaultFormatterOptions()
	require.Equal(t, 2, opts.IndentWidth)
	require.True(t, opts.InsertFinal)
}

func TestIndentString(t *testing.T) {
	text := "line1\nline2\nline3"
	indented := formatter.IndentString(text, 2, 2)

	expected := "    line1\n    line2\n    line3"

	require.Equal(t, expected, indented)
}

func TestIndentString_WithEmptyLines(t *testing.T) {
	text := "line1\n\nline3"
	indented := formatter.IndentString(text, 1, 2)
	require.Contains(t, indented, "line1")
	require.Contains(t, indented, "line3")
}

func TestFormatter_Format_InlineNoExtraSpaces(t *testing.T) {
	f := formatter.NewFormatter()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no space around child element",
			input:    `<button> <i data-lucide="sun"></i> </button>`,
			expected: "<button><i data-lucide=\"sun\"></i></button>\n",
		},
		{
			name:     "no space around text content",
			input:    `<span> hello </span>`,
			expected: "<span>hello</span>\n",
		},
		{
			name:     "preserve space between siblings",
			input:    `<p> <span>a</span> <span>b</span> </p>`,
			expected: "<p><span>a</span> <span>b</span></p>\n",
		},
		{
			name:     "nested inline elements trimmed",
			input:    `<a href="/"> <i class="icon"></i> <span>Label</span> </a>`,
			expected: "<a href=\"/\"><i class=\"icon\"></i> <span>Label</span></a>\n",
		},
		{
			name:     "text with inline element no extra space",
			input:    `<p>Hello <strong>world</strong>!</p>`,
			expected: "<p>Hello <strong>world</strong>!</p>\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			formatted, err := f.Format(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, formatted)
		})
	}
}

// Tests for formatter handling of partial fragments
func TestFormatter_Format_PartialFragment_SingleElement(t *testing.T) {
	f := formatter.NewFormatter()

	content := `<div class="card">
  <h1>Title</h1>
  <p>Content</p>
</div>`

	formatted, err := f.Format(content)
	require.NoError(t, err)
	// Should not wrap in extra divs
	require.True(t, len(formatted) > 0)
	require.Contains(t, formatted, "card")
	// Should not have outer wrapper
	require.NotContains(t, formatted, "<div>\n  <div class=\"card\">")
}

func TestFormatter_Format_PartialFragment_MultipleElements(t *testing.T) {
	f := formatter.NewFormatter()

	content := `<li v-for="item in items" :key="item.id">
  <span>{{ item.name }}</span>
</li>
<li v-for="item in items" :key="item.id">
  <span>{{ item.name }}</span>
</li>`

	formatted, err := f.Format(content)
	require.NoError(t, err)
	require.Contains(t, formatted, "<li")
	require.Contains(t, formatted, "v-for")
	lines := 0
	for i := 0; i < len(formatted); i++ {
		if formatted[i] == '\n' {
			lines++
		}
	}
	require.Greater(t, lines, 1, "should format multiple list items")
}

// Tests for formatter handling of full documents
func TestFormatter_Format_FullDocument_PreservesDoctype(t *testing.T) {
	f := formatter.NewFormatter()

	content := `<!DOCTYPE html>
<html lang="en">
<head>
  <title>Test</title>
</head>
<body>
  <h1>Hello</h1>
</body>
</html>`

	formatted, err := f.Format(content)
	require.NoError(t, err)
	require.True(t, len(formatted) > 0)
	require.Equal(t, "<!DOCTYPE html>", formatted[:15], "should preserve DOCTYPE")
	require.Contains(t, formatted, "<html")
}

func TestFormatter_Format_FullDocument_WithHead(t *testing.T) {
	f := formatter.NewFormatter()

	content := `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>Test Page</title>
</head>
<body>
<h1>Content</h1>
</body>
</html>`

	formatted, err := f.Format(content)
	require.NoError(t, err)
	require.Contains(t, formatted, "<meta")
	require.Contains(t, formatted, "<title")
	require.Contains(t, formatted, "<!DOCTYPE")
}

// Tests for multi-line attribute rendering (lines > 100 chars with multiple attributes)
func TestFormatter_Format_MultiLineAttributes(t *testing.T) {
	f := formatter.NewFormatter()

	// Long line with multiple attributes triggers multi-line rendering
	content := `<div class="very-long-class-name-that-makes-line-exceed-limit" data-attribute="another-very-long-value" id="some-id" style="color: red;">content</div>`

	formatted, err := f.Format(content)
	require.NoError(t, err)
	require.Contains(t, formatted, "class=")
	require.Contains(t, formatted, "data-attribute=")
	require.Contains(t, formatted, "content")
}

// Tests for pre element formatting (renderPreContent)
func TestFormatter_Format_PreElement(t *testing.T) {
	f := formatter.NewFormatter()

	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:     "pre with text",
			input:    `<pre>  code here  </pre>`,
			contains: []string{"<pre>", "</pre>", "code here"},
		},
		{
			name:     "pre with nested code element",
			input:    `<pre><code>function() { return 1; }</code></pre>`,
			contains: []string{"<pre>", "<code>", "</code>", "</pre>"},
		},
		{
			name:     "pre with HTML entities",
			input:    `<pre>a < b && c > d</pre>`,
			contains: []string{"&lt;", "&gt;", "&amp;"},
		},
		{
			name:     "pre with comment",
			input:    `<pre>text<!-- comment -->more</pre>`,
			contains: []string{"<!--", "comment", "-->"},
		},
		{
			name:     "pre with void element",
			input:    `<pre>line1<br>line2</pre>`,
			contains: []string{"<br>", "line1", "line2"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			formatted, err := f.Format(tc.input)
			require.NoError(t, err)
			for _, s := range tc.contains {
				require.Contains(t, formatted, s)
			}
		})
	}
}

// Tests for escapeText with template expressions
func TestFormatter_Format_EscapeText(t *testing.T) {
	f := formatter.NewFormatter()

	tests := []struct {
		name     string
		input    string
		contains []string
		excludes []string
	}{
		{
			name:     "escape ampersand",
			input:    `<p>A & B</p>`,
			contains: []string{"&amp;"},
		},
		{
			name:     "escape less than",
			input:    `<p>A < B</p>`,
			contains: []string{"&lt;"},
		},
		{
			name:     "escape greater than",
			input:    `<p>A > B</p>`,
			contains: []string{"&gt;"},
		},
		{
			name:     "preserve template expression",
			input:    `<p>{{ a < b }}</p>`,
			contains: []string{"{{ a < b }}"},
			excludes: []string{"&lt;"},
		},
		{
			name:     "template expression with comparison",
			input:    `<span>{{ count > 0 && visible }}</span>`,
			contains: []string{"{{ count > 0 && visible }}"},
		},
		{
			name:     "mixed text and template",
			input:    `<p>Result: {{ x < y }} is less & more</p>`,
			contains: []string{"{{ x < y }}", "&amp;"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			formatted, err := f.Format(tc.input)
			require.NoError(t, err)
			for _, s := range tc.contains {
				require.Contains(t, formatted, s)
			}
			for _, s := range tc.excludes {
				require.NotContains(t, formatted, s)
			}
		})
	}
}

// Tests for phrasing containers (p, h1-h6, td, th, etc.)
func TestFormatter_Format_PhrasingContainers(t *testing.T) {
	f := formatter.NewFormatter()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "paragraph with inline elements",
			input:    `<p><strong>bold</strong> and <em>italic</em></p>`,
			expected: "<p><strong>bold</strong> and <em>italic</em></p>\n",
		},
		{
			name:     "h1 with inline",
			input:    `<h1><span>Title</span></h1>`,
			expected: "<h1><span>Title</span></h1>\n",
		},
		{
			name:     "h2 with text and inline",
			input:    `<h2>Heading <code>code</code></h2>`,
			expected: "<h2>Heading <code>code</code></h2>\n",
		},
		{
			name:     "table cell with inline",
			input:    `<td><a href="#">link</a></td>`,
			expected: "<td><a href=\"#\">link</a></td>\n",
		},
		{
			name:     "th with inline",
			input:    `<th><strong>Header</strong></th>`,
			expected: "<th><strong>Header</strong></th>\n",
		},
		{
			name:     "dt with inline",
			input:    `<dt><code>term</code></dt>`,
			expected: "<dt><code>term</code></dt>\n",
		},
		{
			name:     "dd with inline",
			input:    `<dd><em>definition</em></dd>`,
			expected: "<dd><em>definition</em></dd>\n",
		},
		{
			name:     "figcaption with inline",
			input:    `<figcaption><em>caption</em></figcaption>`,
			expected: "<figcaption><em>caption</em></figcaption>\n",
		},
		{
			name:     "summary with inline",
			input:    `<summary><strong>Details</strong></summary>`,
			expected: "<summary><strong>Details</strong></summary>\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			formatted, err := f.Format(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, formatted)
		})
	}
}

// Tests for block elements with element children (should not inline)
func TestFormatter_Format_BlockElements(t *testing.T) {
	f := formatter.NewFormatter()

	tests := []struct {
		name        string
		input       string
		shouldBreak bool
	}{
		{
			name:        "div with child div",
			input:       `<div><div>inner</div></div>`,
			shouldBreak: true,
		},
		{
			name:        "nav with child ul",
			input:       `<nav><ul><li>item</li></ul></nav>`,
			shouldBreak: true,
		},
		{
			name:        "section with child p",
			input:       `<section><p>text</p></section>`,
			shouldBreak: true,
		},
		{
			name:        "article with nested structure",
			input:       `<article><header>title</header></article>`,
			shouldBreak: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			formatted, err := f.Format(tc.input)
			require.NoError(t, err)
			if tc.shouldBreak {
				// Block elements with element children should have newlines
				lines := 0
				for _, c := range formatted {
					if c == '\n' {
						lines++
					}
				}
				require.Greater(t, lines, 1, "block with element children should break across lines")
			}
		})
	}
}

// Tests for inline elements with nested inline elements
func TestFormatter_Format_NestedInlineElements(t *testing.T) {
	f := formatter.NewFormatter()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "span with nested em and strong",
			input:    `<span><em><strong>text</strong></em></span>`,
			expected: "<span><em><strong>text</strong></em></span>\n",
		},
		{
			name:     "a with nested i",
			input:    `<a href="#"><i class="icon"></i></a>`,
			expected: "<a href=\"#\"><i class=\"icon\"></i></a>\n",
		},
		{
			name:     "label with nested span",
			input:    `<label><span>Name:</span></label>`,
			expected: "<label><span>Name:</span></label>\n",
		},
		{
			name:     "code with nested var",
			input:    `<code><var>x</var></code>`,
			expected: "<code><var>x</var></code>\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			formatted, err := f.Format(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, formatted)
		})
	}
}

// Tests for allChildrenAreInline with block child in inline parent
func TestFormatter_Format_InlineWithBlockChild(t *testing.T) {
	f := formatter.NewFormatter()

	// An inline element containing a block element should break
	content := `<span><div>block in inline</div></span>`

	formatted, err := f.Format(content)
	require.NoError(t, err)
	// Should break across multiple lines because div is block
	lines := 0
	for _, c := range formatted {
		if c == '\n' {
			lines++
		}
	}
	require.Greater(t, lines, 1, "inline with block child should break")
}

// Tests for normalizeInlineText edge cases
func TestFormatter_Format_NormalizeInlineText(t *testing.T) {
	f := formatter.NewFormatter()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "collapse multiple spaces",
			input:    `<p>hello    world</p>`,
			expected: "<p>hello world</p>\n",
		},
		{
			name:     "preserve space between inline elements",
			input:    `<p><span>a</span>   <span>b</span></p>`,
			expected: "<p><span>a</span> <span>b</span></p>\n",
		},
		{
			name:     "newlines become spaces",
			input:    "<p>line1\nline2</p>",
			expected: "<p>line1 line2</p>\n",
		},
		{
			name:     "tabs become spaces",
			input:    "<p>col1\tcol2</p>",
			expected: "<p>col1 col2</p>\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			formatted, err := f.Format(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, formatted)
		})
	}
}

// Tests for script/style element formatting
func TestFormatter_Format_ScriptStyleElements(t *testing.T) {
	f := formatter.NewFormatter()

	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:     "script with content",
			input:    "<script>\nconsole.log('hello');\n</script>",
			contains: []string{"<script>", "</script>", "console.log"},
		},
		{
			name:     "script empty",
			input:    `<script src="app.js"></script>`,
			contains: []string{"<script", "></script>"},
		},
		{
			name:     "style with content",
			input:    "<style>\n.foo { color: red; }\n</style>",
			contains: []string{"<style>", "</style>", ".foo"},
		},
		{
			name:     "style empty",
			input:    `<style></style>`,
			contains: []string{"<style></style>"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			formatted, err := f.Format(tc.input)
			require.NoError(t, err)
			for _, s := range tc.contains {
				require.Contains(t, formatted, s)
			}
		})
	}
}

// Tests for trimRawContent edge cases
func TestFormatter_Format_TrimRawContent(t *testing.T) {
	f := formatter.NewFormatter()

	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:     "script with leading blank lines",
			input:    "<script>\n\n\ncode();\n</script>",
			contains: []string{"code();"},
		},
		{
			name:     "script with trailing blank lines",
			input:    "<script>\ncode();\n\n\n</script>",
			contains: []string{"code();"},
		},
		{
			name:     "script all whitespace",
			input:    "<script>   \n\n   </script>",
			contains: []string{"<script></script>"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			formatted, err := f.Format(tc.input)
			require.NoError(t, err)
			for _, s := range tc.contains {
				require.Contains(t, formatted, s)
			}
		})
	}
}

// Tests for void elements
func TestFormatter_Format_VoidElements(t *testing.T) {
	f := formatter.NewFormatter()

	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:     "br element",
			input:    `<p>line1<br>line2</p>`,
			contains: []string{"<br>"},
		},
		{
			name:     "hr element",
			input:    `<div><hr></div>`,
			contains: []string{"<hr>"},
		},
		{
			name:     "img element",
			input:    `<img src="test.png" alt="test">`,
			contains: []string{"<img", "src=", "alt="},
		},
		{
			name:     "input element",
			input:    `<input type="text" name="field">`,
			contains: []string{"<input", "type=", "name="},
		},
		{
			name:     "meta element",
			input:    `<meta charset="UTF-8">`,
			contains: []string{"<meta", "charset="},
		},
		{
			name:     "link element",
			input:    `<link rel="stylesheet" href="style.css">`,
			contains: []string{"<link", "rel=", "href="},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			formatted, err := f.Format(tc.input)
			require.NoError(t, err)
			for _, s := range tc.contains {
				require.Contains(t, formatted, s)
			}
		})
	}
}

// Tests for comment nodes
func TestFormatter_Format_Comments(t *testing.T) {
	f := formatter.NewFormatter()

	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:     "simple comment",
			input:    `<!-- this is a comment -->`,
			contains: []string{"<!--", "this is a comment", "-->"},
		},
		{
			name:     "comment in element",
			input:    `<div><!-- comment --></div>`,
			contains: []string{"<!--", "comment", "-->"},
		},
		{
			name:     "inline comment in paragraph",
			input:    `<p>text<!-- inline -->more</p>`,
			contains: []string{"<!--", "inline", "-->"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			formatted, err := f.Format(tc.input)
			require.NoError(t, err)
			for _, s := range tc.contains {
				require.Contains(t, formatted, s)
			}
		})
	}
}

// Tests that table-scoped fragments are parsed with the correct context
// so that elements like <td>, <th>, <tr> are not stripped by the HTML5 parser.
func TestFormatter_Format_TableFragments(t *testing.T) {
	f := formatter.NewFormatter()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "td fragment",
			input:    `<td>cell</td>`,
			expected: "<td>cell</td>\n",
		},
		{
			name:     "th fragment",
			input:    `<th>header</th>`,
			expected: "<th>header</th>\n",
		},
		{
			name:     "tr fragment with cells",
			input:    `<tr><td>a</td><td>b</td></tr>`,
			expected: "<tr>\n  <td>a</td>\n  <td>b</td>\n</tr>\n",
		},
		{
			name:     "tbody fragment",
			input:    `<tbody><tr><td>x</td></tr></tbody>`,
			expected: "<tbody>\n  <tr>\n    <td>x</td>\n  </tr>\n</tbody>\n",
		},
		{
			name:     "thead fragment",
			input:    `<thead><tr><th>col</th></tr></thead>`,
			expected: "<thead>\n  <tr>\n    <th>col</th>\n  </tr>\n</thead>\n",
		},
		{
			name:     "caption fragment",
			input:    `<caption>Table Title</caption>`,
			expected: "<caption>Table Title</caption>\n",
		},
		{
			name:     "td with attributes",
			input:    `<td class="active">value</td>`,
			expected: "<td class=\"active\">value</td>\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			formatted, err := f.Format(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, formatted)
		})
	}
}

// Tests for full document without DOCTYPE
func TestFormatter_Format_FullDocumentWithoutDoctype(t *testing.T) {
	f := formatter.NewFormatter()

	content := `<html lang="en">
<head><title>Test</title></head>
<body><p>Content</p></body>
</html>`

	formatted, err := f.Format(content)
	require.NoError(t, err)
	require.Contains(t, formatted, "<html")
	require.Contains(t, formatted, "<head>")
	require.Contains(t, formatted, "<body>")
}
