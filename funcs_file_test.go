package vuego_test

import (
	"testing"

	"github.com/titpetric/vuego/tests/assert"
)

func TestFileFunc(t *testing.T) {
	files := map[string]string{
		"content.txt":        "Hello from file",
		"data.txt":           "Variable file content",
		"nested/content.txt": "Nested file content",
		"snippet.html":       "<strong>Bold text</strong>",
	}

	tests := []struct {
		name     string
		source   string
		data     map[string]any
		contains string
	}{
		{
			name:     "a literal filename",
			source:   `<div>{{ file("content.txt") }}</div>`,
			data:     map[string]any{},
			contains: "Hello from file",
		},
		{
			name:     "a variable filename",
			source:   `<div>{{ file(filename) }}</div>`,
			data:     map[string]any{"filename": "data.txt"},
			contains: "Variable file content",
		},
		{
			name:     "a path expression filename",
			source:   `<div>{{ file(config.path) }}</div>`,
			data:     map[string]any{"config": map[string]any{"path": "data.txt"}},
			contains: "Variable file content",
		},
		{
			name:     "a nested path",
			source:   `<div>{{ file("nested/content.txt") }}</div>`,
			data:     map[string]any{},
			contains: "Nested file content",
		},
		{
			name:     "beside other text",
			source:   `<div>Content: {{ file("data.txt") }}</div>`,
			data:     map[string]any{},
			contains: "Variable file content",
		},
		// The content is HTML-escaped like any other interpolated value.
		{
			name:     "HTML content is escaped",
			source:   `<div>{{ file("snippet.html") }}</div>`,
			data:     map[string]any{},
			contains: "&lt;strong&gt;",
		},
		{
			name:     "in a script tag the content is not escaped",
			source:   `<script>const s = {{ file("snippet.html") }};</script>`,
			data:     map[string]any{},
			contains: "<strong>Bold text</strong>",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sources := map[string]string{"test.vuego": test.source}
			for name, content := range files {
				sources[name] = content
			}

			got, err := renderFuncs(t, sources, test.data, nil)
			assert.NoError(t, err)
			assert.Contains(t, got, test.contains)
		})
	}

	t.Run("a v-html directive inserts the file as markup", func(t *testing.T) {
		got, err := renderFuncs(t, map[string]string{
			"test.vuego":   `<div v-html="file(html.path)"></div>`,
			"snippet.html": `<span>html snippet</span>`,
		}, map[string]any{"html": map[string]any{"path": "snippet.html"}}, nil)

		assert.NoError(t, err)
		assert.Contains(t, got, "<span>html snippet</span>")
	})

	t.Run("a bound attribute takes the file content as its value", func(t *testing.T) {
		got, err := renderFuncs(t, map[string]string{
			"test.vuego": `<div :data-content="file(dataPath)"></div>`,
			"data.txt":   "attribute value",
		}, map[string]any{"dataPath": "data.txt"}, nil)

		assert.NoError(t, err)
		assert.Contains(t, got, `data-content="attribute value"`)
	})

	failures := []struct {
		name   string
		source string
	}{
		{name: "a missing file", source: `<div>{{ file("missing.txt") }}</div>`},
		{name: "an empty filename", source: `<div>{{ file("") }}</div>`},
	}

	for _, test := range failures {
		t.Run(test.name, func(t *testing.T) {
			_, err := renderFuncs(t, map[string]string{"test.vuego": test.source}, map[string]any{}, nil)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "file")
		})
	}

	t.Run("without a filesystem the error names the file", func(t *testing.T) {
		tpl := `<div>{{ file("content.txt") }}</div>`

		_, err := renderFuncs(t, map[string]string{"test.vuego": tpl}, map[string]any{}, nil)
		assert.Error(t, err)
	})
}

func TestJSONFileFunc(t *testing.T) {
	// jsonFile returns the decoded mapping. A call cannot be the left side of a
	// pipe, so a template either renders the value whole or pipes the filename
	// into the function.
	files := map[string]string{
		"config.json": `{"name": "test"}`,
		"list.json":   `{"items": ["a", "b"]}`,
		"broken.json": `{"name":`,
	}

	withFiles := func(source string) map[string]string {
		sources := map[string]string{"test.vuego": source}
		for name, content := range files {
			sources[name] = content
		}
		return sources
	}

	t.Run("a decoded mapping is rendered", func(t *testing.T) {
		got, err := renderFuncs(t, withFiles(`<p>{{ jsonFile("config.json") }}</p>`), map[string]any{}, nil)
		assert.NoError(t, err)
		assert.Contains(t, got, "name:test")
	})

	t.Run("the decoded value is a map", func(t *testing.T) {
		got, err := renderFuncs(t, withFiles(`<p>{{ path | jsonFile | type }}</p>`), map[string]any{"path": "config.json"}, nil)
		assert.NoError(t, err)
		assert.Contains(t, got, "map[string]interface {}")
	})

	t.Run("a decoded mapping is countable", func(t *testing.T) {
		got, err := renderFuncs(t, withFiles(`<p>{{ path | jsonFile | len }}</p>`), map[string]any{"path": "list.json"}, nil)
		assert.NoError(t, err)
		assert.Contains(t, got, "1")
	})

	t.Run("a variable filename resolves", func(t *testing.T) {
		got, err := renderFuncs(t, withFiles(`<p>{{ path | jsonFile }}</p>`), map[string]any{"path": "config.json"}, nil)
		assert.NoError(t, err)
		assert.Contains(t, got, "name:test")
	})

	t.Run("invalid JSON names the file", func(t *testing.T) {
		_, err := renderFuncs(t, withFiles(`<p>{{ jsonFile("broken.json") }}</p>`), map[string]any{}, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error decoding json from file broken.json")
	})

	t.Run("a missing file is an error", func(t *testing.T) {
		_, err := renderFuncs(t, withFiles(`<p>{{ jsonFile("missing.json") }}</p>`), map[string]any{}, nil)
		assert.Error(t, err)
	})
}

func TestYAMLFileFunc(t *testing.T) {
	files := map[string]string{
		"config.yml": "name: test\n",
		"list.yml":   "items:\n  - a\n  - b\n",
		"broken.yml": "\tname: [",
	}

	withFiles := func(source string) map[string]string {
		sources := map[string]string{"test.vuego": source}
		for name, content := range files {
			sources[name] = content
		}
		return sources
	}

	t.Run("a decoded mapping is rendered", func(t *testing.T) {
		got, err := renderFuncs(t, withFiles(`<p>{{ yamlFile("config.yml") }}</p>`), map[string]any{}, nil)
		assert.NoError(t, err)
		assert.Contains(t, got, "name:test")
	})

	t.Run("the decoded value is a map", func(t *testing.T) {
		got, err := renderFuncs(t, withFiles(`<p>{{ path | yamlFile | type }}</p>`), map[string]any{"path": "config.yml"}, nil)
		assert.NoError(t, err)
		assert.Contains(t, got, "map[string]interface {}")
	})

	t.Run("a decoded mapping is countable", func(t *testing.T) {
		got, err := renderFuncs(t, withFiles(`<p>{{ path | yamlFile | len }}</p>`), map[string]any{"path": "list.yml"}, nil)
		assert.NoError(t, err)
		assert.Contains(t, got, "1")
	})

	t.Run("a variable filename resolves", func(t *testing.T) {
		got, err := renderFuncs(t, withFiles(`<p>{{ path | yamlFile }}</p>`), map[string]any{"path": "config.yml"}, nil)
		assert.NoError(t, err)
		assert.Contains(t, got, "name:test")
	})

	t.Run("invalid YAML names the file", func(t *testing.T) {
		_, err := renderFuncs(t, withFiles(`<p>{{ yamlFile("broken.yml") }}</p>`), map[string]any{}, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error decoding yaml from file broken.yml")
	})

	t.Run("a missing file is an error", func(t *testing.T) {
		_, err := renderFuncs(t, withFiles(`<p>{{ yamlFile("missing.yml") }}</p>`), map[string]any{}, nil)
		assert.Error(t, err)
	})
}
