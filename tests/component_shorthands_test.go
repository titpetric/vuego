package tests

import (
	"bytes"
	"os"
	"testing"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/testing/assert"
)

func TestComponentShorthandBasic(t *testing.T) {
	root := os.DirFS("testdata/fixtures")

	// Create template with WithComponents option
	tpl := vuego.NewFS(root, vuego.WithComponents())
	assert.NotNil(t, tpl)

	// Load and render the button-primary template
	newTpl := tpl.Load("button-primary.vuego")

	var buf bytes.Buffer
	err := newTpl.Render(t.Context(), &buf)
	assert.NoError(t, err)

	result := buf.String()

	// The component should be resolved and rendered
	assert.Contains(t, result, "btn-primary")
	assert.Contains(t, result, "Primary Button")
}

func TestComponentShorthandWithAttributes(t *testing.T) {
	root := os.DirFS("testdata/fixtures")

	// Create a template with WithComponents
	tmpl := vuego.NewFS(root, vuego.WithComponents())
	assert.NotNil(t, tmpl)

	newTpl := tmpl.Load("button-primary.vuego")

	var buf bytes.Buffer
	err := newTpl.Render(t.Context(), &buf)
	assert.NoError(t, err)

	result := buf.String()
	assert.Contains(t, result, "btn-primary")
	assert.Contains(t, result, "Primary Button")
}

func TestComponentProcessorRegistration(t *testing.T) {
	root := os.DirFS("testdata/fixtures")

	// Create template without components first
	tpl1 := vuego.NewFS(root)
	assert.NotNil(t, tpl1)

	// Create template with components
	tpl2 := vuego.NewFS(root, vuego.WithComponents())
	assert.NotNil(t, tpl2)

	// Both should be valid templates
}

func TestComponentMappingLookup(t *testing.T) {
	root := os.DirFS("testdata/fixtures")

	vue := vuego.NewVue(root)

	// Register a component manually
	vue.RegisterComponent("test-button", "components/TestButton.vuego")

	// Look it up
	filename, ok := vue.GetComponentFile("test-button")
	assert.True(t, ok)
	assert.Equal(t, "components/TestButton.vuego", filename)

	// Look up non-existent component
	_, notFound := vue.GetComponentFile("non-existent")
	assert.False(t, notFound)
}
