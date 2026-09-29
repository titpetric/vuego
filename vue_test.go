package vuego_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io/fs"
	"maps"
	"os"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"golang.org/x/net/html"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/tests/assert"
)

// The helpers below are shared by every test file in this package.

// renderFragment renders source as a template fragment with data and returns
// the output.
func renderFragment(t *testing.T, source string, data any) string {
	t.Helper()

	const filename = "template.vuego"
	vue := vuego.NewVue(fstest.MapFS{filename: {Data: []byte(source)}})

	var buf bytes.Buffer
	assert.NoError(t, vue.RenderFragment(t.Context(), &buf, filename, data))
	return buf.String()
}

// renderFragmentErr renders source as a template fragment and returns the
// output together with the render error.
func renderFragmentErr(t *testing.T, source string, data any) (string, error) {
	t.Helper()

	const filename = "template.vuego"
	vue := vuego.NewVue(fstest.MapFS{filename: {Data: []byte(source)}})

	var buf bytes.Buffer
	err := vue.RenderFragment(t.Context(), &buf, filename, data)
	return buf.String(), err
}

// assertRenderHTML renders source as a template fragment with data and compares
// the resulting DOM against want.
func assertRenderHTML(t *testing.T, source string, data any, want string) {
	t.Helper()

	assert.EqualHTML(t, []byte(want), []byte(renderFragment(t, source, data)), []byte(source), nil)
}

// renderFuncs renders test.vuego out of files as a full page, with data and any
// extra template functions, and returns the output together with the error.
func renderFuncs(t *testing.T, files map[string]string, data any, funcs vuego.FuncMap) (string, error) {
	t.Helper()

	fsys := fstest.MapFS{}
	for name, source := range files {
		fsys[name] = &fstest.MapFile{Data: []byte(source)}
	}

	vue := vuego.NewVue(fsys)
	if funcs != nil {
		vue.Funcs(funcs)
	}

	var buf bytes.Buffer
	err := vue.Render(t.Context(), &buf, "test.vuego", data)
	return buf.String(), err
}

// renderPage renders source as a full page template named test.vuego.
func renderPage(t *testing.T, source string, data any) string {
	t.Helper()

	got, err := renderFuncs(t, map[string]string{"test.vuego": source}, data, nil)
	assert.NoError(t, err)
	return got
}

// renderCtx renders source as a full page under the given context, so a
// template function taking a context.Context can be observed.
func renderCtx(t *testing.T, ctx context.Context, source string, data any) (string, error) {
	t.Helper()

	vue := vuego.NewVue(fstest.MapFS{"test.vuego": {Data: []byte(source)}})

	var buf bytes.Buffer
	err := vue.Render(ctx, &buf, "test.vuego", data)
	return buf.String(), err
}

// renderInclude renders parent, which is expected to include the other named
// templates, and returns the output.
func renderInclude(t *testing.T, parent string, templates map[string]string, data any) string {
	t.Helper()

	fsys := fstest.MapFS{"parent.vuego": {Data: []byte(parent)}}
	for name, source := range templates {
		fsys[name] = &fstest.MapFile{Data: []byte(source)}
	}

	var buf bytes.Buffer
	assert.NoError(t, vuego.NewVue(fsys).RenderFragment(t.Context(), &buf, "parent.vuego", data))
	return buf.String()
}

// nodeText returns the concatenated text of a node list and its descendants.
func nodeText(nodes []*html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	for _, n := range nodes {
		walk(n)
	}
	return b.String()
}

func TestNewVue(t *testing.T) {
	t.Run("carries the default funcmap", func(t *testing.T) {
		assert.Contains(t, renderFragment(t, `<p>{{ name | upper }}</p>`, map[string]any{"name": "ada"}), "ADA")
	})

	t.Run("a nil filesystem is reported on render", func(t *testing.T) {
		var buf bytes.Buffer
		err := vuego.NewVue(nil).Render(t.Context(), &buf, "any.vuego", nil)
		assert.Error(t, err)
	})
}

func TestVue_Render(t *testing.T) {
	t.Run("renders a page from the filesystem", func(t *testing.T) {
		var buf bytes.Buffer
		vue := vuego.NewVue(os.DirFS("testdata/pages"))
		assert.NoError(t, vue.Render(t.Context(), &buf, "Index.vuego", nil))
		assert.NotEmpty(t, buf.String())
	})

	t.Run("front matter overrides the passed data", func(t *testing.T) {
		vue := vuego.NewVue(fstest.MapFS{
			"page.vuego": {Data: []byte("---\ntitle: FromFrontMatter\n---\n<h1>{{ title }}</h1>")},
		})

		var buf bytes.Buffer
		assert.NoError(t, vue.Render(t.Context(), &buf, "page.vuego", map[string]any{"title": "FromData"}))
		assert.Contains(t, buf.String(), "FromFrontMatter")
		assert.NotContains(t, buf.String(), "FromData")
	})

	t.Run("reports a missing file", func(t *testing.T) {
		var buf bytes.Buffer
		vue := vuego.NewVue(fstest.MapFS{})
		assert.Error(t, vue.Render(t.Context(), &buf, "missing.vuego", nil))
	})
}

func TestVue_Funcs(t *testing.T) {
	t.Run("merges into the default funcmap", func(t *testing.T) {
		const filename = "template.vuego"
		vue := vuego.NewVue(fstest.MapFS{filename: {Data: []byte(`<p>{{ name | shout }}</p>`)}})
		vue.Funcs(vuego.FuncMap{"shout": func(v any) any { return strings.ToUpper(v.(string)) + "!" }})

		var buf bytes.Buffer
		assert.NoError(t, vue.RenderFragment(t.Context(), &buf, filename, map[string]any{"name": "ada"}))
		assert.Contains(t, buf.String(), "ADA!")
	})

	t.Run("overwrites a built-in of the same name", func(t *testing.T) {
		const filename = "template.vuego"
		vue := vuego.NewVue(fstest.MapFS{filename: {Data: []byte(`<p>{{ name | upper }}</p>`)}})
		vue.Funcs(vuego.FuncMap{"upper": func(any) any { return "replaced" }})

		var buf bytes.Buffer
		assert.NoError(t, vue.RenderFragment(t.Context(), &buf, filename, map[string]any{"name": "ada"}))
		assert.Contains(t, buf.String(), "replaced")
	})

	t.Run("returns the receiver for chaining", func(t *testing.T) {
		vue := vuego.NewVue(nil)
		assert.Equal(t, vue, vue.Funcs(vuego.FuncMap{}))
	})
}

func TestVue_DefaultFuncMap(t *testing.T) {
	funcs := vuego.NewVue(nil).DefaultFuncMap()

	t.Run("holds every documented built-in", func(t *testing.T) {
		want := []string{
			"default", "escape", "file", "formatDate", "formatTime", "int",
			"json", "jsonFile", "jsonPretty", "len", "lower", "string",
			"title", "trim", "type", "upper", "yamlFile",
		}

		names := slices.Sorted(maps.Keys(funcs))
		assert.Equal(t, want, names)
	})

	t.Run("each entry is a func", func(t *testing.T) {
		for name, fn := range funcs {
			assert.NotNil(t, fn, name)
		}
	})

	t.Run("a fresh map is returned each call", func(t *testing.T) {
		vue := vuego.NewVue(nil)

		first := vue.DefaultFuncMap()
		delete(first, "upper")

		assert.NotNil(t, vue.DefaultFuncMap()["upper"])
	})
}

func TestVue_RegisterComponent(t *testing.T) {
	t.Run("maps a tag name to a filename", func(t *testing.T) {
		vue := vuego.NewVue(nil)
		vue.RegisterComponent("button-primary", "components/ButtonPrimary.vuego")

		filename, ok := vue.GetComponentFile("button-primary")
		assert.True(t, ok)
		assert.Equal(t, "components/ButtonPrimary.vuego", filename)
	})

	t.Run("a later registration replaces an earlier one", func(t *testing.T) {
		vue := vuego.NewVue(nil)
		vue.RegisterComponent("card", "components/Card.vuego")
		vue.RegisterComponent("card", "components/CardV2.vuego")

		filename, ok := vue.GetComponentFile("card")
		assert.True(t, ok)
		assert.Equal(t, "components/CardV2.vuego", filename)
	})

	t.Run("returns the receiver for chaining", func(t *testing.T) {
		vue := vuego.NewVue(nil)
		assert.Equal(t, vue, vue.RegisterComponent("a", "A.vuego"))
	})
}

func TestVue_GetComponentFile(t *testing.T) {
	vue := vuego.NewVue(nil)
	vue.RegisterComponent("button-primary", "components/ButtonPrimary.vuego")

	t.Run("registered tag", func(t *testing.T) {
		filename, ok := vue.GetComponentFile("button-primary")
		assert.True(t, ok)
		assert.Equal(t, "components/ButtonPrimary.vuego", filename)
	})

	t.Run("unregistered tag", func(t *testing.T) {
		filename, ok := vue.GetComponentFile("div")
		assert.False(t, ok)
		assert.Equal(t, "", filename)
	})

	t.Run("lookup is case sensitive", func(t *testing.T) {
		_, ok := vue.GetComponentFile("Button-Primary")
		assert.False(t, ok)
	})
}

func TestFixtures(t *testing.T) {
	fixtures := os.DirFS("testdata/fixtures")

	// Create a template with components loaded recursively
	tpl := vuego.NewFS(fixtures, vuego.WithComponents(), vuego.WithLessProcessor())

	templates, err := fs.Glob(fixtures, "*.vuego")
	assert.NoError(t, err)

	for _, template := range templates {
		want, err := fs.ReadFile(fixtures, strings.ReplaceAll(template, ".vuego", ".html"))
		assert.NoError(t, err)

		dataBytes, err := fs.ReadFile(fixtures, strings.ReplaceAll(template, ".vuego", ".json"))
		assert.NoError(t, err)

		templateBytes, err := fs.ReadFile(fixtures, template)
		assert.NoError(t, err)

		data := map[string]any{}
		err = json.NewDecoder(bytes.NewReader(dataBytes)).Decode(&data)
		assert.True(t, err == nil, "error decoding template json: %s", template)

		t.Run(template, func(t *testing.T) {
			var got bytes.Buffer
			loaded := tpl.Load(template).Fill(data)
			assert.NoError(t, loaded.Render(t.Context(), &got))
			assert.EqualHTML(t, want, got.Bytes(), templateBytes, dataBytes)
		})
	}
}

func TestVue_RenderFragment(t *testing.T) {
	t.Run("escapes interpolated attribute values", func(t *testing.T) {
		got := renderFragment(t, `<div data-tooltip="{{value}}"></div>`, map[string]any{"value": "A & B"})
		assert.EqualHTML(t, []byte(`<div data-tooltip="A &amp; B"></div>`), []byte(got), nil, nil)
	})

	// A falsey bound value drops the attribute, as Vue does.
	falsey := []struct {
		name     string
		source   string
		data     map[string]any
		attr     string
		rendered bool
	}{
		{name: "false boolean", source: `<input :checked="isDone" />`, data: map[string]any{"isDone": false}, attr: "checked"},
		{name: "true boolean", source: `<input :checked="isDone" />`, data: map[string]any{"isDone": true}, attr: "checked", rendered: true},
		{name: "empty string", source: `<input :disabled="reason" />`, data: map[string]any{"reason": ""}, attr: "disabled"},
		{name: "the string false", source: `<input :disabled="reason" />`, data: map[string]any{"reason": "false"}, attr: "disabled"},
		{name: "zero", source: `<input :disabled="count" />`, data: map[string]any{"count": 0}, attr: "disabled"},
		{name: "nil", source: `<input :disabled="val" />`, data: map[string]any{"val": nil}, attr: "disabled"},
		{name: "non-empty string", source: `<input :disabled="reason" />`, data: map[string]any{"reason": "User not ready"}, attr: "disabled", rendered: true},
		{name: "v-bind prefix", source: `<input v-bind:required="isRequired" />`, data: map[string]any{"isRequired": false}, attr: "required"},
	}

	for _, test := range falsey {
		t.Run("bound attribute "+test.name, func(t *testing.T) {
			got := renderFragment(t, test.source, test.data)
			if test.rendered {
				assert.Contains(t, got, test.attr)
			} else {
				assert.NotContains(t, got, test.attr)
			}
		})
	}
}

func TestVue_StructFieldAccess(t *testing.T) {
	type Address struct {
		City string
	}

	type Person struct {
		Name    string
		Age     int
		Address Address
	}

	t.Run("fields are reachable without a data prefix", func(t *testing.T) {
		got := renderFragment(t, `<div>Name: {{Name}}, Age: {{Age}}</div>`, Person{Name: "Alice", Age: 30})
		assert.Contains(t, got, "Name: Alice")
		assert.Contains(t, got, "Age: 30")
	})

	t.Run("nested fields resolve", func(t *testing.T) {
		got := renderFragment(t, `<div>{{Address.City}}</div>`, Person{Address: Address{City: "New York"}})
		assert.Contains(t, got, "New York")
	})

	t.Run("a map is used as given", func(t *testing.T) {
		got := renderFragment(t, `<div>{{Name}}</div>`, map[string]any{"Name": "Bob"})
		assert.Contains(t, got, "Bob")
	})
}
