package vuego_test

import (
	"bytes"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/diff"
)

func TestVue_Interpolate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]any
		expected string
	}{
		{
			name:     "simple variable",
			template: "<div>{{ name }}</div>",
			data:     map[string]any{"name": "Alice"},
			expected: "<div>Alice</div>",
		},
		{
			name:     "missing variable",
			template: "<div>{{ missing }}</div>",
			data:     map[string]any{},
			expected: "<div></div>",
		},
		{
			name:     "no interpolation",
			template: "<div>plain text</div>",
			data:     map[string]any{},
			expected: "<div>plain text</div>",
		},
		{
			name:     "html escaping",
			template: "<div>{{ value }}</div>",
			data:     map[string]any{"value": "<script>alert('xss')</script>"},
			expected: "<div>&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;</div>",
		},
		{
			name:     "multiple interpolations",
			template: "<p>{{ first }} {{ last }}</p>",
			data:     map[string]any{"first": "John", "last": "Doe"},
			expected: "<p>John Doe</p>",
		},
		{
			name:     "null value",
			template: "<div>{{ value }}</div>",
			data:     map[string]any{"value": nil},
			expected: "<div></div>",
		},
		{
			name:     "numeric value",
			template: "<div>{{ number }}</div>",
			data:     map[string]any{"number": 42},
			expected: "<div>42</div>",
		},
		{
			name:     "boolean value",
			template: "<div>{{ flag }}</div>",
			data:     map[string]any{"flag": true},
			expected: "<div>true</div>",
		},
		{
			name:     "whitespace in interpolation",
			template: "<div>{{   value   }}</div>",
			data:     map[string]any{"value": "test"},
			expected: "<div>test</div>",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			templateBytes := []byte(tc.template)
			fs := fstest.MapFS{
				"test.vuego": &fstest.MapFile{Data: templateBytes},
			}
			vue := vuego.NewVue(fs)

			var buf bytes.Buffer
			err := vue.RenderFragment(&buf, "test.vuego", tc.data)
			require.NoError(t, err)
			require.True(t, diff.EqualHTML(t, []byte(tc.expected), buf.Bytes(), nil, nil))
		})
	}
}
