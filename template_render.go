package vuego

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"golang.org/x/net/html"

	"github.com/titpetric/vuego/internal/helpers"
)

// Render processes the loaded template and writes the output to w.
// Requires that Load() has been called first.
func (t *template) Render(ctx context.Context, w io.Writer) error {
	// Check if context is already cancelled before starting
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := t.Err(); err != nil {
		return err
	}
	if !t.filenameLoaded {
		return fmt.Errorf("no template loaded; call Load() first")
	}

	return t.vue.Render(w, t.filename, t.stack.EnvMap())
}

// RenderFile processes the template file and writes the output to w.
// Uses internal buffering so w is unmodified if an error occurs.
// The context can be used to cancel the rendering operation.
func (t *template) RenderFile(ctx context.Context, w io.Writer, filename string) error {
	return t.Load(filename).Render(ctx, w)
}

// RenderString processes a template string and writes the output to w.
// Uses internal buffering so w is unmodified if an error occurs.
// The context can be used to cancel the rendering operation.
func (t *template) RenderString(ctx context.Context, w io.Writer, templateStr string) error {
	return t.RenderByte(ctx, w, []byte(templateStr))
}

// RenderByte processes template bytes and writes the output to w.
// Uses internal buffering so w is unmodified if an error occurs.
// The context can be used to cancel the rendering operation.
func (t *template) RenderByte(ctx context.Context, w io.Writer, templateData []byte) error {
	return t.RenderReader(ctx, w, bytes.NewReader(templateData))
}

// RenderReader processes template data from a reader and writes the output to w.
// Uses internal buffering so w is unmodified if an error occurs.
// The context can be used to cancel the rendering operation.
func (t *template) RenderReader(ctx context.Context, w io.Writer, r io.Reader) error {
	// Check if context is already cancelled before starting
	if err := ctx.Err(); err != nil {
		return err
	}

	// Parse the template from reader as a fragment
	body := helpers.GetBodyNode()
	dom, err := html.ParseFragment(r, body)
	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}

	// Create VueContext with filename from loaded template
	vueCtx := NewVueContext(t.filename, &VueContextOptions{
		Stack:      t.stack.Copy(),
		Processors: t.vue.nodeProcessors,
	})

	// Buffer the output to ensure w is unmodified on error
	buf := &bytes.Buffer{}
	if err := t.vue.renderNodesWithContext(vueCtx, buf, dom); err != nil {
		return err
	}

	// Check context cancellation before writing
	if err := ctx.Err(); err != nil {
		return err
	}

	// Only write to w if rendering succeeded
	_, err = buf.WriteTo(w)
	return err
}
