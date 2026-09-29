package vuego_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/tests/assert"
)

func TestCallFunc(t *testing.T) {
	funcs := vuego.FuncMap{
		"greet":    func(s string) string { return "Hello, " + s },
		"needsOne": func(x int) int { return x },
		"toString": func(s string) string { return "value: " + s },
		"noop":     func(string) {},
	}

	tests := []struct {
		name     string
		source   string
		data     map[string]any
		contains string
	}{
		{name: "a variable argument", source: `<p>{{ greet(name) }}</p>`, data: map[string]any{"name": "Alice"}, contains: "Hello, Alice"},
		{name: "a literal argument", source: `<p>{{ greet("World") }}</p>`, data: map[string]any{}, contains: "Hello, World"},
		// A nil argument becomes the zero value of the parameter type.
		{name: "a nil argument is the zero value", source: `<p>{{ greet(nothing) }}</p>`, data: map[string]any{"nothing": nil}, contains: "Hello,"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := renderFuncs(t, map[string]string{"test.vuego": test.source}, test.data, funcs)
			assert.NoError(t, err)
			assert.Contains(t, got, test.contains)
		})
	}

	t.Run("a func returning nothing is still called", func(t *testing.T) {
		called := false
		got, err := renderFuncs(t,
			map[string]string{"test.vuego": `<p>{{ name | record }}</p>`},
			map[string]any{"name": "test"},
			vuego.FuncMap{"record": func(string) { called = true }},
		)
		assert.NoError(t, err)
		assert.True(t, called)
		assert.NotContains(t, got, "test")
	})

	t.Run("a func returning a value and a nil error yields the value", func(t *testing.T) {
		got, err := renderFuncs(t,
			map[string]string{"test.vuego": `<p>{{ divide(10, 2) }}</p>`},
			map[string]any{},
			vuego.FuncMap{"divide": func(a, b int) (int, error) {
				if b == 0 {
					return 0, fmt.Errorf("division by zero")
				}
				return a / b, nil
			}},
		)
		assert.NoError(t, err)
		assert.Contains(t, got, "5")
	})
}

func TestCallFunc_ArgumentCount(t *testing.T) {
	tests := []struct {
		name   string
		source string
		funcs  vuego.FuncMap
		want   string
	}{
		{
			name:   "too many arguments",
			source: `<p>{{ needsOne(1, 2) }}</p>`,
			funcs:  vuego.FuncMap{"needsOne": func(x int) int { return x }},
			want:   "in test.vuego: in expression '{{ needsOne(1, 2) }}': needsOne(): function expects 1 arguments, got 2",
		},
		{
			name:   "too few for a variadic func",
			source: `<p>{{ needsTwo(1) }}</p>`,
			funcs:  vuego.FuncMap{"needsTwo": func(a, b int, rest ...int) int { return a + b }},
			want:   "in test.vuego: in expression '{{ needsTwo(1) }}': needsTwo(): function expects at least 2 arguments, got 1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := renderFuncs(t, map[string]string{"test.vuego": test.source}, map[string]any{}, test.funcs)
			assert.Error(t, err)
			assert.Equal(t, test.want, err.Error())
		})
	}

	t.Run("a variadic func takes any number of arguments", func(t *testing.T) {
		funcs := vuego.FuncMap{"join": func(parts ...string) string { return strings.Join(parts, ",") }}

		got, err := renderFuncs(t, map[string]string{"test.vuego": `<p>{{ join("a", "b", "c") }}</p>`}, map[string]any{}, funcs)
		assert.NoError(t, err)
		assert.Contains(t, got, "a,b,c")
	})
}

func TestConvertValue(t *testing.T) {
	// A template argument arrives as a string, an int, a float or a bool, and
	// is converted to the parameter type where a conversion exists.
	tests := []struct {
		name     string
		source   string
		funcs    vuego.FuncMap
		contains string
	}{
		{name: "string to int", source: `<p>{{ convert("21") }}</p>`, funcs: vuego.FuncMap{"convert": func(n int) int { return n * 2 }}, contains: "42"},
		{name: "string to int8", source: `<p>{{ convert("42") }}</p>`, funcs: vuego.FuncMap{"convert": func(n int8) int8 { return n }}, contains: "42"},
		{name: "string to int16", source: `<p>{{ convert("1000") }}</p>`, funcs: vuego.FuncMap{"convert": func(n int16) int16 { return n }}, contains: "1000"},
		{name: "string to int32", source: `<p>{{ convert("100000") }}</p>`, funcs: vuego.FuncMap{"convert": func(n int32) int32 { return n }}, contains: "100000"},
		{name: "string to int64", source: `<p>{{ convert("9223372036854775806") }}</p>`, funcs: vuego.FuncMap{"convert": func(n int64) int64 { return n }}, contains: "9223372036854775806"},
		{name: "string to uint", source: `<p>{{ convert("100") }}</p>`, funcs: vuego.FuncMap{"convert": func(n uint) uint { return n * 2 }}, contains: "200"},
		{name: "string to uint8", source: `<p>{{ convert("200") }}</p>`, funcs: vuego.FuncMap{"convert": func(n uint8) uint8 { return n }}, contains: "200"},
		{name: "string to uint16", source: `<p>{{ convert("50000") }}</p>`, funcs: vuego.FuncMap{"convert": func(n uint16) uint16 { return n }}, contains: "50000"},
		{name: "string to uint32", source: `<p>{{ convert("4000000000") }}</p>`, funcs: vuego.FuncMap{"convert": func(n uint32) uint32 { return n }}, contains: "4000000000"},
		{name: "string to uint64", source: `<p>{{ convert("9223372036854775806") }}</p>`, funcs: vuego.FuncMap{"convert": func(n uint64) uint64 { return n }}, contains: "9223372036854775806"},
		{name: "string to float32", source: `<p>{{ convert("3.14") }}</p>`, funcs: vuego.FuncMap{"convert": func(f float32) float32 { return f }}, contains: "3.14"},
		{name: "string to float64", source: `<p>{{ convert("10.5") }}</p>`, funcs: vuego.FuncMap{"convert": func(f float64) float64 { return f / 2 }}, contains: "5.25"},
		{name: "string to bool", source: `<p>{{ convert("true") }}</p>`, funcs: vuego.FuncMap{"convert": func(b bool) string { return fmt.Sprintf("%v", b) }}, contains: "true"},

		{name: "int to string", source: `<p>{{ convert(42) }}</p>`, funcs: vuego.FuncMap{"convert": func(s string) string { return "value: " + s }}, contains: "value: 42"},
		{name: "float to string", source: `<p>{{ convert(3.14) }}</p>`, funcs: vuego.FuncMap{"convert": func(s string) string { return "value: " + s }}, contains: "value: 3.14"},
		{name: "bool to string", source: `<p>{{ convert(true) }}</p>`, funcs: vuego.FuncMap{"convert": func(s string) string { return "bool: " + s }}, contains: "bool: true"},
		{name: "int to float", source: `<p>{{ convert(10) }}</p>`, funcs: vuego.FuncMap{"convert": func(f float64) float64 { return f / 2 }}, contains: "5"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := renderFuncs(t, map[string]string{"test.vuego": test.source}, map[string]any{}, test.funcs)
			assert.NoError(t, err)
			assert.Contains(t, got, test.contains)
		})
	}

	failures := []struct {
		name   string
		source string
		funcs  vuego.FuncMap
		want   string
	}{
		{
			name:   "a word is not an int",
			source: `<p>{{ convert("not-a-number") }}</p>`,
			funcs:  vuego.FuncMap{"convert": func(n int) int { return n }},
			want:   `in test.vuego: in expression '{{ convert("not-a-number") }}': convert(): cannot convert argument 0 from string to int`,
		},
		{
			name:   "a word is not a float",
			source: `<p>{{ convert("not-a-float") }}</p>`,
			funcs:  vuego.FuncMap{"convert": func(f float64) float64 { return f }},
			want:   `in test.vuego: in expression '{{ convert("not-a-float") }}': convert(): cannot convert argument 0 from string to float64`,
		},
		{
			name:   "a word is not a bool",
			source: `<p>{{ convert("maybe") }}</p>`,
			funcs:  vuego.FuncMap{"convert": func(b bool) bool { return b }},
			want:   `in test.vuego: in expression '{{ convert("maybe") }}': convert(): cannot convert argument 0 from string to bool`,
		},
		{
			name:   "a slice is not an int",
			source: `<p>{{ items | double }}</p>`,
			funcs:  vuego.FuncMap{"double": func(n int) int { return n * 2 }},
			want:   "in test.vuego: in expression '{{ items | double }}': double(): cannot convert argument 0 from []string to int",
		},
		{
			name:   "a map is not an int",
			source: `<p>{{ double(obj) }}</p>`,
			funcs:  vuego.FuncMap{"double": func(n int) int { return n * 2 }},
			want:   "in test.vuego: in expression '{{ double(obj) }}': double(): cannot convert argument 0 from map[string]interface {} to int",
		},
	}

	data := map[string]any{
		"items": []string{"a", "b", "c"},
		"obj":   map[string]any{"nested": "value"},
	}

	for _, test := range failures {
		t.Run(test.name, func(t *testing.T) {
			_, err := renderFuncs(t, map[string]string{"test.vuego": test.source}, data, test.funcs)
			assert.Error(t, err)
			assert.Equal(t, test.want, err.Error())
		})
	}
}

func TestCallFunc_Error(t *testing.T) {
	t.Run("an error from the func stops the render", func(t *testing.T) {
		_, err := renderFuncs(t,
			map[string]string{"test.vuego": `<p>{{ divide(10, 0) }}</p>`},
			map[string]any{},
			vuego.FuncMap{"divide": func(a, b int) (int, error) {
				if b == 0 {
					return 0, fmt.Errorf("division by zero")
				}
				return a / b, nil
			}},
		)
		assert.Error(t, err)
		assert.Equal(t, "in test.vuego: in expression '{{ divide(10, 0) }}': divide(): division by zero", err.Error())
	})

	t.Run("an error names the template, the expression and the func", func(t *testing.T) {
		failing := vuego.FuncMap{"loadSVG": func(path string) (string, error) {
			return "", fmt.Errorf("failed to load SVG file '%s'", path)
		}}

		_, err := renderFuncs(t, map[string]string{"test.vuego": `<div>{{ loadSVG("missing.svg") }}</div>`}, map[string]any{}, failing)
		assert.Error(t, err)
		assert.Equal(t, `in test.vuego: in expression '{{ loadSVG("missing.svg") }}': loadSVG(): failed to load SVG file 'missing.svg'`, err.Error())
	})

	t.Run("an error through a pipe carries the piped value", func(t *testing.T) {
		failing := vuego.FuncMap{"loadSVG": func(path string) (string, error) {
			return "", fmt.Errorf("failed to load SVG file '%s'", path)
		}}

		_, err := renderFuncs(t,
			map[string]string{"test.vuego": `<div>{{ path | loadSVG }}</div>`},
			map[string]any{"path": "icons/check.svg"},
			failing,
		)
		assert.Error(t, err)
		assert.Equal(t, `in test.vuego: in expression '{{ path | loadSVG }}': loadSVG(): failed to load SVG file 'icons/check.svg'`, err.Error())
	})

	t.Run("a func returning a value and no error renders the value", func(t *testing.T) {
		got, err := renderFuncs(t,
			map[string]string{"test.vuego": `<div>{{ loadSVG("icons/star.svg") }}</div>`},
			map[string]any{},
			vuego.FuncMap{"loadSVG": func(string) (string, error) { return "svg content", nil }},
		)
		assert.NoError(t, err)
		assert.Contains(t, got, "svg content")
	})
}

func TestCallFunc_ContextParameter(t *testing.T) {
	t.Run("a *VueContext first parameter is injected", func(t *testing.T) {
		got, err := renderFuncs(t,
			map[string]string{"test.vuego": `<p>{{ hasContext() }}</p>`},
			map[string]any{},
			vuego.FuncMap{"hasContext": func(ctx *vuego.VueContext) string {
				if ctx == nil {
					return "no-context"
				}
				return "has-context"
			}},
		)
		assert.NoError(t, err)
		assert.Contains(t, got, "has-context")
	})

	t.Run("a *VueContext parameter sits before the declared arguments", func(t *testing.T) {
		got, err := renderFuncs(t,
			map[string]string{"test.vuego": `<p>{{ greet("Alice") }}</p>`},
			map[string]any{},
			vuego.FuncMap{"greet": func(ctx *vuego.VueContext, name string) string {
				if ctx == nil {
					return "no context"
				}
				return "Hello, " + name
			}},
		)
		assert.NoError(t, err)
		assert.Contains(t, got, "Hello, Alice")
	})

	t.Run("the injected context names the template being rendered", func(t *testing.T) {
		got, err := renderFuncs(t,
			map[string]string{"test.vuego": `<p>{{ templateName() }}</p>`},
			map[string]any{},
			vuego.FuncMap{"templateName": func(ctx *vuego.VueContext) string { return ctx.FromFilename }},
		)
		assert.NoError(t, err)
		assert.Contains(t, got, "test.vuego")
	})

	t.Run("a context func may also return an error", func(t *testing.T) {
		got, err := renderFuncs(t,
			map[string]string{"test.vuego": `<p>{{ requireContext() }}</p>`},
			map[string]any{},
			vuego.FuncMap{"requireContext": func(ctx *vuego.VueContext) (string, error) {
				if ctx == nil {
					return "", fmt.Errorf("context is required")
				}
				return "success", nil
			}},
		)
		assert.NoError(t, err)
		assert.Contains(t, got, "success")
	})

	t.Run("a context.Context method bound in the data sees the render context", func(t *testing.T) {
		const source = `<div v-if="Show()">visible</div><div v-if="!Show()">hidden</div>`

		gate := &contextGate{}

		t.Run("value set", func(t *testing.T) {
			ctx := context.WithValue(t.Context(), showKey{}, true)

			got, err := renderCtx(t, ctx, source, map[string]any{"Show": gate.Show})
			assert.NoError(t, err)
			assert.Contains(t, got, "visible")
			assert.NotContains(t, got, "hidden")
		})

		t.Run("value unset", func(t *testing.T) {
			got, err := renderCtx(t, t.Context(), source, map[string]any{"Show": gate.Show})
			assert.NoError(t, err)
			assert.NotContains(t, got, "visible")
			assert.Contains(t, got, "hidden")
		})
	})
}

type showKey struct{}

// contextGate exposes a method whose first parameter is a context.Context, the
// shape a caller binds into the template data.
type contextGate struct{}

func (c *contextGate) Show(ctx context.Context) bool {
	v, _ := ctx.Value(showKey{}).(bool)
	return v
}
