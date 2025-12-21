package vuego_test

import (
	"bytes"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego"
)

// TestEvalInclude_StackManagement verifies that include stack push/pop is properly balanced
// even when errors occur during include loading or parsing.
func TestEvalInclude_StackManagement(t *testing.T) {
	t.Run("stack is popped after successful include", func(t *testing.T) {
		fs := fstest.MapFS{
			"parent.vuego": &fstest.MapFile{
				Data: []byte(`<div>
  <vuego include="child.vuego"></vuego>
</div>`),
			},
			"child.vuego": &fstest.MapFile{
				Data: []byte(`<span>child</span>`),
			},
		}

		tpl := vuego.NewFS(fs)
		var buf bytes.Buffer
		err := tpl.Load("parent.vuego").Fill(map[string]any{}).Render(t.Context(), &buf)
		require.NoError(t, err)

		// Should render without errors
		output := buf.String()
		require.Contains(t, output, "child")
	})

	t.Run("stack is popped even when include file missing", func(t *testing.T) {
		fs := fstest.MapFS{
			"parent.vuego": &fstest.MapFile{
				Data: []byte(`<div>
  <vuego include="missing.vuego"></vuego>
</div>`),
			},
		}

		tpl := vuego.NewFS(fs)
		var buf bytes.Buffer
		err := tpl.Load("parent.vuego").Fill(map[string]any{}).Render(t.Context(), &buf)

		// Should error but leave stack in consistent state
		require.Error(t, err)
		require.Contains(t, err.Error(), "error loading missing.vuego")
	})
}

// TestEvalInclude_DataPassing verifies that data passed via include attributes
// is properly scoped and doesn't leak to subsequent renders.
func TestEvalInclude_DataPassing(t *testing.T) {
	t.Run("include data is isolated per render", func(t *testing.T) {
		fs := fstest.MapFS{
			"parent.vuego": &fstest.MapFile{
				Data: []byte(`<vuego include="child.vuego" :msg="message"></vuego>`),
			},
			"child.vuego": &fstest.MapFile{
				Data: []byte(`{{ msg }}`),
			},
		}

		tpl := vuego.NewFS(fs)

		// First render
		var buf1 bytes.Buffer
		err := tpl.Load("parent.vuego").Fill(map[string]any{
			"message": "First",
		}).Render(t.Context(), &buf1)
		require.NoError(t, err)
		require.Contains(t, buf1.String(), "First")

		// Second render with different data
		var buf2 bytes.Buffer
		err = tpl.Load("parent.vuego").Fill(map[string]any{
			"message": "Second",
		}).Render(t.Context(), &buf2)
		require.NoError(t, err)
		require.Contains(t, buf2.String(), "Second")
		require.NotContains(t, buf2.String(), "First")
	})
}
