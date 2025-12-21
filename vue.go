package vuego

import (
	"io"
	"io/fs"
	"sync"

	"golang.org/x/net/html"

	"github.com/titpetric/vuego/internal/helpers"
	"github.com/titpetric/vuego/internal/parser"
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

// Funcs merges custom template functions into the existing funcmap, overwriting any existing keys.
// Returns the Vue instance for chaining.
func (v *Vue) Funcs(funcMap FuncMap) *Vue {
	for k, fn := range funcMap {
		v.funcMap[k] = fn
	}
	return v
}

// RenderNodes evaluates and renders HTML nodes with the given data.
// This is the core rendering function used by all public render methods.
func (v *Vue) RenderNodes(w io.Writer, nodes []*html.Node, data any) error {
	dataMap := toMapData(data)

	ctx := NewVueContext("", &VueContextOptions{
		Stack:      NewStackWithData(dataMap, data),
		Processors: v.nodeProcessors,
	})

	return v.renderNodesWithContext(ctx, w, nodes)
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
// Front-matter data in the template is authoritative and overrides passed data.
// Render is safe to call concurrently from multiple goroutines.
func (v *Vue) Render(w io.Writer, filename string, data any) error {
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
	vueCtx := NewVueContext(filename, &VueContextOptions{
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
// The cache stores only the nodes without front-matter data (front-matter is re-extracted on each load).
func (v *Vue) loadCachedWithFrontMatter(filename string) (map[string]any, []*html.Node, error) {
	v.templateMu.RLock()
	if cached, ok := v.templateCache[filename]; ok {
		v.templateMu.RUnlock()
		// Load again to extract front-matter from the file
		frontMatter, _, err := v.loader.loadFragment(filename)
		if err != nil {
			return nil, nil, err
		}
		return frontMatter, cached, nil
	}
	v.templateMu.RUnlock()

	frontMatter, templateBytes, err := v.loader.loadFragment(filename)
	if err != nil {
		return nil, nil, err
	}

	dom, err := parser.ParseTemplateBytes(templateBytes)
	if err != nil {
		return nil, nil, err
	}

	v.templateMu.Lock()
	v.templateCache[filename] = dom
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
func (v *Vue) RenderFragment(w io.Writer, filename string, data any) error {
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
	vueCtx := NewVueContext(filename, &VueContextOptions{
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
