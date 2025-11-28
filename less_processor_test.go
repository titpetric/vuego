package vuego_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/titpetric/vuego"
)

// TestLessProcessor_LessCompilation tests LESS compilation in <style type="text/css+less"> tags.
func TestLessProcessor_LessCompilation(t *testing.T) {
	// Create a test filesystem with a simple template
	templateFS := os.DirFS("testdata/nodeprocessor")

	// Create Vue instance and register the LESS processor
	v := vuego.NewVue(templateFS)
	v.RegisterNodeProcessor(vuego.NewLessProcessor())

	// Render the template
	var buf bytes.Buffer
	err := v.Render(&buf, "less.html", map[string]any{})
	require.NoError(t, err)

	// Verify the output contains compiled CSS instead of LESS
	output := buf.String()
	t.Logf("Output:\n%s", output)
	require.Contains(t, output, "<style>")
	require.NotContains(t, output, `type="text/css+less"`)
	require.Contains(t, output, "color: red;")
}

// TestLessProcessor_LessVariables tests LESS variable compilation.
func TestLessProcessor_LessVariables(t *testing.T) {
	templateFS := os.DirFS("testdata/nodeprocessor")

	v := vuego.NewVue(templateFS)
	v.RegisterNodeProcessor(vuego.NewLessProcessor())

	var buf bytes.Buffer
	err := v.Render(&buf, "less_variables.html", map[string]any{})
	require.NoError(t, err)

	output := buf.String()
	require.Contains(t, output, "<style>")
	// LESS variables should be compiled to actual values
	require.Contains(t, output, "#ff0000") // @primary-color: #ff0000 should be compiled
}

// TestLessProcessor_NoLessTag tests that normal style tags are unaffected.
func TestLessProcessor_NoLessTag(t *testing.T) {
	templateFS := os.DirFS("testdata/nodeprocessor")

	v := vuego.NewVue(templateFS)
	v.RegisterNodeProcessor(vuego.NewLessProcessor())

	var buf bytes.Buffer
	err := v.Render(&buf, "mixed_styles.html", map[string]any{})
	require.NoError(t, err)

	output := buf.String()
	// Normal style tags should remain unchanged
	require.Contains(t, output, `<style type="text/css">`)
	// LESS tags should be compiled
	require.Contains(t, output, "<style>")
}
