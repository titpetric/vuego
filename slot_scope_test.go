package vuego_test

import (
	"testing"

	"golang.org/x/net/html"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/tests/assert"
)

func TestNewSlotScope(t *testing.T) {
	t.Run("starts with no slots", func(t *testing.T) {
		scope := vuego.NewSlotScope()
		assert.NotNil(t, scope)
		assert.Len(t, scope.Slots, 0)
	})

	t.Run("the map is ready for writes", func(t *testing.T) {
		scope := vuego.NewSlotScope()
		scope.SetSlot("header", &vuego.SlotContent{})
		assert.Len(t, scope.Slots, 1)
	})
}

func TestSlotScope_SetSlot(t *testing.T) {
	t.Run("stores content under a name", func(t *testing.T) {
		content := &vuego.SlotContent{Nodes: []*html.Node{{Type: html.TextNode, Data: "x"}}}

		scope := vuego.NewSlotScope()
		scope.SetSlot("header", content)

		assert.Equal(t, content, scope.Slots["header"])
	})

	t.Run("a second write replaces the first", func(t *testing.T) {
		first := &vuego.SlotContent{Props: map[string]any{"n": 1}}
		second := &vuego.SlotContent{Props: map[string]any{"n": 2}}

		scope := vuego.NewSlotScope()
		scope.SetSlot("header", first)
		scope.SetSlot("header", second)

		assert.Len(t, scope.Slots, 1)
		assert.Equal(t, second, scope.GetSlot("header"))
	})

	t.Run("the default slot is a name like any other", func(t *testing.T) {
		scope := vuego.NewSlotScope()
		scope.SetSlot("default", &vuego.SlotContent{})
		scope.SetSlot("footer", &vuego.SlotContent{})

		assert.Len(t, scope.Slots, 2)
	})
}

func TestSlotScope_GetSlot(t *testing.T) {
	content := &vuego.SlotContent{
		Nodes: []*html.Node{{Type: html.ElementNode, Data: "span"}},
		Props: map[string]any{"item": "value"},
	}

	scope := vuego.NewSlotScope()
	scope.SetSlot("header", content)

	t.Run("a stored slot", func(t *testing.T) {
		got := scope.GetSlot("header")
		assert.Equal(t, content, got)
		assert.Equal(t, "value", got.Props["item"])
	})

	t.Run("an unknown name is nil", func(t *testing.T) {
		assert.Nil(t, scope.GetSlot("footer"))
	})

	t.Run("the empty name is nil when unset", func(t *testing.T) {
		assert.Nil(t, scope.GetSlot(""))
	})
}

func TestSlotContent(t *testing.T) {
	t.Run("carries nodes, props and the template node", func(t *testing.T) {
		templateNode := &html.Node{Type: html.ElementNode, Data: "template"}
		nodes := []*html.Node{{Type: html.TextNode, Data: "body"}}

		content := &vuego.SlotContent{
			Nodes:        nodes,
			Props:        map[string]any{"item": 1},
			TemplateNode: templateNode,
		}

		assert.Equal(t, nodes, content.Nodes)
		assert.Equal(t, 1, content.Props["item"])
		assert.Equal(t, templateNode, content.TemplateNode)
	})

	t.Run("the zero value is usable", func(t *testing.T) {
		var content vuego.SlotContent
		assert.Nil(t, content.Nodes)
		assert.Nil(t, content.Props)
		assert.Nil(t, content.TemplateNode)
	})
}
