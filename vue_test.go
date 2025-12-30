package vuego_test

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/internal/helpers"
)

func TestVue_Render(t *testing.T) {
	var buf bytes.Buffer

	vue := vuego.NewVue(os.DirFS("testdata/pages"))
	err := vue.Render(&buf, "Index.vuego", nil)
	require.NoError(t, err)

	t.Log(buf.String())
}

func TestFixtures(t *testing.T) {
	fixtures := os.DirFS("testdata/fixtures")

	// Create a template with components loaded recursively
	tpl := vuego.NewFS(fixtures, vuego.WithComponents())

	templates, err := fs.Glob(fixtures, "*.vuego")
	require.NoError(t, err)

	for _, template := range templates {
		want, err := fs.ReadFile(fixtures, strings.ReplaceAll(template, ".vuego", ".html"))
		require.NoError(t, err)

		dataBytes, err := fs.ReadFile(fixtures, strings.ReplaceAll(template, ".vuego", ".json"))
		require.NoError(t, err)

		templateBytes, err := fs.ReadFile(fixtures, template)
		require.NoError(t, err)

		data := map[string]any{}
		require.NoError(t, json.NewDecoder(bytes.NewReader(dataBytes)).Decode(&data))

		t.Run(template, func(t *testing.T) {
			var got bytes.Buffer
			loaded := tpl.Load(template).Fill(data)
			require.NoError(t, loaded.Render(t.Context(), &got))
			if !helpers.EqualHTML(t, want, got.Bytes(), templateBytes, dataBytes) {
				t.Logf("HTML mismatch in %s", template)
				t.Logf("Expected:\n%s", string(want))
				t.Logf("Got:\n%s", got.String())
				t.FailNow()
			}
		})
	}
}

func TestVue_Escaping(t *testing.T) {
	const template = "template.vuego"

	templateFS := &fstest.MapFS{
		template: {Data: []byte(`<div data-tooltip="{{value}}"></div>`)},
	}
	data := map[string]any{
		"value": "A & B",
	}

	vue := vuego.NewVue(templateFS)

	want := []byte(`<div data-tooltip="A &amp; B"></div>`)

	t.Run(template, func(t *testing.T) {
		var got bytes.Buffer
		require.NoError(t, vue.RenderFragment(&got, template, data))
		require.True(t, helpers.EqualHTML(t, want, got.Bytes(), nil, nil))

		t.Logf("-- Escape result: %s", got.String())
	})
}

func TestVue_BoundAttributesFalsey(t *testing.T) {
	const template = "template.vuego"

	t.Run("false boolean attribute is removed", func(t *testing.T) {
		templateFS := &fstest.MapFS{
			template: {Data: []byte(`<input :checked="isDone" />`)},
		}
		data := map[string]any{
			"isDone": false,
		}

		vue := vuego.NewVue(templateFS)
		var got bytes.Buffer
		require.NoError(t, vue.RenderFragment(&got, template, data))
		// Should not contain "checked" attribute
		require.NotContains(t, got.String(), "checked")
	})

	t.Run("true boolean attribute is rendered", func(t *testing.T) {
		templateFS := &fstest.MapFS{
			template: {Data: []byte(`<input :checked="isDone" />`)},
		}
		data := map[string]any{
			"isDone": true,
		}

		vue := vuego.NewVue(templateFS)
		var got bytes.Buffer
		require.NoError(t, vue.RenderFragment(&got, template, data))
		// Should contain "checked" attribute
		require.Contains(t, got.String(), "checked")
	})

	t.Run("empty string is falsey", func(t *testing.T) {
		templateFS := &fstest.MapFS{
			template: {Data: []byte(`<input :disabled="reason" />`)},
		}
		data := map[string]any{
			"reason": "",
		}

		vue := vuego.NewVue(templateFS)
		var got bytes.Buffer
		require.NoError(t, vue.RenderFragment(&got, template, data))
		require.NotContains(t, got.String(), "disabled")
	})

	t.Run("string 'false' is falsey", func(t *testing.T) {
		templateFS := &fstest.MapFS{
			template: {Data: []byte(`<input :disabled="reason" />`)},
		}
		data := map[string]any{
			"reason": "false",
		}

		vue := vuego.NewVue(templateFS)
		var got bytes.Buffer
		require.NoError(t, vue.RenderFragment(&got, template, data))
		require.NotContains(t, got.String(), "disabled")
	})

	t.Run("zero is falsey", func(t *testing.T) {
		templateFS := &fstest.MapFS{
			template: {Data: []byte(`<input :disabled="count" />`)},
		}
		data := map[string]any{
			"count": 0,
		}

		vue := vuego.NewVue(templateFS)
		var got bytes.Buffer
		require.NoError(t, vue.RenderFragment(&got, template, data))
		require.NotContains(t, got.String(), "disabled")
	})

	t.Run("nil is falsey", func(t *testing.T) {
		templateFS := &fstest.MapFS{
			template: {Data: []byte(`<input :disabled="val" />`)},
		}
		data := map[string]any{
			"val": nil,
		}

		vue := vuego.NewVue(templateFS)
		var got bytes.Buffer
		require.NoError(t, vue.RenderFragment(&got, template, data))
		require.NotContains(t, got.String(), "disabled")
	})

	t.Run("truthy string is rendered", func(t *testing.T) {
		templateFS := &fstest.MapFS{
			template: {Data: []byte(`<input :disabled="reason" />`)},
		}
		data := map[string]any{
			"reason": "User not ready",
		}

		vue := vuego.NewVue(templateFS)
		var got bytes.Buffer
		require.NoError(t, vue.RenderFragment(&got, template, data))
		require.Contains(t, got.String(), "disabled")
	})

	t.Run("v-bind: prefix also respects falsey values", func(t *testing.T) {
		templateFS := &fstest.MapFS{
			template: {Data: []byte(`<input v-bind:required="isRequired" />`)},
		}
		data := map[string]any{
			"isRequired": false,
		}

		vue := vuego.NewVue(templateFS)
		var got bytes.Buffer
		require.NoError(t, vue.RenderFragment(&got, template, data))
		require.NotContains(t, got.String(), "required")
	})
}

func TestVue_StructFieldAccess(t *testing.T) {
	const template = "template.vuego"

	t.Run("struct fields accessible without data prefix", func(t *testing.T) {
		templateFS := &fstest.MapFS{
			template: {Data: []byte(`<div>Name: {{Name}}, Age: {{Age}}</div>`)},
		}

		type Person struct {
			Name string
			Age  int
		}

		person := Person{Name: "Alice", Age: 30}
		vue := vuego.NewVue(templateFS)
		var got bytes.Buffer
		require.NoError(t, vue.RenderFragment(&got, template, person))
		require.Contains(t, got.String(), "Name: Alice")
		require.Contains(t, got.String(), "Age: 30")
	})

	t.Run("nested struct fields work", func(t *testing.T) {
		templateFS := &fstest.MapFS{
			template: {Data: []byte(`<div>{{Address.City}}</div>`)},
		}

		type Address struct {
			City string
		}

		type Person struct {
			Address Address
		}

		person := Person{Address: Address{City: "New York"}}
		vue := vuego.NewVue(templateFS)
		var got bytes.Buffer
		require.NoError(t, vue.RenderFragment(&got, template, person))
		require.Contains(t, got.String(), "New York")
	})

	t.Run("map data still takes precedence over struct fields", func(t *testing.T) {
		templateFS := &fstest.MapFS{
			template: {Data: []byte(`<div>{{Name}}</div>`)},
		}

		data := map[string]any{
			"Name": "Bob",
		}

		vue := vuego.NewVue(templateFS)
		var got bytes.Buffer
		// When passing a map, struct fields are not used
		require.NoError(t, vue.RenderFragment(&got, template, data))
		require.Contains(t, got.String(), "Bob")
	})
}
