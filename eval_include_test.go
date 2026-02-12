package vuego_test

import (
	"bytes"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"

	"github.com/titpetric/vuego"
)

// TestEvalInclude_StackManagement verifies that include stack push/pop is properly balanced
// even when errors occur during include loading or parsing.
func TestEvalInclude_StackManagement(t *testing.T) {
	t.Run("stack is popped after successful include", func(t *testing.T) {
		fs := fstest.MapFS{
			"parent.vuego": &fstest.MapFile{
				Data: []byte(`<div>
  <template include="child.vuego"></template>
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
  <template include="missing.vuego"></template>
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

// TestEvalInclude_InheritedSlots verifies that an included component receives
// slots from __slotScope__ in the parent data, and that explicitly defined slots
// in the include tag take priority over inherited ones.
func TestEvalInclude_InheritedSlots(t *testing.T) {
	t.Run("inherited slot is used when include has no explicit slot", func(t *testing.T) {
		fs := fstest.MapFS{
			"page.vuego": &fstest.MapFile{
				Data: []byte(`<template include="component.vuego"></template>`),
			},
			"component.vuego": &fstest.MapFile{
				Data: []byte(`<div><slot name="extra"><p>default extra</p></slot></div>`),
			},
		}

		slotScope := vuego.NewSlotScope()
		slotScope.SetSlot("extra", &vuego.SlotContent{
			Nodes: []*html.Node{
				{Type: html.TextNode, Data: "injected content"},
			},
		})

		tpl := vuego.NewFS(fs)
		var buf bytes.Buffer
		err := tpl.Load("page.vuego").Fill(map[string]any{
			"__slotScope__": slotScope,
		}).Render(t.Context(), &buf)
		require.NoError(t, err)

		output := buf.String()
		require.Contains(t, output, "injected content")
		require.NotContains(t, output, "default extra")
	})

	t.Run("explicit slot in include takes priority over inherited slot", func(t *testing.T) {
		fs := fstest.MapFS{
			"page.vuego": &fstest.MapFile{
				Data: []byte(`<template include="component.vuego"><template #extra>explicit content</template></template>`),
			},
			"component.vuego": &fstest.MapFile{
				Data: []byte(`<div><slot name="extra"><p>default extra</p></slot></div>`),
			},
		}

		slotScope := vuego.NewSlotScope()
		slotScope.SetSlot("extra", &vuego.SlotContent{
			Nodes: []*html.Node{
				{Type: html.TextNode, Data: "inherited content"},
			},
		})

		tpl := vuego.NewFS(fs)
		var buf bytes.Buffer
		err := tpl.Load("page.vuego").Fill(map[string]any{
			"__slotScope__": slotScope,
		}).Render(t.Context(), &buf)
		require.NoError(t, err)

		output := buf.String()
		require.Contains(t, output, "explicit content")
		require.NotContains(t, output, "inherited content")
		require.NotContains(t, output, "default extra")
	})
}

// TestEvalInclude_DataPassing verifies that data passed via include attributes
// is properly scoped and doesn't leak to subsequent renders.
func TestEvalInclude_DataPassing(t *testing.T) {
	t.Run("include data is isolated per render", func(t *testing.T) {
		fs := fstest.MapFS{
			"parent.vuego": &fstest.MapFile{
				Data: []byte(`<template include="child.vuego" :msg="message"></template>`),
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
