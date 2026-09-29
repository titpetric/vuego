package vuego

import (
	"context"
	"io"
	"io/fs"
	"sync"
	"time"

	"golang.org/x/net/html"

	"github.com/titpetric/vuego/internal/helpers"
	"github.com/titpetric/vuego/internal/parser"
	ireflect "github.com/titpetric/vuego/internal/reflect"
)

// templateCacheEntry stores cached template data with modification tracking.
type templateCacheEntry struct {
	dom         []*html.Node
	frontMatter map[string]any
	modTime     time.Time
}

// Vue is the main template renderer for .vuego templates.
// After initialization, Vue is safe for concurrent use by multiple goroutines.
type Vue struct {
	templateFS fs.FS
	loader     *Loader
	renderer   Renderer
	funcMap    FuncMap
	exprEval   *ExprEvaluator

	// Template cache to avoid re-parsing the same template
	templateCache map[string]*templateCacheEntry
	templateMu    sync.RWMutex

	// Custom node processors for post-processing rendered DOM
	nodeProcessors []NodeProcessor

	// Component shorthand mapping: maps kebab-case tag names to component filenames
	// e.g., "button-primary" -> "components/ButtonPrimary.vuego"
	componentMap map[string]string

	// initialData is pre-loaded data from WithData() that seeds the template stack.
	// This is equivalent to calling Fill() on the base template before any New()/Load().
	initialData map[string]any
}

// NewVue creates a new Vue backed by the given filesystem.
// The returned Vue is safe for concurrent use by multiple goroutines.
func NewVue(templateFS fs.FS) *Vue {
	v := &Vue{
		templateFS:    templateFS,
		loader:        NewLoader(templateFS),
		renderer:      NewRenderer(),
		exprEval:      NewExprEvaluator(),
		templateCache: make(map[string]*templateCacheEntry),
		componentMap:  make(map[string]string),
	}
	v.funcMap = v.DefaultFuncMap()
	return v
}

// Funcs merges custom template functions into the existing funcmap, overwriting any existing keys.
// Returns the Vue instance for chaining.
func (v *Vue) Funcs(funcMap FuncMap) *Vue {
	for k, fn := range funcMap {
		v.funcMap[k] = fn
	}
	return v
}

// DefaultFuncMap returns the built-in template functions. The keys are
// documented in docs/funcmap.md.
func (v *Vue) DefaultFuncMap() FuncMap {
	return FuncMap{
		"type":       typeFunc,
		"upper":      upperFunc,
		"lower":      lowerFunc,
		"title":      titleFunc,
		"formatTime": formatTimeFunc,
		"formatDate": formatDateFunc,
		"default":    defaultFunc,
		"len":        lenFunc,
		"trim":       trimFunc,
		"escape":     escapeFunc,
		"int":        intFunc,
		"string":     stringFunc,
		"json":       jsonFunc,
		"jsonPretty": jsonPrettyFunc,
		"file":       fileFunc(v),
		"jsonFile":   jsonFileFunc(v),
		"yamlFile":   yamlFileFunc(v),
	}
}

// RegisterNodeProcessor adds a custom node processor to the Vue instance.
// Processors are applied in order after template evaluation completes.
func (v *Vue) RegisterNodeProcessor(processor NodeProcessor) *Vue {
	if v.nodeProcessors == nil {
		v.nodeProcessors = make([]NodeProcessor, 0)
	}
	v.nodeProcessors = append(v.nodeProcessors, processor)
	return v
}

// RegisterComponent registers a component shorthand mapping.
// The tagName (e.g., "button-primary") will be resolved to the given filename (e.g., "components/ButtonPrimary.vuego").
func (v *Vue) RegisterComponent(tagName, filename string) *Vue {
	v.componentMap[tagName] = filename
	return v
}

// GetComponentFile looks up a component file by its kebab-case tag name.
// Returns the filename and true if found, otherwise returns empty string and false.
func (v *Vue) GetComponentFile(tagName string) (string, bool) {
	filename, ok := v.componentMap[tagName]
	return filename, ok
}

// renderNodesWithContext is an internal method that evaluates and renders nodes with a pre-configured context.
func (v *Vue) renderNodesWithContext(ctx VueContext, w io.Writer, nodes []*html.Node) error {
	nodeCopy := make([]*html.Node, 0, len(nodes))
	for i := 0; i < len(nodes); i++ {
		nodeCopy = append(nodeCopy, helpers.DeepCloneNode(nodes[i]))
	}

	if err := v.preProcessNodes(ctx, nodeCopy); err != nil {
		return err
	}

	result, err := v.evaluate(ctx, nodeCopy, 0)
	if err != nil {
		return err
	}

	if err := v.postProcessNodes(ctx, result); err != nil {
		return err
	}

	return v.render(w, result)
}

// toMapData converts any value to map[string]any for use as template context.
// If data is already a map[string]any, it's returned as-is.
// If data is a struct, it's converted to a map using JSON tags.
// Otherwise, returns an empty map (which will still allow field access via Stack.rootData fallback).
func toMapData(data any) map[string]any {
	if data == nil {
		return make(map[string]any)
	}
	if m, ok := data.(map[string]any); ok {
		return m
	}
	// Try to convert struct to map using JSON tags
	if m := ireflect.StructToMap(data); len(m) > 0 {
		return m
	}
	// Return empty map; fields will be accessible via Stack.rootData fallback
	return make(map[string]any)
}

// Render processes a full-page template file and writes the output to w.
// Front-matter data in the template is authoritative and overrides passed data.
// Render is safe to call concurrently from multiple goroutines.
func (v *Vue) Render(ctx context.Context, w io.Writer, filename string, data any) error {
	frontMatter, dom, err := v.loadCachedWithFrontMatter(filename)
	if err != nil {
		return err
	}

	// Merge front-matter data into the provided data (front-matter is authoritative)
	dataMap := toMapData(data)
	for k, v := range frontMatter {
		dataMap[k] = v
	}

	// Create context for v-once attribute tracking
	vueCtx := NewVueContext(ctx, filename, &VueContextOptions{
		Stack:      NewStackWithData(dataMap, data),
		Processors: v.nodeProcessors,
	})

	// Assign unique IDs to all v-once elements for tracking across deep clones
	for _, node := range dom {
		assignSeenAttrs(&vueCtx, node)
	}

	// Use renderNodesWithContext with pre-configured context
	return v.renderNodesWithContext(vueCtx, w, dom)
}

// loadCachedWithFrontMatter returns cached template nodes and front-matter data, or loads and caches them.
// The cache stores DOM nodes, front-matter, and file modification time for invalidation.
func (v *Vue) loadCachedWithFrontMatter(filename string) (map[string]any, []*html.Node, error) {
	// Get current file modification time
	var currentModTime time.Time
	if v.templateFS != nil {
		if info, err := fs.Stat(v.templateFS, filename); err == nil {
			currentModTime = info.ModTime()
		}
	}

	v.templateMu.RLock()
	cached, ok := v.templateCache[filename]
	if ok && (currentModTime.IsZero() || cached.modTime.Equal(currentModTime)) {
		// Cache hit and file hasn't changed (or we can't check mtime)
		v.templateMu.RUnlock()
		return cached.frontMatter, cached.dom, nil
	}
	v.templateMu.RUnlock()

	// Cache miss or file changed - reload
	frontMatter, templateBytes, err := v.loader.loadFragment(filename)
	if err != nil {
		return nil, nil, err
	}

	dom, err := parser.ParseTemplateBytes(templateBytes)
	if err != nil {
		return nil, nil, err
	}

	v.templateMu.Lock()
	v.templateCache[filename] = &templateCacheEntry{
		dom:         dom,
		frontMatter: frontMatter,
		modTime:     currentModTime,
	}
	v.templateMu.Unlock()

	return frontMatter, dom, nil
}

// assignSeenAttrs recursively assigns unique IDs to all v-once elements in the tree
func assignSeenAttrs(ctx *VueContext, node *html.Node) {
	if node.Type == html.ElementNode {
		if helpers.HasAttr(node, "v-once") {
			id := ctx.nextSeenID()
			helpers.SetAttr(node, "v-once-id", id)
		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		assignSeenAttrs(ctx, c)
	}
}

// RenderFragment processes a template fragment file and writes the output to w.
// Front-matter data in the template is authoritative and overrides passed data.
// RenderFragment is safe to call concurrently from multiple goroutines.
func (v *Vue) RenderFragment(ctx context.Context, w io.Writer, filename string, data any) error {
	frontMatter, templateBytes, err := v.loader.loadFragment(filename)
	if err != nil {
		return err
	}

	dom, err := parser.ParseTemplateBytes(templateBytes)
	if err != nil {
		return err
	}

	// Merge front-matter data into the provided data (front-matter is authoritative)
	dataMap := toMapData(data)
	for k, v := range frontMatter {
		dataMap[k] = v
	}

	// Create context for v-once attribute tracking
	vueCtx := NewVueContext(ctx, filename, &VueContextOptions{
		Stack:      NewStackWithData(dataMap, data),
		Processors: v.nodeProcessors,
	})

	// Assign unique IDs to all v-once elements for tracking across deep clones
	for _, node := range dom {
		assignSeenAttrs(&vueCtx, node)
	}

	// Use RenderNodes with pre-configured context
	return v.renderNodesWithContext(vueCtx, w, dom)
}

func (v *Vue) render(w io.Writer, nodes []*html.Node) error {
	for _, node := range nodes {
		if err := renderNode(w, node, 0); err != nil {
			return err
		}
	}
	return nil
}
