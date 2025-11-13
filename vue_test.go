package vuego_test

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/internal/helpers"
)

func TestVue_Render(t *testing.T) {
	var buf bytes.Buffer

	vue := vuego.NewVue(os.DirFS("testdata/pages"))
	err := vue.Render(&buf, "Index.vuego", nil)
	require.NoError(t, err)

	t.Log(buf.String())
}

func TestFixtures(t *testing.T) {
	fixtures := os.DirFS("testdata/fixtures")
	vue := vuego.NewVue(fixtures)
	templates, err := fs.Glob(fixtures, "*.vuego")
	require.NoError(t, err)

	for _, template := range templates {
		want, err := fs.ReadFile(fixtures, strings.ReplaceAll(template, ".vuego", ".html"))
		require.NoError(t, err)

		dataBytes, err := fs.ReadFile(fixtures, strings.ReplaceAll(template, ".vuego", ".json"))
		require.NoError(t, err)

		templateBytes, err := fs.ReadFile(fixtures, template)
		require.NoError(t, err)

		data := map[string]any{}
		require.NoError(t, json.NewDecoder(bytes.NewReader(dataBytes)).Decode(&data))

		t.Run(template, func(t *testing.T) {
			var got bytes.Buffer
			require.NoError(t, vue.RenderFragment(&got, template, data))
			require.True(t, helpers.CompareHTML(t, want, got.Bytes(), templateBytes, dataBytes))
		})
	}
}
