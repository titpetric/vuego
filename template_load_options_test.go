package vuego_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego"
)

func TestLoadWithLessProcessor(t *testing.T) {
	root := os.DirFS("testdata")

	tpl := vuego.NewFS(root, vuego.WithLessProcessor())

	require.NotNil(t, tpl)
}

func TestLoadWithMultipleProcessors(t *testing.T) {
	root := os.DirFS("testdata")

	tpl := vuego.NewFS(root, vuego.WithLessProcessor(), vuego.WithLessProcessor())

	require.NotNil(t, tpl)
}

func TestLoadWithoutProcessors(t *testing.T) {
	root := os.DirFS("testdata")

	tpl := vuego.NewFS(root)

	require.NotNil(t, tpl)
}

func TestTemplateFuncs(t *testing.T) {
	root := os.DirFS("testdata")

	tpl := vuego.NewFS(root)

	funcMap := vuego.FuncMap{
		"customFunc": func(s string) string {
			return "custom: " + s
		},
	}

	result := tpl.Funcs(funcMap)

	// Verify chaining returns the template
	require.Equal(t, tpl, result)
}

func TestTemplateChaining(t *testing.T) {
	root := os.DirFS("testdata")

	tpl := vuego.NewFS(root, vuego.WithLessProcessor())

	// Verify chaining works
	result := tpl.
		Assign("key", "value").
		Funcs(vuego.FuncMap{"test": func() string { return "test" }})

	require.Equal(t, tpl, result)
	require.Equal(t, "value", tpl.Get("key"))
}

func TestLoadTemplateRender(t *testing.T) {
	root := os.DirFS("testdata")

	tpl := vuego.NewFS(root, vuego.WithLessProcessor())

	// Assign required attributes - Button component requires name, variant, and title
	tpl.Assign("name", "TestButton").
		Assign("variant", "primary").
		Assign("title", "Click me")

	var buf bytes.Buffer
	err := tpl.Load("pages/components/Button.vuego").Render(t.Context(), &buf)
	require.NoError(t, err)
	require.Greater(t, buf.Len(), 0)
}
