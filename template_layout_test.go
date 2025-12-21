package vuego_test

import (
	"bytes"
	"embed"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/titpetric/vuego"
)

//go:embed all:testdata
var testdata embed.FS

func TestRenderLayout(t *testing.T) {
	fsys, err := fs.Sub(testdata, "testdata")
	assert.NoError(t, err)

	renderer := vuego.NewFS(fsys)
	data := map[string]any{
		"content": "Test Content",
	}

	t.Run("blog.vuego", func(t *testing.T) {
		var buf bytes.Buffer
		err := renderer.Load("blog.vuego").Fill(data).Layout(t.Context(), &buf)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, ">Layout: base<")
		assert.Contains(t, output, ">Test Content<")
	})

	t.Run("index.vuego", func(t *testing.T) {
		var buf bytes.Buffer
		err := renderer.Load("index.vuego").Fill(data).Layout(t.Context(), &buf)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, ">Layout: base<")
		assert.Contains(t, output, ">Layout: post<")
		assert.Contains(t, output, ">Test Content<")
	})

	t.Run("no_layout_defaults_to_base", func(t *testing.T) {
		var buf bytes.Buffer
		// fixtures/frontmatter-basic.vuego has no layout param
		err := renderer.Load("fixtures/frontmatter-basic.vuego").Fill(data).Layout(t.Context(), &buf)
		assert.NoError(t, err)

		output := buf.String()
		// When no layout param exists, should still wrap in layouts/base.vuego
		assert.Contains(t, output, ">Layout: base<")
	})
}
