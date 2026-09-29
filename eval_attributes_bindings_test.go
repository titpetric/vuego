package vuego_test

import (
	"testing"
)

func TestVue_ObjectBinding_Class(t *testing.T) {
	// For :class an object literal is read as a set of flags: a key is written
	// when its value is truthy.
	tests := []struct {
		name   string
		source string
		data   map[string]any
		want   string
	}{
		{
			name:   "literal booleans",
			source: `<div :class="{active: true, disabled: false}"></div>`,
			data:   map[string]any{},
			want:   `<div class="active"></div>`,
		},
		{
			name:   "quoted keys",
			source: `<div :class="{'active': true, 'disabled': false}"></div>`,
			data:   map[string]any{},
			want:   `<div class="active"></div>`,
		},
		{
			name:   "variable values",
			source: `<div :class="{active: isActive, error: hasError}"></div>`,
			data:   map[string]any{"isActive": true, "hasError": false},
			want:   `<div class="active"></div>`,
		},
		{
			name:   "every truthy key is written in order",
			source: `<div :class="{btn: true, primary: true, large: true}"></div>`,
			data:   map[string]any{},
			want:   `<div class="btn primary large"></div>`,
		},
		{
			name:   "no truthy key drops the attribute",
			source: `<div :class="{active: false, disabled: false}"></div>`,
			data:   map[string]any{},
			want:   `<div></div>`,
		},
		{
			name:   "an empty object drops the attribute",
			source: `<div :class="{}"></div>`,
			data:   map[string]any{},
			want:   `<div></div>`,
		},
		{
			name:   "a static class is kept and appended to",
			source: `<div class="static-class" :class="{dynamic: true}"></div>`,
			data:   map[string]any{},
			want:   `<div class="static-class dynamic"></div>`,
		},
		{
			name:   "zero is falsey",
			source: `<div :class="{active: count}"></div>`,
			data:   map[string]any{"count": 0},
			want:   `<div></div>`,
		},
		{
			name:   "a non-zero number is truthy",
			source: `<div :class="{active: count}"></div>`,
			data:   map[string]any{"count": 1},
			want:   `<div class="active"></div>`,
		},
		{
			name:   "negation of true",
			source: `<div :class="{active: status, inactive: !status}"></div>`,
			data:   map[string]any{"status": true},
			want:   `<div class="active"></div>`,
		},
		{
			name:   "negation of false",
			source: `<div :class="{active: status, inactive: !status}"></div>`,
			data:   map[string]any{"status": false},
			want:   `<div class="inactive"></div>`,
		},
		{
			name:   "a missing variable is falsey",
			source: `<div :class="{active: status?, inactive: !(type(status) == 'bool' && status)}"></div>`,
			data:   map[string]any{},
			want:   `<div class="inactive"></div>`,
		},
		{
			name:   "whitespace around keys and colons is trimmed",
			source: `<div :class="{ active : isActive , error : hasError }"></div>`,
			data:   map[string]any{"isActive": true, "hasError": false},
			want:   `<div class="active"></div>`,
		},
		{
			name:   "literals and variables mix",
			source: `<div :class="{btn: true, active: isActive, disabled: hasError}"></div>`,
			data:   map[string]any{"isActive": true, "hasError": false},
			want:   `<div class="btn active"></div>`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertRenderHTML(t, test.source, test.data, test.want)
		})
	}
}

func TestVue_ObjectBinding_Style(t *testing.T) {
	// For :style an object literal becomes CSS declarations, with camelCase
	// property names converted to kebab-case.
	tests := []struct {
		name   string
		source string
		data   map[string]any
		want   string
	}{
		{
			name:   "string values",
			source: `<div :style="{display: 'block', color: 'red'}"></div>`,
			data:   map[string]any{},
			want:   `<div style="display:block;color:red;"></div>`,
		},
		{
			name:   "one property",
			source: `<div :style="{display: 'none'}"></div>`,
			data:   map[string]any{},
			want:   `<div style="display:none;"></div>`,
		},
		{
			name:   "variable values",
			source: `<div :style="{color: textColor, fontSize: size}"></div>`,
			data:   map[string]any{"textColor": "blue", "size": "16px"},
			want:   `<div style="color:blue;font-size:16px;"></div>`,
		},
		{
			name:   "lengths are kept verbatim",
			source: `<div :style="{width: '100px', height: '50px'}"></div>`,
			data:   map[string]any{},
			want:   `<div style="width:100px;height:50px;"></div>`,
		},
		{
			name:   "a numeric string value is kept",
			source: `<div :style="{margin: '0', padding: '10px'}"></div>`,
			data:   map[string]any{},
			want:   `<div style="margin:0;padding:10px;"></div>`,
		},
		{
			name:   "camelCase becomes kebab-case",
			source: `<div :style="{fontSize: '16px', backgroundColor: 'blue', marginTop: '10px'}"></div>`,
			data:   map[string]any{},
			want:   `<div style="font-size:16px;background-color:blue;margin-top:10px;"></div>`,
		},
		{
			name:   "camelCase with variable values",
			source: `<div :style="{fontSize: size, fontWeight: weight}"></div>`,
			data:   map[string]any{"size": "14px", "weight": "bold"},
			want:   `<div style="font-size:14px;font-weight:bold;"></div>`,
		},
		{
			name:   "an empty object drops the attribute",
			source: `<div :style="{}"></div>`,
			data:   map[string]any{},
			want:   `<div></div>`,
		},
		{
			name:   "a static style is merged with",
			source: `<div style="padding: 10px;" :style="{color: 'red'}"></div>`,
			data:   map[string]any{},
			want:   `<div style="padding:10px;color:red;"></div>`,
		},
		{
			name:   "the bound value wins on a conflict",
			source: `<div style="color: blue;" :style="{color: 'red'}"></div>`,
			data:   map[string]any{},
			want:   `<div style="color:red;"></div>`,
		},
		{
			name:   "a static style and a bound property both survive",
			source: `<div style="color: red;" :style="{fontSize: size}"></div>`,
			data:   map[string]any{"size": "14px"},
			want:   `<div style="color:red;font-size:14px;"></div>`,
		},
		{
			name:   "v-show adds a display declaration beside a static class",
			source: `<div v-show="visible" class="box"></div>`,
			data:   map[string]any{"visible": false},
			want:   `<div class="box" style="display:none;"></div>`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertRenderHTML(t, test.source, test.data, test.want)
		})
	}
}

func TestVue_ObjectBinding_Other(t *testing.T) {
	// For any other attribute the values are joined with a space.
	tests := []struct {
		name   string
		source string
		data   map[string]any
		want   string
	}{
		{
			name:   "values are joined",
			source: `<div :data-flags="{a: 'one', b: 'two'}"></div>`,
			data:   map[string]any{},
			want:   `<div data-flags="a:one b:two"></div>`,
		},
		{
			name:   "an empty object drops the attribute",
			source: `<div :data-flags="{}"></div>`,
			data:   map[string]any{},
			want:   `<div></div>`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertRenderHTML(t, test.source, test.data, test.want)
		})
	}
}
