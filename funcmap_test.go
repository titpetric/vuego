package vuego_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/tests/assert"
)

func TestFuncMap(t *testing.T) {
	t.Run("a declared FuncMap is callable", func(t *testing.T) {
		var funcs vuego.FuncMap = map[string]any{
			"greet": func(v any) any { return "Hello, " + v.(string) },
		}

		got, err := renderFuncs(t,
			map[string]string{"test.vuego": `<p>{{ name | greet }}</p>`},
			map[string]any{"name": "Alice"},
			funcs,
		)
		assert.NoError(t, err)
		assert.Contains(t, got, "Hello, Alice")
	})

	t.Run("an unknown filter is an error naming the filter", func(t *testing.T) {
		_, err := renderFuncs(t,
			map[string]string{"test.vuego": `<p>{{ name | nonexistent }}</p>`},
			map[string]any{"name": "test"},
			nil,
		)
		assert.Error(t, err)
		assert.Equal(t, "in test.vuego: in expression '{{ name | nonexistent }}': function 'nonexistent' not found", err.Error())
	})
}

func TestBuiltinFilters(t *testing.T) {
	// Each built-in filter is applied to a value and its output read back out
	// of the rendered text.
	tests := []struct {
		name     string
		source   string
		data     map[string]any
		contains string
		absent   string
	}{
		{name: "type", source: `<p>{{ value | type }}</p>`, data: map[string]any{"value": 42}, contains: "int"},
		{name: "upper", source: `<p>{{ name | upper }}</p>`, data: map[string]any{"name": "john doe"}, contains: "JOHN DOE"},
		{name: "lower", source: `<p>{{ name | lower }}</p>`, data: map[string]any{"name": "JOHN DOE"}, contains: "john doe"},
		{name: "title", source: `<p>{{ name | title }}</p>`, data: map[string]any{"name": "john doe"}, contains: "John Doe"},
		{name: "upper leaves a non-string alone", source: `<p>{{ value | upper }}</p>`, data: map[string]any{"value": 42}, contains: "42"},
		{name: "trim", source: `<p>{{ text | trim }}</p>`, data: map[string]any{"text": "  hello world  "}, contains: "hello world", absent: "  hello world  "},
		{name: "len of a slice", source: `<p>{{ items | len }}</p>`, data: map[string]any{"items": []*http.Request{nil, nil, nil}}, contains: "3"},
		{name: "len of a string", source: `<p>{{ text | len }}</p>`, data: map[string]any{"text": "abcd"}, contains: "4"},
		{name: "len of a map", source: `<p>{{ obj | len }}</p>`, data: map[string]any{"obj": map[string]any{"a": 1, "b": 2}}, contains: "2"},
		{name: "len of a scalar is zero", source: `<p>{{ value | len }}</p>`, data: map[string]any{"value": 42}, contains: "0"},
		{name: "int from a string", source: `<p>{{ value | int }}</p>`, data: map[string]any{"value": "42"}, contains: "42"},
		{name: "int from a float truncates", source: `<p>{{ value | int }}</p>`, data: map[string]any{"value": 3.9}, contains: "3"},
		{name: "int from a non-number is zero", source: `<p>{{ value | int }}</p>`, data: map[string]any{"value": "words"}, contains: "0"},
		{name: "string from a number", source: `<p>{{ count | string }}</p>`, data: map[string]any{"count": 42}, contains: "42"},
		{name: "default fills an empty value", source: `<p>{{ value | default("fallback") }}</p>`, data: map[string]any{"value": ""}, contains: "fallback"},
		{name: "default leaves a set value", source: `<p>{{ name | default("Anonymous") }}</p>`, data: map[string]any{"name": "Bob"}, contains: "Bob", absent: "Anonymous"},
		// escape exists for raw contexts: inside a script tag the renderer
		// does not escape, so the filter is what produces the entities.
		{name: "escape in a script tag", source: `<script>const s = "{{ value | escape }}";</script>`, data: map[string]any{"value": "<b>"}, contains: "&lt;b&gt;"},
		{name: "json", source: `<script>const v = {{ value | json }};</script>`, data: map[string]any{"value": map[string]any{"a": 1}}, contains: `{"a":1}`},
		{name: "jsonPretty", source: `<script>const v = {{ value | jsonPretty }};</script>`, data: map[string]any{"value": map[string]any{"a": 1}}, contains: `"a": 1`},

		// Chains apply left to right.
		{name: "two filters", source: `<p>{{ name | lower | title }}</p>`, data: map[string]any{"name": "JOHN DOE"}, contains: "John Doe"},
		{name: "three filters", source: `<p>{{ text | trim | lower | title }}</p>`, data: map[string]any{"text": "  HELLO WORLD  "}, contains: "Hello World"},
		{name: "default then upper", source: `<p>{{ name | default("guest") | upper }}</p>`, data: map[string]any{"name": ""}, contains: "GUEST"},

		// Filters apply to path expressions and inside v-for.
		{name: "a path expression", source: `<p>{{ user.name | upper }}</p>`, data: map[string]any{"user": map[string]any{"name": "alice"}}, contains: "ALICE"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := renderPage(t, test.source, test.data)
			assert.Contains(t, got, test.contains)
			if test.absent != "" {
				assert.NotContains(t, got, test.absent)
			}
		})
	}

	t.Run("formatTime uses the layout it is given", func(t *testing.T) {
		now := time.Now()
		got := renderPage(t, `<p>{{ timestamp | formatTime("2006-01-02") }}</p>`, map[string]any{"timestamp": now})
		assert.Contains(t, got, now.Format("2006-01-02"))
	})

	t.Run("formatTime parses an RFC3339 string", func(t *testing.T) {
		got := renderPage(t, `<p>{{ timestamp | formatTime("2006-01-02") }}</p>`, map[string]any{"timestamp": "2024-03-01T10:00:00Z"})
		assert.Contains(t, got, "2024-03-01")
	})

	t.Run("formatDate takes a layout too", func(t *testing.T) {
		now := time.Now()
		got := renderPage(t, `<p>{{ timestamp | formatDate("Jan 2 2006") }}</p>`, map[string]any{"timestamp": now})
		assert.Contains(t, got, now.Format("Jan 2 2006"))
	})

	t.Run("a filter runs per v-for item", func(t *testing.T) {
		got := renderPage(t,
			"<ul>\n  <li v-for=\"item in items\">{{ item.name | upper }}</li>\n</ul>",
			map[string]any{"items": []map[string]any{{"name": "apple"}, {"name": "banana"}}},
		)
		assert.Contains(t, got, "APPLE")
		assert.Contains(t, got, "BANANA")
	})

	t.Run("a filter can drive v-if", func(t *testing.T) {
		assert.Contains(t,
			renderPage(t, `<div v-if="len(items)"><p>Has items</p></div>`, map[string]any{"items": []string{"a", "b", "c"}}),
			"Has items",
		)
		assert.NotContains(t,
			renderPage(t, `<div v-if="len(items)"><p>Has items</p></div>`, map[string]any{"items": []string{}}),
			"Has items",
		)
	})
}

func TestBuiltinFilters_JSONEscaping(t *testing.T) {
	// JSON is written raw inside a script tag and HTML-escaped everywhere else.
	data := map[string]any{"data": map[string]any{"status": "active", "count": 42}}

	raw := []struct {
		name   string
		source string
	}{
		{name: "script with json", source: `<script>const d = {{ data | json }};</script>`},
		{name: "script with jsonPretty", source: `<script>const d = {{ data | jsonPretty }};</script>`},
	}

	for _, test := range raw {
		t.Run(test.name, func(t *testing.T) {
			got := renderPage(t, test.source, data)
			assert.Contains(t, got, `"status"`)
			assert.Contains(t, got, `"active"`)
			assert.NotContains(t, got, "&#34;")
		})
	}

	escaped := []struct {
		name   string
		source string
	}{
		{name: "code tag", source: `<code>{{ data | json }}</code>`},
		{name: "pre tag", source: `<pre>{{ data | json }}</pre>`},
		{name: "pre and code", source: `<pre><code>{{ data | json }}</code></pre>`},
		{name: "div", source: `<div>{{ data | json }}</div>`},
	}

	for _, test := range escaped {
		t.Run(test.name, func(t *testing.T) {
			got := renderPage(t, test.source, data)
			assert.Contains(t, got, "&#34;status&#34;")
			assert.Contains(t, got, "&#34;active&#34;")
			assert.NotContains(t, got, `"status": "active"`)
		})
	}

	t.Run("a fragment escapes JSON too", func(t *testing.T) {
		got := renderFragment(t, `{{ data | json }}`, map[string]any{"data": map[string]any{"key": "value"}})
		assert.Contains(t, got, "&#34;key&#34;")
		assert.Contains(t, got, "&#34;value&#34;")
	})

	t.Run("a nested structure keeps its keys", func(t *testing.T) {
		got := renderPage(t,
			`<script>const payload = {{ payload | jsonPretty }};</script>`,
			map[string]any{"payload": map[string]any{
				"user":   map[string]any{"name": "Charlie", "tags": []string{"admin", "user"}},
				"status": "active",
			}},
		)
		assert.Contains(t, got, `"name": "Charlie"`)
		assert.Contains(t, got, `"status": "active"`)
		assert.NotContains(t, got, "&#34;")
	})
}
