package vuego

import (
	"io"
	"io/fs"
	"sync"

	"golang.org/x/net/html"

	"github.com/titpetric/vuego/internal/helpers"
)

// Vue is the main template renderer for .vuego templates.
// After initialization, Vue is safe for concurrent use by multiple goroutines.
type Vue struct {
	templateFS fs.FS
	loader     *Loader
	renderer   Renderer
	funcMap    FuncMap
	exprEval   *ExprEvaluator

	// Template cache to avoid re-parsing the same template
	templateCache map[string][]*html.Node
	templateMu    sync.RWMutex

	// Custom node processors for post-processing rendered DOM
	nodeProcessors []NodeProcessor
}

// NewVue creates a new Vue backed by the given filesystem.
// The returned Vue is safe for concurrent use by multiple goroutines.
func NewVue(templateFS fs.FS) *Vue {
	v := &Vue{
		templateFS:    templateFS,
		loader:        NewLoader(templateFS),
		renderer:      NewRenderer(),
		exprEval:      NewExprEvaluator(),
		templateCache: make(map[string][]*html.Node),
	}
	v.funcMap = v.DefaultFuncMap()
	return v
}

// Funcs sets custom template functions. Returns the Vue instance for chaining.
func (v *Vue) Funcs(funcMap FuncMap) *Vue {
	v.funcMap = funcMap
	return v
}

// toMapData converts any value to map[string]any for use as template context.
// If data is already a map[string]any, it's returned as-is.
// Otherwise, returns an empty map (struct fields will be resolved via fallback in Stack.Lookup).
func toMapData(data any) map[string]any {
	if data == nil {
		return make(map[string]any)
	}
	if m, ok := data.(map[string]any); ok {
		return m
	}
	// Return empty map; struct fields will be accessible via Stack.rootData fallback
	return make(map[string]any)
}

// Render processes a full-page template file and writes the output to w.
// Render is safe to call concurrently from multiple goroutines.
func (v *Vue) Render(w io.Writer, filename string, data any) error {
	dom, err := v.loadCached(filename)
	if err != nil {
		return err
	}

	// Convert data to map[string]any and create context with original data for struct fallback
	dataMap := toMapData(data)
	ctx := NewVueContext(filename, &VueContextOptions{
		Data:         dataMap,
		OriginalData: data,
		Processors:   v.nodeProcessors,
	})

	if err := v.preProcessNodes(ctx, dom); err != nil {
		return err
	}

	result, err := v.evaluate(ctx, dom, 0)
	if err != nil {
		return err
	}

	// Apply custom node processors for post-processing
	if err := v.postProcessNodes(ctx, result); err != nil {
		return err
	}

	return v.render(w, result)
}

// loadCached returns a cached template or loads and caches it.
func (v *Vue) loadCached(filename string) ([]*html.Node, error) {
	v.templateMu.RLock()
	if cached, ok := v.templateCache[filename]; ok {
		v.templateMu.RUnlock()
		return cached, nil
	}
	v.templateMu.RUnlock()

	dom, err := v.loader.Load(filename)
	if err != nil {
		return nil, err
	}

	v.templateMu.Lock()
	v.templateCache[filename] = dom
	v.templateMu.Unlock()

	return dom, nil
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
// RenderFragment is safe to call concurrently from multiple goroutines.
func (v *Vue) RenderFragment(w io.Writer, filename string, data any) error {
	dom, err := v.loader.LoadFragment(filename)
	if err != nil {
		return err
	}

	// Convert data to map[string]any and create context with original data for struct fallback
	dataMap := toMapData(data)
	ctx := NewVueContext(filename, &VueContextOptions{
		Data:         dataMap,
		OriginalData: data,
		Processors:   v.nodeProcessors,
	})

	if err := v.preProcessNodes(ctx, dom); err != nil {
		return err
	}

	// Assign unique IDs to all v-once elements for tracking across deep clones
	for _, node := range dom {
		assignSeenAttrs(&ctx, node)
	}

	result, err := v.evaluate(ctx, dom, 0)
	if err != nil {
		return err
	}

	// Apply custom node processors for post-processing
	if err := v.postProcessNodes(ctx, result); err != nil {
		return err
	}

	return v.render(w, result)
}

func (v *Vue) render(w io.Writer, nodes []*html.Node) error {
	for _, node := range nodes {
		if err := renderNode(w, node, 0); err != nil {
			return err
		}
	}
	return nil
}
