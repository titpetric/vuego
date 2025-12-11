package parser_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego/internal/parser"
)

func TestParseTemplateBytes(t *testing.T) {
	t.Run("fragment without html tag", func(t *testing.T) {
		templateBytes := []byte("<div>content</div>")
		nodes, err := parser.ParseTemplateBytes(templateBytes)
		require.NoError(t, err)
		require.NotNil(t, nodes)
		require.Greater(t, len(nodes), 0)
	})

	t.Run("full document with html tag", func(t *testing.T) {
		templateBytes := []byte("<html><body><h1>Test</h1></body></html>")
		nodes, err := parser.ParseTemplateBytes(templateBytes)
		require.NoError(t, err)
		require.NotNil(t, nodes)
		require.Greater(t, len(nodes), 0)
	})

	t.Run("invalid html", func(t *testing.T) {
		templateBytes := []byte("<<<>>>")
		nodes, err := parser.ParseTemplateBytes(templateBytes)
		require.NoError(t, err)
		require.NotNil(t, nodes)
	})
}
