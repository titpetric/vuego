package vuego_test

import (
	"context"
	"testing"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/tests/assert"
)

func newContext(t *testing.T, filename string, data map[string]any) vuego.VueContext {
	t.Helper()

	return vuego.NewVueContext(t.Context(), filename, &vuego.VueContextOptions{
		Stack: vuego.NewStack(data),
	})
}

func TestVueContextOptions(t *testing.T) {
	t.Run("the stack is the one the context resolves against", func(t *testing.T) {
		stack := vuego.NewStack(map[string]any{"name": "ada"})

		ctx := vuego.NewVueContext(t.Context(), "page.vuego", &vuego.VueContextOptions{Stack: stack})
		assert.Equal(t, stack, ctx.Stack())
	})

	t.Run("processors are allocated per context", func(t *testing.T) {
		var calls []string
		processor := &recordingProcessor{name: "a", calls: &calls}

		ctx := vuego.NewVueContext(t.Context(), "page.vuego", &vuego.VueContextOptions{
			Stack:      vuego.NewStack(nil),
			Processors: []vuego.NodeProcessor{processor},
		})

		assert.Len(t, ctx.Processors, 1)
		// New() returns a render-scoped copy, not the registered value.
		assert.True(t, ctx.Processors[0] != vuego.NodeProcessor(processor), "processor was not copied")
	})

	t.Run("no processors yields none", func(t *testing.T) {
		assert.Len(t, newContext(t, "page.vuego", nil).Processors, 0)
	})
}

func TestNewVueContext(t *testing.T) {
	t.Run("seeds the template chain and the current directory", func(t *testing.T) {
		ctx := newContext(t, "layouts/pages/index.vuego", nil)

		assert.Equal(t, "layouts/pages/index.vuego", ctx.FromFilename)
		assert.Equal(t, "layouts/pages", ctx.CurrentDir)
		assert.Equal(t, []string{"layouts/pages/index.vuego"}, ctx.TemplateStack)
		assert.Len(t, ctx.TagStack, 0)
	})

	t.Run("a bare filename has the current directory as its directory", func(t *testing.T) {
		assert.Equal(t, ".", newContext(t, "index.vuego", nil).CurrentDir)
	})
}

func TestVueContext_WithTemplate(t *testing.T) {
	base := newContext(t, "pages/index.vuego", map[string]any{"name": "ada"})

	t.Run("extends the inclusion chain", func(t *testing.T) {
		got := base.WithTemplate("components/Card.vuego")

		assert.Equal(t, []string{"pages/index.vuego", "components/Card.vuego"}, got.TemplateStack)
		assert.Equal(t, "components/Card.vuego", got.FromFilename)
		assert.Equal(t, "components", got.CurrentDir)
	})

	t.Run("leaves the receiver alone", func(t *testing.T) {
		_ = base.WithTemplate("components/Card.vuego")
		assert.Equal(t, []string{"pages/index.vuego"}, base.TemplateStack)
	})

	t.Run("shares the stack", func(t *testing.T) {
		got := base.WithTemplate("components/Card.vuego")
		assert.Equal(t, base.Stack(), got.Stack())
	})

	t.Run("chains further", func(t *testing.T) {
		got := base.WithTemplate("a.vuego").WithTemplate("b.vuego")
		assert.Equal(t, []string{"pages/index.vuego", "a.vuego", "b.vuego"}, got.TemplateStack)
	})
}

func TestVueContext_FormatTemplateChain(t *testing.T) {
	t.Run("a single template is its own chain", func(t *testing.T) {
		assert.Equal(t, "index.vuego", newContext(t, "index.vuego", nil).FormatTemplateChain())
	})

	t.Run("an inclusion chain is joined with arrows", func(t *testing.T) {
		ctx := newContext(t, "index.vuego", nil).WithTemplate("Card.vuego").WithTemplate("Button.vuego")
		assert.Equal(t, "index.vuego -> Card.vuego -> Button.vuego", ctx.FormatTemplateChain())
	})
}

func TestVueContext_PushTag(t *testing.T) {
	t.Run("appends to the tag stack", func(t *testing.T) {
		ctx := newContext(t, "index.vuego", nil)
		ctx.PushTag("div")
		ctx.PushTag("script")

		assert.Equal(t, []string{"div", "script"}, ctx.TagStack)
	})
}

func TestVueContext_PopTag(t *testing.T) {
	t.Run("removes the last tag", func(t *testing.T) {
		ctx := newContext(t, "index.vuego", nil)
		ctx.PushTag("div")
		ctx.PushTag("script")
		ctx.PopTag()

		assert.Equal(t, []string{"div"}, ctx.TagStack)
	})

	t.Run("an empty stack is left empty", func(t *testing.T) {
		ctx := newContext(t, "index.vuego", nil)
		ctx.PopTag()
		assert.Len(t, ctx.TagStack, 0)
	})
}

func TestVueContext_CurrentTag(t *testing.T) {
	t.Run("an empty stack has no current tag", func(t *testing.T) {
		assert.Equal(t, "", newContext(t, "index.vuego", nil).CurrentTag())
	})

	t.Run("returns the innermost tag", func(t *testing.T) {
		ctx := newContext(t, "index.vuego", nil)
		ctx.PushTag("div")
		ctx.PushTag("script")

		assert.Equal(t, "script", ctx.CurrentTag())

		ctx.PopTag()
		assert.Equal(t, "div", ctx.CurrentTag())
	})
}

func TestVueContext_Context(t *testing.T) {
	t.Run("returns the context it was built with", func(t *testing.T) {
		type key struct{}

		parent := context.WithValue(context.Background(), key{}, "value")
		ctx := vuego.NewVueContext(parent, "index.vuego", &vuego.VueContextOptions{
			Stack: vuego.NewStack(nil),
		})

		assert.Equal(t, "value", ctx.Context().Value(key{}))
	})

	t.Run("survives WithTemplate", func(t *testing.T) {
		type key struct{}

		parent := context.WithValue(context.Background(), key{}, "value")
		ctx := vuego.NewVueContext(parent, "index.vuego", &vuego.VueContextOptions{
			Stack: vuego.NewStack(nil),
		})

		assert.Equal(t, "value", ctx.WithTemplate("Card.vuego").Context().Value(key{}))
	})
}

func TestVueContext_Stack(t *testing.T) {
	t.Run("resolves through the stack it holds", func(t *testing.T) {
		ctx := newContext(t, "index.vuego", map[string]any{"name": "ada"})

		got, ok := ctx.Stack().Lookup("name")
		assert.True(t, ok)
		assert.Equal(t, "ada", got)
	})

	t.Run("writes through the returned stack are visible", func(t *testing.T) {
		ctx := newContext(t, "index.vuego", map[string]any{})
		ctx.Stack().Set("added", 1)

		got, ok := ctx.Stack().Lookup("added")
		assert.True(t, ok)
		assert.Equal(t, 1, got)
	})
}

func TestVueContext_ExprEnv(t *testing.T) {
	t.Run("holds the stack values", func(t *testing.T) {
		env := newContext(t, "index.vuego", map[string]any{"name": "ada", "n": 1}).ExprEnv()

		assert.Equal(t, "ada", env["name"])
		assert.Equal(t, 1, env["n"])
	})

	t.Run("a context-taking func is bound to the context", func(t *testing.T) {
		type key struct{}

		parent := context.WithValue(context.Background(), key{}, "bound")
		ctx := vuego.NewVueContext(parent, "index.vuego", &vuego.VueContextOptions{
			Stack: vuego.NewStack(map[string]any{
				"whoami": func(c context.Context) string {
					v, _ := c.Value(key{}).(string)
					return v
				},
			}),
		})

		fn, ok := ctx.ExprEnv()["whoami"].(func() string)
		assert.True(t, ok, "whoami is wrapped to take no arguments")
		assert.Equal(t, "bound", fn())
	})

	t.Run("a func without a context parameter is left alone", func(t *testing.T) {
		ctx := newContext(t, "index.vuego", map[string]any{
			"double": func(n int) int { return n * 2 },
		})

		fn, ok := ctx.ExprEnv()["double"].(func(int) int)
		assert.True(t, ok, "double keeps its signature")
		assert.Equal(t, 4, fn(2))
	})

	t.Run("a non-func value is left alone", func(t *testing.T) {
		env := newContext(t, "index.vuego", map[string]any{"n": 1}).ExprEnv()
		assert.Equal(t, 1, env["n"])
	})
}
