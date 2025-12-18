package vuego_test

import (
	"bytes"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/internal/helpers"
)

func TestVue_EvalVText(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]any
		expected string
	}{
		{
			name:     "simple v-text",
			template: "<div v-text=\"text\"></div>",
			data:     map[string]any{"text": "plain text"},
			expected: "<div>plain text</div>",
		},
		{
			name:     "v-text escapes HTML",
			template: "<div v-text=\"text\"></div>",
			data:     map[string]any{"text": "<span>inner</span>"},
			expected: "<div>&lt;span&gt;inner&lt;/span&gt;</div>",
		},
		{
			name:     "v-text with empty string",
			template: "<div v-text=\"text\">original</div>",
			data:     map[string]any{"text": ""},
			expected: "<div></div>",
		},
		{
			name:     "v-text missing variable",
			template: "<div v-text=\"missing\">fallback</div>",
			data:     map[string]any{},
			expected: "<div>fallback</div>",
		},
		{
			name:     "v-text replaces children",
			template: "<div v-text=\"text\"><p>original</p></div>",
			data:     map[string]any{"text": "replaced"},
			expected: "<div>replaced</div>",
		},
		{
			name:     "v-text with special characters",
			template: "<div v-text=\"text\"></div>",
			data:     map[string]any{"text": "a & b < c > d"},
			expected: "<div>a &amp; b &lt; c &gt; d</div>",
		},
		{
			name:     "v-text with quotes",
			template: "<div v-text=\"text\"></div>",
			data:     map[string]any{"text": `"quoted" value`},
			expected: "<div>&quot;quoted&quot; value</div>",
		},
		{
			name:     "v-text same as interpolation",
			template: "<span v-text=\"msg\"></span>",
			data:     map[string]any{"msg": "hello"},
			expected: "<span>hello</span>",
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

func TestVue_EvalVText_PreservesWhitespace(t *testing.T) {
	// Test that v-text content preserves its original whitespace without re-indenting
	tests := []struct {
		name     string
		template string
		data     map[string]any
		expected string // exact expected output (not normalized)
	}{
		{
			name:     "v-text preserves newlines in text",
			template: "<div v-text=\"text\"></div>",
			data:     map[string]any{"text": "line1\nline2"},
			expected: "<div>line1\nline2</div>\n",
		},
		{
			name:     "v-text escapes HTML with newlines",
			template: "<div v-text=\"text\"></div>",
			data: map[string]any{
				"text": "<span>\n  content\n</span>",
			},
			// v-text content is escaped and output as-is without re-indenting
			expected: "<div>&lt;span&gt;\n  content\n&lt;/span&gt;</div>\n",
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
			// Compare exact output to verify whitespace is preserved
			require.Equal(t, tc.expected, buf.String())
		})
	}
}
