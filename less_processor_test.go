package vuego_test

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"golang.org/x/net/html"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/tests/assert"
)

var _ vuego.NodeProcessor = &vuego.LessProcessor{}

func TestLessProcessorError(t *testing.T) {
	t.Run("carries the reason and the cause", func(t *testing.T) {
		cause := errors.New("unexpected token")
		err := &vuego.LessProcessorError{Err: cause, Reason: "failed to parse LESS"}

		assert.Equal(t, cause, err.Err)
		assert.Equal(t, "failed to parse LESS", err.Reason)
	})
}

func TestLessProcessorError_Error(t *testing.T) {
	tests := []struct {
		name   string
		reason string
		cause  string
		want   string
	}{
		{name: "parse failure", reason: "failed to parse LESS", cause: "unexpected token", want: "failed to parse LESS: unexpected token"},
		{name: "render failure", reason: "failed to render LESS to CSS", cause: "unknown variable", want: "failed to render LESS to CSS: unknown variable"},
		{name: "empty reason", reason: "", cause: "boom", want: ": boom"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := &vuego.LessProcessorError{Err: errors.New(test.cause), Reason: test.reason}
			assert.Equal(t, test.want, err.Error())
		})
	}
}

func TestNewLessProcessor(t *testing.T) {
	t.Run("without a filesystem", func(t *testing.T) {
		assert.NotNil(t, vuego.NewLessProcessor())
	})

	t.Run("with a filesystem for @import", func(t *testing.T) {
		assert.NotNil(t, vuego.NewLessProcessor(fstest.MapFS{}))
	})
}

func TestLessProcessor_New(t *testing.T) {
	t.Run("returns a separate allocation", func(t *testing.T) {
		processor := vuego.NewLessProcessor()

		scoped := processor.New()
		assert.NotNil(t, scoped)
		assert.True(t, scoped != vuego.NodeProcessor(processor), "the processor was not copied")
	})

	t.Run("the copy compiles the same input", func(t *testing.T) {
		nodes := parseNodes(t, `<style type="text/css+less">@c: red; a { color: @c; }</style>`)

		scoped := vuego.NewLessProcessor().New()
		assert.NoError(t, scoped.PostProcess(nodes))
		assert.Contains(t, renderToString(t, nodes), "color: red")
	})
}

func TestLessProcessor_PreProcess(t *testing.T) {
	t.Run("leaves the nodes alone", func(t *testing.T) {
		nodes := parseNodes(t, `<style type="text/css+less">@c: red; a { color: @c; }</style>`)
		before := renderToString(t, nodes)

		assert.NoError(t, vuego.NewLessProcessor().PreProcess(nodes))
		assert.Equal(t, before, renderToString(t, nodes))
	})

	t.Run("accepts no nodes", func(t *testing.T) {
		assert.NoError(t, vuego.NewLessProcessor().PreProcess(nil))
	})
}

func TestLessProcessor_PostProcess(t *testing.T) {
	t.Run("compiles a less style tag", func(t *testing.T) {
		nodes := parseNodes(t, `<style type="text/css+less">@c: red; a { color: @c; }</style>`)

		assert.NoError(t, vuego.NewLessProcessor().PostProcess(nodes))

		got := renderToString(t, nodes)
		assert.Contains(t, got, `type="text/css"`)
		assert.NotContains(t, got, "text/css+less")
		assert.Contains(t, got, "color: red")
	})

	t.Run("leaves a plain style tag alone", func(t *testing.T) {
		nodes := parseNodes(t, `<style>a { color: red; }</style>`)

		assert.NoError(t, vuego.NewLessProcessor().PostProcess(nodes))
		assert.Contains(t, renderToString(t, nodes), "a { color: red; }")
	})

	t.Run("an empty less tag is left empty", func(t *testing.T) {
		nodes := parseNodes(t, `<style type="text/css+less"></style>`)

		assert.NoError(t, vuego.NewLessProcessor().PostProcess(nodes))
		assert.Contains(t, renderToString(t, nodes), "text/css+less")
	})

	t.Run("finds a nested less tag", func(t *testing.T) {
		nodes := parseNodes(t, `<div><section><style type="text/css+less">@c: blue; a { color: @c; }</style></section></div>`)

		assert.NoError(t, vuego.NewLessProcessor().PostProcess(nodes))
		assert.Contains(t, renderToString(t, nodes), "color: blue")
	})

	// The lessgo parser is tolerant: an unterminated rule yields no output
	// rather than an error, so the compiled tag ends up empty.
	t.Run("malformed LESS compiles to nothing", func(t *testing.T) {
		nodes := parseNodes(t, `<style type="text/css+less">a { color: </style>`)

		assert.NoError(t, vuego.NewLessProcessor().PostProcess(nodes))

		got := renderToString(t, nodes)
		assert.Contains(t, got, `type="text/css"`)
		assert.NotContains(t, got, "color")
	})

	t.Run("an unresolved variable is passed through", func(t *testing.T) {
		nodes := parseNodes(t, `<style type="text/css+less">a { color: @undefined; }</style>`)

		assert.NoError(t, vuego.NewLessProcessor().PostProcess(nodes))
		assert.Contains(t, renderToString(t, nodes), "color: @undefined")
	})

	t.Run("accepts no nodes", func(t *testing.T) {
		assert.NoError(t, vuego.NewLessProcessor().PostProcess(nil))
	})
}

func TestLessProcessor_Render(t *testing.T) {
	templateFS := os.DirFS("testdata/nodeprocessor")

	newVue := func() *vuego.Vue {
		vue := vuego.NewVue(templateFS)
		vue.RegisterNodeProcessor(vuego.NewLessProcessor())
		return vue
	}

	tests := []struct {
		name     string
		filename string
		contains []string
		absent   []string
	}{
		{
			name:     "compiles a less style tag",
			filename: "less.html",
			contains: []string{`<style type="text/css">`, "color: red;"},
			absent:   []string{`type="text/css+less"`},
		},
		{
			name:     "resolves less variables",
			filename: "less_variables.html",
			contains: []string{`<style type="text/css">`, "#ff0000"},
			absent:   []string{`type="text/css+less"`},
		},
		{
			name:     "leaves plain style tags alone",
			filename: "mixed_styles.html",
			contains: []string{`<style type="text/css">`, "color: #333;"},
			absent:   []string{`type="text/css+less"`},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			assert.NoError(t, newVue().Render(t.Context(), &buf, test.filename, map[string]any{}))

			for _, want := range test.contains {
				assert.Contains(t, buf.String(), want)
			}
			for _, unwanted := range test.absent {
				assert.NotContains(t, buf.String(), unwanted)
			}
		})
	}
}

// renderToString renders nodes with the standard html package so a processor's
// effect on the tree can be asserted without going through a full render.
func renderToString(t *testing.T, nodes []*html.Node) string {
	t.Helper()

	var b strings.Builder
	for _, node := range nodes {
		assert.NoError(t, html.Render(&b, node))
	}
	return b.String()
}
