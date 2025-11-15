package vuego

import (
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"
	"sync"

	"github.com/titpetric/vuego/internal/helpers"
	"golang.org/x/net/html"
)

// Vue is the main template renderer for .vuego templates.
// After initialization, Vue is safe for concurrent use by multiple goroutines.
type Vue struct {
	templateFS fs.FS
	loader     *Component
	funcMap    FuncMap
	exprEval   *ExprEvaluator

	// Template cache to avoid re-parsing the same template
	templateCache map[string][]*html.Node
	templateMu    sync.RWMutex
}

// VueContext carries template inclusion context and request-scoped state used during evaluation.
// Each render operation gets its own VueContext, making concurrent rendering safe.
type VueContext struct {
	// Variable scope and data resolution
	stack *Stack

	// Template inclusion chain context
	BaseDir       string
	CurrentDir    string
	FromFilename  string
	TemplateStack []string

	// HTML rendering state
	TagStack []string

	// v-once element tracking for deep clones
	seen         map[string]bool
	seenCounter int
}

// NewVueContext returns a VueContext initialized for the given template filename with initial data.
func NewVueContext(fromFilename string, data map[string]any) VueContext {
	return NewVueContextWithData(fromFilename, data, nil)
}

// NewVueContextWithData returns a VueContext with both map data and original root data for struct field fallback.
func NewVueContextWithData(fromFilename string, data map[string]any, originalData any) VueContext {
	return VueContext{
		stack:         NewStackWithData(data, originalData),
		CurrentDir:    path.Dir(fromFilename),
		FromFilename:  fromFilename,
		TemplateStack: []string{fromFilename},
		TagStack:      []string{},
		seen:          make(map[string]bool),
	}
}

// WithTemplate returns a copy of the context extended with filename in the inclusion chain.
func (ctx VueContext) WithTemplate(filename string) VueContext {
	newStack := make([]string, len(ctx.TemplateStack)+1)
	copy(newStack, ctx.TemplateStack)
	newStack[len(ctx.TemplateStack)] = filename
	return VueContext{
		stack:         ctx.stack, // Share the same stack across template chain
		BaseDir:       ctx.BaseDir,
		CurrentDir:    path.Dir(filename),
		FromFilename:  filename,
		TemplateStack: newStack,
		TagStack:      ctx.TagStack, // Share the same tag stack
		seen:          ctx.seen,     // Share the v-once tracking map
	}
}

// FormatTemplateChain returns the template inclusion chain formatted for error messages.
func (ctx VueContext) FormatTemplateChain() string {
	if len(ctx.TemplateStack) <= 1 {
		return ctx.FromFilename
	}
	return strings.Join(ctx.TemplateStack, " -> ")
}

// PushTag adds a tag to the tag stack.
func (ctx *VueContext) PushTag(tag string) {
	ctx.TagStack = append(ctx.TagStack, tag)
}

// PopTag removes the current tag from the tag stack.
func (ctx *VueContext) PopTag() {
	if len(ctx.TagStack) > 0 {
		ctx.TagStack = ctx.TagStack[:len(ctx.TagStack)-1]
	}
}

// CurrentTag returns the current parent tag, or empty string if no tag is on the stack.
func (ctx VueContext) CurrentTag() string {
	if len(ctx.TagStack) == 0 {
		return ""
	}
	return ctx.TagStack[len(ctx.TagStack)-1]
}

// nextSeenID returns a unique ID for tracking v-once elements across deep clones.
func (ctx *VueContext) nextSeenID() string {
	ctx.seenCounter++
	return fmt.Sprint(ctx.seenCounter)
}

// NewVue creates a new Vue backed by the given filesystem.
// The returned Vue is safe for concurrent use by multiple goroutines.
func NewVue(templateFS fs.FS) *Vue {
	return &Vue{
		templateFS:    templateFS,
		loader:        NewComponent(templateFS),
		funcMap:       DefaultFuncMap(),
		exprEval:      NewExprEvaluator(),
		templateCache: make(map[string][]*html.Node),
	}
}

// Funcs sets custom template functions. Returns the Vue instance for chaining.
func (v *Vue) Funcs(funcMap FuncMap) *Vue {
	if v.funcMap == nil {
		v.funcMap = make(FuncMap)
	}
	for name, fn := range funcMap {
		v.funcMap[name] = fn
	}
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
	ctx := NewVueContextWithData(filename, dataMap, data)

	result, err := v.evaluate(ctx, dom, 0)
	if err != nil {
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
	ctx := NewVueContextWithData(filename, dataMap, data)

	// Assign unique IDs to all v-once elements for tracking across deep clones
	for _, node := range dom {
		assignSeenAttrs(&ctx, node)
	}

	result, err := v.evaluate(ctx, dom, 0)
	if err != nil {
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

var indentCache = [256]string{}

func init() {
	for i := 0; i < len(indentCache); i++ {
		indentCache[i] = strings.Repeat(" ", i)
	}
}

func getIndent(indent int) string {
	if indent < len(indentCache) {
		return indentCache[indent]
	}
	return strings.Repeat(" ", indent)
}

// shouldEscapeTextNode checks if a text node needs HTML escaping.
// Returns false if the text appears to be already HTML-escaped (from interpolation),
// true if it contains raw HTML special characters that need escaping.
func shouldEscapeTextNode(data string) bool {
	// If the text contains HTML entity references like &lt; &amp; &#39; etc,
	// it's likely from interpolation and already escaped
	if strings.Contains(data, "&") && strings.Contains(data, ";") {
		return false
	}
	// Check if text contains unescaped HTML special characters
	return strings.ContainsAny(data, "<>&\"'")
}

func renderNode(w io.Writer, node *html.Node, indent int) error {
	ctx := VueContext{}
	return renderNodeWithContext(w, node, indent, &ctx)
}

func renderNodeWithContext(w io.Writer, node *html.Node, indent int, ctx *VueContext) error {
	switch node.Type {
	case html.TextNode:
		if strings.TrimSpace(node.Data) == "" {
			return nil
		}
		spaces := getIndent(indent)
		parentTag := ctx.CurrentTag()
		// Skip HTML escaping inside script and style tags
		if parentTag == "script" || parentTag == "style" {
			_, _ = w.Write([]byte(spaces + node.Data))
		} else if shouldEscapeTextNode(node.Data) {
			_, _ = w.Write([]byte(spaces + html.EscapeString(node.Data)))
		} else {
			_, _ = w.Write([]byte(spaces + node.Data))
		}

	case html.ElementNode:
		// Count children without allocating slice
		childCount := 0
		firstChild := node.FirstChild
		for c := firstChild; c != nil; c = c.NextSibling {
			childCount++
		}

		spaces := getIndent(indent)
		tagName := node.Data
		// compact single-entry text nodes
		if childCount == 0 {
			_, _ = w.Write([]byte(spaces + "<" + tagName + renderAttrs(node.Attr) + "></" + tagName + ">\n"))
		} else if childCount == 1 && firstChild.Type == html.TextNode {
			_, _ = w.Write([]byte(spaces + "<" + tagName + renderAttrs(node.Attr) + ">"))
			// Skip HTML escaping inside script and style tags
			if tagName == "script" || tagName == "style" {
				_, _ = w.Write([]byte(firstChild.Data))
			} else if shouldEscapeTextNode(firstChild.Data) {
				_, _ = w.Write([]byte(html.EscapeString(firstChild.Data)))
			} else {
				_, _ = w.Write([]byte(firstChild.Data))
			}
			_, _ = w.Write([]byte("</" + tagName + ">\n"))
		} else {
			_, _ = w.Write([]byte(spaces + "<" + tagName + renderAttrs(node.Attr) + ">\n"))
			ctx.PushTag(tagName)
			for c := firstChild; c != nil; c = c.NextSibling {
				if err := renderNodeWithContext(w, c, indent+2, ctx); err != nil {
					return err
				}
			}
			ctx.PopTag()
			_, _ = w.Write([]byte(spaces + "</" + tagName + ">\n"))
		}
	}

	return nil
}

// shouldIgnoreAttr returns true if an attribute should be skipped in HTML output.
// This includes Vue directives and binding attributes that are only used during evaluation.
func shouldIgnoreAttr(key string) bool {
	// Vue directives that should not appear in final HTML
	switch key {
	case "v-if", "v-else-if", "v-else", "v-for", "v-pre", "v-html", "v-show", "v-once", "v-once-id":
		return true
	}
	// v-bind: and : prefixed attributes are processed and shouldn't appear in output
	if strings.HasPrefix(key, "v-bind:") || strings.HasPrefix(key, ":") {
		return true
	}
	return false
}

func renderAttrs(attrs []html.Attribute) string {
	if len(attrs) == 0 {
		return ""
	}
	var sb strings.Builder
	for _, a := range attrs {
		// Skip Vue directives and internal binding attributes
		if shouldIgnoreAttr(a.Key) {
			continue
		}
		sb.WriteByte(' ')
		sb.WriteString(a.Key)
		sb.WriteByte('=')
		sb.WriteByte('"')
		sb.WriteString(a.Val)
		sb.WriteByte('"')
	}
	return sb.String()
}
