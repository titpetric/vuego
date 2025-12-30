package vuego_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego"
)

func TestComponentShorthandBasic(t *testing.T) {
	root := os.DirFS("../testdata/fixtures")

	// Create template with WithComponents option
	tpl := vuego.NewFS(root, vuego.WithComponents())
	require.NotNil(t, tpl)

	// Load and render the button-primary template
	newTpl := tpl.Load("button-primary.vuego")

	var buf bytes.Buffer
	err := newTpl.Render(t.Context(), &buf)
	require.NoError(t, err)

	result := buf.String()

	// The component should be resolved and rendered
	require.Contains(t, result, "btn-primary")
	require.Contains(t, result, "Primary Button")
}

func TestComponentShorthandWithAttributes(t *testing.T) {
	root := os.DirFS("../testdata/fixtures")

	// Create a template with WithComponents
	tmpl := vuego.NewFS(root, vuego.WithComponents())
	require.NotNil(t, tmpl)

	newTpl := tmpl.Load("button-primary.vuego")

	var buf bytes.Buffer
	err := newTpl.Render(t.Context(), &buf)
	require.NoError(t, err)

	result := buf.String()
	require.Contains(t, result, "btn-primary")
	require.Contains(t, result, "Primary Button")
}

func TestComponentProcessorRegistration(t *testing.T) {
	root := os.DirFS("../testdata/fixtures")

	// Create template without components first
	tpl1 := vuego.NewFS(root)
	require.NotNil(t, tpl1)

	// Create template with components
	tpl2 := vuego.NewFS(root, vuego.WithComponents())
	require.NotNil(t, tpl2)

	// Both should be valid templates
}

func TestComponentMappingLookup(t *testing.T) {
	root := os.DirFS("../testdata/fixtures")

	vue := vuego.NewVue(root)

	// Register a component manually
	vue.RegisterComponent("test-button", "components/TestButton.vuego")

	// Look it up
	filename, ok := vue.GetComponentFile("test-button")
	require.True(t, ok)
	require.Equal(t, "components/TestButton.vuego", filename)

	// Look up non-existent component
	_, notFound := vue.GetComponentFile("non-existent")
	require.False(t, notFound)
}
