package vuego_test

import (
	"testing"
	"testing/fstest"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/tests/assert"
)

func TestNewLoader(t *testing.T) {
	t.Run("keeps the filesystem it was given", func(t *testing.T) {
		fsys := fstest.MapFS{"test.vuego": &fstest.MapFile{Data: []byte("<div>test</div>")}}

		loader := vuego.NewLoader(fsys)
		assert.NotNil(t, loader)
		assert.Equal(t, fsys, loader.FS)
	})

	t.Run("a nil filesystem is reported on use, not on construction", func(t *testing.T) {
		loader := vuego.NewLoader(nil)
		assert.NotNil(t, loader)

		_, err := loader.Load("any.vuego")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no filesystem configured")
	})
}

func TestLoader_Stat(t *testing.T) {
	loader := vuego.NewLoader(fstest.MapFS{
		"exists.html": &fstest.MapFile{Data: []byte("<div>test</div>")},
	})

	t.Run("file exists", func(t *testing.T) {
		assert.NoError(t, loader.Stat("exists.html"))
	})

	t.Run("file not found", func(t *testing.T) {
		assert.Error(t, loader.Stat("missing.html"))
	})
}

func TestLoader_Load(t *testing.T) {
	loader := vuego.NewLoader(fstest.MapFS{
		"doc.html": &fstest.MapFile{Data: []byte("<html><body><h1>Test</h1></body></html>")},
	})

	t.Run("valid document", func(t *testing.T) {
		nodes, err := loader.Load("doc.html")
		assert.NoError(t, err)
		assert.Greater(t, len(nodes), 0)
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := loader.Load("missing.html")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error reading missing.html")
	})

	t.Run("malformed markup still parses", func(t *testing.T) {
		loader := vuego.NewLoader(fstest.MapFS{
			"bad.html": &fstest.MapFile{Data: []byte("<<<>>>")},
		})
		nodes, err := loader.Load("bad.html")
		assert.NoError(t, err)
		assert.NotNil(t, nodes)
	})
}

func TestLoader_LoadFragment(t *testing.T) {
	loader := vuego.NewLoader(fstest.MapFS{
		"fragment.html": &fstest.MapFile{Data: []byte("<div>content</div>")},
		"document.html": &fstest.MapFile{Data: []byte("<html><body><div>doc</div></body></html>")},
	})

	t.Run("fragment without an html tag", func(t *testing.T) {
		nodes, err := loader.LoadFragment("fragment.html")
		assert.NoError(t, err)
		assert.Greater(t, len(nodes), 0)
		assert.Equal(t, "div", nodes[0].Data)
	})

	t.Run("full document", func(t *testing.T) {
		nodes, err := loader.LoadFragment("document.html")
		assert.NoError(t, err)
		assert.NotNil(t, nodes)
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := loader.LoadFragment("missing.html")
		assert.Error(t, err)
	})

	// Front matter is stripped from what the parser sees.
	frontMatter := []struct {
		name     string
		content  string
		wantTag  string
		wantText string
	}{
		{
			name:     "front matter is removed",
			content:  "---\ntitle: Test\n---\n<div>body</div>",
			wantTag:  "div",
			wantText: "body",
		},
		{
			name:     "no closing delimiter leaves the content alone",
			content:  "---\ntitle: Test\n<div>body</div>",
			wantTag:  "",
			wantText: "---\ntitle: Test\n",
		},
		{
			name:     "content that does not open with --- has no front matter",
			content:  "<div>body</div>",
			wantTag:  "div",
			wantText: "body",
		},
	}

	for _, test := range frontMatter {
		t.Run(test.name, func(t *testing.T) {
			loader := vuego.NewLoader(fstest.MapFS{
				"page.vuego": &fstest.MapFile{Data: []byte(test.content)},
			})

			nodes, err := loader.LoadFragment("page.vuego")
			assert.NoError(t, err)
			assert.Greater(t, len(nodes), 0)
			if test.wantTag != "" {
				assert.Equal(t, test.wantTag, nodes[0].Data)
			}
			assert.Contains(t, nodeText(nodes), test.wantText)
		})
	}

	t.Run("front matter that is not YAML is an error", func(t *testing.T) {
		loader := vuego.NewLoader(fstest.MapFS{
			"page.vuego": &fstest.MapFile{Data: []byte("---\n\ttitle: [\n---\n<div>body</div>")},
		})

		_, err := loader.LoadFragment("page.vuego")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "front-matter YAML")
	})
}
