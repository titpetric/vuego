package vuego_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/testing/assert"
)

func TestLoadWithLessProcessor(t *testing.T) {
	root := os.DirFS("testdata")

	tpl := vuego.NewFS(root, vuego.WithLessProcessor())

	assert.NotNil(t, tpl)
}

func TestLoadWithMultipleProcessors(t *testing.T) {
	root := os.DirFS("testdata")

	tpl := vuego.NewFS(root, vuego.WithLessProcessor(), vuego.WithLessProcessor())

	assert.NotNil(t, tpl)
}

func TestLoadWithoutProcessors(t *testing.T) {
	root := os.DirFS("testdata")

	tpl := vuego.NewFS(root)

	assert.NotNil(t, tpl)
}

func TestTemplateFuncs(t *testing.T) {
	root := os.DirFS("testdata")

	funcMap := vuego.FuncMap{
		"customFunc": func(s string) string {
			return "custom: " + s
		},
	}

	tpl := vuego.NewFS(root, vuego.WithFuncs(funcMap))

	// Verify template is created
	assert.NotNil(t, tpl)
}

func TestTemplateChaining(t *testing.T) {
	root := os.DirFS("testdata")

	tpl := vuego.NewFS(root, vuego.WithLessProcessor(), vuego.WithFuncs(vuego.FuncMap{"test": func() string { return "test" }}))

	// Verify chaining works
	result := tpl.Assign("key", "value")

	assert.Equal(t, tpl, result)
	assert.Equal(t, "value", tpl.Get("key"))
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
	assert.NoError(t, err)
	assert.Greater(t, buf.Len(), 0)
}
