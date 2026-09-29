package vuego_test

import (
	"bytes"
	"testing"
	"testing/fstest"

	"golang.org/x/net/html"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/tests/assert"
)

// recordingProcessor records the calls it receives and rewrites <mark> tags to
// <em> in PostProcess.
type recordingProcessor struct {
	name  string
	calls *[]string
}

func (p *recordingProcessor) New() vuego.NodeProcessor {
	// Render-scoped allocation: the copy shares the call log so a test can see
	// what the scoped instance did.
	return &recordingProcessor{name: p.name, calls: p.calls}
}

func (p *recordingProcessor) PreProcess(nodes []*html.Node) error {
	*p.calls = append(*p.calls, p.name+":pre")
	return nil
}

func (p *recordingProcessor) PostProcess(nodes []*html.Node) error {
	*p.calls = append(*p.calls, p.name+":post")
	for _, node := range nodes {
		renameTag(node, "mark", "em")
	}
	return nil
}

func renameTag(node *html.Node, from, to string) {
	if node.Type == html.ElementNode && node.Data == from {
		node.Data = to
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		renameTag(c, from, to)
	}
}

// errorProcessor fails in the phase named by failOn.
type errorProcessor struct {
	failOn string
	err    error
}

func (p *errorProcessor) New() vuego.NodeProcessor { return p }

func (p *errorProcessor) PreProcess([]*html.Node) error {
	if p.failOn == "pre" {
		return p.err
	}
	return nil
}

func (p *errorProcessor) PostProcess([]*html.Node) error {
	if p.failOn == "post" {
		return p.err
	}
	return nil
}

func TestNodeProcessor(t *testing.T) {
	const filename = "template.vuego"

	newVue := func(source string) *vuego.Vue {
		return vuego.NewVue(fstest.MapFS{filename: {Data: []byte(source)}})
	}

	t.Run("PreProcess and PostProcess both run", func(t *testing.T) {
		var calls []string
		vue := newVue(`<div><mark>x</mark></div>`)
		vue.RegisterNodeProcessor(&recordingProcessor{name: "a", calls: &calls})

		var buf bytes.Buffer
		assert.NoError(t, vue.RenderFragment(t.Context(), &buf, filename, nil))
		assert.Equal(t, []string{"a:pre", "a:post"}, calls)
	})

	t.Run("PostProcess edits reach the output", func(t *testing.T) {
		var calls []string
		vue := newVue(`<div><mark>x</mark></div>`)
		vue.RegisterNodeProcessor(&recordingProcessor{name: "a", calls: &calls})

		var buf bytes.Buffer
		assert.NoError(t, vue.RenderFragment(t.Context(), &buf, filename, nil))
		assert.Contains(t, buf.String(), "<em>x</em>")
		assert.NotContains(t, buf.String(), "<mark>")
	})

	t.Run("processors run in registration order", func(t *testing.T) {
		var calls []string
		vue := newVue(`<div>x</div>`)
		vue.RegisterNodeProcessor(&recordingProcessor{name: "first", calls: &calls})
		vue.RegisterNodeProcessor(&recordingProcessor{name: "second", calls: &calls})

		var buf bytes.Buffer
		assert.NoError(t, vue.RenderFragment(t.Context(), &buf, filename, nil))
		assert.Equal(t, []string{"first:pre", "second:pre", "first:post", "second:post"}, calls)
	})

	t.Run("a PreProcess error stops the render", func(t *testing.T) {
		vue := newVue(`<div>x</div>`)
		vue.RegisterNodeProcessor(&errorProcessor{failOn: "pre", err: errWriteFailed})

		var buf bytes.Buffer
		err := vue.RenderFragment(t.Context(), &buf, filename, nil)
		assert.ErrorIs(t, err, errWriteFailed)
		assert.Equal(t, "", buf.String())
	})

	t.Run("a PostProcess error stops the render", func(t *testing.T) {
		vue := newVue(`<div>x</div>`)
		vue.RegisterNodeProcessor(&errorProcessor{failOn: "post", err: errWriteFailed})

		var buf bytes.Buffer
		err := vue.RenderFragment(t.Context(), &buf, filename, nil)
		assert.ErrorIs(t, err, errWriteFailed)
		assert.Equal(t, "", buf.String())
	})
}

func TestVue_RegisterNodeProcessor(t *testing.T) {
	t.Run("returns the receiver for chaining", func(t *testing.T) {
		var calls []string
		vue := vuego.NewVue(nil)
		assert.Equal(t, vue, vue.RegisterNodeProcessor(&recordingProcessor{name: "a", calls: &calls}))
	})

	t.Run("New is called per render so state is not shared", func(t *testing.T) {
		const filename = "template.vuego"

		var calls []string
		vue := vuego.NewVue(fstest.MapFS{filename: {Data: []byte(`<div>x</div>`)}})
		vue.RegisterNodeProcessor(&recordingProcessor{name: "a", calls: &calls})

		var first, second bytes.Buffer
		assert.NoError(t, vue.RenderFragment(t.Context(), &first, filename, nil))
		assert.NoError(t, vue.RenderFragment(t.Context(), &second, filename, nil))
		assert.Equal(t, []string{"a:pre", "a:post", "a:pre", "a:post"}, calls)
	})
}

func TestVue_ComponentTagResolution(t *testing.T) {
	// A registered component tag is rewritten to a template include before any
	// processor runs.
	fsys := fstest.MapFS{
		"page.vuego":                 {Data: []byte(`<div><my-card /></div>`)},
		"components/MyCard.vuego":    {Data: []byte(`<article>card</article>`)},
		"components/nested/Hi.vuego": {Data: []byte(`<span>hi</span>`)},
	}

	t.Run("a registered tag renders its component", func(t *testing.T) {
		vue := vuego.NewVue(fsys)
		vue.RegisterComponent("my-card", "components/MyCard.vuego")

		var buf bytes.Buffer
		assert.NoError(t, vue.RenderFragment(t.Context(), &buf, "page.vuego", nil))
		assert.Contains(t, buf.String(), "<article>card</article>")
		assert.NotContains(t, buf.String(), "my-card")
	})

	t.Run("an unregistered tag renders as-is", func(t *testing.T) {
		vue := vuego.NewVue(fsys)

		var buf bytes.Buffer
		assert.NoError(t, vue.RenderFragment(t.Context(), &buf, "page.vuego", nil))
		assert.Contains(t, buf.String(), "my-card")
	})

	t.Run("WithComponents registers nested files under a kebab-case name", func(t *testing.T) {
		tpl := vuego.NewFS(fsys, vuego.WithComponents())

		var buf bytes.Buffer
		assert.NoError(t, tpl.RenderString(t.Context(), &buf, `<div><my-card></my-card><nested-hi></nested-hi></div>`))
		assert.Contains(t, buf.String(), "<article>card</article>")
		assert.Contains(t, buf.String(), "<span>hi</span>")
	})
}
