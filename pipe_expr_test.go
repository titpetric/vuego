package vuego_test

import (
	"testing"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/tests/assert"
)

func TestPipeExpr(t *testing.T) {
	funcs := vuego.FuncMap{
		"add":     func(a, b int) int { return a + b },
		"concat":  func(a, b string) string { return a + b },
		"process": func(s string) string { return s },
		"double":  func(n int) int { return n * 2 },
		"half":    func(n float64) float64 { return n / 2 },
		"check":   func(b bool) string { return map[bool]string{true: "yes", false: "no"}[b] },
	}

	tests := []struct {
		name     string
		source   string
		data     map[string]any
		contains string
	}{
		// A filter chain starts from a resolved value.
		{name: "one filter", source: `<p>{{ name | upper }}</p>`, data: map[string]any{"name": "ada"}, contains: "ADA"},
		{name: "two filters", source: `<p>{{ name | lower | title }}</p>`, data: map[string]any{"name": "ADA LOVELACE"}, contains: "Ada Lovelace"},
		{name: "a filter with a double-quoted argument", source: `<p>{{ value | default("fallback") }}</p>`, data: map[string]any{"value": ""}, contains: "fallback"},
		{name: "a filter with a single-quoted argument", source: `<p>{{ items | default('empty') }}</p>`, data: map[string]any{"items": nil}, contains: "empty"},
		{name: "a missing variable reaches the first filter as nil", source: `<p>{{ missing | default("fallback") }}</p>`, data: map[string]any{}, contains: "fallback"},
		{name: "a chain over a missing variable", source: `<p>{{ undefined | default("empty") | upper }}</p>`, data: map[string]any{}, contains: "EMPTY"},
		{name: "a filter on a path expression", source: `<p>{{ user.name | upper }}</p>`, data: map[string]any{"user": map[string]any{"name": "bob"}}, contains: "BOB"},

		// A call with no initial value is a segment of its own.
		{name: "a built-in call", source: `<p>{{ len(items) }}</p>`, data: map[string]any{"items": []string{"a", "b", "c"}}, contains: "3"},
		{name: "a call with a variable argument", source: `<p>{{ upper(first) }}</p>`, data: map[string]any{"first": "hello"}, contains: "HELLO"},
		{name: "a call with two literal arguments", source: `<p>{{ add(5, 10) }}</p>`, data: map[string]any{}, contains: "15"},
		{name: "a call with two variable arguments", source: `<p>{{ add(x, y) }}</p>`, data: map[string]any{"x": 7, "y": 3}, contains: "10"},

		// Argument literals keep their type.
		{name: "a double-quoted argument", source: `<p>{{ process("hello world") }}</p>`, data: map[string]any{}, contains: "hello world"},
		{name: "a single-quoted argument", source: `<p>{{ process('hello world') }}</p>`, data: map[string]any{}, contains: "hello world"},
		{name: "an integer argument", source: `<p>{{ double(42) }}</p>`, data: map[string]any{}, contains: "84"},
		{name: "a float argument", source: `<p>{{ half(3.14) }}</p>`, data: map[string]any{}, contains: "1.57"},
		{name: "a true argument", source: `<p>{{ check(true) }}</p>`, data: map[string]any{}, contains: "yes"},
		{name: "a false argument", source: `<p>{{ check(false) }}</p>`, data: map[string]any{}, contains: "no"},
		// An unresolvable bare word is passed through as a string.
		{name: "an unresolved bare word is a string", source: `<p>{{ process(unquoted) }}</p>`, data: map[string]any{}, contains: "unquoted"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := renderFuncs(t, map[string]string{"test.vuego": test.source}, test.data, funcs)
			assert.NoError(t, err)
			assert.Contains(t, got, test.contains)
		})
	}

	errors := []struct {
		name   string
		source string
		data   map[string]any
		want   string
	}{
		{
			name:   "an unknown filter in a chain",
			source: `<p>{{ name | upper | badfilter }}</p>`,
			data:   map[string]any{"name": "test"},
			want:   "in test.vuego: in expression '{{ name | upper | badfilter }}': function 'badfilter' not found",
		},
		{
			name:   "an unknown direct call",
			source: `<p>{{ unknownFunc(10) }}</p>`,
			data:   map[string]any{},
			want:   "in test.vuego: in expression '{{ unknownFunc(10) }}': function 'unknownFunc' not found",
		},
	}

	for _, test := range errors {
		t.Run(test.name, func(t *testing.T) {
			_, err := renderFuncs(t, map[string]string{"test.vuego": test.source}, test.data, funcs)
			assert.Error(t, err)
			assert.Equal(t, test.want, err.Error())
		})
	}
}

func TestPipeExpr_Expression(t *testing.T) {
	// An expression holding an operator is evaluated whole rather than split on
	// the pipe character.
	tests := []struct {
		name     string
		source   string
		data     map[string]any
		contains string
	}{
		{name: "arithmetic", source: `<p>{{ a + b }}</p>`, data: map[string]any{"a": 1, "b": 2}, contains: "3"},
		{name: "comparison", source: `<p>{{ count > 5 }}</p>`, data: map[string]any{"count": 10}, contains: "true"},
		{name: "logical and", source: `<p>{{ ok && ready }}</p>`, data: map[string]any{"ok": true, "ready": true}, contains: "true"},
		{name: "logical or", source: `<p>{{ ok || ready }}</p>`, data: map[string]any{"ok": false, "ready": true}, contains: "true"},
		{name: "a call inside an expression", source: `<p>{{ len(items) > 2 }}</p>`, data: map[string]any{"items": []int{1, 2, 3}}, contains: "true"},
		{name: "a ternary", source: `<p>{{ ok ? "yes" : "no" }}</p>`, data: map[string]any{"ok": true}, contains: "yes"},
		{name: "a double pipe is an operator, not two filters", source: `<p>{{ a || b }}</p>`, data: map[string]any{"a": false, "b": "used"}, contains: "used"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Contains(t, renderPage(t, test.source, test.data), test.contains)
		})
	}

	t.Run("an unparsable expression is reported", func(t *testing.T) {
		_, err := renderFuncs(t, map[string]string{"test.vuego": `<p>{{ a == }}</p>`}, map[string]any{"a": 1}, nil)
		assert.Error(t, err)
	})
}

func TestPipeExpr_InAttributes(t *testing.T) {
	tests := []struct {
		name   string
		source string
		data   map[string]any
		want   string
	}{
		{
			name:   "interpolation in a static attribute",
			source: `<img src="{{ image }}" />`,
			data:   map[string]any{"image": "cat.png"},
			want:   `<img src="cat.png" />`,
		},
		{
			name:   "a filter in a static attribute",
			source: `<img alt="{{ caption | upper }}" />`,
			data:   map[string]any{"caption": "a cat"},
			want:   `<img alt="A CAT" />`,
		},
		{
			name:   "a call in a static attribute",
			source: `<div data-count="{{ len(items) }}"></div>`,
			data:   map[string]any{"items": []int{1, 2}},
			want:   `<div data-count="2"></div>`,
		},
		{
			name:   "two interpolations in one attribute",
			source: `<div data-pair="{{ a }}-{{ b }}"></div>`,
			data:   map[string]any{"a": "x", "b": "y"},
			want:   `<div data-pair="x-y"></div>`,
		},
		{
			name:   "a path expression in an attribute",
			source: `<a href="{{ post.url }}">link</a>`,
			data:   map[string]any{"post": map[string]any{"url": "/p/1"}},
			want:   `<a href="/p/1">link</a>`,
		},
		{
			name:   "a default filter for a missing variable",
			source: `<div title="{{ missing | default('none') }}"></div>`,
			data:   map[string]any{},
			want:   `<div title="none"></div>`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertRenderHTML(t, test.source, test.data, test.want)
		})
	}
}
