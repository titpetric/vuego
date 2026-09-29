package diff_test

import (
	"testing"

	yaml "gopkg.in/yaml.v3"

	"github.com/titpetric/vuego/diff"
	"github.com/titpetric/vuego/tests/assert"
)

func TestDomToYAML(t *testing.T) {
	// decode round-trips the YAML so assertions read the structure rather
	// than the serialiser's key order.
	decode := func(t *testing.T, in string) map[string]any {
		t.Helper()
		var out map[string]any
		assert.NoError(t, yaml.Unmarshal([]byte(diff.DomToYAML([]byte(in))), &out))
		return out
	}

	t.Run("element carries its tag", func(t *testing.T) {
		doc := decode(t, "<p>text</p>")
		assert.Equal(t, "html", doc["tag"])
	})

	t.Run("attributes sit beside the tag", func(t *testing.T) {
		out := diff.DomToYAML([]byte(`<div class="a" id="x"></div>`))
		assert.Contains(t, out, "class: a")
		assert.Contains(t, out, "id: x")
	})

	t.Run("text children become text entries", func(t *testing.T) {
		out := diff.DomToYAML([]byte("<p>hello</p>"))
		assert.Contains(t, out, "text: hello")
	})

	t.Run("whitespace-only text children are dropped", func(t *testing.T) {
		out := diff.DomToYAML([]byte("<div>\n   \n</div>"))
		assert.NotContains(t, out, "text:")
	})

	t.Run("comments become comment entries", func(t *testing.T) {
		out := diff.DomToYAML([]byte("<div><!-- note --></div>"))
		assert.Contains(t, out, "comment: note")
	})

	t.Run("childless elements carry no children key", func(t *testing.T) {
		out := diff.DomToYAML([]byte("<hr>"))
		assert.NotContains(t, out, "children:\n    - tag: hr\n      children:")
	})

	t.Run("nesting is preserved", func(t *testing.T) {
		doc := decode(t, "<div><p><span>a</span></p></div>")

		// html -> [head, body] -> div -> p -> span
		htmlChildren, ok := doc["children"].([]any)
		assert.True(t, ok, "html has children")
		assert.Len(t, htmlChildren, 2)

		body, ok := htmlChildren[1].(map[string]any)
		assert.True(t, ok, "second child is a map")
		assert.Equal(t, "body", body["tag"])

		bodyChildren, ok := body["children"].([]any)
		assert.True(t, ok, "body has children")
		div, ok := bodyChildren[0].(map[string]any)
		assert.True(t, ok, "body child is a map")
		assert.Equal(t, "div", div["tag"])
	})

	t.Run("empty input yields the implied document", func(t *testing.T) {
		doc := decode(t, "")
		assert.Equal(t, "html", doc["tag"])
	})
}
