package vuego

import (
	"bytes"
	"context"
	"io"
	"io/fs"

	"golang.org/x/net/html"
)

// LoadOption is a functional option for configuring Load()
type LoadOption func(*Vue)

// Template represents a loaded and prepared vuego template.
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

	// Render processes the template and writes output to w.
	// If an error occurs, w is unmodified (uses internal buffering).
	// The context can be used to cancel the rendering operation.
	Render(ctx context.Context, w io.Writer) error
}

// template is the internal implementation of Template.
type template struct {
	vue         *Vue
	filename    string
	dom         []*html.Node
	vars        map[string]any
	frontMatter map[string]any
}

// Load loads a template from the given filesystem and filename with optional configurations.
// The template's initial variables are populated from front-matter data.
// Supports functional options to register processors and configure template rendering.
func Load(templateFS fs.FS, filename string, opts ...LoadOption) (Template, error) {
	vue := NewVue(templateFS)

	// Apply functional options
	for _, opt := range opts {
		opt(vue)
	}

	// Load the template with front-matter
	frontMatter, dom, err := vue.loader.loadFragmentInternal(filename)
	if err != nil {
		return nil, err
	}

	// Initialize variables with front-matter data
	vars := make(map[string]any)
	if frontMatter != nil {
		for k, v := range frontMatter {
			vars[k] = v
		}
	}

	return &template{
		vue:         vue,
		filename:    filename,
		dom:         dom,
		vars:        vars,
		frontMatter: frontMatter,
	}, nil
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

// Render processes the template and writes the output to w.
// Uses internal buffering so w is unmodified if an error occurs.
// The context can be used to cancel the rendering operation.
func (t *template) Render(ctx context.Context, w io.Writer) error {
	// Check if context is already cancelled before starting
	if err := ctx.Err(); err != nil {
		return err
	}

	vueCtx := NewVueContext(t.filename, &VueContextOptions{
		Data:         t.vars,
		OriginalData: nil,
		Processors:   t.vue.nodeProcessors,
	})

	if err := t.vue.preProcessNodes(vueCtx, t.dom); err != nil {
		return err
	}

	// Check context cancellation before evaluation
	if err := ctx.Err(); err != nil {
		return err
	}

	result, err := t.vue.evaluate(vueCtx, t.dom, 0)
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
