package vuego_test

import (
	"testing"

	"github.com/titpetric/vuego/tests/assert"
)

func TestVue_Interpolate(t *testing.T) {
	tests := []struct {
		name   string
		source string
		data   map[string]any
		want   string
	}{
		{
			name:   "a variable",
			source: "<div>{{ name }}</div>",
			data:   map[string]any{"name": "Alice"},
			want:   "<div>Alice</div>",
		},
		{
			name:   "a missing variable is empty",
			source: "<div>{{ missing }}</div>",
			data:   map[string]any{},
			want:   "<div></div>",
		},
		{
			name:   "nil is empty",
			source: "<div>{{ value }}</div>",
			data:   map[string]any{"value": nil},
			want:   "<div></div>",
		},
		{
			name:   "text with no interpolation is left alone",
			source: "<div>plain text</div>",
			data:   map[string]any{},
			want:   "<div>plain text</div>",
		},
		{
			name:   "a number",
			source: "<div>{{ number }}</div>",
			data:   map[string]any{"number": 42},
			want:   "<div>42</div>",
		},
		{
			name:   "a bool",
			source: "<div>{{ flag }}</div>",
			data:   map[string]any{"flag": true},
			want:   "<div>true</div>",
		},
		{
			name:   "two interpolations",
			source: "<p>{{ first }} {{ last }}</p>",
			data:   map[string]any{"first": "John", "last": "Doe"},
			want:   "<p>John Doe</p>",
		},
		{
			name:   "whitespace inside the braces is trimmed",
			source: "<div>{{   value   }}</div>",
			data:   map[string]any{"value": "test"},
			want:   "<div>test</div>",
		},
		{
			name:   "a path expression",
			source: "<div>{{ user.name }}</div>",
			data:   map[string]any{"user": map[string]any{"name": "Alice"}},
			want:   "<div>Alice</div>",
		},

		// Interpolated values are HTML-escaped.
		{
			name:   "markup is escaped",
			source: "<div>{{ value }}</div>",
			data:   map[string]any{"value": "<script>alert('xss')</script>"},
			want:   "<div>&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;</div>",
		},
		{
			name:   "an ampersand is escaped",
			source: "<div>{{ value }}</div>",
			data:   map[string]any{"value": "a & b"},
			want:   "<div>a &amp; b</div>",
		},

		// Ternary and comparison.
		{
			name:   "a ternary taking the true branch",
			source: "<div>{{ flag ? 'yes' : 'no' }}</div>",
			data:   map[string]any{"flag": true},
			want:   "<div>yes</div>",
		},
		{
			name:   "a ternary taking the false branch",
			source: "<div>{{ flag ? 'yes' : 'no' }}</div>",
			data:   map[string]any{"flag": false},
			want:   "<div>no</div>",
		},
		{
			name:   "a ternary on a comparison",
			source: "<div>{{ age > 18 ? 'adult' : 'minor' }}</div>",
			data:   map[string]any{"age": 25},
			want:   "<div>adult</div>",
		},
		{
			name:   "a comparison renders its boolean",
			source: "<p>Status: {{ status == 'active' }}</p>",
			data:   map[string]any{"status": "active"},
			want:   "<p>Status: true</p>",
		},
		{
			name:   "a comparison that does not hold",
			source: "<p>Status: {{ status == 'active' }}</p>",
			data:   map[string]any{"status": "inactive"},
			want:   "<p>Status: false</p>",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertRenderHTML(t, test.source, test.data, test.want)
		})
	}
}

func TestVue_Interpolate_RawContexts(t *testing.T) {
	// Inside a script or style tag the value is written as it is, because
	// escaping there would break the code the value lands in.
	raw := []struct {
		name     string
		source   string
		data     map[string]any
		contains string
		absent   []string
	}{
		{
			name:     "a plain string in a script tag",
			source:   `<script>const msg = "{{ message }}";</script>`,
			data:     map[string]any{"message": "hello world"},
			contains: `const msg = "hello world";`,
		},
		{
			name:     "markup in a script tag",
			source:   `<script>const value = "{{ htmlContent }}";</script>`,
			data:     map[string]any{"htmlContent": `<div class="test">content</div>`},
			contains: `<div class="test">content</div>`,
			absent:   []string{"&lt;", "&gt;", "&#34;"},
		},
		{
			name:     "punctuation in a script tag",
			source:   `<script>const val = '{{ val }}';</script>`,
			data:     map[string]any{"val": `test & check < > " '`},
			contains: `test & check < > " '`,
			absent:   []string{"&amp;", "&lt;", "&gt;"},
		},
		{
			name:     "a value in a style tag",
			source:   `<style>a { color: {{ color }}; }</style>`,
			data:     map[string]any{"color": "red"},
			contains: "a { color: red; }",
		},
	}

	for _, test := range raw {
		t.Run(test.name, func(t *testing.T) {
			got := renderPage(t, test.source, test.data)
			assert.Contains(t, got, test.contains)
			for _, unwanted := range test.absent {
				assert.NotContains(t, got, unwanted)
			}
		})
	}

	t.Run("escaping still applies outside a script tag", func(t *testing.T) {
		got := renderFragment(t, `<div>{{ value }}</div>`, map[string]any{"value": `<script>alert('xss')</script>`})
		assert.Contains(t, got, "&lt;script&gt;")
		assert.NotContains(t, got, "<script>")
	})
}
