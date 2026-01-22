package vuego

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/titpetric/vuego/internal/helpers"
)

// LoadOption is a functional option for configuring Load().
type LoadOption func(*Vue)

// New creates a new Template for rendering strings, bytes, or readers without a filesystem.
// Use this when templates are provided as strings/bytes rather than loaded from files.
// To render from files, use NewFS(fs) or New(WithFS(fs)) instead.
// The returned Template can be used for variable assignment and rendering.
func New(opts ...LoadOption) Template {
	return NewFS(nil, opts...)
}

// View is a type safety shim to bind a template file to a data model type.
// The returned template should be rendered and discarded.
func View[V any](renderer Template, filename string, data V) Template {
	return renderer.Load(filename).Fill(data)
}

// WithFS returns a LoadOption that sets the filesystem for template loading.
func WithFS(templateFS fs.FS) LoadOption {
	return func(vue *Vue) {
		vue.templateFS = templateFS
		vue.loader = NewLoader(templateFS)
	}
}

// WithFuncs returns a LoadOption that merges custom template functions into the existing funcmap.
func WithFuncs(funcMap FuncMap) LoadOption {
	return func(vue *Vue) {
		vue.Funcs(funcMap)
	}
}

// Template represents a prepared vuego template.
// It allows variable assignment and rendering with internal buffering.
type Template interface {
	TemplateConstructors
	TemplateState
	TemplateRendering
	TemplateRenderingDetail
}

// TemplateConstructors creates new template rendering contexts.
// The results provide request scoped data allocations and should be discarded after use.
type TemplateConstructors interface {
	New() Template
	Load(filename string) Template
}

// TemplateState bundles the interface for template state management.
type TemplateState interface {
	Fill(vars any) Template
	Assign(key string, value any) Template
	Get(key string) string
}

// TemplateRendering bundles the interface for the render functions.
type TemplateRendering interface {
	Render(ctx context.Context, w io.Writer) error
}

// TemplateRenderingDetail the interface for stateless render functions.
type TemplateRenderingDetail interface {
	RenderFile(ctx context.Context, w io.Writer, filename string) error
	RenderString(ctx context.Context, w io.Writer, templateStr string) error
	RenderByte(ctx context.Context, w io.Writer, templateData []byte) error
	RenderReader(ctx context.Context, w io.Writer, r io.Reader) error
}

// template is the internal implementation of Template.
type template struct {
	vue   *Vue
	stack *Stack

	// filename and error for Template.Load
	err            error
	frontMatter    map[string]any
	templateBytes  []byte
	filename       string
	filenameLoaded bool
}

var _ Template = &template{}

// NewFS creates a new Template with access to the given filesystem and optional configurations.
// The returned Template can be used to render files, strings, or bytes with variable assignment.
func NewFS(templateFS fs.FS, opts ...LoadOption) Template {
	vue := NewVue(templateFS)

	// Apply functional options
	for _, opt := range opts {
		opt(vue)
	}

	return &template{
		vue:   vue,
		stack: NewStack(nil),
	}
}

// New will create an empty loaded copy safe for concurrent use.
// It provides an implementation of Template but provides a typed return
// that's open for further modification in Load().
func (t *template) New() Template {
	return t.new()
}

func (t *template) new() *template {
	return &template{
		stack: t.stack.Copy(),
		vue:   t.vue,
	}
}

// WithLessProcessor returns a LoadOption that registers a LESS processor.
func WithLessProcessor() LoadOption {
	return func(vue *Vue) {
		vue.RegisterNodeProcessor(NewLessProcessor(vue.templateFS))
	}
}

// WithProcessor returns a LoadOption that registers a custom node processor.
func WithProcessor(processor NodeProcessor) LoadOption {
	return func(vue *Vue) {
		vue.RegisterNodeProcessor(processor)
	}
}

// WithComponents returns a LoadOption that registers all component shorthands.
// It recursively loads all .vuego files from the components folder and subfolders.
// File paths are mapped to kebab-case tag names using directory path and filename.
func WithComponents() LoadOption {
	return func(vue *Vue) {
		// Walk the components directory recursively
		err := fs.WalkDir(vue.templateFS, "components", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				// Skip directories that don't exist or can't be accessed
				return nil
			}

			// Skip directories, only process files
			if d.IsDir() {
				return nil
			}

			// Only process .vuego files
			if !strings.HasSuffix(d.Name(), ".vuego") {
				return nil
			}

			// Extract the relative path from components folder
			relPath := strings.TrimPrefix(path, "components/")
			relPath = strings.TrimSuffix(relPath, ".vuego")

			// Convert to kebab-case for the tag name
			// Split by path separators and convert each part
			parts := strings.Split(relPath, string(filepath.Separator))
			for i, part := range parts {
				parts[i] = helpers.CamelToKebab(part)
			}
			tagName := strings.Join(parts, "-")

			// Store the mapping
			vue.RegisterComponent(tagName, path)

			return nil
		})

		// Silently ignore if components directory doesn't exist
		_ = err
	}
}

// Fill sets all variables from the map, preserving any front-matter that was loaded.
func (t *template) Fill(vars any) Template {
	dataMap := toMapData(vars)
	// Merge loaded front-matter into data (front-matter takes precedence)
	for k, v := range t.frontMatter {
		dataMap[k] = v
	}
	t.stack = NewStackWithData(dataMap, vars)
	return t
}

// Assign sets a single variable.
func (t *template) Assign(key string, value any) Template {
	t.stack.Set(key, value)
	return t
}

// Err returns the internal error if any.
func (t *template) Err() error {
	return t.err
}

// SetErr sets the internal error. Used from Load().
func (t *template) SetErr(err error) {
	t.err = err
}

// Load will load a template file and front matter, extending VueContext with the loaded DOM.
// Front matter data is merged into the template's data and available via Get().
//
// Load returns a new allocation from a base template which may have data filled.
// The returned value is a throw-away after Render() is invoked.
func (t *template) Load(filename string) Template {
	tpl := t.new()

	// Load the template with front-matter and raw template bytes
	tpl.frontMatter, tpl.templateBytes, tpl.err = t.vue.loader.loadFragment(filename)
	tpl.filename = filename
	tpl.filenameLoaded = true

	for k, v := range tpl.frontMatter {
		tpl.Assign(k, v)
	}

	return tpl
}

// Get retrieves a variable value as a string.
func (t *template) Get(key string) string {
	val, ok := t.stack.Lookup(key)
	if !ok || val == nil {
		return ""
	}
	if v, ok := val.(string); ok {
		return v
	}
	if v, ok := val.(bool); ok {
		return fmt.Sprint(v)
	}
	result := fmt.Sprint(val)
	return result
}
