package vuego_test

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
	"github.com/titpetric/vuego"
)

func TestNewComponent(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{Data: []byte("<div>test</div>")},
	}

	comp := vuego.NewComponent(fs)
	require.NotNil(t, comp)
}

func TestComponent_Stat(t *testing.T) {
	fs := fstest.MapFS{
		"exists.html": &fstest.MapFile{Data: []byte("<div>test</div>")},
	}

	comp := vuego.NewComponent(fs)

	t.Run("file exists", func(t *testing.T) {
		err := comp.Stat("exists.html")
		require.NoError(t, err)
	})

	t.Run("file not found", func(t *testing.T) {
		err := comp.Stat("missing.html")
		require.Error(t, err)
	})
}

func TestComponent_Load(t *testing.T) {
	fs := fstest.MapFS{
		"doc.html": &fstest.MapFile{Data: []byte("<html><body><h1>Test</h1></body></html>")},
	}

	comp := vuego.NewComponent(fs)

	t.Run("valid document", func(t *testing.T) {
		nodes, err := comp.Load("doc.html")
		require.NoError(t, err)
		require.NotNil(t, nodes)
		require.Greater(t, len(nodes), 0)
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := comp.Load("missing.html")
		require.Error(t, err)
	})

	t.Run("invalid html", func(t *testing.T) {
		fs := fstest.MapFS{
			"bad.html": &fstest.MapFile{Data: []byte("<<<>>>")},
		}
		comp := vuego.NewComponent(fs)
		nodes, err := comp.Load("bad.html")
		require.NoError(t, err)
		require.NotNil(t, nodes)
	})
}

func TestComponent_LoadFragment(t *testing.T) {
	fs := fstest.MapFS{
		"fragment.html": &fstest.MapFile{Data: []byte("<div>content</div>")},
		"document.html": &fstest.MapFile{Data: []byte("<html><body><div>doc</div></body></html>")},
	}

	comp := vuego.NewComponent(fs)

	t.Run("fragment without html tag", func(t *testing.T) {
		nodes, err := comp.LoadFragment("fragment.html")
		require.NoError(t, err)
		require.NotNil(t, nodes)
		require.Greater(t, len(nodes), 0)
	})

	t.Run("full document falls back to Load", func(t *testing.T) {
		nodes, err := comp.LoadFragment("document.html")
		require.NoError(t, err)
		require.NotNil(t, nodes)
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := comp.LoadFragment("missing.html")
		require.Error(t, err)
	})
}
