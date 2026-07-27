package vuego_test

import (
	"bytes"
	"os"
	"testing"
	"testing/fstest"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/testing/assert"
)

func TestSlot_Default(t *testing.T) {
	root := os.DirFS("testdata/fixtures")
	vue := vuego.NewVue(root)
	vue.RegisterComponent("slot-button", "components/SlotButton.vuego")

	var buf bytes.Buffer
	err := vue.RenderFragment(t.Context(), &buf, "slot-default.vuego", nil)
	assert.NoError(t, err)

	result := buf.String()
	assert.Contains(t, result, "btn")
	assert.Contains(t, result, "Click me")
	assert.NotContains(t, result, "Submit") // Should not use fallback
}

func TestSlot_Fallback(t *testing.T) {
	root := os.DirFS("testdata/fixtures")
	vue := vuego.NewVue(root)
	vue.RegisterComponent("slot-button", "components/SlotButton.vuego")

	var buf bytes.Buffer
	err := vue.RenderFragment(t.Context(), &buf, "slot-fallback.vuego", nil)
	assert.NoError(t, err)

	result := buf.String()
	assert.Contains(t, result, "btn")
	assert.Contains(t, result, "Submit") // Should use fallback
}

func TestSlot_Named(t *testing.T) {
	root := os.DirFS("testdata/fixtures")
	vue := vuego.NewVue(root)
	vue.RegisterComponent("modal", "components/Modal.vuego")

	var buf bytes.Buffer
	err := vue.RenderFragment(t.Context(), &buf, "slot-named.vuego", nil)
	assert.NoError(t, err)

	result := buf.String()
	assert.Contains(t, result, "modal-header")
	assert.Contains(t, result, "Header Content")
	assert.Contains(t, result, "modal-body")
	assert.Contains(t, result, "Body Content")
	assert.Contains(t, result, "modal-footer")
	assert.Contains(t, result, "Footer Content")
}

func TestSlot_Scoped(t *testing.T) {
	root := os.DirFS("testdata/fixtures")
	vue := vuego.NewVue(root)
	vue.RegisterComponent("list", "components/List.vuego")

	var buf bytes.Buffer
	err := vue.RenderFragment(t.Context(), &buf, "slot-scoped.vuego", map[string]any{
		"products": []string{"Apple", "Banana", "Cherry"},
	})
	assert.NoError(t, err)

	result := buf.String()
	assert.Contains(t, result, "Apple (#0)")
	assert.Contains(t, result, "Banana (#1)")
	assert.Contains(t, result, "Cherry (#2)")
}

func TestSlot_NamedScoped(t *testing.T) {
	root := os.DirFS("testdata/fixtures")
	vue := vuego.NewVue(root)
	vue.RegisterComponent("tabs", "components/Tabs.vuego")

	var buf bytes.Buffer
	err := vue.RenderFragment(t.Context(), &buf, "slot-named-scoped.vuego", nil)
	assert.NoError(t, err)

	result := buf.String()
	assert.Contains(t, result, "tabs-header")
	assert.Contains(t, result, "Header: Welcome")
	assert.Contains(t, result, "tabs-content")
	assert.Contains(t, result, "Items: 3")
}

func TestSlot_SimpleScoped(t *testing.T) {
	root := os.DirFS("testdata/fixtures")
	vue := vuego.NewVue(root)
	vue.RegisterComponent("simple-slot", "components/SimpleSlot.vuego")

	var buf bytes.Buffer
	err := vue.RenderFragment(t.Context(), &buf, "simple-scoped.vuego", map[string]any{
		"message": "Hello",
	})
	assert.NoError(t, err)

	result := buf.String()
	assert.Contains(t, result, "Message: Hello")
}

func TestSlot_InheritedFromLayout(t *testing.T) {
	inlineFS := &fstest.MapFS{
		"page.vuego": &fstest.MapFile{Data: []byte(`---
layout: slotted
---
<template #sidebar><nav>Side Nav</nav></template>`)},
		"layouts/slotted.vuego": &fstest.MapFile{Data: []byte(`<div class="sidebar"><slot name="sidebar"><p>Default sidebar</p></slot></div>
<div class="main" v-html="content"></div>`)},
	}

	renderer := vuego.NewFS(inlineFS)

	var buf bytes.Buffer
	err := renderer.Load("page.vuego").Fill(nil).Render(t.Context(), &buf)
	assert.NoError(t, err)

	result := buf.String()
	assert.Contains(t, result, "Side Nav", "inherited slot content should be rendered")
	assert.NotContains(t, result, "Default sidebar", "fallback slot content should not be used")
	assert.Contains(t, result, "sidebar", "layout structure with sidebar class should be present")
}
