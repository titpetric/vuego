package vuego_test

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego"
)

func TestVue_Funcs_PipeChaining(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{
			Data: []byte(`<p>{{ name | upper }}</p>`),
		},
	}

	vue := vuego.NewVue(fs)
	data := map[string]any{
		"name": "john doe",
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.vuego", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "JOHN DOE")
}

func TestVue_Funcs_MultiplePipes(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{
			Data: []byte(`<p>{{ name | lower | title }}</p>`),
		},
	}

	vue := vuego.NewVue(fs)
	data := map[string]any{
		"name": "JOHN DOE",
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.vuego", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "John Doe")
}

func TestVue_Funcs_PipeWithArgs(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{
			Data: []byte(`<p>{{ value | default("fallback") }}</p>`),
		},
	}

	vue := vuego.NewVue(fs)
	data := map[string]any{
		"value": "",
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.vuego", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "fallback")
}

func TestVue_Funcs_CustomFunc(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{
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
	err := vue.Render(&buf, "test.vuego", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "Hello, Alice")
}

func TestVue_Funcs_FormatTime(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{
			Data: []byte(`<p>{{ timestamp | formatTime("2006-01-02") }}</p>`),
		},
	}

	vue := vuego.NewVue(fs)
	now := time.Now()
	data := map[string]any{
		"timestamp": now,
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.vuego", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), now.Format("2006-01-02"))
}

func TestVue_Funcs_LenFilter(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{
			Data: []byte(`<p>{{ items | len }}</p>`),
		},
	}

	vue := vuego.NewVue(fs)
	data := map[string]any{
		"items": []*http.Request{nil, nil, nil},
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.vuego", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "3")
}

func TestVue_Funcs_TrimFilter(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{
			Data: []byte(`<p>{{ text | trim }}</p>`),
		},
	}

	vue := vuego.NewVue(fs)
	data := map[string]any{
		"text": "  hello world  ",
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.vuego", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "hello world")
	require.NotContains(t, buf.String(), "  hello world  ")
}

func TestVue_Funcs_IntFilter(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{
			Data: []byte(`<p>{{ value | int }}</p>`),
		},
	}

	vue := vuego.NewVue(fs)
	data := map[string]any{
		"value": "42",
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.vuego", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "42")
}

func TestVue_Funcs_ComplexChain(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{
			Data: []byte(`<p>{{ text | trim | lower | title }}</p>`),
		},
	}

	vue := vuego.NewVue(fs)
	data := map[string]any{
		"text": "  HELLO WORLD  ",
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.vuego", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "Hello World")
}

func TestVue_Funcs_VIfWithFunction(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{
			Data: []byte(`<div v-if="len(items)"><p>Has items</p></div>`),
		},
	}

	vue := vuego.NewVue(fs)
	data := map[string]any{
		"items": []string{"a", "b", "c"},
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.vuego", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "Has items")
}

func TestVue_Funcs_VIfWithFunctionFalse(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{
			Data: []byte(`<div v-if="len(items)"><p>Has items</p></div>`),
		},
	}

	vue := vuego.NewVue(fs)
	data := map[string]any{
		"items": []string{},
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.vuego", data)
	require.NoError(t, err)
	require.NotContains(t, buf.String(), "Has items")
}

func TestVue_Funcs_NestedData(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{
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
	err := vue.Render(&buf, "test.vuego", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "ALICE")
}

func TestVue_Funcs_DefaultWithValue(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{
			Data: []byte(`<p>{{ name | default("Anonymous") }}</p>`),
		},
	}

	vue := vuego.NewVue(fs)
	data := map[string]any{
		"name": "Bob",
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.vuego", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "Bob")
	require.NotContains(t, buf.String(), "Anonymous")
}

func TestVue_Funcs_ChainWithDefault(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{
			Data: []byte(`<p>{{ name | default("guest") | upper }}</p>`),
		},
	}

	vue := vuego.NewVue(fs)
	data := map[string]any{
		"name": "",
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.vuego", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "GUEST")
}

func TestVue_Funcs_CustomFuncWithArgs(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{
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
	err := vue.Render(&buf, "test.vuego", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "15")
}

func TestVue_Funcs_StringFilter(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{
			Data: []byte(`<p>{{ count | string }}</p>`),
		},
	}

	vue := vuego.NewVue(fs)
	data := map[string]any{
		"count": 42,
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.vuego", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "42")
}

func TestVue_Funcs_MissingFilter(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{
			Data: []byte(`<p>{{ name | nonexistent }}</p>`),
		},
	}

	vue := vuego.NewVue(fs)
	data := map[string]any{
		"name": "test",
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.vuego", data)
	require.Error(t, err)
	require.Equal(t, "in test.vuego: in expression '{{ name | nonexistent }}': function 'nonexistent' not found", err.Error())
}

func TestVue_Funcs_VForWithFilters(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{
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
	err := vue.Render(&buf, "test.vuego", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "APPLE")
	require.Contains(t, buf.String(), "BANANA")
}

func TestVue_Funcs_TypedArguments(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{
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
	err := vue.Render(&buf, "test.vuego", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), now.Format("2006-01-02"))
}

func TestVue_Funcs_TypedMultipleArgs(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{
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
	err := vue.Render(&buf, "test.vuego", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "8")
}

func TestVue_Funcs_StringToIntConversion(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{
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
	err := vue.Render(&buf, "test.vuego", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "50")
}

func TestVue_Funcs_VariadicFunction(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{
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
	err := vue.Render(&buf, "test.vuego", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "a,b,c")
}

func TestVue_Funcs_ErrorReturn(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{
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
	err := vue.Render(&buf, "test.vuego", data)
	require.Error(t, err)
	require.Equal(t, "in test.vuego: in expression '{{ divide(10, 0) }}': divide(): division by zero", err.Error())
}

func TestVue_Funcs_NoReturnValue(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{
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
	err := vue.Render(&buf, "test.vuego", data)
	require.NoError(t, err)
	require.True(t, called)
}

func TestVue_Funcs_TypeConversionError(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{
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
	err := vue.Render(&buf, "test.vuego", data)
	require.Error(t, err)
	require.Equal(t, "in test.vuego: in expression '{{ items | double }}': double(): cannot convert argument 0 from []string to int", err.Error())
}

func TestVue_Funcs_StringToIntConversionVariations(t *testing.T) {
	t.Run("string to int8", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<p>{{ convert("42") }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"convert": func(n int8) int8 {
				return n
			},
		})

		data := map[string]any{}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.NoError(t, err)
		require.Contains(t, buf.String(), "42")
	})

	t.Run("string to int16", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<p>{{ convert("1000") }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"convert": func(n int16) int16 {
				return n
			},
		})

		data := map[string]any{}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.NoError(t, err)
		require.Contains(t, buf.String(), "1000")
	})

	t.Run("string to int32", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<p>{{ convert("100000") }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"convert": func(n int32) int32 {
				return n
			},
		})

		data := map[string]any{}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.NoError(t, err)
		require.Contains(t, buf.String(), "100000")
	})

	t.Run("string to int64", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<p>{{ convert("9223372036854775806") }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"convert": func(n int64) int64 {
				return n
			},
		})

		data := map[string]any{}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.NoError(t, err)
		require.Contains(t, buf.String(), "9223372036854775806")
	})

	t.Run("string to uint", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<p>{{ convert("42") }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"convert": func(n uint) uint {
				return n
			},
		})

		data := map[string]any{}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.NoError(t, err)
		require.Contains(t, buf.String(), "42")
	})

	t.Run("string to uint8", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<p>{{ convert("200") }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"convert": func(n uint8) uint8 {
				return n
			},
		})

		data := map[string]any{}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.NoError(t, err)
		require.Contains(t, buf.String(), "200")
	})

	t.Run("string to uint16", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<p>{{ convert("50000") }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"convert": func(n uint16) uint16 {
				return n
			},
		})

		data := map[string]any{}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.NoError(t, err)
		require.Contains(t, buf.String(), "50000")
	})

	t.Run("string to uint32", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<p>{{ convert("4000000000") }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"convert": func(n uint32) uint32 {
				return n
			},
		})

		data := map[string]any{}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.NoError(t, err)
		require.Contains(t, buf.String(), "4000000000")
	})

	t.Run("string to uint64", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<p>{{ convert("9223372036854775806") }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"convert": func(n uint64) uint64 {
				return n
			},
		})

		data := map[string]any{}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.NoError(t, err)
		require.Contains(t, buf.String(), "9223372036854775806")
	})

	t.Run("string to float32", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<p>{{ convert("3.14") }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"convert": func(f float32) float32 {
				return f
			},
		})

		data := map[string]any{}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.NoError(t, err)
		require.Contains(t, buf.String(), "3.14")
	})

	t.Run("string to float64", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<p>{{ convert("2.71828") }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"convert": func(f float64) float64 {
				return f
			},
		})

		data := map[string]any{}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.NoError(t, err)
		require.Contains(t, buf.String(), "2.71828")
	})

	t.Run("string to bool", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<p>{{ convert("true") }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"convert": func(b bool) string {
				return fmt.Sprintf("%v", b)
			},
		})

		data := map[string]any{}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.NoError(t, err)
		require.Contains(t, buf.String(), "true")
	})

	t.Run("int to string via function", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<p>{{ toStr(42) }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"toStr": func(n int) string {
				return fmt.Sprintf("%d", n)
			},
		})

		data := map[string]any{}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.NoError(t, err)
		require.Contains(t, buf.String(), "42")
	})

	t.Run("float to string via function", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<p>{{ toStr(3.14159) }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"toStr": func(f float64) string {
				return fmt.Sprintf("%v", f)
			},
		})

		data := map[string]any{}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.NoError(t, err)
		require.Contains(t, buf.String(), "3.14159")
	})

	t.Run("bool to string via function", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<p>{{ toStr(true) }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"toStr": func(b bool) string {
				return fmt.Sprintf("%v", b)
			},
		})

		data := map[string]any{}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.NoError(t, err)
		require.Contains(t, buf.String(), "true")
	})

	t.Run("invalid string to int", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<p>{{ convert("not-a-number") }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"convert": func(n int) int {
				return n
			},
		})

		data := map[string]any{}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.Error(t, err)
		require.Equal(t, "in test.vuego: in expression '{{ convert(\"not-a-number\") }}': convert(): cannot convert argument 0 from string to int", err.Error())
	})

	t.Run("invalid string to float", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<p>{{ convert("not-a-float") }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"convert": func(f float64) float64 {
				return f
			},
		})

		data := map[string]any{}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.Error(t, err)
		require.Equal(t, "in test.vuego: in expression '{{ convert(\"not-a-float\") }}': convert(): cannot convert argument 0 from string to float64", err.Error())
	})

	t.Run("invalid string to bool", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<p>{{ convert("maybe") }}</p>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"convert": func(b bool) bool {
				return b
			},
		})

		data := map[string]any{}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.Error(t, err)
		require.Equal(t, "in test.vuego: in expression '{{ convert(\"maybe\") }}': convert(): cannot convert argument 0 from string to bool", err.Error())
	})
}

func TestVue_Funcs_JsonFilter(t *testing.T) {
	t.Run("simple object in script tag", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<script>const data = {{ user | json }};</script>`),
			},
		}

		vue := vuego.NewVue(fs)
		data := map[string]any{
			"user": map[string]any{
				"name": "Alice",
				"age":  30,
			},
		}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.NoError(t, err)
		output := buf.String()
		// JSON is not HTML-escaped inside script tags
		require.Contains(t, output, `"age": 30`)
		require.Contains(t, output, `"name": "Alice"`)
		require.NotContains(t, output, `&#34;`)
	})

	t.Run("array in script tag", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<script>const items = {{ items | json }};</script>`),
			},
		}

		vue := vuego.NewVue(fs)
		data := map[string]any{
			"items": []string{"apple", "banana", "cherry"},
		}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.NoError(t, err)
		output := buf.String()
		require.Contains(t, output, `"apple"`)
		require.Contains(t, output, `"banana"`)
		require.Contains(t, output, `"cherry"`)
		require.NotContains(t, output, `&#34;`)
	})

	t.Run("string value in script tag", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<script>const name = {{ name | json }};</script>`),
			},
		}

		vue := vuego.NewVue(fs)
		data := map[string]any{
			"name": "Bob",
		}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.NoError(t, err)
		output := buf.String()
		require.Contains(t, output, `"Bob"`)
		require.NotContains(t, output, `&#34;`)
	})

	t.Run("number value in script tag", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<script>const count = {{ count | json }};</script>`),
			},
		}

		vue := vuego.NewVue(fs)
		data := map[string]any{
			"count": 42,
		}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.NoError(t, err)
		// Numbers don't need escaping
		require.Contains(t, buf.String(), `42`)
	})

	t.Run("nested structure in script tag", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<script>const payload = {{ payload | json }};</script>`),
			},
		}

		vue := vuego.NewVue(fs)
		data := map[string]any{
			"payload": map[string]any{
				"user": map[string]any{
					"name": "Charlie",
					"tags": []string{"admin", "user"},
				},
				"status": "active",
			},
		}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.NoError(t, err)
		output := buf.String()
		// Check unescaped JSON
		require.Contains(t, output, `"name": "Charlie"`)
		require.Contains(t, output, `"status": "active"`)
		require.NotContains(t, output, `&#34;`)
	})

	t.Run("fragment still escapes JSON outside script tag", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`{{ data | json }}`),
			},
		}

		vue := vuego.NewVue(fs)
		data := map[string]any{
			"data": map[string]any{"key": "value"},
		}

		var buf bytes.Buffer
		err := vue.RenderFragment(&buf, "test.vuego", data)
		require.NoError(t, err)
		output := buf.String()
		// JSON is HTML-escaped in fragments (not in script tags)
		require.Contains(t, output, `&#34;key&#34;`)
		require.Contains(t, output, `&#34;value&#34;`)
	})

	t.Run("JSON escaped in code tag", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<code>{{ data | json }}</code>`),
			},
		}

		vue := vuego.NewVue(fs)
		data := map[string]any{
			"data": map[string]any{"key": "value"},
		}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.NoError(t, err)
		output := buf.String()
		// JSON should be HTML-escaped inside code tags
		require.Contains(t, output, `&#34;key&#34;`)
		require.Contains(t, output, `&#34;value&#34;`)
		require.NotContains(t, output, `"key": "value"`)
	})

	t.Run("JSON escaped in pre tag", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<pre>{{ data | json }}</pre>`),
			},
		}

		vue := vuego.NewVue(fs)
		data := map[string]any{
			"data": map[string]any{
				"name": "Alice",
				"age":  30,
			},
		}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.NoError(t, err)
		output := buf.String()
		// JSON should be HTML-escaped inside pre tags
		require.Contains(t, output, `&#34;name&#34;`)
		require.Contains(t, output, `&#34;Alice&#34;`)
		require.Contains(t, output, `&#34;age&#34;`)
		require.NotContains(t, output, `"name": "Alice"`)
	})

	t.Run("JSON escaped in pre+code tags", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<pre><code>{{ data | json }}</code></pre>`),
			},
		}

		vue := vuego.NewVue(fs)
		data := map[string]any{
			"data": map[string]any{
				"status": "active",
				"count":  42,
			},
		}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.NoError(t, err)
		output := buf.String()
		// JSON should be HTML-escaped inside pre+code tags
		require.Contains(t, output, `&#34;status&#34;`)
		require.Contains(t, output, `&#34;active&#34;`)
		require.Contains(t, output, `&#34;count&#34;`)
		require.NotContains(t, output, `"status": "active"`)
	})

	t.Run("JSON unescaped in script tag (comparison)", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<script>const data = {{ data | json }};</script>`),
			},
		}

		vue := vuego.NewVue(fs)
		data := map[string]any{
			"data": map[string]any{
				"status": "active",
				"count":  42,
			},
		}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.NoError(t, err)
		output := buf.String()
		// JSON should NOT be escaped inside script tags
		require.Contains(t, output, `"status": "active"`)
		require.Contains(t, output, `"count": 42`)
		require.NotContains(t, output, `&#34;`)
	})
}

func TestVue_Funcs_LoadSVG(t *testing.T) {
	t.Run("loadSVG returns file content as string", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<div>{{ loadSVG("icons/star.svg") }}</div>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"loadSVG": func(path string) (string, error) {
				return "svg content", nil
			},
		})
		data := map[string]any{}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.NoError(t, err)
		require.Contains(t, buf.String(), "svg content")
	})

	t.Run("loadSVG breaks rendering on missing file", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<div>{{ loadSVG("missing.svg") }}</div>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"loadSVG": func(path string) (string, error) {
				return "", fmt.Errorf("failed to load SVG file '%s'", path)
			},
		})
		data := map[string]any{}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.Error(t, err)
		require.Equal(t, "in test.vuego: in expression '{{ loadSVG(\"missing.svg\") }}': loadSVG(): failed to load SVG file 'missing.svg'", err.Error())
	})

	t.Run("loadSVG error propagates through pipe", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<div>{{ path | loadSVG }}</div>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"loadSVG": func(path string) (string, error) {
				return "", fmt.Errorf("failed to load SVG file '%s'", path)
			},
		})
		data := map[string]any{
			"path": "icons/check.svg",
		}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.Error(t, err)
		require.Equal(t, "in test.vuego: in expression '{{ path | loadSVG }}': loadSVG(): failed to load SVG file 'icons/check.svg'", err.Error())
	})

	t.Run("loadSVG error context includes template name", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<div>{{ loadSVG("icons/nonexistent.svg") }}</div>`),
			},
		}

		vue := vuego.NewVue(fs).Funcs(vuego.FuncMap{
			"loadSVG": func(path string) (string, error) {
				return "", fmt.Errorf("failed to load SVG file '%s'", path)
			},
		})
		data := map[string]any{}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.Error(t, err)
		require.Equal(t, "in test.vuego: in expression '{{ loadSVG(\"icons/nonexistent.svg\") }}': loadSVG(): failed to load SVG file 'icons/nonexistent.svg'", err.Error())
	})
}

func TestVue_InterpolateUnescapedInScriptTag(t *testing.T) {
	t.Run("plain string with quotes in script tag", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<script>const msg = "{{ message }}";</script>`),
			},
		}

		vue := vuego.NewVue(fs)
		data := map[string]any{
			"message": "hello world",
		}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.NoError(t, err)
		output := buf.String()
		// Interpolated values inside script tags should be unescaped
		require.Contains(t, output, `const msg = "hello world";`)
	})

	t.Run("string with HTML entities in script tag", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<script>const value = "{{ htmlContent }}";</script>`),
			},
		}

		vue := vuego.NewVue(fs)
		data := map[string]any{
			"htmlContent": `<div class="test">content</div>`,
		}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.NoError(t, err)
		output := buf.String()
		// Should NOT be HTML-escaped inside script tags
		require.Contains(t, output, `<div class="test">content</div>`)
		require.NotContains(t, output, `&lt;`)
		require.NotContains(t, output, `&gt;`)
		require.NotContains(t, output, `&#34;`)
	})

	t.Run("string with special chars in script tag", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<script>const val = '{{ val }}';</script>`),
			},
		}

		vue := vuego.NewVue(fs)
		data := map[string]any{
			"val": `test & check < > " '`,
		}

		var buf bytes.Buffer
		err := vue.Render(&buf, "test.vuego", data)
		require.NoError(t, err)
		output := buf.String()
		// Should NOT be HTML-escaped inside script tags
		require.Contains(t, output, `test & check < > " '`)
		require.NotContains(t, output, `&amp;`)
		require.NotContains(t, output, `&lt;`)
		require.NotContains(t, output, `&gt;`)
	})

	t.Run("escaping still applies outside script tags", func(t *testing.T) {
		fs := fstest.MapFS{
			"test.vuego": &fstest.MapFile{
				Data: []byte(`<div>{{ value }}</div>`),
			},
		}

		vue := vuego.NewVue(fs)
		data := map[string]any{
			"value": `<script>alert('xss')</script>`,
		}

		var buf bytes.Buffer
		err := vue.RenderFragment(&buf, "test.vuego", data)
		require.NoError(t, err)
		output := buf.String()
		// Should still be HTML-escaped outside script tags
		require.Contains(t, output, `&lt;script&gt;`)
		require.NotContains(t, output, `<script>`)
	})
}
