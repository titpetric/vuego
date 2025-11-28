package vuego_test

import (
	"bytes"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/internal/helpers"
)

func TestVue_EvalVHtml(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]any
		expected string
	}{
		{
			name:     "simple v-html",
			template: "<div v-html=\"html\"></div>",
			data:     map[string]any{"html": "<span>inner</span>"},
			expected: "<div><span>inner</span></div>",
		},
		{
			name:     "v-html with empty string",
			template: "<div v-html=\"html\">original</div>",
			data:     map[string]any{"html": ""},
			expected: "<div></div>",
		},
		{
			name:     "v-html missing variable",
			template: "<div v-html=\"missing\">fallback</div>",
			data:     map[string]any{},
			expected: "<div>fallback</div>",
		},
		{
			name:     "v-html replaces children",
			template: "<div v-html=\"content\"><p>original</p></div>",
			data:     map[string]any{"content": "<b>replaced</b>"},
			expected: "<div><b>replaced</b></div>",
		},
		{
			name:     "v-html with multiple elements",
			template: "<div v-html=\"html\"></div>",
			data:     map[string]any{"html": "<p>one</p><p>two</p>"},
			expected: "<div><p>one</p><p>two</p></div>",
		},
		{
			name:     "v-html with nested html",
			template: "<div v-html=\"html\"></div>",
			data:     map[string]any{"html": "<div><span><b>nested</b></span></div>"},
			expected: "<div><div><span><b>nested</b></span></div></div>",
		},
		{
			name:     "v-html with text content",
			template: "<div v-html=\"html\"></div>",
			data:     map[string]any{"html": "plain text"},
			expected: "<div>plain text</div>",
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
