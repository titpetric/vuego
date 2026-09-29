package vuego_test

import (
	"bytes"
	"errors"
	"testing"
	"testing/fstest"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/tests/assert"
)

// errWriteFailed is the sentinel the processor and writer test doubles in this
// package return.
var errWriteFailed = errors.New("write failed")

// Vue satisfies Component, which is what the interface exists for: a caller
// can hold a renderer without depending on the concrete type.
var _ vuego.Component = (*vuego.Vue)(nil)

func TestComponent(t *testing.T) {
	const filename = "page.vuego"

	fsys := fstest.MapFS{filename: {Data: []byte(`<h1>{{ title }}</h1>`)}}

	t.Run("a Vue renders through the interface", func(t *testing.T) {
		var component vuego.Component = vuego.NewVue(fsys)

		var buf bytes.Buffer
		assert.NoError(t, component.Render(t.Context(), &buf, filename, map[string]any{"title": "Hello"}))
		assert.Contains(t, buf.String(), "Hello")
	})

	t.Run("a render error reaches the caller", func(t *testing.T) {
		var component vuego.Component = vuego.NewVue(fsys)

		var buf bytes.Buffer
		assert.Error(t, component.Render(t.Context(), &buf, "missing.vuego", nil))
	})
}
