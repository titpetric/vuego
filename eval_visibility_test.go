package vuego_test

import (
	"bytes"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/internal/helpers"
)

func TestVue_EvalVShow(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]any
		expected string
	}{
		{
			name:     "v-show with truthy condition",
			template: `<div v-show="visible"></div>`,
			data:     map[string]any{"visible": true},
			expected: `<div></div>`,
		},
		{
			name:     "v-show with falsey condition",
			template: `<div v-show="visible"></div>`,
			data:     map[string]any{"visible": false},
			expected: `<div style="display:none;"></div>`,
		},
		{
			name:     "v-show hides when condition is false",
			template: `<p v-show="show">Content</p>`,
			data:     map[string]any{"show": false},
			expected: `<p style="display:none;">Content</p>`,
		},
		{
			name:     "v-show preserves existing styles",
			template: `<div v-show="visible" style="color: red; font-size: 14px;"></div>`,
			data:     map[string]any{"visible": false},
			expected: `<div style="color:red;font-size:14px;display:none;"></div>`,
		},
		{
			name:     "v-show with expression truthy",
			template: `<span v-show="true"></span>`,
			data:     map[string]any{},
			expected: `<span></span>`,
		},
		{
			name:     "v-show with expression falsey",
			template: `<span v-show="count > 0"></span>`,
			data:     map[string]any{"count": 0},
			expected: `<span style="display:none;"></span>`,
		},
		{
			name:     "v-show with missing variable",
			template: `<div v-show="missing"></div>`,
			data:     map[string]any{},
			expected: `<div style="display:none;"></div>`,
		},
		{
			name:     "v-show overwrites display property",
			template: `<div v-show="visible" style="display: flex;"></div>`,
			data:     map[string]any{"visible": false},
			expected: `<div style="display:none;"></div>`,
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
