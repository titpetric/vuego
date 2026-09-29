package vuego_test

import (
	"bytes"
	"os"
	"testing"
	"testing/fstest"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/tests/assert"
)

func TestLoadOption(t *testing.T) {
	// A LoadOption is applied by NewFS in the order given, so a later option
	// sees what an earlier one did.
	t.Run("options are applied in order", func(t *testing.T) {
		tpl := vuego.NewFS(nil,
			vuego.WithFuncs(vuego.FuncMap{"tag": func() string { return "first" }}),
			vuego.WithFuncs(vuego.FuncMap{"tag": func() string { return "second" }}),
		)

		var buf bytes.Buffer
		assert.NoError(t, tpl.RenderString(t.Context(), &buf, `<p>{{ tag() }}</p>`))
		assert.Contains(t, buf.String(), "second")
	})

	t.Run("no options leaves the defaults in place", func(t *testing.T) {
		tpl := vuego.NewFS(nil)

		var buf bytes.Buffer
		assert.NoError(t, tpl.RenderString(t.Context(), &buf, `<p>{{ name | upper }}</p>`))
		assert.Contains(t, buf.String(), "")
	})

	t.Run("New passes options through to NewFS", func(t *testing.T) {
		tpl := vuego.New(vuego.WithFuncs(vuego.FuncMap{"greet": func() string { return "hi" }}))

		var buf bytes.Buffer
		assert.NoError(t, tpl.RenderString(t.Context(), &buf, `<p>{{ greet() }}</p>`))
		assert.Contains(t, buf.String(), "hi")
	})
}

func TestWithFS(t *testing.T) {
	fsys := fstest.MapFS{"page.vuego": {Data: []byte(`<h1>from fs</h1>`)}}

	t.Run("a template renders from the filesystem", func(t *testing.T) {
		tpl := vuego.New(vuego.WithFS(fsys))

		var buf bytes.Buffer
		assert.NoError(t, tpl.Load("page.vuego").Render(t.Context(), &buf))
		assert.Contains(t, buf.String(), "from fs")
	})

	t.Run("replaces a filesystem passed to NewFS", func(t *testing.T) {
		tpl := vuego.NewFS(
			fstest.MapFS{"page.vuego": {Data: []byte(`<h1>first</h1>`)}},
			vuego.WithFS(fsys),
		)

		var buf bytes.Buffer
		assert.NoError(t, tpl.Load("page.vuego").Render(t.Context(), &buf))
		assert.Contains(t, buf.String(), "from fs")
	})

	t.Run("without it a file cannot be loaded", func(t *testing.T) {
		var buf bytes.Buffer
		assert.Error(t, vuego.New().Load("page.vuego").Render(t.Context(), &buf))
	})
}

func TestWithFuncs(t *testing.T) {
	t.Run("a custom func is callable", func(t *testing.T) {
		tpl := vuego.New(vuego.WithFuncs(vuego.FuncMap{
			"customFunc": func(s string) string { return "custom: " + s },
		}))

		var buf bytes.Buffer
		assert.NoError(t, tpl.RenderString(t.Context(), &buf, `<p>{{ name | customFunc }}</p>`))
		assert.Contains(t, buf.String(), "custom: ")
	})

	t.Run("the built-ins are still present", func(t *testing.T) {
		tpl := vuego.New(vuego.WithFuncs(vuego.FuncMap{"noop": func(v any) any { return v }}))

		var buf bytes.Buffer
		assert.NoError(t, tpl.Fill(map[string]any{"name": "ada"}).RenderString(t.Context(), &buf, `<p>{{ name | upper }}</p>`))
		assert.Contains(t, buf.String(), "ADA")
	})

	t.Run("a built-in can be replaced", func(t *testing.T) {
		tpl := vuego.New(vuego.WithFuncs(vuego.FuncMap{"upper": func(any) any { return "replaced" }}))

		var buf bytes.Buffer
		assert.NoError(t, tpl.Fill(map[string]any{"name": "ada"}).RenderString(t.Context(), &buf, `<p>{{ name | upper }}</p>`))
		assert.Contains(t, buf.String(), "replaced")
	})
}

func TestWithLessProcessor(t *testing.T) {
	root := os.DirFS("testdata")

	const lessStyle = `<style type="text/css+less">@c: red; a { color: @c; }</style>`

	t.Run("a less style tag is compiled to CSS", func(t *testing.T) {
		tpl := vuego.NewFS(root, vuego.WithLessProcessor())

		var buf bytes.Buffer
		assert.NoError(t, tpl.RenderString(t.Context(), &buf, lessStyle))
		assert.Contains(t, buf.String(), `type="text/css"`)
		assert.Contains(t, buf.String(), "color: red")
		assert.NotContains(t, buf.String(), "text/css+less")
	})

	t.Run("without it the less source is left in place", func(t *testing.T) {
		tpl := vuego.NewFS(root)

		var buf bytes.Buffer
		assert.NoError(t, tpl.RenderString(t.Context(), &buf, lessStyle))
		assert.Contains(t, buf.String(), "text/css+less")
		assert.Contains(t, buf.String(), "@c: red")
	})

	t.Run("registering it twice still compiles once", func(t *testing.T) {
		tpl := vuego.NewFS(root, vuego.WithLessProcessor(), vuego.WithLessProcessor())

		var buf bytes.Buffer
		assert.NoError(t, tpl.RenderString(t.Context(), &buf, lessStyle))
		assert.Contains(t, buf.String(), "color: red")
	})
}

func TestWithProcessor(t *testing.T) {
	t.Run("the processor runs during a render", func(t *testing.T) {
		var calls []string
		tpl := vuego.New(vuego.WithProcessor(&recordingProcessor{name: "a", calls: &calls}))

		var buf bytes.Buffer
		assert.NoError(t, tpl.RenderString(t.Context(), &buf, `<div><mark>x</mark></div>`))
		assert.Equal(t, []string{"a:pre", "a:post"}, calls)
		assert.Contains(t, buf.String(), "<em>x</em>")
	})

	t.Run("several processors run in order", func(t *testing.T) {
		var calls []string
		tpl := vuego.New(
			vuego.WithProcessor(&recordingProcessor{name: "first", calls: &calls}),
			vuego.WithProcessor(&recordingProcessor{name: "second", calls: &calls}),
		)

		var buf bytes.Buffer
		assert.NoError(t, tpl.RenderString(t.Context(), &buf, `<div>x</div>`))
		assert.Equal(t, []string{"first:pre", "second:pre", "first:post", "second:post"}, calls)
	})

	t.Run("a processor error reaches the caller", func(t *testing.T) {
		tpl := vuego.New(vuego.WithProcessor(&errorProcessor{failOn: "post", err: errWriteFailed}))

		var buf bytes.Buffer
		assert.ErrorIs(t, tpl.RenderString(t.Context(), &buf, `<div>x</div>`), errWriteFailed)
	})
}

func TestWithComponents(t *testing.T) {
	fsys := fstest.MapFS{
		"components/MyCard.vuego":    {Data: []byte(`<article>card</article>`)},
		"components/nested/Hi.vuego": {Data: []byte(`<span>hi</span>`)},
		"components/notes.txt":       {Data: []byte(`ignored`)},
	}

	t.Run("a top-level component gets a kebab-case tag", func(t *testing.T) {
		tpl := vuego.NewFS(fsys, vuego.WithComponents())

		var buf bytes.Buffer
		assert.NoError(t, tpl.RenderString(t.Context(), &buf, `<my-card></my-card>`))
		assert.Contains(t, buf.String(), "<article>card</article>")
	})

	t.Run("a nested component joins its directory into the tag", func(t *testing.T) {
		tpl := vuego.NewFS(fsys, vuego.WithComponents())

		var buf bytes.Buffer
		assert.NoError(t, tpl.RenderString(t.Context(), &buf, `<nested-hi></nested-hi>`))
		assert.Contains(t, buf.String(), "<span>hi</span>")
	})

	t.Run("a non-vuego file is not registered", func(t *testing.T) {
		tpl := vuego.NewFS(fsys, vuego.WithComponents())

		var buf bytes.Buffer
		assert.NoError(t, tpl.RenderString(t.Context(), &buf, `<notes></notes>`))
		assert.NotContains(t, buf.String(), "ignored")
	})

	t.Run("a filesystem with no components directory is not an error", func(t *testing.T) {
		tpl := vuego.NewFS(fstest.MapFS{}, vuego.WithComponents())

		var buf bytes.Buffer
		assert.NoError(t, tpl.RenderString(t.Context(), &buf, `<div>x</div>`))
	})

	t.Run("a nil filesystem is not an error", func(t *testing.T) {
		tpl := vuego.New(vuego.WithComponents())

		var buf bytes.Buffer
		assert.NoError(t, tpl.RenderString(t.Context(), &buf, `<div>x</div>`))
	})
}

func TestView(t *testing.T) {
	type IndexData struct {
		Title string `json:"title"`
	}

	t.Run("binds a filename to a data type", func(t *testing.T) {
		tpl := vuego.NewFS(fstest.MapFS{
			"index.vuego": {Data: []byte(`<h1>{{ title }}</h1>`)},
		})

		var buf bytes.Buffer
		assert.NoError(t, vuego.View(tpl, "index.vuego", IndexData{Title: "Hello"}).Render(t.Context(), &buf))
		assert.Contains(t, buf.String(), "Hello")
	})

	t.Run("a missing file surfaces on render", func(t *testing.T) {
		tpl := vuego.NewFS(fstest.MapFS{})

		var buf bytes.Buffer
		assert.Error(t, vuego.View(tpl, "missing.vuego", IndexData{}).Render(t.Context(), &buf))
	})
}
