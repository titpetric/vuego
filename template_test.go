package vuego_test

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/testing/assert"
)

func TestTemplate(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.NewFS(templateFS)
	assert.NotNil(t, tmpl)

	t.Run("type binding view", func(t *testing.T) {
		type ViewData struct {
			Tooltip string `json:"tooltip"`
		}
		var buf bytes.Buffer

		model := ViewData{"Vuego"}
		view := vuego.View[ViewData](tmpl, "attribute-bind.vuego", model)
		err := view.Render(t.Context(), &buf)

		assert.NoError(t, err)
		assert.NotEmpty(t, buf.String())
	})
}

func TestTemplate_Fill(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.NewFS(templateFS)

	// Fill should be chainable
	result := tmpl.Fill(map[string]any{
		"message": "Hello World",
	})
	assert.NotNil(t, result)
	assert.Equal(t, tmpl, result)

	// Verify the variables were set
	assert.Equal(t, "Hello World", tmpl.Get("message"))
}

func TestTemplate_Assign(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.NewFS(templateFS)

	// Assign should be chainable
	result := tmpl.Assign("name", "Alice")
	assert.NotNil(t, result)
	assert.Equal(t, tmpl, result)

	// Verify the variable was set
	assert.Equal(t, "Alice", tmpl.Get("name"))
}

func TestTemplate_Get(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.NewFS(templateFS)

	// Set and get
	tmpl.Assign("key", "value")
	assert.Equal(t, "value", tmpl.Get("key"))

	// Get non-existent variable
	assert.Equal(t, "", tmpl.Get("nonexistent"))

	// String value
	tmpl.Assign("message", "Hello World")
	assert.Equal(t, "Hello World", tmpl.Get("message"))

	// Non-string value returns string (number, bool)
	tmpl.Assign("count", 42)
	tmpl.Assign("isTrue", true)
	tmpl.Assign("isFalse", false)

	assert.Equal(t, "42", tmpl.Get("count"))
	assert.Equal(t, "true", tmpl.Get("isTrue"))
	assert.Equal(t, "false", tmpl.Get("isFalse"))
}

func TestTemplate_Render(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.NewFS(templateFS)

	tmpl.Assign("message", "Hello")

	buf := &bytes.Buffer{}
	err := tmpl.Load("interpolation-basic.vuego").Render(t.Context(), buf)
	assert.NoError(t, err)

	// Should have rendered something
	assert.Greater(t, buf.Len(), 0)
}

func TestTemplate_RenderCancelledContext(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.NewFS(templateFS)

	tmpl.Assign("message", "Hello")

	// Create a cancelled context
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	buf := &bytes.Buffer{}
	err := tmpl.Load("interpolation-basic.vuego").Render(ctx, buf)
	assert.Equal(t, context.Canceled, err)
}

func TestTemplate_FrontMatterVariables(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.NewFS(templateFS)

	// Render the template to load front-matter
	buf := &bytes.Buffer{}
	err := tmpl.Load("frontmatter-basic.vuego").Render(t.Context(), buf)
	assert.NoError(t, err)

	// Note: Front-matter variables are merged during rendering, not during Load
	// This test demonstrates that rendering works with front-matter
	assert.Greater(t, buf.Len(), 0)
}

func TestTemplate_Chaining(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	data := map[string]any{
		"message": "Hello",
	}

	tmpl := vuego.NewFS(templateFS)

	// Test method chaining, Assign first
	tmpl.Assign("foo", "bar").Assign("baz", "qux").Fill(data)

	// Fill resets state.
	assert.Equal(t, "", tmpl.Get("foo"))
	assert.Equal(t, "", tmpl.Get("baz"))
	assert.Equal(t, "Hello", tmpl.Get("message"))

	// Test method chaining, Fill first
	tmpl.Fill(data).Assign("foo", "bar").Assign("baz", "qux")

	// Fill resets state.
	assert.Equal(t, "bar", tmpl.Get("foo"))
	assert.Equal(t, "qux", tmpl.Get("baz"))
	assert.Equal(t, "Hello", tmpl.Get("message"))
}

func TestTemplate_FillOverrides(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.NewFS(templateFS)

	// Set initial value
	tmpl.Assign("name", "Alice")
	assert.Equal(t, "Alice", tmpl.Get("name"))

	// Fill with new values
	tmpl.Fill(map[string]any{
		"name": "Bob",
	})
	assert.Equal(t, "Bob", tmpl.Get("name"))
}

func TestTemplate_RenderWithFrontMatter(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.NewFS(templateFS)

	buf := &bytes.Buffer{}
	err := tmpl.Load("frontmatter-basic.vuego").Render(t.Context(), buf)
	assert.NoError(t, err)

	// Should have rendered HTML with the title from front-matter
	assert.Greater(t, buf.Len(), 0)
	output := buf.String()
	assert.Contains(t, output, "From Front-Matter")
}

func TestTemplate_RenderString(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.NewFS(templateFS)
	tmpl.Assign("message", "Hello from String")

	buf := &bytes.Buffer{}
	err := tmpl.RenderString(t.Context(), buf, "<p>{{ message }}</p>")
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Hello from String")
}

func TestTemplate_RenderByte(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.NewFS(templateFS)
	tmpl.Assign("message", "Hello from Bytes")

	buf := &bytes.Buffer{}
	templateBytes := []byte("<p>{{ message }}</p>")
	err := tmpl.RenderByte(t.Context(), buf, templateBytes)
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Hello from Bytes")
}

func TestTemplate_RenderReader(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.NewFS(templateFS)
	tmpl.Assign("message", "Hello from Reader")

	buf := &bytes.Buffer{}
	reader := bytes.NewReader([]byte("<p>{{ message }}</p>"))
	err := tmpl.RenderReader(t.Context(), buf, reader)
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Hello from Reader")
}

func TestTemplate_RenderReaderFromFile(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.NewFS(templateFS)
	tmpl.Assign("message", "Hello from File Reader")

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "template_*.html")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("<p>{{ message }}</p>")
	assert.NoError(t, err)
	tmpFile.Close()

	// Reopen for reading
	file, err := os.Open(tmpFile.Name())
	assert.NoError(t, err)
	defer file.Close()

	buf := &bytes.Buffer{}
	err = tmpl.RenderReader(t.Context(), buf, file)
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Hello from File Reader")
}

func TestNew(t *testing.T) {
	tmpl := vuego.New()
	assert.NotNil(t, tmpl)
}

func TestNew_WithFS(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.New(vuego.WithFS(templateFS))
	assert.NotNil(t, tmpl)

	tmpl.Assign("message", "Hello")

	buf := &bytes.Buffer{}
	err := tmpl.Load("interpolation-basic.vuego").Render(t.Context(), buf)
	assert.NoError(t, err)
	assert.Greater(t, buf.Len(), 0)
}

func TestNew_RenderWithoutFilesystemFails(t *testing.T) {
	tmpl := vuego.New()
	tmpl.Assign("message", "Hello")

	buf := &bytes.Buffer{}
	err := tmpl.Load("interpolation-basic.vuego").Render(t.Context(), buf)
	assert.Equal(t, "error reading interpolation-basic.vuego: no filesystem configured", err.Error())
}

func TestNew_RenderStringWithoutFilesystem(t *testing.T) {
	tmpl := vuego.New()
	tmpl.Assign("message", "Hello from String")

	buf := &bytes.Buffer{}
	err := tmpl.RenderString(t.Context(), buf, "<p>{{ message }}</p>")
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Hello from String")
}

func TestLoadConfig(t *testing.T) {
	templateFS := os.DirFS("testdata/LoadConfig")

	tmpl := vuego.NewFS(templateFS)
	assert.NotNil(t, tmpl)

	buf := &bytes.Buffer{}
	err := tmpl.Load("index.vuego").Render(t.Context(), buf)
	assert.NoError(t, err)

	output := buf.String()
	// data/theme.yml overrides root theme.yml
	assert.Contains(t, output, "Data Theme")
	assert.Contains(t, output, "From data folder")
	// data/menu.yml is loaded
	assert.Contains(t, output, "Home")
	assert.Contains(t, output, "About")
}

func TestLoadConfig_MergesWithFill(t *testing.T) {
	templateFS := os.DirFS("testdata/LoadConfig")

	tmpl := vuego.NewFS(templateFS)

	// Fill should override LoadConfig values
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

	buf := &bytes.Buffer{}
	err := tmpl.Load("index.vuego").Fill(data).Render(t.Context(), buf)
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Override Title")
	assert.Contains(t, output, "Override Subtitle")
}
