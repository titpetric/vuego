package vuego_test

import (
	"bytes"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego"
)

// Article is a sample model type to test with
type Article struct {
	Title string
	Body  string
}

func TestVue_TypedAttributePassthrough(t *testing.T) {
	tests := []struct {
		name     string
		parent   string
		child    string
		data     map[string]any
		expected string
	}{
		{
			name:   "slice of structs preserves type through binding",
			parent: `<vuego include="child.vuego" :articles="articles"></vuego>`,
			child:  `<template>{{ articles | type }}</template>`,
			data: map[string]any{
				"articles": []Article{
					{Title: "Article 1", Body: "Content 1"},
					{Title: "Article 2", Body: "Content 2"},
				},
			},
			expected: `[]vuego_test.Article`,
		},
		{
			name:   "map preserves type through binding",
			parent: `<vuego include="child.vuego" :data="myMap"></vuego>`,
			child:  `<template>{{ data | type }}</template>`,
			data: map[string]any{
				"myMap": map[string]string{"key": "value"},
			},
			expected: `map[string]string`,
		},
		{
			name:   "int preserves type through binding",
			parent: `<vuego include="child.vuego" :count="count"></vuego>`,
			child:  `<template>{{ count | type }}</template>`,
			data: map[string]any{
				"count": 42,
			},
			expected: `int`,
		},
		{
			name:   "complex object type preservation",
			parent: `<vuego include="child.vuego" :article="article"></vuego>`,
			child:  `<template>{{ article | type }}</template>`,
			data: map[string]any{
				"article": Article{Title: "Test", Body: "Test body"},
			},
			expected: `vuego_test.Article`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fs := fstest.MapFS{
				"parent.vuego": &fstest.MapFile{Data: []byte(tc.parent)},
				"child.vuego":  &fstest.MapFile{Data: []byte(tc.child)},
			}
			vue := vuego.NewVue(fs)

			var buf bytes.Buffer
			err := vue.RenderFragment(&buf, "parent.vuego", tc.data)
			require.NoError(t, err)

			output := buf.String()
			require.Contains(t, output, tc.expected, "expected type %s but got %s", tc.expected, output)
		})
	}
}
