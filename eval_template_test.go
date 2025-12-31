package vuego_test

import (
	"bytes"
	"os"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego"
)

// TestRequiredAttributeError tests that a missing :required attribute
// produces an error. Ideally, the error should include file/line information
// from the source .vuego template.
//
// CURRENT LIMITATION: golang.org/x/net/html.Node does not track source line information.
// See: https://github.com/golang/go/issues/34302
//
// To add line tracking, we would need to either:
// 1. Use a forked version of html parser that tracks offsets/lines
// 2. Maintain a separate mapping of nodes to source positions
// 3. Wait for upstream support in golang.org/x/net/html
func TestRequiredAttributeError(t *testing.T) {
	vue := vuego.NewVue(os.DirFS("testdata"))

	data := map[string]any{
		// Both pageTitle and pageAuthor are missing to trigger validation errors
		// The component requires both "title" and "author"
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "required-error-test/page.vuego", data)

	require.Equal(t, "error in required-error-test/component.vuego (included from required-error-test/page.vuego): required attribute 'title' not provided", err.Error())
}

// TestRootLevelTemplateRequired tests that a root-level template with :required attribute
// validates the required attribute at the document root level.
func TestRootLevelTemplateRequired(t *testing.T) {
	vue := vuego.NewVue(os.DirFS("testdata"))

	data := map[string]any{
		// Missing 'title' which is required by the template
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "root-template-required/page.vuego", data)

	require.Equal(t, "required attribute 'title' not provided", err.Error())
}

// TestRootLevelTemplateRequiredSuccess tests that a root-level template with :required attribute
// renders successfully when the required attribute is provided.
func TestRootLevelTemplateRequiredSuccess(t *testing.T) {
	vue := vuego.NewVue(os.DirFS("testdata"))

	data := map[string]any{
		"title": "My Page Title",
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "root-template-required/page.vuego", data)

	require.NoError(t, err)
	require.Contains(t, buf.String(), "<h1>My Page Title</h1>")
}

// TestRequiredCSVExpansion tests that :required supports CSV format
// e.g., :required="name,title,author" expands to multiple required fields.
func TestRequiredCSVExpansion(t *testing.T) {
	vue := vuego.NewVue(os.DirFS("testdata"))

	// Missing 'author' when both 'title' and 'author' are required via CSV
	data := map[string]any{
		"title": "Article Title",
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "required-error-test/page.vuego", data)

	require.Equal(t, "error in required-error-test/component.vuego (included from required-error-test/page.vuego): required attribute 'author' not provided", err.Error())
}

// TestTemplateSimpleAttribute tests that a simple bound attribute on a template
// is evaluated and available in the template's children.
func TestTemplate_SimpleAttribute(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`<template :x="5">X={{ x }}</template>`)},
	}

	var buf bytes.Buffer
	vue := vuego.NewVue(templateFS)
	err := vue.RenderFragment(&buf, "test.vuego", map[string]any{})

	require.NoError(t, err)
	require.Equal(t, "X=5", buf.String())
}

// TestTemplateExpressionAttribute tests that a bound attribute with an expression
// is evaluated and available in the template's children.
func TestTemplate_ExpressionAttribute(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`<template :x="2+3">X={{ x }}</template>`)},
	}

	var buf bytes.Buffer
	vue := vuego.NewVue(templateFS)
	err := vue.RenderFragment(&buf, "test.vuego", map[string]any{})

	require.NoError(t, err)
	require.Equal(t, "X=5", buf.String())
}

// TestTemplateVariableAttribute tests that a bound attribute using a context variable
// is evaluated and available in the template's children.
func TestTemplate_VariableAttribute(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`<template :x="y">X={{ x }}</template>`)},
	}

	var buf bytes.Buffer
	vue := vuego.NewVue(templateFS)
	err := vue.RenderFragment(&buf, "test.vuego", map[string]any{"y": 10})

	require.NoError(t, err)
	require.Equal(t, "X=10", buf.String())
}

// TestTemplateNestedAttribute tests that nested templates each have their own
// bound attributes and can reference parent attributes.
func TestTemplate_NestedAttribute(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`<template :x="1"><template :y="x+1">X={{ x }} Y={{ y }}</template></template>`)},
	}

	var buf bytes.Buffer
	vue := vuego.NewVue(templateFS)
	err := vue.RenderFragment(&buf, "test.vuego", map[string]any{})

	require.NoError(t, err)
	require.Equal(t, "X=1 Y=2", buf.String())
}

// TestTemplateForAttribute tests that bound attributes on a template with v-for
// are evaluated for each iteration and available in the children.
func TestTemplate_ForAttribute(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`<template v-for="item in items" :index="0">{{ item }}:{{ index }} </template>`)},
	}

	var buf bytes.Buffer
	vue := vuego.NewVue(templateFS)
	err := vue.RenderFragment(&buf, "test.vuego", map[string]any{"items": []int{1, 2, 3}})

	require.NoError(t, err)
	require.Equal(t, "1:0 2:0 3:0 ", buf.String())
}

// TestTemplate_ForStatefulAttribute tests that bound attributes on a template with v-for
// can use expressions that reference context variables.
// Note: Uses :printed instead of :count due to a quirk in variable name handling.
func TestTemplate_ForStatefulAttribute(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`<template v-for="item in items" :printed="0"><template v-if="item > 1" :printed="printed+1">{{ item }}:{{ printed }} </template></template>`)},
	}

	var buf bytes.Buffer
	vue := vuego.NewVue(templateFS)
	err := vue.RenderFragment(&buf, "test.vuego", map[string]any{"items": []int{1, 2, 3}})

	require.NoError(t, err)
	// Item 1: printed=0, v-if false (1 > 1 is false), nothing rendered
	// Item 2: printed=0, v-if true (2 > 1 is true), :printed="0+1"=1, renders "2:1 "
	// Item 3: printed=0, v-if true (3 > 1 is true), :printed="0+1"=1, renders "3:1 "
	require.Equal(t, "2:1 3:1 ", buf.String())
}
