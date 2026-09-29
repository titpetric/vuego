package vuego_test

import (
	"bytes"
	"os"
	"testing"
	"testing/fstest"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/tests/assert"
)

func TestNew(t *testing.T) {
	t.Run("renders without a filesystem", func(t *testing.T) {
		var buf bytes.Buffer
		assert.NoError(t, vuego.New().RenderString(t.Context(), &buf, "<p>text</p>"))
		assert.Contains(t, buf.String(), "text")
	})
}

func TestNewFS(t *testing.T) {
	t.Run("loads from the filesystem it is given", func(t *testing.T) {
		tpl := vuego.NewFS(fstest.MapFS{"page.vuego": {Data: []byte(`<h1>page</h1>`)}})

		var buf bytes.Buffer
		assert.NoError(t, tpl.Load("page.vuego").Render(t.Context(), &buf))
		assert.Contains(t, buf.String(), "page")
	})

	t.Run("seeds the data from theme.yml and data/", func(t *testing.T) {
		tpl := vuego.NewFS(os.DirFS("testdata/LoadConfig"))

		var buf bytes.Buffer
		assert.NoError(t, tpl.RenderString(t.Context(), &buf, `<p>{{ theme.header.title }}</p>`))
		assert.Contains(t, buf.String(), "Data Theme")
	})
}

func TestTemplate(t *testing.T) {
	fixtures := os.DirFS("testdata/fixtures")

	t.Run("a typed view binds a data model to a file", func(t *testing.T) {
		type ViewData struct {
			Tooltip string `json:"tooltip"`
		}

		var buf bytes.Buffer
		view := vuego.View[ViewData](vuego.NewFS(fixtures), "attribute-bind.vuego", ViewData{Tooltip: "Vuego"})
		assert.NoError(t, view.Render(t.Context(), &buf))
		assert.Contains(t, buf.String(), "Vuego")
	})

	t.Run("New returns an independent copy", func(t *testing.T) {
		base := vuego.NewFS(fixtures).Assign("shared", "base")

		derived := base.New().Assign("own", "derived")

		assert.Equal(t, "base", derived.Get("shared"))
		assert.Equal(t, "", base.Get("own"))
	})
}

func TestTemplate_Fill(t *testing.T) {
	fixtures := os.DirFS("testdata/fixtures")

	t.Run("sets the variables and is chainable", func(t *testing.T) {
		tpl := vuego.NewFS(fixtures)

		assert.Equal(t, tpl, tpl.Fill(map[string]any{"message": "Hello World"}))
		assert.Equal(t, "Hello World", tpl.Get("message"))
	})

	t.Run("replaces what Assign set", func(t *testing.T) {
		tpl := vuego.NewFS(fixtures)
		tpl.Assign("name", "Alice")
		tpl.Fill(map[string]any{"name": "Bob"})

		assert.Equal(t, "Bob", tpl.Get("name"))
	})

	t.Run("discards keys Assign set that Fill does not carry", func(t *testing.T) {
		tpl := vuego.NewFS(fixtures)
		tpl.Assign("foo", "bar").Assign("baz", "qux").Fill(map[string]any{"message": "Hello"})

		assert.Equal(t, "", tpl.Get("foo"))
		assert.Equal(t, "", tpl.Get("baz"))
		assert.Equal(t, "Hello", tpl.Get("message"))
	})

	t.Run("a struct is keyed by json tag", func(t *testing.T) {
		type data struct {
			Message string `json:"message"`
		}

		tpl := vuego.NewFS(fixtures).Fill(data{Message: "Hello"})
		assert.Equal(t, "Hello", tpl.Get("message"))
	})
}

func TestTemplate_Assign(t *testing.T) {
	fixtures := os.DirFS("testdata/fixtures")

	t.Run("sets one variable and is chainable", func(t *testing.T) {
		tpl := vuego.NewFS(fixtures)

		assert.Equal(t, tpl, tpl.Assign("name", "Alice"))
		assert.Equal(t, "Alice", tpl.Get("name"))
	})

	t.Run("assignments after Fill are kept", func(t *testing.T) {
		tpl := vuego.NewFS(fixtures)
		tpl.Fill(map[string]any{"message": "Hello"}).Assign("foo", "bar").Assign("baz", "qux")

		assert.Equal(t, "bar", tpl.Get("foo"))
		assert.Equal(t, "qux", tpl.Get("baz"))
		assert.Equal(t, "Hello", tpl.Get("message"))
	})

	t.Run("a second assignment replaces the first", func(t *testing.T) {
		tpl := vuego.NewFS(fixtures)
		tpl.Assign("name", "Alice").Assign("name", "Bob")

		assert.Equal(t, "Bob", tpl.Get("name"))
	})
}

func TestTemplate_Get(t *testing.T) {
	tpl := vuego.NewFS(os.DirFS("testdata/fixtures"))
	tpl.Assign("message", "Hello World")
	tpl.Assign("count", 42)
	tpl.Assign("isTrue", true)
	tpl.Assign("isFalse", false)
	tpl.Assign("nothing", nil)
	tpl.Assign("list", []string{"a", "b"})

	tests := []struct {
		key  string
		want string
	}{
		{key: "message", want: "Hello World"},
		{key: "count", want: "42"},
		{key: "isTrue", want: "true"},
		{key: "isFalse", want: "false"},
		{key: "list", want: "[a b]"},
		{key: "nothing", want: ""},
		{key: "nonexistent", want: ""},
	}

	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			assert.Equal(t, test.want, tpl.Get(test.key))
		})
	}
}

func TestTemplate_Load(t *testing.T) {
	fixtures := os.DirFS("testdata/fixtures")

	t.Run("front matter becomes template data", func(t *testing.T) {
		tpl := vuego.NewFS(fixtures).Load("frontmatter-basic.vuego")
		assert.Equal(t, "From Front-Matter", tpl.Get("title"))
	})

	t.Run("returns a copy, leaving the receiver unloaded", func(t *testing.T) {
		base := vuego.NewFS(fixtures)
		_ = base.Load("frontmatter-basic.vuego")

		var buf bytes.Buffer
		assert.Error(t, base.Render(t.Context(), &buf))
	})

	t.Run("data assigned on the base is carried over", func(t *testing.T) {
		tpl := vuego.NewFS(fixtures).Assign("title", "Hello").Load("interpolation-basic.vuego")

		var buf bytes.Buffer
		assert.NoError(t, tpl.Render(t.Context(), &buf))
		assert.Contains(t, buf.String(), "Hello")
	})
}

func TestLoadConfig(t *testing.T) {
	templateFS := os.DirFS("testdata/LoadConfig")

	t.Run("data/ overrides the root theme.yml", func(t *testing.T) {
		var buf bytes.Buffer
		assert.NoError(t, vuego.NewFS(templateFS).Load("index.vuego").Render(t.Context(), &buf))

		assert.Contains(t, buf.String(), "Data Theme")
		assert.Contains(t, buf.String(), "From data folder")
	})

	t.Run("every yaml file under data/ is loaded", func(t *testing.T) {
		var buf bytes.Buffer
		assert.NoError(t, vuego.NewFS(templateFS).Load("index.vuego").Render(t.Context(), &buf))

		assert.Contains(t, buf.String(), "Home")
		assert.Contains(t, buf.String(), "About")
	})

	t.Run("Fill overrides the loaded config", func(t *testing.T) {
		type viewData struct {
			Theme struct {
				Header struct {
					Title    string `json:"title"`
					Subtitle string `json:"subtitle"`
				} `json:"header"`
			} `json:"theme"`
		}

		data := viewData{}
		data.Theme.Header.Title = "Override Title"
		data.Theme.Header.Subtitle = "Override Subtitle"

		var buf bytes.Buffer
		assert.NoError(t, vuego.NewFS(templateFS).Load("index.vuego").Fill(data).Render(t.Context(), &buf))

		assert.Contains(t, buf.String(), "Override Title")
		assert.Contains(t, buf.String(), "Override Subtitle")
	})

	t.Run("a filesystem without config files is not an error", func(t *testing.T) {
		tpl := vuego.NewFS(fstest.MapFS{"page.vuego": {Data: []byte(`<p>x</p>`)}})

		var buf bytes.Buffer
		assert.NoError(t, tpl.Load("page.vuego").Render(t.Context(), &buf))
	})

	t.Run("invalid yaml is skipped", func(t *testing.T) {
		tpl := vuego.NewFS(fstest.MapFS{
			"theme.yml":  {Data: []byte("\ttitle: [")},
			"page.vuego": {Data: []byte(`<p>x</p>`)},
		})

		var buf bytes.Buffer
		assert.NoError(t, tpl.Load("page.vuego").Render(t.Context(), &buf))
		assert.Equal(t, "", tpl.Get("title"))
	})
}
