package formatter_test

import (
	"testing"

	"github.com/titpetric/vuego/tests/assert"
)

func TestEscapeText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "ampersand", input: `<p>A &amp; B</p>`, want: "<p>A &amp; B</p>\n"},
		{name: "less than", input: `<p>A &lt; B</p>`, want: "<p>A &lt; B</p>\n"},
		{name: "greater than", input: `<p>A &gt; B</p>`, want: "<p>A &gt; B</p>\n"},
		{name: "bare less than is escaped", input: `<p>A < B</p>`, want: "<p>A &lt; B</p>\n"},
		// Template expressions are copied through, so operators inside {{ }}
		// keep the syntax the evaluator expects.
		{name: "expression with less than", input: `<p>{{ a < b }}</p>`, want: "<p>{{ a < b }}</p>\n"},
		{name: "expression with greater than and and", input: `<span>{{ count > 0 && visible }}</span>`, want: "<span>{{ count > 0 && visible }}</span>\n"},
		{name: "text around an expression is escaped", input: `<p>Result: {{ x < y }} is less & more</p>`, want: "<p>Result: {{ x < y }} is less &amp; more</p>\n"},
		{name: "two expressions", input: `<p>{{ a < b }} and {{ c > d }}</p>`, want: "<p>{{ a < b }} and {{ c > d }}</p>\n"},
		// An unterminated {{ has no matching }}, so it is escaped like any
		// other text.
		{name: "unterminated expression is escaped", input: `<p>{{ a < b</p>`, want: "<p>{{ a &lt; b</p>\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, format(t, test.input))
		})
	}
}

func TestNormalizeInlineText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "runs of spaces collapse", input: `<p>hello    world</p>`, want: "<p>hello world</p>\n"},
		{name: "space between inline elements is kept", input: `<p><span>a</span>   <span>b</span></p>`, want: "<p><span>a</span> <span>b</span></p>\n"},
		{name: "newlines become spaces", input: "<p>line1\nline2</p>", want: "<p>line1 line2</p>\n"},
		{name: "tabs become spaces", input: "<p>col1\tcol2</p>", want: "<p>col1 col2</p>\n"},
		{name: "mixed whitespace collapses", input: "<p>a \t\n b</p>", want: "<p>a b</p>\n"},
		{name: "leading and trailing whitespace is trimmed off the parent", input: "<p>\n  text\n</p>", want: "<p>text</p>\n"},
		{name: "whitespace-only between elements becomes one space", input: "<p><span>a</span>\n\n<span>b</span></p>", want: "<p><span>a</span> <span>b</span></p>\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, format(t, test.input))
		})
	}
}
