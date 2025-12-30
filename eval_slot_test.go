package vuego_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego"
)

func TestSlot_Default(t *testing.T) {
	root := os.DirFS("testdata/fixtures")
	vue := vuego.NewVue(root)
	vue.RegisterComponent("slot-button", "components/SlotButton.vuego")

	var buf bytes.Buffer
	err := vue.RenderFragment(&buf, "slot-default.vuego", nil)
	require.NoError(t, err)

	result := buf.String()
	require.Contains(t, result, "btn")
	require.Contains(t, result, "Click me")
	require.NotContains(t, result, "Submit") // Should not use fallback
}

func TestSlot_Fallback(t *testing.T) {
	root := os.DirFS("testdata/fixtures")
	vue := vuego.NewVue(root)
	vue.RegisterComponent("slot-button", "components/SlotButton.vuego")

	var buf bytes.Buffer
	err := vue.RenderFragment(&buf, "slot-fallback.vuego", nil)
	require.NoError(t, err)

	result := buf.String()
	require.Contains(t, result, "btn")
	require.Contains(t, result, "Submit") // Should use fallback
}

func TestSlot_Named(t *testing.T) {
	root := os.DirFS("testdata/fixtures")
	vue := vuego.NewVue(root)
	vue.RegisterComponent("modal", "components/Modal.vuego")

	var buf bytes.Buffer
	err := vue.RenderFragment(&buf, "slot-named.vuego", nil)
	require.NoError(t, err)

	result := buf.String()
	require.Contains(t, result, "modal-header")
	require.Contains(t, result, "Header Content")
	require.Contains(t, result, "modal-body")
	require.Contains(t, result, "Body Content")
	require.Contains(t, result, "modal-footer")
	require.Contains(t, result, "Footer Content")
}

func TestSlot_Scoped(t *testing.T) {
	root := os.DirFS("testdata/fixtures")
	vue := vuego.NewVue(root)
	vue.RegisterComponent("list", "components/List.vuego")

	var buf bytes.Buffer
	err := vue.RenderFragment(&buf, "slot-scoped.vuego", map[string]any{
		"products": []string{"Apple", "Banana", "Cherry"},
	})
	require.NoError(t, err)

	result := buf.String()
	require.Contains(t, result, "Apple (#0)")
	require.Contains(t, result, "Banana (#1)")
	require.Contains(t, result, "Cherry (#2)")
}

func TestSlot_NamedScoped(t *testing.T) {
	root := os.DirFS("testdata/fixtures")
	vue := vuego.NewVue(root)
	vue.RegisterComponent("tabs", "components/Tabs.vuego")

	var buf bytes.Buffer
	err := vue.RenderFragment(&buf, "slot-named-scoped.vuego", nil)
	require.NoError(t, err)

	result := buf.String()
	require.Contains(t, result, "tabs-header")
	require.Contains(t, result, "Header: Welcome")
	require.Contains(t, result, "tabs-content")
	require.Contains(t, result, "Items: 3")
}

func TestSlot_SimpleScoped(t *testing.T) {
	root := os.DirFS("testdata/fixtures")
	vue := vuego.NewVue(root)
	vue.RegisterComponent("simple-slot", "components/SimpleSlot.vuego")

	var buf bytes.Buffer
	err := vue.RenderFragment(&buf, "simple-scoped.vuego", map[string]any{
		"message": "Hello",
	})
	require.NoError(t, err)

	result := buf.String()
	require.Contains(t, result, "Message: Hello")
}
