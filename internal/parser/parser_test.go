package parser_test

import (
	"testing"

	"github.com/titpetric/vuego/internal/parser"
	"github.com/titpetric/vuego/testing/assert"
)

func TestParseTemplateBytes(t *testing.T) {
	t.Run("fragment without html tag", func(t *testing.T) {
		templateBytes := []byte("<div>content</div>")
		nodes, err := parser.ParseTemplateBytes(templateBytes)
		assert.NoError(t, err)
		assert.NotNil(t, nodes)
		assert.Greater(t, len(nodes), 0)
	})

	t.Run("full document with html tag", func(t *testing.T) {
		templateBytes := []byte("<html><body><h1>Test</h1></body></html>")
		nodes, err := parser.ParseTemplateBytes(templateBytes)
		assert.NoError(t, err)
		assert.NotNil(t, nodes)
		assert.Greater(t, len(nodes), 0)
	})

	t.Run("invalid html", func(t *testing.T) {
		templateBytes := []byte("<<<>>>")
		nodes, err := parser.ParseTemplateBytes(templateBytes)
		assert.NoError(t, err)
		assert.NotNil(t, nodes)
	})
}
