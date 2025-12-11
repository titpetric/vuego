package vuego_test

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego"
)

func TestLoad(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.NewFS(templateFS)
	require.NotNil(t, tmpl)
}

func TestTemplate_Fill(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.NewFS(templateFS)

	// Fill should be chainable
	result := tmpl.Fill(map[string]any{
		"message": "Hello World",
	})
	require.NotNil(t, result)
	require.Equal(t, tmpl, result)

	// Verify the variables were set
	require.Equal(t, "Hello World", tmpl.Get("message"))
}

func TestTemplate_Assign(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.NewFS(templateFS)

	// Assign should be chainable
	result := tmpl.Assign("name", "Alice")
	require.NotNil(t, result)
	require.Equal(t, tmpl, result)

	// Verify the variable was set
	require.Equal(t, "Alice", tmpl.Get("name"))
}

func TestTemplate_Get(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.NewFS(templateFS)

	// Set and get
	tmpl.Assign("key", "value")
	require.Equal(t, "value", tmpl.Get("key"))

	// Get non-existent variable
	require.Equal(t, "", tmpl.Get("nonexistent"))

	// String value
	tmpl.Assign("message", "Hello World")
	require.Equal(t, "Hello World", tmpl.Get("message"))

	// Non-string value returns string (number, bool)
	tmpl.Assign("count", 42)
	tmpl.Assign("isTrue", true)
	tmpl.Assign("isFalse", false)

	require.Equal(t, "42", tmpl.Get("count"))
	require.Equal(t, "true", tmpl.Get("isTrue"))
	require.Equal(t, "false", tmpl.Get("isFalse"))
}

func TestTemplate_Render(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.NewFS(templateFS)

	tmpl.Assign("message", "Hello")

	buf := &bytes.Buffer{}
	err := tmpl.Load("interpolation-basic.vuego").Render(context.Background(), buf)
	require.NoError(t, err)

	// Should have rendered something
	require.Greater(t, buf.Len(), 0)
}

func TestTemplate_RenderCancelledContext(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.NewFS(templateFS)

	tmpl.Assign("message", "Hello")

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	buf := &bytes.Buffer{}
	err := tmpl.Load("interpolation-basic.vuego").Render(ctx, buf)
	require.Equal(t, context.Canceled, err)
}

func TestTemplate_FrontMatterVariables(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.NewFS(templateFS)

	// Render the template to load front-matter
	buf := &bytes.Buffer{}
	err := tmpl.Load("frontmatter-basic.vuego").Render(context.Background(), buf)
	require.NoError(t, err)

	// Note: Front-matter variables are merged during rendering, not during Load
	// This test demonstrates that rendering works with front-matter
	require.Greater(t, buf.Len(), 0)
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
	require.Equal(t, "", tmpl.Get("foo"))
	require.Equal(t, "", tmpl.Get("baz"))
	require.Equal(t, "Hello", tmpl.Get("message"))

	// Test method chaining, Fill first
	tmpl.Fill(data).Assign("foo", "bar").Assign("baz", "qux")

	// Fill resets state.
	require.Equal(t, "bar", tmpl.Get("foo"))
	require.Equal(t, "qux", tmpl.Get("baz"))
	require.Equal(t, "Hello", tmpl.Get("message"))
}

func TestTemplate_FillOverrides(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.NewFS(templateFS)

	// Set initial value
	tmpl.Assign("name", "Alice")
	require.Equal(t, "Alice", tmpl.Get("name"))

	// Fill with new values
	tmpl.Fill(map[string]any{
		"name": "Bob",
	})
	require.Equal(t, "Bob", tmpl.Get("name"))
}

func TestTemplate_RenderWithFrontMatter(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.NewFS(templateFS)

	buf := &bytes.Buffer{}
	err := tmpl.Load("frontmatter-basic.vuego").Render(context.Background(), buf)
	require.NoError(t, err)

	// Should have rendered HTML with the title from front-matter
	require.Greater(t, buf.Len(), 0)
	output := buf.String()
	require.Contains(t, output, "From Front-Matter")
}

func TestTemplate_RenderString(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.NewFS(templateFS)
	tmpl.Assign("message", "Hello from String")

	buf := &bytes.Buffer{}
	err := tmpl.RenderString(context.Background(), buf, "<p>{{ message }}</p>")
	require.NoError(t, err)

	output := buf.String()
	require.Contains(t, output, "Hello from String")
}

func TestTemplate_RenderByte(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.NewFS(templateFS)
	tmpl.Assign("message", "Hello from Bytes")

	buf := &bytes.Buffer{}
	templateBytes := []byte("<p>{{ message }}</p>")
	err := tmpl.RenderByte(context.Background(), buf, templateBytes)
	require.NoError(t, err)

	output := buf.String()
	require.Contains(t, output, "Hello from Bytes")
}

func TestTemplate_RenderReader(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.NewFS(templateFS)
	tmpl.Assign("message", "Hello from Reader")

	buf := &bytes.Buffer{}
	reader := bytes.NewReader([]byte("<p>{{ message }}</p>"))
	err := tmpl.RenderReader(context.Background(), buf, reader)
	require.NoError(t, err)

	output := buf.String()
	require.Contains(t, output, "Hello from Reader")
}

func TestTemplate_RenderReaderFromFile(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.NewFS(templateFS)
	tmpl.Assign("message", "Hello from File Reader")

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "template_*.html")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("<p>{{ message }}</p>")
	require.NoError(t, err)
	tmpFile.Close()

	// Reopen for reading
	file, err := os.Open(tmpFile.Name())
	require.NoError(t, err)
	defer file.Close()

	buf := &bytes.Buffer{}
	err = tmpl.RenderReader(context.Background(), buf, file)
	require.NoError(t, err)

	output := buf.String()
	require.Contains(t, output, "Hello from File Reader")
}

func TestNew(t *testing.T) {
	tmpl := vuego.New()
	require.NotNil(t, tmpl)
}

func TestNew_WithFS(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.New(vuego.WithFS(templateFS))
	require.NotNil(t, tmpl)

	tmpl.Assign("message", "Hello")

	buf := &bytes.Buffer{}
	err := tmpl.Load("interpolation-basic.vuego").Render(context.Background(), buf)
	require.NoError(t, err)
	require.Greater(t, buf.Len(), 0)
}

func TestNew_RenderWithoutFilesystemFails(t *testing.T) {
	tmpl := vuego.New()
	tmpl.Assign("message", "Hello")

	buf := &bytes.Buffer{}
	err := tmpl.Load("interpolation-basic.vuego").Render(context.Background(), buf)
	require.Equal(t, "error reading interpolation-basic.vuego: no filesystem configured", err.Error())
}

func TestNew_RenderStringWithoutFilesystem(t *testing.T) {
	tmpl := vuego.New()
	tmpl.Assign("message", "Hello from String")

	buf := &bytes.Buffer{}
	err := tmpl.RenderString(context.Background(), buf, "<p>{{ message }}</p>")
	require.NoError(t, err)

	output := buf.String()
	require.Contains(t, output, "Hello from String")
}
