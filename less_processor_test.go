package vuego_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/testing/assert"
)

var _ vuego.NodeProcessor = &vuego.LessProcessor{}

// TestLessProcessor_LessCompilation tests LESS compilation in <style type="text/css+less"> tags.
func TestLessProcessor_LessCompilation(t *testing.T) {
	// Create a test filesystem with a simple template
	templateFS := os.DirFS("testdata/nodeprocessor")

	// Create Vue instance and register the LESS processor
	v := vuego.NewVue(templateFS)
	v.RegisterNodeProcessor(vuego.NewLessProcessor())

	// Render the template
	var buf bytes.Buffer
	err := v.Render(t.Context(), &buf, "less.html", map[string]any{})
	assert.NoError(t, err)

	// Verify the output contains compiled CSS instead of LESS
	output := buf.String()
	t.Logf("Output:\n%s", output)
	assert.Contains(t, output, `<style type="text/css">`)
	assert.NotContains(t, output, `type="text/css+less"`)
	assert.Contains(t, output, "color: red;")
}

// TestLessProcessor_LessVariables tests LESS variable compilation.
func TestLessProcessor_LessVariables(t *testing.T) {
	templateFS := os.DirFS("testdata/nodeprocessor")

	v := vuego.NewVue(templateFS)
	v.RegisterNodeProcessor(vuego.NewLessProcessor())

	var buf bytes.Buffer
	err := v.Render(t.Context(), &buf, "less_variables.html", map[string]any{})
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `<style type="text/css">`)
	// LESS variables should be compiled to actual values
	assert.Contains(t, output, "#ff0000") // @primary-color: #ff0000 should be compiled
}

// TestLessProcessor_NoLessTag tests that normal style tags are unaffected.
func TestLessProcessor_NoLessTag(t *testing.T) {
	templateFS := os.DirFS("testdata/nodeprocessor")

	v := vuego.NewVue(templateFS)
	v.RegisterNodeProcessor(vuego.NewLessProcessor())

	var buf bytes.Buffer
	err := v.Render(t.Context(), &buf, "mixed_styles.html", map[string]any{})
	assert.NoError(t, err)

	output := buf.String()
	// Normal style tags should remain unchanged, LESS tags should be compiled to type="text/css"
	assert.Contains(t, output, `<style type="text/css">`)
	assert.NotContains(t, output, `type="text/css+less"`)
	// Verify LESS compilation worked (LESS variables should be compiled)
	assert.Contains(t, output, "color: #333;")
}
