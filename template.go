package vuego

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// New creates a Template that renders strings, bytes or readers without a
// filesystem. Call NewFS(fs) or New(WithFS(fs)) to render files instead.
// The returned Template holds assigned variables and renders them.
func New(opts ...LoadOption) Template {
	return NewFS(nil, opts...)
}

// View binds a template file to a data model type.
// Render the returned template once, then discard it.
func View[V any](renderer Template, filename string, data V) Template {
	return renderer.Load(filename).Fill(data)
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

// NewFS creates a Template over the given filesystem, applying each option in
// order. The returned Template renders files, strings or bytes and holds
// assigned variables.
func NewFS(templateFS fs.FS, opts ...LoadOption) Template {
	vue := NewVue(templateFS)

	// Apply functional options
	for _, opt := range opts {
		opt(vue)
	}

	// Implicitly load config (theme.yml + data/*.yml) from the filesystem
	loadConfig(vue)

	tpl := &template{
		vue:   vue,
		stack: NewStack(nil),
	}

	// Apply initial data if available (equivalent to Fill before New/Load)
	if vue.initialData != nil {
		tpl.Fill(vue.initialData)
	}

	return tpl
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

// loadConfig loads YAML config files from the filesystem into vue.initialData.
// It loads root-level theme.yml, then all YAML files from the data/ directory.
// Files in data/ override root-level values with the same keys.
func loadConfig(vue *Vue) {
	if vue.templateFS == nil {
		return
	}

	if vue.initialData == nil {
		vue.initialData = make(map[string]any)
	}

	// Helper to load and merge a YAML file
	loadYAML := func(path string) {
		content, err := fs.ReadFile(vue.templateFS, path)
		if err != nil {
			return // File doesn't exist or can't be read, skip silently
		}

		var data map[string]any
		if err := yaml.Unmarshal(content, &data); err != nil {
			return // Invalid YAML, skip silently
		}

		// Merge into initialData
		for k, v := range data {
			vue.initialData[k] = v
		}
	}

	// Load root-level theme.yml first
	loadYAML("theme.yml")

	// Load all YAML files from data/ directory (overrides root-level values)
	entries, err := fs.ReadDir(vue.templateFS, "data")
	if err != nil {
		return // data/ directory doesn't exist, skip silently
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml") {
			loadYAML("data/" + name)
		}
	}
}

// Fill sets all variables from the map, preserving any front-matter that was loaded.
func (t *template) Fill(vars any) Template {
	// Start with auto-loaded config (lowest precedence)
	dataMap := map[string]any{}
	if t.vue.initialData != nil {
		for k, v := range t.vue.initialData {
			dataMap[k] = v
		}
	}
	// Merge passed data (overrides config)
	passedData := toMapData(vars)
	for k, v := range passedData {
		dataMap[k] = v
	}
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
