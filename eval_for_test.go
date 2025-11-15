package vuego_test

import (
	"bytes"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/internal/helpers"
)

func TestVue_EvalFor(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]any
		expected string
	}{
		{
			name:     "simple v-for with item variable",
			template: "<ul><li v-for=\"item in items\">{{ item }}</li></ul>",
			data:     map[string]any{"items": []string{"a", "b", "c"}},
			expected: "<ul><li>a</li><li>b</li><li>c</li></ul>",
		},
		{
			name:     "v-for with index and value",
			template: "<ul><li v-for=\"(i, item) in items\">{{ i }}: {{ item }}</li></ul>",
			data:     map[string]any{"items": []string{"a", "b"}},
			expected: "<ul><li>0: a</li><li>1: b</li></ul>",
		},
		{
			name:     "v-for with nested content",
			template: "<div v-for=\"item in items\"><span>{{ item.name }}</span></div>",
			data: map[string]any{"items": []map[string]any{
				{"name": "Alice"},
				{"name": "Bob"},
			}},
			expected: "<div><span>Alice</span></div><div><span>Bob</span></div>",
		},
		{
			name:     "v-for with empty list",
			template: "<ul><li v-for=\"item in items\">{{ item }}</li></ul>",
			data:     map[string]any{"items": []string{}},
			expected: "<ul></ul>",
		},
		{
			name:     "v-for with single item",
			template: "<div v-for=\"item in items\">{{ item }}</div>",
			data:     map[string]any{"items": []int{42}},
			expected: "<div>42</div>",
		},
		{
			name:     "v-for with numbers",
			template: "<span v-for=\"n in nums\">{{ n }}</span>",
			data:     map[string]any{"nums": []int{1, 2, 3}},
			expected: "<span>1</span><span>2</span><span>3</span>",
		},
		{
			name:     "v-for with nested loops",
			template: "<div v-for=\"item in items\"><span v-for=\"tag in item.tags\">{{ tag }}</span></div>",
			data: map[string]any{"items": []map[string]any{
				{"tags": []string{"a", "b"}},
				{"tags": []string{"c"}},
			}},
			expected: "<div><span>a</span><span>b</span></div><div><span>c</span></div>",
		},
		{
			name:     "v-for removes v-for attribute",
			template: "<li v-for=\"item in items\" v-for=\"skip-this\">{{ item }}</li>",
			data:     map[string]any{"items": []string{"x"}},
			expected: "<li>x</li>",
		},
		{
			name:     "v-for with struct field access and negation",
			template: "<span v-for=\"item in items\"><button v-if=\"!item.disabled\">{{ item.name }}</button></span>",
			data: map[string]any{"items": []map[string]any{
				{"name": "Page 1", "disabled": false},
				{"name": "Page 2", "disabled": true},
				{"name": "Page 3", "disabled": false},
			}},
			expected: "<span><button>Page 1</button></span><span></span><span><button>Page 3</button></span>",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			template := []byte(tc.template)
			fs := fstest.MapFS{
				"test.vuego": &fstest.MapFile{Data: template},
			}
			vue := vuego.NewVue(fs)

			var buf bytes.Buffer
			err := vue.RenderFragment(&buf, "test.vuego", tc.data)
			require.NoError(t, err)
			require.True(t, helpers.CompareHTML(t, []byte(tc.expected), buf.Bytes(), nil, nil))
		})
	}
}

func TestParseFor(t *testing.T) {
	// Note: parseFor is not exported, so we test it indirectly through eval
	// This test ensures v-for parsing works correctly
	template := []byte("<div v-for=\"item in items\">{{ item }}</div>")
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{Data: template},
	}
	vue := vuego.NewVue(fs)

	var buf bytes.Buffer
	err := vue.RenderFragment(&buf, "test.vuego", map[string]any{"items": []string{"a"}})
	require.NoError(t, err)
	require.True(t, helpers.CompareHTML(t, []byte("<div>a</div>"), buf.Bytes(), nil, nil))
}
