package vuego_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/titpetric/vuego"
)

func TestVue_Funcs_PipeChaining(t *testing.T) {
	fs := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte(`<p>{{ name | upper }}</p>`),
		},
	}

	vue := vuego.NewVue(fs)
	data := map[string]any{
		"name": "john doe",
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.html", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "JOHN DOE")
}

func TestVue_Funcs_MultiplePipes(t *testing.T) {
	fs := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte(`<p>{{ name | lower | title }}</p>`),
		},
	}

	vue := vuego.NewVue(fs)
	data := map[string]any{
		"name": "JOHN DOE",
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.html", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "John Doe")
}

func TestVue_Funcs_PipeWithArgs(t *testing.T) {
	fs := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte(`<p>{{ value | default("fallback") }}</p>`),
		},
	}

	vue := vuego.NewVue(fs)
	data := map[string]any{
		"value": "",
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.html", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "fallback")
}

func TestVue_Funcs_CustomFunc(t *testing.T) {
	fs := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte(`<p>{{ name | greet }}</p>`),
		},
	}

	vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
		"greet": func(v any) any {
			return "Hello, " + v.(string)
		},
	})

	data := map[string]any{
		"name": "Alice",
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.html", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "Hello, Alice")
}

func TestVue_Funcs_FormatTime(t *testing.T) {
	fs := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte(`<p>{{ timestamp | formatTime("2006-01-02") }}</p>`),
		},
	}

	vue := vuego.NewVue(fs)
	now := time.Now()
	data := map[string]any{
		"timestamp": now,
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.html", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), now.Format("2006-01-02"))
}

func TestVue_Funcs_LenFilter(t *testing.T) {
	fs := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte(`<p>{{ items | len }}</p>`),
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
}

func TestVue_Funcs_TrimFilter(t *testing.T) {
	fs := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte(`<p>{{ text | trim }}</p>`),
		},
	}

	vue := vuego.NewVue(fs)
	data := map[string]any{
		"text": "  hello world  ",
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.html", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "hello world")
	require.NotContains(t, buf.String(), "  hello world  ")
}

func TestVue_Funcs_IntFilter(t *testing.T) {
	fs := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte(`<p>{{ value | int }}</p>`),
		},
	}

	vue := vuego.NewVue(fs)
	data := map[string]any{
		"value": "42",
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.html", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "42")
}

func TestVue_Funcs_ComplexChain(t *testing.T) {
	fs := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte(`<p>{{ text | trim | lower | title }}</p>`),
		},
	}

	vue := vuego.NewVue(fs)
	data := map[string]any{
		"text": "  HELLO WORLD  ",
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.html", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "Hello World")
}

func TestVue_Funcs_VIfWithFunction(t *testing.T) {
	fs := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte(`<div v-if="len(items)"><p>Has items</p></div>`),
		},
	}

	vue := vuego.NewVue(fs)
	data := map[string]any{
		"items": []string{"a", "b", "c"},
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.html", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "Has items")
}

func TestVue_Funcs_VIfWithFunctionFalse(t *testing.T) {
	fs := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte(`<div v-if="len(items)"><p>Has items</p></div>`),
		},
	}

	vue := vuego.NewVue(fs)
	data := map[string]any{
		"items": []string{},
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.html", data)
	require.NoError(t, err)
	require.NotContains(t, buf.String(), "Has items")
}

func TestVue_Funcs_NestedData(t *testing.T) {
	fs := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte(`<p>{{ user.name | upper }}</p>`),
		},
	}

	vue := vuego.NewVue(fs)
	data := map[string]any{
		"user": map[string]any{
			"name": "alice",
		},
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.html", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "ALICE")
}

func TestVue_Funcs_DefaultWithValue(t *testing.T) {
	fs := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte(`<p>{{ name | default("Anonymous") }}</p>`),
		},
	}

	vue := vuego.NewVue(fs)
	data := map[string]any{
		"name": "Bob",
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.html", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "Bob")
	require.NotContains(t, buf.String(), "Anonymous")
}

func TestVue_Funcs_ChainWithDefault(t *testing.T) {
	fs := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte(`<p>{{ name | default("guest") | upper }}</p>`),
		},
	}

	vue := vuego.NewVue(fs)
	data := map[string]any{
		"name": "",
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.html", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "GUEST")
}

func TestVue_Funcs_CustomFuncWithArgs(t *testing.T) {
	fs := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte(`<p>{{ value | multiply(3) }}</p>`),
		},
	}

	vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
		"multiply": func(v int, multiplier int) int {
			return v * multiplier
		},
	})

	data := map[string]any{
		"value": 5,
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.html", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "15")
}

func TestVue_Funcs_StringFilter(t *testing.T) {
	fs := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte(`<p>{{ count | string }}</p>`),
		},
	}

	vue := vuego.NewVue(fs)
	data := map[string]any{
		"count": 42,
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.html", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "42")
}

func TestVue_Funcs_NoFilterChain(t *testing.T) {
	fs := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte(`<p>{{ name }}</p>`),
		},
	}

	vue := vuego.NewVue(fs)
	data := map[string]any{
		"name": "test",
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.html", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "test")
}

func TestVue_Funcs_MissingFilter(t *testing.T) {
	fs := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte(`<p>{{ name | nonexistent }}</p>`),
		},
	}

	vue := vuego.NewVue(fs)
	data := map[string]any{
		"name": "test",
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.html", data)
	require.Error(t, err)
	require.Equal(t, "in test.html: in expression '{{ name | nonexistent }}': function 'nonexistent' not found", err.Error())
}

func TestVue_Funcs_VForWithFilters(t *testing.T) {
	fs := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte(strings.TrimSpace(`
<ul>
  <li v-for="item in items">{{ item.name | upper }}</li>
</ul>
			`)),
		},
	}

	vue := vuego.NewVue(fs)
	data := map[string]any{
		"items": []map[string]any{
			{"name": "apple"},
			{"name": "banana"},
		},
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.html", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "APPLE")
	require.Contains(t, buf.String(), "BANANA")
}

func TestVue_Funcs_TypedArguments(t *testing.T) {
	fs := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte(`<p>{{ timestamp | formatDate }}</p>`),
		},
	}

	vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
		"formatDate": func(t *time.Time) string {
			if t == nil {
				return ""
			}
			return t.Format("2006-01-02")
		},
	})

	now := time.Now()
	data := map[string]any{
		"timestamp": &now,
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.html", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), now.Format("2006-01-02"))
}

func TestVue_Funcs_TypedMultipleArgs(t *testing.T) {
	fs := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte(`<p>{{ add(5, 3) }}</p>`),
		},
	}

	vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
		"add": func(a int, b int) int {
			return a + b
		},
	})

	data := map[string]any{}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.html", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "8")
}

func TestVue_Funcs_StringToIntConversion(t *testing.T) {
	fs := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte(`<p>{{ multiply("10", 5) }}</p>`),
		},
	}

	vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
		"multiply": func(a int, b int) int {
			return a * b
		},
	})

	data := map[string]any{}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.html", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "50")
}

func TestVue_Funcs_VariadicFunction(t *testing.T) {
	fs := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte(`<p>{{ join("a", "b", "c") }}</p>`),
		},
	}

	vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
		"join": func(parts ...string) string {
			return strings.Join(parts, ",")
		},
	})

	data := map[string]any{}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.html", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "a,b,c")
}

func TestVue_Funcs_ErrorReturn(t *testing.T) {
	fs := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte(`<p>{{ divide(10, 0) }}</p>`),
		},
	}

	vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
		"divide": func(a, b int) (int, error) {
			if b == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return a / b, nil
		},
	})

	data := map[string]any{}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.html", data)
	require.Error(t, err)
	require.Equal(t, "in test.html: in expression '{{ divide(10, 0) }}': divide(): division by zero", err.Error())
}

func TestVue_Funcs_NoReturnValue(t *testing.T) {
	fs := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte(`<p>{{ name | noop }}</p>`),
		},
	}

	called := false
	vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
		"noop": func(v string) {
			called = true
		},
	})

	data := map[string]any{
		"name": "test",
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.html", data)
	require.NoError(t, err)
	require.True(t, called)
}

func TestVue_Funcs_TypeConversionError(t *testing.T) {
	fs := fstest.MapFS{
		"test.html": &fstest.MapFile{
			Data: []byte(`<p>{{ items | double }}</p>`),
		},
	}

	vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
		"double": func(n int) int {
			return n * 2
		},
	})

	data := map[string]any{
		"items": []string{"a", "b", "c"},
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.html", data)
	require.Error(t, err)
	require.Equal(t, "in test.html: in expression '{{ items | double }}': double(): cannot convert argument 0 from []string to int", err.Error())
}
