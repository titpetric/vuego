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

func TestVue_ObjectBinding_Class(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]any
		expected string
	}{
		{
			name:     "object binding with truthy values",
			template: `<div :class="{active: true, disabled: false}"></div>`,
			data:     map[string]any{},
			expected: `<div class="active"></div>`,
		},
		{
			name:     "object binding with variable values",
			template: `<div :class="{active: isActive, error: hasError}"></div>`,
			data:     map[string]any{"isActive": true, "hasError": false},
			expected: `<div class="active"></div>`,
		},
		{
			name:     "all classes included when truthy",
			template: `<div :class="{btn: true, primary: true, large: true}"></div>`,
			data:     map[string]any{},
			expected: `<div class="btn primary large"></div>`,
		},
		{
			name:     "no classes when all falsey",
			template: `<div :class="{active: false, disabled: false}"></div>`,
			data:     map[string]any{},
			expected: `<div></div>`,
		},
		{
			name:     "class object with static class",
			template: `<div class="static-class" :class="{dynamic: true}"></div>`,
			data:     map[string]any{},
			expected: `<div class="static-class dynamic"></div>`,
		},
		{
			name:     "class object with zero is falsey",
			template: `<div :class="{active: count}"></div>`,
			data:     map[string]any{"count": 0},
			expected: `<div></div>`,
		},
		{
			name:     "class object with non-zero is truthy",
			template: `<div :class="{active: count}"></div>`,
			data:     map[string]any{"count": 1},
			expected: `<div class="active"></div>`,
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

func TestVue_ObjectBinding_Style(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]any
		expected string
	}{
		{
			name:     "style object with string values",
			template: `<div :style="{display: 'block', color: 'red'}"></div>`,
			data:     map[string]any{},
			expected: `<div style="display:block;color:red;"></div>`,
		},
		{
			name:     "style object with variable values",
			template: `<div :style="{color: textColor, fontSize: size}"></div>`,
			data:     map[string]any{"textColor": "blue", "size": "16px"},
			expected: `<div style="color:blue;font-size:16px;"></div>`,
		},
		{
			name:     "style object with single property",
			template: `<div :style="{display: 'none'}"></div>`,
			data:     map[string]any{},
			expected: `<div style="display:none;"></div>`,
		},
		{
			name:     "style object combined with static style",
			template: `<div style="padding: 10px;" :style="{color: 'red'}"></div>`,
			data:     map[string]any{},
			expected: `<div style="padding:10px;color:red;"></div>`,
		},
		{
			name:     "style object with numeric values",
			template: `<div :style="{width: '100px', height: '50px'}"></div>`,
			data:     map[string]any{},
			expected: `<div style="width:100px;height:50px;"></div>`,
		},
		{
			name:     "style object overwrites conflicting static style",
			template: `<div style="color: blue;" :style="{color: 'red'}"></div>`,
			data:     map[string]any{},
			expected: `<div style="color:red;"></div>`,
		},
		{
			name:     "camelCase properties converted to kebab-case",
			template: `<div :style="{fontSize: '16px', backgroundColor: 'blue', marginTop: '10px'}"></div>`,
			data:     map[string]any{},
			expected: `<div style="font-size:16px;background-color:blue;margin-top:10px;"></div>`,
		},
		{
			name:     "camelCase with variable values",
			template: `<div :style="{fontSize: size, fontWeight: weight}"></div>`,
			data:     map[string]any{"size": "14px", "weight": "bold"},
			expected: `<div style="font-size:14px;font-weight:bold;"></div>`,
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

func TestVue_CombinedStyleAttributes(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]any
		expected string
	}{
		{
			name:     "static style with bound style",
			template: `<div style="color: red;" :style="{fontSize: size}"></div>`,
			data:     map[string]any{"size": "14px"},
			expected: `<div style="color:red;font-size:14px;"></div>`,
		},
		{
			name:     "v-show with static class",
			template: `<div v-show="visible" class="box"></div>`,
			data:     map[string]any{"visible": false},
			expected: `<div class="box" style="display:none;"></div>`,
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
