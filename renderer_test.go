package vuego_test

import (
	"bytes"
	"strings"
	"testing"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/tests/assert"
)

// parseNodes parses source as a body fragment and returns its nodes.
func parseNodes(t *testing.T, source string) []*html.Node {
	t.Helper()

	nodes, err := html.ParseFragment(strings.NewReader(source), &html.Node{
		Type:     html.ElementNode,
		DataAtom: atom.Body,
		Data:     "body",
	})
	assert.NoError(t, err)
	return nodes
}

// renderNodes renders source through a Renderer and returns the output.
func renderNodes(t *testing.T, source string) string {
	t.Helper()

	var buf bytes.Buffer
	assert.NoError(t, vuego.NewRenderer().Render(t.Context(), &buf, parseNodes(t, source)))
	return buf.String()
}

func TestNewRenderer(t *testing.T) {
	t.Run("returns a usable Renderer", func(t *testing.T) {
		renderer := vuego.NewRenderer()
		assert.NotNil(t, renderer)

		var buf bytes.Buffer
		assert.NoError(t, renderer.Render(t.Context(), &buf, nil))
		assert.Equal(t, "", buf.String())
	})

	t.Run("each call returns a separate value", func(t *testing.T) {
		assert.NotNil(t, vuego.NewRenderer())
		assert.NotNil(t, vuego.NewRenderer())
	})
}

func TestRenderer(t *testing.T) {
	t.Run("renders every node it is given", func(t *testing.T) {
		got := renderNodes(t, "<p>a</p><p>b</p>")
		assert.Equal(t, "<p>a</p>\n<p>b</p>\n", got)
	})

	t.Run("indents nested elements by two spaces", func(t *testing.T) {
		got := renderNodes(t, "<div><section><p>deep</p></section></div>")
		assert.Equal(t, "<div>\n  <section>\n    <p>deep</p>\n  </section>\n</div>\n", got)
	})

	t.Run("keeps attributes", func(t *testing.T) {
		assert.Contains(t, renderNodes(t, `<div class="a" id="x">y</div>`), `<div class="a" id="x">y</div>`)
	})

	// Vue directives are evaluation-only and never reach the output.
	directives := []string{
		"v-if", "v-keep", "v-else-if", "v-else", "v-for", "v-pre",
		"v-html", "v-text", "v-show", "v-once", "v-once-id",
	}

	for _, directive := range directives {
		t.Run("drops "+directive, func(t *testing.T) {
			got := renderNodes(t, `<span `+directive+`="x">text</span>`)
			assert.NotContains(t, got, directive)
			assert.Contains(t, got, "text")
		})
	}

	t.Run("a bracketed attribute name loses its brackets", func(t *testing.T) {
		got := renderNodes(t, `<div [v-if]="kept">x</div>`)
		assert.Contains(t, got, `v-if="kept"`)
		assert.NotContains(t, got, "[v-if]")
	})

	escaping := []struct {
		name   string
		source string
		want   string
	}{
		{name: "an ampersand in an attribute is escaped", source: `<div title="a & b"></div>`, want: `title="a &amp; b"`},
		{name: "an escaped entity is not escaped twice", source: `<div title="a &amp; b"></div>`, want: `title="a &amp; b"`},
		{name: "a bare less-than in text is escaped", source: `<p>a < b</p>`, want: "a &lt; b"},
		{name: "text holding an entity is not escaped twice", source: `<p>a &lt; b</p>`, want: "a &lt; b"},
	}

	for _, test := range escaping {
		t.Run(test.name, func(t *testing.T) {
			assert.Contains(t, renderNodes(t, test.source), test.want)
		})
	}

	t.Run("an attribute value already holding an entity is left alone", func(t *testing.T) {
		// html.Parse decodes entities, so the node is built by hand to keep
		// the raw value the renderer sees.
		node := &html.Node{
			Type:     html.ElementNode,
			DataAtom: atom.Div,
			Data:     "div",
			Attr:     []html.Attribute{{Key: "title", Val: "a &#38; b"}},
		}

		var buf bytes.Buffer
		assert.NoError(t, vuego.NewRenderer().Render(t.Context(), &buf, []*html.Node{node}))
		assert.Contains(t, buf.String(), `title="a &#38; b"`)
	})

	t.Run("script content is not escaped", func(t *testing.T) {
		got := renderNodes(t, "<script>if (a < b && c > d) {}</script>")
		assert.Contains(t, got, "if (a < b && c > d) {}")
		assert.NotContains(t, got, "&lt;")
	})

	t.Run("style content is not escaped", func(t *testing.T) {
		got := renderNodes(t, "<style>a > b { color: red }</style>")
		assert.Contains(t, got, "a > b { color: red }")
		assert.NotContains(t, got, "&gt;")
	})

	t.Run("a template tag is unwrapped", func(t *testing.T) {
		got := renderNodes(t, "<template><p>inner</p></template>")
		assert.Equal(t, "<p>inner</p>\n", got)
		assert.NotContains(t, got, "<template")
	})

	t.Run("v-keep keeps the template tag and drops the directive", func(t *testing.T) {
		got := renderNodes(t, `<template v-keep class="wrap"><p>inner</p></template>`)
		assert.Contains(t, got, `<template class="wrap">`)
		assert.Contains(t, got, "<p>inner</p>")
		assert.Contains(t, got, "</template>")
		assert.NotContains(t, got, "v-keep")
	})

	t.Run("whitespace-only text nodes are dropped", func(t *testing.T) {
		assert.Equal(t, "<div>\n  <p>a</p>\n</div>\n", renderNodes(t, "<div>\n\n   <p>a</p>\n\n</div>"))
	})

	t.Run("an empty element stays on one line", func(t *testing.T) {
		assert.Equal(t, "<div></div>\n", renderNodes(t, "<div></div>"))
	})
}
