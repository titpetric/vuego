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

	tmpl, err := vuego.Load(templateFS, "frontmatter-basic.vuego")
	require.NoError(t, err)
	require.NotNil(t, tmpl)
}

func TestTemplate_Fill(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl, err := vuego.Load(templateFS, "interpolation-basic.vuego")
	require.NoError(t, err)

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

	tmpl, err := vuego.Load(templateFS, "interpolation-basic.vuego")
	require.NoError(t, err)

	// Assign should be chainable
	result := tmpl.Assign("name", "Alice")
	require.NotNil(t, result)
	require.Equal(t, tmpl, result)

	// Verify the variable was set
	require.Equal(t, "Alice", tmpl.GetVar("name"))
}

func TestTemplate_GetVar(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl, err := vuego.Load(templateFS, "interpolation-basic.vuego")
	require.NoError(t, err)

	// Set and get
	tmpl.Assign("key", "value")
	require.Equal(t, "value", tmpl.GetVar("key"))

	// Get non-existent variable
	require.Nil(t, tmpl.GetVar("nonexistent"))
}

func TestTemplate_GetString(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl, err := vuego.Load(templateFS, "interpolation-basic.vuego")
	require.NoError(t, err)

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

	tmpl, err := vuego.Load(templateFS, "interpolation-basic.vuego")
	require.NoError(t, err)

	tmpl.Assign("message", "Hello")

	buf := &bytes.Buffer{}
	err = tmpl.Render(context.Background(), buf)
	require.NoError(t, err)

	// Should have rendered something
	require.Greater(t, buf.Len(), 0)
}

func TestTemplate_RenderCancelledContext(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl, err := vuego.Load(templateFS, "interpolation-basic.vuego")
	require.NoError(t, err)

	tmpl.Assign("message", "Hello")

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	buf := &bytes.Buffer{}
	err = tmpl.Render(ctx, buf)
	require.Equal(t, context.Canceled, err)
}

func TestTemplate_FrontMatterVariables(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl, err := vuego.Load(templateFS, "frontmatter-basic.vuego")
	require.NoError(t, err)

	// Variables from front-matter should be available
	title := tmpl.GetVar("title")
	require.Equal(t, "From Front-Matter", title)

	count := tmpl.GetVar("count")
	require.Equal(t, 42, count)
}

func TestTemplate_Chaining(t *testing.T) {
	templateFS := os.DirFS("testdata/fixtures")

	tmpl, err := vuego.Load(templateFS, "interpolation-basic.vuego")
	require.NoError(t, err)

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

	tmpl, err := vuego.Load(templateFS, "interpolation-basic.vuego")
	require.NoError(t, err)

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

	tmpl, err := vuego.Load(templateFS, "frontmatter-basic.vuego")
	require.NoError(t, err)

	// Front-matter variables should be in the template
	require.Equal(t, "From Front-Matter", tmpl.GetVar("title"))

	buf := &bytes.Buffer{}
	err = tmpl.Render(context.Background(), buf)
	require.NoError(t, err)

	// Should have rendered HTML with the title
	require.Greater(t, buf.Len(), 0)
	output := buf.String()
	require.Contains(t, output, "From Front-Matter")
}
