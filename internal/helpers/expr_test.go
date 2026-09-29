package helpers_test

import (
	"testing"

	"github.com/titpetric/vuego/internal/helpers"
	"github.com/titpetric/vuego/tests/assert"
)

func TestIsFunctionCall(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{input: "func()", want: true},
		{input: "func(arg)", want: true},
		{input: "func(arg1, arg2)", want: true},
		{input: "  func()", want: true},
		{input: "_func()", want: true},
		{input: "func123()", want: true},
		{input: "", want: false},
		{input: "name", want: false},
		{input: "()", want: false},
		{input: "my-func()", want: false},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			assert.Equal(t, test.want, helpers.IsFunctionCall(test.input))
		})
	}
}

func TestContainsPipe(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{input: "name | upper", want: true},
		{input: "| upper", want: true},
		{input: "name | upper | lower", want: true},
		// The check is a character scan, so a pipe inside a literal counts.
		{input: `"text | literal"`, want: true},
		{input: "", want: false},
		{input: "name", want: false},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			assert.Equal(t, test.want, helpers.ContainsPipe(test.input))
		})
	}
}

func TestIsComplexExpr(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{input: "a == b", want: true},
		{input: "a === b", want: true},
		{input: "a != b", want: true},
		{input: "a !== b", want: true},
		{input: "a <= b", want: true},
		{input: "a >= b", want: true},
		{input: "a && b", want: true},
		{input: "a || b", want: true},
		{input: "a==b", want: true},
		{input: "a + b", want: true},
		{input: "a - b", want: true},
		{input: "a * b", want: true},
		{input: "a / b", want: true},
		{input: "a % b", want: true},
		{input: "a < b", want: true},
		{input: "a > b", want: true},
		{input: "ok ? yes : no", want: true},
		{input: "", want: false},
		{input: "name", want: false},
		{input: "user.name", want: false},
		// Arithmetic operators need surrounding spaces, so a path or a
		// hyphenated name is not complex.
		{input: "foo/bar", want: false},
		{input: "foo-bar", want: false},
		{input: "a+b", want: false},
		// A ternary needs both halves.
		{input: "ok ? yes", want: false},
		{input: "a : b", want: false},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			assert.Equal(t, test.want, helpers.IsComplexExpr(test.input))
		})
	}
}

func TestNormalizeComparisonOperators(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "a === b", want: "a == b"},
		{input: "a !== b", want: "a != b"},
		{input: "a === b && c !== d", want: "a == b && c != d"},
		{input: "a==b", want: "a==b"},
		{input: "a != b", want: "a != b"},
		{input: "a <= b", want: "a <= b"},
		{input: "", want: ""},
		{input: "name", want: "name"},
		{input: "a === b ? 'y' : 'n'", want: "a == b ? 'y' : 'n'"},
		{input: "user.role === 'admin'", want: "user.role == 'admin'"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			assert.Equal(t, test.want, helpers.NormalizeComparisonOperators(test.input))
		})
	}
}
