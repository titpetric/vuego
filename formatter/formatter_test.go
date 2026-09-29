package formatter_test

import (
	"strings"
	"testing"

	"github.com/titpetric/vuego/formatter"
	"github.com/titpetric/vuego/tests/assert"
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

// format runs the default formatter over input. It is shared by every test
// file in this package.
func format(t *testing.T, input string) string {
	t.Helper()

	out, err := formatter.NewFormatter().Format(input)
	assert.NoError(t, err)
	return out
}

func TestNewFormatter(t *testing.T) {
	t.Run("carries the default options", func(t *testing.T) {
		out, err := formatter.NewFormatter().Format("<div><p>a</p></div>")
		assert.NoError(t, err)

		// IndentWidth 2 and InsertFinal true.
		assert.Equal(t, "<div>\n  <p>a</p>\n</div>\n", out)
	})
}

func TestNewFormatterWithOptions(t *testing.T) {
	tests := []struct {
		name string
		opts formatter.FormatterOptions
		want string
	}{
		{
			name: "default width and final newline",
			opts: formatter.FormatterOptions{IndentWidth: 2, InsertFinal: true},
			want: "<div>\n  <p>a</p>\n</div>\n",
		},
		{
			name: "four space indent",
			opts: formatter.FormatterOptions{IndentWidth: 4, InsertFinal: true},
			want: "<div>\n    <p>a</p>\n</div>\n",
		},
		{
			name: "no indent",
			opts: formatter.FormatterOptions{IndentWidth: 0, InsertFinal: true},
			want: "<div>\n<p>a</p>\n</div>\n",
		},
		{
			name: "no final newline",
			opts: formatter.FormatterOptions{IndentWidth: 2, InsertFinal: false},
			want: "<div>\n  <p>a</p>\n</div>",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out, err := formatter.NewFormatterWithOptions(test.opts).Format("<div><p>a</p></div>")
			assert.NoError(t, err)
			assert.Equal(t, test.want, out)
		})
	}
}

func TestDefaultFormatterOptions(t *testing.T) {
	opts := formatter.DefaultFormatterOptions()
	assert.Equal(t, 2, opts.IndentWidth)
	assert.True(t, opts.InsertFinal)
}

func TestFormatString(t *testing.T) {
	t.Run("formats with the default options", func(t *testing.T) {
		out, err := formatter.FormatString("<div><span>Test</span></div>")
		assert.NoError(t, err)
		assert.Equal(t, "<div>\n  <span>Test</span>\n</div>\n", out)
	})

	t.Run("matches NewFormatter", func(t *testing.T) {
		out, err := formatter.FormatString(example)
		assert.NoError(t, err)
		assert.Equal(t, format(t, example), out)
	})
}

func TestIndentString(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		level int
		width int
		want  string
	}{
		{
			name:  "two levels of two spaces",
			text:  "line1\nline2\nline3",
			level: 2,
			width: 2,
			want:  "    line1\n    line2\n    line3",
		},
		{
			name:  "blank lines stay blank",
			text:  "line1\n\nline3",
			level: 1,
			width: 2,
			want:  "  line1\n\n  line3",
		},
		{
			name:  "zero level leaves the text alone",
			text:  "line1\nline2",
			level: 0,
			width: 2,
			want:  "line1\nline2",
		},
		{
			name:  "zero width leaves the text alone",
			text:  "line1\nline2",
			level: 3,
			width: 0,
			want:  "line1\nline2",
		},
		{
			name:  "existing indentation is kept",
			text:  "a\n  b",
			level: 1,
			width: 2,
			want:  "  a\n    b",
		},
		{
			name:  "empty text",
			text:  "",
			level: 1,
			width: 2,
			want:  "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, formatter.IndentString(test.text, test.level, test.width))
		})
	}
}

func TestFormatter_Format(t *testing.T) {
	t.Run("keeps frontmatter and the DOM of the input", func(t *testing.T) {
		formatted := format(t, example)

		assert.Contains(t, formatted, "></template>", "Tags without content need to be one-lined (template tag).")
		assert.Contains(t, formatted, "></script>", "Tags without content need to be one-lined (script tag).")
		assert.EqualHTML(t, []byte(example), []byte(formatted), []byte(example), nil)
	})

	frontmatter := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "frontmatter is preserved verbatim",
			input: "---\ntitle: Test\n---\n<div>a</div>",
			want:  "---\ntitle: Test\n---\n<div>a</div>\n",
		},
		{
			name:  "an unterminated block is body text",
			input: "---\ntitle: Test\n<div>a</div>",
			want:  "---\ntitle: Test\n<div>a</div>\n",
		},
		{
			name:  "content that does not open with --- has no frontmatter",
			input: "<div>a</div>",
			want:  "<div>a</div>\n",
		},
	}

	for _, test := range frontmatter {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, format(t, test.input))
		})
	}

	t.Run("a vuego template keeps its directives", func(t *testing.T) {
		formatted := format(t, "<template v-for=\"item in items\">\n  <div>{{ item.name }}</div>\n</template>")
		assert.Contains(t, formatted, `v-for="item in items"`)
		assert.Contains(t, formatted, "{{ item.name }}")
	})

	t.Run("long attribute lists stay on the tag", func(t *testing.T) {
		formatted := format(t, `<div class="very-long-class-name-that-makes-line-exceed-limit" data-attribute="another-very-long-value" id="some-id" style="color: red;">content</div>`)
		assert.Contains(t, formatted, "class=")
		assert.Contains(t, formatted, "data-attribute=")
		assert.Contains(t, formatted, "content")
	})
}

func TestFormatter_Format_FullDocument(t *testing.T) {
	t.Run("preserves the doctype", func(t *testing.T) {
		formatted := format(t, "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n  <title>Test</title>\n</head>\n<body>\n  <h1>Hello</h1>\n</body>\n</html>")
		assert.True(t, strings.HasPrefix(formatted, "<!DOCTYPE html>\n"), formatted)
		assert.Contains(t, formatted, "<html")
	})

	t.Run("formats head and body", func(t *testing.T) {
		formatted := format(t, "<!DOCTYPE html>\n<html>\n<head>\n<meta charset=\"UTF-8\">\n<title>Test Page</title>\n</head>\n<body>\n<h1>Content</h1>\n</body>\n</html>")
		assert.Contains(t, formatted, "<meta")
		assert.Contains(t, formatted, "<title")
		assert.Contains(t, formatted, "<!DOCTYPE")
	})

	t.Run("without a doctype", func(t *testing.T) {
		formatted := format(t, "<html lang=\"en\">\n<head><title>Test</title></head>\n<body><p>Content</p></body>\n</html>")
		assert.NotContains(t, formatted, "<!DOCTYPE")
		assert.Contains(t, formatted, "<html")
		assert.Contains(t, formatted, "<head>")
		assert.Contains(t, formatted, "<body>")
	})
}

func TestFormatter_Format_Fragment(t *testing.T) {
	t.Run("a single element gets no html or body wrapper", func(t *testing.T) {
		formatted := format(t, "<div class=\"card\">\n  <h1>Title</h1>\n  <p>Content</p>\n</div>")
		assert.Equal(t, "<div class=\"card\">\n  <h1>Title</h1>\n  <p>Content</p>\n</div>\n", formatted)
	})

	t.Run("sibling elements are each formatted", func(t *testing.T) {
		formatted := format(t, "<li v-for=\"item in items\" :key=\"item.id\">\n  <span>{{ item.name }}</span>\n</li>\n<li v-for=\"item in items\" :key=\"item.id\">\n  <span>{{ item.name }}</span>\n</li>")
		assert.Equal(t, 2, strings.Count(formatted, "<li"))
		assert.Equal(t, 2, strings.Count(formatted, "v-for"))
	})

	// Table-scoped elements need an ancestor context node, or the HTML5
	// parser drops them.
	tables := []struct {
		name  string
		input string
		want  string
	}{
		{name: "td", input: `<td>cell</td>`, want: "<td>cell</td>\n"},
		{name: "th", input: `<th>header</th>`, want: "<th>header</th>\n"},
		{name: "tr with cells", input: `<tr><td>a</td><td>b</td></tr>`, want: "<tr>\n  <td>a</td>\n  <td>b</td>\n</tr>\n"},
		{name: "tbody", input: `<tbody><tr><td>x</td></tr></tbody>`, want: "<tbody>\n  <tr>\n    <td>x</td>\n  </tr>\n</tbody>\n"},
		{name: "thead", input: `<thead><tr><th>col</th></tr></thead>`, want: "<thead>\n  <tr>\n    <th>col</th>\n  </tr>\n</thead>\n"},
		{name: "tfoot", input: `<tfoot><tr><td>sum</td></tr></tfoot>`, want: "<tfoot>\n  <tr>\n    <td>sum</td>\n  </tr>\n</tfoot>\n"},
		{name: "caption", input: `<caption>Table Title</caption>`, want: "<caption>Table Title</caption>\n"},
		{name: "colgroup", input: `<colgroup><col></colgroup>`, want: "<colgroup>\n  <col>\n</colgroup>\n"},
		{name: "td with attributes", input: `<td class="active">value</td>`, want: "<td class=\"active\">value</td>\n"},
		// A tag whose name merely starts with a table tag name must not pick
		// up the table context.
		{name: "theme is not thead", input: `<theme>x</theme>`, want: "<theme>x</theme>\n"},
	}

	for _, test := range tables {
		t.Run("table scope "+test.name, func(t *testing.T) {
			assert.Equal(t, test.want, format(t, test.input))
		})
	}
}
