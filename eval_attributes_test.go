package vuego_test

import (
	"bytes"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/internal/helpers"
)

func TestVue_EvalAttributes_BoundAndInterpolated(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]any
		expected string
	}{
		{
			name:     "colon prefix binding",
			template: `<div :id="id"></div>`,
			data:     map[string]any{"id": "test-div"},
			expected: `<div id="test-div"></div>`,
		},
		{
			name:     "v-bind prefix binding",
			template: `<div v-bind:class="class"></div>`,
			data:     map[string]any{"class": "active"},
			expected: `<div class="active"></div>`,
		},
		{
			name:     "bound attribute with falsey value",
			template: `<button :disabled="disabled"></button>`,
			data:     map[string]any{"disabled": false},
			expected: `<button></button>`,
		},
		{
			name:     "bound attribute with truthy value",
			template: `<button :disabled="disabled"></button>`,
			data:     map[string]any{"disabled": true},
			expected: `<button disabled="true"></button>`,
		},
		{
			name:     "normal attribute with interpolation",
			template: `<div class="item-{{ id }}"></div>`,
			data:     map[string]any{"id": "123"},
			expected: `<div class="item-123"></div>`,
		},
		{
			name:     "normal attribute without interpolation",
			template: `<div class="static"></div>`,
			data:     map[string]any{},
			expected: `<div class="static"></div>`,
		},
		{
			name:     "bound attribute missing variable",
			template: `<div :title="missing"></div>`,
			data:     map[string]any{},
			expected: `<div></div>`,
		},
		{
			name:     "empty string is falsey",
			template: `<input :required="value" />`,
			data:     map[string]any{"value": ""},
			expected: `<input />`,
		},
		{
			name:     "zero is falsey",
			template: `<span :data-count="count"></span>`,
			data:     map[string]any{"count": 0},
			expected: `<span></span>`,
		},
		{
			name:     "non-zero is truthy",
			template: `<span :data-count="count"></span>`,
			data:     map[string]any{"count": 5},
			expected: `<span data-count="5"></span>`,
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
