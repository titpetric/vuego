package vuego_test

import (
	"bytes"
	"fmt"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
	"github.com/titpetric/vuego"
)

// TestVue_EvalPipe - comprehensive tests for pipe evaluation with high cognitive complexity
func TestVue_EvalPipe(t *testing.T) {
	t.Run("direct function call without initial value", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ len(items) }}</p>`),
			},
		}

		vue := vuego.NewVue(fs)
		data := map[string]any{
			"items": []string{"a", "b", "c"},
		}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", data)
		require.NoError(t, err)
		require.Contains(t, buf.String(), "3")
	})

	t.Run("direct function call with multiple arguments", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ add(5, 10) }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"add": func(a, b int) int {
				return a + b
			},
		})

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", map[string]any{})
		require.NoError(t, err)
		require.Contains(t, buf.String(), "15")
	})

	t.Run("direct function call with variable arguments", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ add(x, y) }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"add": func(a, b int) int {
				return a + b
			},
		})

		data := map[string]any{
			"x": 7,
			"y": 3,
		}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", data)
		require.NoError(t, err)
		require.Contains(t, buf.String(), "10")
	})

	t.Run("direct function call missing from funcmap", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ unknownFunc(10) }}</p>`),
			},
		}

		vue := vuego.NewVue(fs)
		data := map[string]any{}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", data)
		require.Error(t, err)
		require.Contains(t, err.Error(), "function 'unknownFunc' not found")
	})

	t.Run("direct function call result used in expression", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ upper(first) }}</p>`),
			},
		}

		vue := vuego.NewVue(fs)
		data := map[string]any{
			"first": "hello",
		}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", data)
		require.NoError(t, err)
		require.Contains(t, buf.String(), "HELLO")
	})

	t.Run("pipe with nil initial value and default filter", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ missing | default("fallback") }}</p>`),
			},
		}

		vue := vuego.NewVue(fs)
		data := map[string]any{}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", data)
		require.NoError(t, err)
		require.Contains(t, buf.String(), "fallback")
	})

	t.Run("pipe with string quotes in arguments", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ items | default('empty') }}</p>`),
			},
		}

		vue := vuego.NewVue(fs)
		data := map[string]any{
			"items": nil,
		}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", data)
		require.NoError(t, err)
		require.Contains(t, buf.String(), "empty")
	})

	t.Run("chained filters on missing variable", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ undefined | default("empty") | upper }}</p>`),
			},
		}

		vue := vuego.NewVue(fs)
		data := map[string]any{}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", data)
		require.NoError(t, err)
		require.Contains(t, buf.String(), "EMPTY")
	})

	t.Run("filter in chain missing from funcmap", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ name | upper | badfilter }}</p>`),
			},
		}

		vue := vuego.NewVue(fs)
		data := map[string]any{
			"name": "test",
		}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", data)
		require.Error(t, err)
		require.Contains(t, err.Error(), "function 'badfilter' not found")
	})
}

// TestVue_CallFunc - comprehensive tests for function calling with type conversion
func TestVue_CallFunc(t *testing.T) {
	t.Run("function with parameter", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ greet(name) }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"greet": func(s string) string {
				return "Hello, " + s
			},
		})

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", map[string]any{"name": "Alice"})
		require.NoError(t, err)
		require.Contains(t, buf.String(), "Hello, Alice")
	})

	t.Run("wrong argument count", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ needs_one(1, 2) }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"needs_one": func(x int) int {
				return x
			},
		})

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", map[string]any{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "expects 1 arguments, got 2")
	})

	t.Run("string to int conversion in function call", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ double("21") }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"double": func(x int) int {
				return x * 2
			},
		})

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", map[string]any{})
		require.NoError(t, err)
		require.Contains(t, buf.String(), "42")
	})

	t.Run("string parameter to string function", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ greet("World") }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"greet": func(s string) string {
				return "Hello, " + s
			},
		})

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", map[string]any{})
		require.NoError(t, err)
		require.Contains(t, buf.String(), "Hello, World")
	})

	t.Run("float to string conversion", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ process(3.14) }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"process": func(s string) string {
				return "value: " + s
			},
		})

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", map[string]any{})
		require.NoError(t, err)
		require.Contains(t, buf.String(), "value: 3.14")
	})

	t.Run("bool to string conversion", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ process(true) }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"process": func(s string) string {
				return "bool: " + s
			},
		})

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", map[string]any{})
		require.NoError(t, err)
		require.Contains(t, buf.String(), "bool: true")
	})

	t.Run("nil argument handling", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ process(nil_var) }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"process": func(v string) string {
				if v == "" {
					return "empty"
				}
				return v
			},
		})

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", map[string]any{"nil_var": nil})
		require.NoError(t, err)
		require.Contains(t, buf.String(), "empty")
	})

	t.Run("unconvertible type causes error", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ process(complex_obj) }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"process": func(x int) int {
				return x
			},
		})

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", map[string]any{
			"complex_obj": map[string]any{"nested": "value"},
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "cannot convert")
	})

	t.Run("variadic function with multiple args", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ concat("a", "b", "c", "d") }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"concat": func(parts ...string) string {
				result := ""
				for _, p := range parts {
					result += p
				}
				return result
			},
		})

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", map[string]any{})
		require.NoError(t, err)
		require.Contains(t, buf.String(), "abcd")
	})

	t.Run("variadic function with insufficient args", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ needs_two(1) }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"needs_two": func(a, b int, rest ...int) int {
				return a + b
			},
		})

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", map[string]any{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "at least 2 arguments")
	})

	t.Run("function returns multiple values with error", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ safe_divide(10, 2) }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"safe_divide": func(a, b int) (int, error) {
				if b == 0 {
					return 0, fmt.Errorf("division by zero")
				}
				return a / b, nil
			},
		})

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", map[string]any{})
		require.NoError(t, err)
		require.Contains(t, buf.String(), "5")
	})

	t.Run("string to float conversion", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ process("3.14") }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"process": func(f float64) float64 {
				return f * 2
			},
		})

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", map[string]any{})
		require.NoError(t, err)
		require.Contains(t, buf.String(), "6.28")
	})

	t.Run("string to bool conversion", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ process("true") }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"process": func(b bool) string {
				if b {
					return "yes"
				}
				return "no"
			},
		})

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", map[string]any{})
		require.NoError(t, err)
		require.Contains(t, buf.String(), "yes")
	})
}

// TestVue_ResolveArgument - comprehensive tests for argument resolution
func TestVue_ResolveArgument(t *testing.T) {
	t.Run("quoted string with double quotes", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ process("hello world") }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"process": func(s string) string {
				return s
			},
		})

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", map[string]any{})
		require.NoError(t, err)
		require.Contains(t, buf.String(), "hello world")
	})

	t.Run("quoted string with single quotes", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ process('hello world') }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"process": func(s string) string {
				return s
			},
		})

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", map[string]any{})
		require.NoError(t, err)
		require.Contains(t, buf.String(), "hello world")
	})

	t.Run("integer literal", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ double(42) }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"double": func(n int) int {
				return n * 2
			},
		})

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", map[string]any{})
		require.NoError(t, err)
		require.Contains(t, buf.String(), "84")
	})

	t.Run("float literal", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ half(3.14) }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"half": func(n float64) float64 {
				return n / 2
			},
		})

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", map[string]any{})
		require.NoError(t, err)
		require.Contains(t, buf.String(), "1.57")
	})

	t.Run("bool literal true", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ check(true) }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"check": func(b bool) string {
				if b {
					return "yes"
				}
				return "no"
			},
		})

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", map[string]any{})
		require.NoError(t, err)
		require.Contains(t, buf.String(), "yes")
	})

	t.Run("bool literal false", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ check(false) }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"check": func(b bool) string {
				if b {
					return "yes"
				}
				return "no"
			},
		})

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", map[string]any{})
		require.NoError(t, err)
		require.Contains(t, buf.String(), "no")
	})

	t.Run("variable reference", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ upper(name) }}</p>`),
			},
		}

		vue := vuego.NewVue(fs)
		data := map[string]any{"name": "alice"}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", data)
		require.NoError(t, err)
		require.Contains(t, buf.String(), "ALICE")
	})

	t.Run("variable reference with nested path", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ upper(user.name) }}</p>`),
			},
		}

		vue := vuego.NewVue(fs)
		data := map[string]any{
			"user": map[string]any{"name": "bob"},
		}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", data)
		require.NoError(t, err)
		require.Contains(t, buf.String(), "BOB")
	})

	t.Run("unquoted string literal fallback", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ process(unquoted) }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"process": func(s string) string {
				return s
			},
		})

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", map[string]any{})
		require.NoError(t, err)
		require.Contains(t, buf.String(), "unquoted")
	})
}

// TestVue_EvalAttributes - comprehensive tests for attribute evaluation
func TestVue_EvalAttributes(t *testing.T) {
	t.Run("simple attribute interpolation", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<img src="{{ image }}" />`),
			},
		}

		vue := vuego.NewVue(fs)
		data := map[string]any{"image": "test.jpg"}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", data)
		require.NoError(t, err)
		require.Contains(t, buf.String(), `src="test.jpg"`)
	})

	t.Run("attribute with piped filter", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<div class="{{ className | lower }}"></div>`),
			},
		}

		vue := vuego.NewVue(fs)
		data := map[string]any{"className": "PRIMARY"}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", data)
		require.NoError(t, err)
		require.Contains(t, buf.String(), `class="primary"`)
	})

	t.Run("attribute with function call", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<div data-count="{{ len(items) }}"></div>`),
			},
		}

		vue := vuego.NewVue(fs)
		data := map[string]any{"items": []string{"a", "b"}}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", data)
		require.NoError(t, err)
		require.Contains(t, buf.String(), `data-count="2"`)
	})

	t.Run("multiple interpolations in attribute", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<div title="{{ prefix }}: {{ suffix }}"></div>`),
			},
		}

		vue := vuego.NewVue(fs)
		data := map[string]any{
			"prefix": "Hello",
			"suffix": "World",
		}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", data)
		require.NoError(t, err)
		require.Contains(t, buf.String(), `title="Hello: World"`)
	})

	t.Run("attribute with variable referencing nested path", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<div id="{{ user.name }}"></div>`),
			},
		}

		vue := vuego.NewVue(fs)
		data := map[string]any{
			"user": map[string]any{"name": "john"},
		}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", data)
		require.NoError(t, err)
		require.Contains(t, buf.String(), `id="john"`)
	})

	t.Run("attribute with default filter for missing variable", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<div id="{{ missing | default('default-id') }}"></div>`),
			},
		}

		vue := vuego.NewVue(fs)
		data := map[string]any{}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", data)
		require.NoError(t, err)
		require.Contains(t, buf.String(), `id="default-id"`)
	})
}

// TestEscapeFunc - tests for the built-in escape function
func TestEscapeFunc(t *testing.T) {
	t.Run("escape filter available in funcmap", func(t *testing.T) {
		funcMap := vuego.DefaultFuncMap()
		require.NotNil(t, funcMap["escape"])
	})

	t.Run("escape works on simple strings", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ text | escape }}</p>`),
			},
		}

		vue := vuego.NewVue(fs)
		data := map[string]any{"text": "plain text"}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", data)
		require.NoError(t, err)
		require.Contains(t, buf.String(), "plain text")
	})

	t.Run("escape non-string returns string representation", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ number | escape }}</p>`),
			},
		}

		vue := vuego.NewVue(fs)
		data := map[string]any{"number": 42}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", data)
		require.NoError(t, err)
		require.Contains(t, buf.String(), "42")
	})
}

// TestConvertValue - comprehensive type conversion tests
func TestConvertValue(t *testing.T) {
	t.Run("string to int conversion via filter", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ add("5", "3") }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"add": func(a, b int) int {
				return a + b
			},
		})

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", map[string]any{})
		require.NoError(t, err)
		require.Contains(t, buf.String(), "8")
	})

	t.Run("string to uint conversion", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ process("100") }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"process": func(u uint) uint {
				return u * 2
			},
		})

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", map[string]any{})
		require.NoError(t, err)
		require.Contains(t, buf.String(), "200")
	})

	t.Run("invalid string to int conversion", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ process("notanumber") }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"process": func(i int) int {
				return i
			},
		})

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", map[string]any{})
		require.Error(t, err)
	})

	t.Run("string to float64 conversion", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ half("10.5") }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"half": func(f float64) float64 {
				return f / 2
			},
		})

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", map[string]any{})
		require.NoError(t, err)
		require.Contains(t, buf.String(), "5.25")
	})

	t.Run("int to float conversion", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.html": &fstest.MapFile{
				Data: []byte(`<p>{{ process(10) }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"process": func(f float64) float64 {
				return f / 2
			},
		})

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.html", map[string]any{})
		require.NoError(t, err)
		require.Contains(t, buf.String(), "5")
	})
}
