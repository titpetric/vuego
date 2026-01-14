package vuego_test

import (
	"bytes"
	"embed"
	"io/fs"
	"testing"
	"testing/fstest"

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

	t.Run("v_html_on_element_in_first_template", func(t *testing.T) {
		var buf bytes.Buffer
		err := renderer.Load("article-test.vuego").Fill(data).Layout(t.Context(), &buf)
		assert.NoError(t, err)

		output := buf.String()
		t.Logf("Output: %s", output)
		// The <article v-html="content"></article> should render with the article's HTML output
		assert.Contains(t, output, "<article><p>Article content here</p></article>", "article with v-html should contain the rendered page content")
	})

	t.Run("page_with_post_layout_renders_content", func(t *testing.T) {
		var buf bytes.Buffer
		err := renderer.Load("blog-post-page.vuego").Fill(data).Layout(t.Context(), &buf)
		assert.NoError(t, err)

		output := buf.String()
		t.Logf("Output:\n%s", output)
		// blog-post-page.vuego -> layouts/post.vuego (which has <p v-html="content"></p>) -> layouts/base.vuego
		assert.Contains(t, output, "<p>This is a test blog post page</p>", "page content should be rendered in post layout")
	})

	t.Run("relative_layout_path", func(t *testing.T) {
		inlineFS := &fstest.MapFS{
			"pages/page.vuego": &fstest.MapFile{Data: []byte(`---
layout: base.vuego
---
<p>Page content</p>`)},
			"pages/base.vuego": &fstest.MapFile{Data: []byte(`<div class="wrapper" v-html="content"></div>`)},
		}
		inlineRenderer := vuego.NewFS(inlineFS)

		var buf bytes.Buffer
		err := inlineRenderer.Load("pages/page.vuego").Fill(nil).Layout(t.Context(), &buf)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, `<div class="wrapper">`, "relative layout should be resolved")
		assert.Contains(t, output, "<p>Page content</p>", "page content should be rendered")
	})

	t.Run("relative_layout_fallback_to_layouts_dir", func(t *testing.T) {
		inlineFS := &fstest.MapFS{
			"pages/page.vuego": &fstest.MapFile{Data: []byte(`---
layout: main
---
<p>Page content</p>`)},
			"layouts/main.vuego": &fstest.MapFile{Data: []byte(`<main v-html="content"></main>`)},
		}
		inlineRenderer := vuego.NewFS(inlineFS)

		var buf bytes.Buffer
		err := inlineRenderer.Load("pages/page.vuego").Fill(nil).Layout(t.Context(), &buf)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "<main>", "fallback to layouts/ should work")
		assert.Contains(t, output, "<p>Page content</p>", "page content should be rendered")
	})

	t.Run("relative_layout_without_extension", func(t *testing.T) {
		inlineFS := &fstest.MapFS{
			"pages/page.vuego": &fstest.MapFile{Data: []byte(`---
layout: wrapper
---
<p>Content</p>`)},
			"pages/wrapper.vuego": &fstest.MapFile{Data: []byte(`<section v-html="content"></section>`)},
		}
		inlineRenderer := vuego.NewFS(inlineFS)

		var buf bytes.Buffer
		err := inlineRenderer.Load("pages/page.vuego").Fill(nil).Layout(t.Context(), &buf)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "<section>", "relative layout without extension should work")
	})

	t.Run("struct_data_converted_to_map", func(t *testing.T) {
		// Test that structs are automatically converted to maps via toMapData
		type TestData struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		}

		// Create a Vue instance with in-memory filesystem for this test
		inlineFS := &fstest.MapFS{
			"test.vuego": &fstest.MapFile{Data: []byte("<p>{{ title }}</p>\n<p>{{ description }}</p>")},
		}
		inlineRenderer := vuego.NewFS(inlineFS)

		var buf bytes.Buffer
		testData := TestData{
			Title:       "Test Title",
			Description: "Test Description",
		}
		err := inlineRenderer.Load("test.vuego").Fill(testData).Render(t.Context(), &buf)
		assert.NoError(t, err)

		output := buf.String()
		// Struct fields should be accessible via JSON tag names
		assert.Contains(t, output, "<p>Test Title</p>", "title field should be rendered")
		assert.Contains(t, output, "<p>Test Description</p>", "description field should be rendered")
	})
}
