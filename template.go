package vuego

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"

	"golang.org/x/net/html"

	"github.com/titpetric/vuego/internal/helpers"
)

// LoadOption is a functional option for configuring Load()
type LoadOption func(*Vue)

// Template represents a prepared vuego template.
// It allows variable assignment and rendering with internal buffering.
type Template interface {
	// Fill sets all variables from the map at once.
	Fill(vars map[string]any) Template

	// Assign sets a single variable.
	Assign(key string, value any) Template

	// GetVar retrieves a variable value.
	GetVar(key string) any

	// GetString retrieves a variable value as a string.
	// Returns an empty string if the value is not a string.
	GetString(key string) string

	// Funcs sets custom template functions, overwriting default keys. Returns the Template for chaining.
	Funcs(funcMap FuncMap) Template

	// RegisterProcessor registers a node processor. Returns the Template for chaining.
	RegisterProcessor(processor NodeProcessor) Template

	// Render processes the template file and writes output to w.
	// If an error occurs, w is unmodified (uses internal buffering).
	// The context can be used to cancel the rendering operation.
	Render(ctx context.Context, w io.Writer, filename string) error

	// RenderString processes a template string and writes output to w.
	// If an error occurs, w is unmodified (uses internal buffering).
	// The context can be used to cancel the rendering operation.
	RenderString(ctx context.Context, w io.Writer, templateStr string) error

	// RenderByte processes template bytes and writes output to w.
	// If an error occurs, w is unmodified (uses internal buffering).
	// The context can be used to cancel the rendering operation.
	RenderByte(ctx context.Context, w io.Writer, templateData []byte) error

	// RenderReader processes template data from a reader and writes output to w.
	// If an error occurs, w is unmodified (uses internal buffering).
	// The context can be used to cancel the rendering operation.
	RenderReader(ctx context.Context, w io.Writer, r io.Reader) error
}

// template is the internal implementation of Template.
type template struct {
	vue  *Vue
	vars map[string]any
}

// Load creates a new Template with access to the given filesystem and optional configurations.
// The returned Template can be used to render files, strings, or bytes with variable assignment.
func Load(templateFS fs.FS, opts ...LoadOption) Template {
	vue := NewVue(templateFS)

	// Apply functional options
	for _, opt := range opts {
		opt(vue)
	}

	return &template{
		vue:  vue,
		vars: make(map[string]any),
	}
}

// WithLessProcessor returns a LoadOption that registers a LESS processor
func WithLessProcessor() LoadOption {
	return func(vue *Vue) {
		vue.RegisterNodeProcessor(NewLessProcessor(vue.templateFS))
	}
}

// WithProcessor returns a LoadOption that registers a custom node processor
func WithProcessor(processor NodeProcessor) LoadOption {
	return func(vue *Vue) {
		vue.RegisterNodeProcessor(processor)
	}
}

// Fill sets all variables from the map.
func (t *template) Fill(vars map[string]any) Template {
	if vars != nil {
		for k, v := range vars {
			t.vars[k] = v
		}
	}
	return t
}

// Assign sets a single variable.
func (t *template) Assign(key string, value any) Template {
	t.vars[key] = value
	return t
}

// GetVar retrieves a variable value.
func (t *template) GetVar(key string) any {
	return t.vars[key]
}

// GetString retrieves a variable value as a string.
// Returns an empty string if the value is not a string.
func (t *template) GetString(key string) string {
	val := t.vars[key]
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

// Funcs sets custom template functions, overwriting default keys. Returns the Template for chaining.
func (t *template) Funcs(funcMap FuncMap) Template {
	// Merge funcMap into the default funcMap, overwriting keys
	defaultFuncs := t.vue.funcMap
	for k, v := range funcMap {
		defaultFuncs[k] = v
	}
	t.vue.funcMap = defaultFuncs
	return t
}

// RegisterProcessor registers a node processor. Returns the Template for chaining.
func (t *template) RegisterProcessor(processor NodeProcessor) Template {
	t.vue.RegisterNodeProcessor(processor)
	return t
}

// Render processes the template file and writes the output to w.
// Uses internal buffering so w is unmodified if an error occurs.
// The context can be used to cancel the rendering operation.
func (t *template) Render(ctx context.Context, w io.Writer, filename string) error {
	// Check if context is already cancelled before starting
	if err := ctx.Err(); err != nil {
		return err
	}

	// Load the template with front-matter
	frontMatter, dom, err := t.vue.loader.loadFragmentInternal(filename)
	if err != nil {
		return err
	}

	// Merge front-matter with existing variables
	vars := make(map[string]any)
	for k, v := range t.vars {
		vars[k] = v
	}
	if frontMatter != nil {
		for k, v := range frontMatter {
			vars[k] = v
		}
	}

	vueCtx := NewVueContext(filename, &VueContextOptions{
		Data:         vars,
		OriginalData: nil,
		Processors:   t.vue.nodeProcessors,
	})

	if err := t.vue.preProcessNodes(vueCtx, dom); err != nil {
		return err
	}

	// Check context cancellation before evaluation
	if err := ctx.Err(); err != nil {
		return err
	}

	result, err := t.vue.evaluate(vueCtx, dom, 0)
	if err != nil {
		return err
	}

	// Check context cancellation before post-processing
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := t.vue.postProcessNodes(vueCtx, result); err != nil {
		return err
	}

	// Check context cancellation before rendering
	if err := ctx.Err(); err != nil {
		return err
	}

	// Buffer the output to ensure w is unmodified on error
	buf := &bytes.Buffer{}
	if err := t.vue.render(buf, result); err != nil {
		return err
	}

	// Only write to w if rendering succeeded
	_, err = buf.WriteTo(w)
	return err
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

	vueCtx := NewVueContext("<inline>", &VueContextOptions{
		Data:         t.vars,
		OriginalData: nil,
		Processors:   t.vue.nodeProcessors,
	})

	if err := t.vue.preProcessNodes(vueCtx, dom); err != nil {
		return err
	}

	// Check context cancellation before evaluation
	if err := ctx.Err(); err != nil {
		return err
	}

	result, err := t.vue.evaluate(vueCtx, dom, 0)
	if err != nil {
		return err
	}

	// Check context cancellation before post-processing
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := t.vue.postProcessNodes(vueCtx, result); err != nil {
		return err
	}

	// Check context cancellation before rendering
	if err := ctx.Err(); err != nil {
		return err
	}

	// Buffer the output to ensure w is unmodified on error
	buf := &bytes.Buffer{}
	if err := t.vue.render(buf, result); err != nil {
		return err
	}

	// Only write to w if rendering succeeded
	_, err = buf.WriteTo(w)
	return err
}
