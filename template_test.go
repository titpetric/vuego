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

	tmpl := vuego.Load(templateFS)
	require.NotNil(t, tmpl)
}

func TestTemplate_Fill(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.Load(templateFS)

	// Fill should be chainable
	result := tmpl.Fill(map[string]any{
		"message": "Hello World",
	})
	require.NotNil(t, result)
	require.Equal(t, tmpl, result)

	// Verify the variables were set
	require.Equal(t, "Hello World", tmpl.GetVar("message"))
}

func TestTemplate_Assign(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.Load(templateFS)

	// Assign should be chainable
	result := tmpl.Assign("name", "Alice")
	require.NotNil(t, result)
	require.Equal(t, tmpl, result)

	// Verify the variable was set
	require.Equal(t, "Alice", tmpl.GetVar("name"))
}

func TestTemplate_GetVar(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.Load(templateFS)

	// Set and get
	tmpl.Assign("key", "value")
	require.Equal(t, "value", tmpl.GetVar("key"))

	// Get non-existent variable
	require.Nil(t, tmpl.GetVar("nonexistent"))
}

func TestTemplate_GetString(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.Load(templateFS)

	// String value
	tmpl.Assign("message", "Hello World")
	require.Equal(t, "Hello World", tmpl.GetString("message"))

	// Non-string value returns empty string
	tmpl.Assign("count", 42)
	require.Equal(t, "", tmpl.GetString("count"))

	// Non-existent variable returns empty string
	require.Equal(t, "", tmpl.GetString("nonexistent"))
}

func TestTemplate_Render(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.Load(templateFS)

	tmpl.Assign("message", "Hello")

	buf := &bytes.Buffer{}
	err := tmpl.Render(context.Background(), buf, "interpolation-basic.vuego")
	require.NoError(t, err)

	// Should have rendered something
	require.Greater(t, buf.Len(), 0)
}

func TestTemplate_RenderCancelledContext(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.Load(templateFS)

	tmpl.Assign("message", "Hello")

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	buf := &bytes.Buffer{}
	err := tmpl.Render(ctx, buf, "interpolation-basic.vuego")
	require.Equal(t, context.Canceled, err)
}

func TestTemplate_FrontMatterVariables(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.Load(templateFS)

	// Render the template to load front-matter
	buf := &bytes.Buffer{}
	err := tmpl.Render(context.Background(), buf, "frontmatter-basic.vuego")
	require.NoError(t, err)

	// Note: Front-matter variables are merged during rendering, not during Load
	// This test demonstrates that rendering works with front-matter
	require.Greater(t, buf.Len(), 0)
}

func TestTemplate_Chaining(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.Load(templateFS)

	// Test method chaining
	tmpl.Assign("foo", "bar").Assign("baz", "qux").Fill(map[string]any{
		"message": "Hello",
	})

	require.Equal(t, "bar", tmpl.GetVar("foo"))
	require.Equal(t, "qux", tmpl.GetVar("baz"))
	require.Equal(t, "Hello", tmpl.GetVar("message"))
}

func TestTemplate_FillOverrides(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.Load(templateFS)

	// Set initial value
	tmpl.Assign("name", "Alice")
	require.Equal(t, "Alice", tmpl.GetVar("name"))

	// Fill with new values
	tmpl.Fill(map[string]any{
		"name": "Bob",
	})
	require.Equal(t, "Bob", tmpl.GetVar("name"))
}

func TestTemplate_RenderWithFrontMatter(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.Load(templateFS)

	buf := &bytes.Buffer{}
	err := tmpl.Render(context.Background(), buf, "frontmatter-basic.vuego")
	require.NoError(t, err)

	// Should have rendered HTML with the title from front-matter
	require.Greater(t, buf.Len(), 0)
	output := buf.String()
	require.Contains(t, output, "From Front-Matter")
}

func TestTemplate_RenderString(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.Load(templateFS)
	tmpl.Assign("message", "Hello from String")

	buf := &bytes.Buffer{}
	err := tmpl.RenderString(context.Background(), buf, "<p>{{ message }}</p>")
	require.NoError(t, err)

	output := buf.String()
	require.Contains(t, output, "Hello from String")
}

func TestTemplate_RenderByte(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl := vuego.Load(templateFS)
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

	tmpl := vuego.Load(templateFS)
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

	tmpl := vuego.Load(templateFS)
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
	err := tmpl.Render(context.Background(), buf, "interpolation-basic.vuego")
	require.NoError(t, err)
	require.Greater(t, buf.Len(), 0)
}

func TestNew_RenderWithoutFilesystemFails(t *testing.T) {
	tmpl := vuego.New()
	tmpl.Assign("message", "Hello")

	buf := &bytes.Buffer{}
	err := tmpl.Render(context.Background(), buf, "interpolation-basic.vuego")
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
