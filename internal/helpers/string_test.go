package helpers_test

import (
	"testing"

	"github.com/titpetric/vuego/internal/helpers"
	"github.com/titpetric/vuego/tests/assert"
)

func TestIsIdentifier(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{input: "name", want: true},
		{input: "_", want: true},
		{input: "_name", want: true},
		{input: "name123", want: true},
		{input: "_name_123", want: true},
		{input: "my_name_123", want: true},
		{input: "MyName", want: true},
		{input: "a", want: true},
		{input: "", want: false},
		{input: "1name", want: false},
		{input: "-name", want: false},
		{input: "my name", want: false},
		{input: "my-name", want: false},
		{input: "my.name", want: false},
		{input: "my@name", want: false},
		{input: "my|name", want: false},
		{input: "name$var", want: false},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			assert.Equal(t, test.want, helpers.IsIdentifier(test.input))
		})
	}
}

func TestIsIdentifierChar(t *testing.T) {
	tests := []struct {
		name  string
		char  rune
		first bool
		want  bool
	}{
		{name: "lowercase letter", char: 'a', want: true},
		{name: "uppercase letter", char: 'A', want: true},
		{name: "digit", char: '0', want: true},
		{name: "underscore", char: '_', want: true},
		{name: "lowercase letter first", char: 'a', first: true, want: true},
		{name: "underscore first", char: '_', first: true, want: true},
		{name: "digit first", char: '0', first: true, want: false},
		{name: "space", char: ' ', want: false},
		{name: "dash", char: '-', want: false},
		{name: "dollar", char: '$', want: false},
		{name: "at", char: '@', want: false},
		{name: "bang", char: '!', want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, helpers.IsIdentifierChar(test.char, test.first))
		})
	}
}

func TestNeedsHTMLEscape(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "ampersand", input: "a & b", want: true},
		{name: "less than", input: "a < b", want: true},
		{name: "greater than", input: "a > b", want: true},
		{name: "double quote", input: `say "hi"`, want: true},
		{name: "single quote", input: "it's", want: true},
		{name: "tag", input: "<div></div>", want: true},
		{name: "entity", input: "&amp;", want: true},
		{name: "empty string", input: "", want: false},
		{name: "plain text", input: "hello world", want: false},
		{name: "digits and punctuation", input: "1.2, 3-4; 5:6", want: false},
		{name: "multibyte text", input: "Ljubljana, Slovenija", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, helpers.NeedsHTMLEscape(test.input))
		})
	}
}

func TestFormatAttr(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		// Trim whitespace
		{input: "  value  ", want: "value"},
		{input: "\nvalue\n", want: "value"},
		{input: "\tvalue\t", want: "value"},

		// Replace newlines with spaces
		{input: "line1\nline2", want: "line1 line2"},
		{input: "line1\r\nline2", want: "line1 line2"},

		// Reduce multiple spaces to single space
		{input: "multiple  spaces", want: "multiple spaces"},
		{input: "multiple   spaces", want: "multiple spaces"},
		{input: "multiple\t\tspaces", want: "multiple spaces"},
		{input: "multiple  \n  spaces", want: "multiple spaces"},

		// Complex cases
		{input: "  line1\n  line2  ", want: "line1 line2"},
		{input: "\n\n  value  \n\n", want: "value"},
		{input: "a  b\nc\t\td", want: "a b c d"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			assert.Equal(t, test.want, helpers.FormatAttr(test.input))
		})
	}
}

func TestCamelToKebab(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		// PascalCase
		{input: "ButtonPrimary", want: "button-primary"},
		{input: "Button", want: "button"},
		{input: "B", want: "b"},

		// camelCase
		{input: "myVariable", want: "my-variable"},
		{input: "myVar", want: "my-var"},
		{input: "testAPIKey", want: "test-a-p-i-key"},

		// Already kebab-case
		{input: "button-primary", want: "button-primary"},
		{input: "my-variable", want: "my-variable"},

		// Edge cases
		{input: "", want: ""},
		{input: "a", want: "a"},
		{input: "A", want: "a"},
		{input: "ABC", want: "a-b-c"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			assert.Equal(t, test.want, helpers.CamelToKebab(test.input))
		})
	}
}
