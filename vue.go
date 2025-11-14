package vuego

import (
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"

	"golang.org/x/net/html"
)

// Vue is the main template renderer for .vuego templates.
// After initialization, Vue is safe for concurrent use by multiple goroutines.
type Vue struct {
	templateFS fs.FS
	loader     *Component
	funcMap    FuncMap
	exprEval   *ExprEvaluator
}

// VueContext carries template inclusion context and request-scoped state used during evaluation.
// Each render operation gets its own VueContext, making concurrent rendering safe.
type VueContext struct {
	stack         *Stack
	BaseDir       string
	CurrentDir    string
	FromFilename  string
	TemplateStack []string
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
	}
}

// FormatTemplateChain returns the template inclusion chain formatted for error messages.
func (ctx VueContext) FormatTemplateChain() string {
	if len(ctx.TemplateStack) <= 1 {
		return ctx.FromFilename
	}
	return strings.Join(ctx.TemplateStack, " -> ")
}

// NewVue creates a new Vue backed by the given filesystem.
// The returned Vue is safe for concurrent use by multiple goroutines.
func NewVue(templateFS fs.FS) *Vue {
	return &Vue{
		templateFS: templateFS,
		loader:     NewComponent(templateFS),
		funcMap:    DefaultFuncMap(),
		exprEval:   NewExprEvaluator(),
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
	dom, err := v.loader.Load(filename)
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

func renderNode(w io.Writer, node *html.Node, indent int) error {
	spaces := strings.Repeat(" ", indent)

	switch node.Type {
	case html.TextNode:
		if strings.TrimSpace(node.Data) == "" {
			return nil
		}
		_, _ = w.Write([]byte(spaces + node.Data))

	case html.ElementNode:
		var children []*html.Node
		for child := range node.ChildNodes() {
			children = append(children, child)
		}

		// compact single-entry text nodes
		if len(children) == 0 {
			_, _ = w.Write([]byte(spaces + "<" + node.Data + renderAttrs(node.Attr) + "></" + node.Data + ">\n"))
		} else if len(children) == 1 && children[0].Type == html.TextNode {
			_, _ = w.Write([]byte(spaces + "<" + node.Data + renderAttrs(node.Attr) + ">"))
			_, _ = w.Write([]byte(children[0].Data))
			_, _ = w.Write([]byte("</" + node.Data + ">\n"))
		} else {
			_, _ = w.Write([]byte(spaces + "<" + node.Data + renderAttrs(node.Attr) + ">\n"))
			for _, child := range children {
				if err := renderNode(w, child, indent+2); err != nil {
					return err
				}
			}
			_, _ = w.Write([]byte(spaces + "</" + node.Data + ">\n"))
		}
	}

	return nil
}

func renderAttrs(attrs []html.Attribute) string {
	var sb strings.Builder
	for _, a := range attrs {
		sb.WriteString(fmt.Sprintf(` %s="%s"`, a.Key, a.Val))
	}
	return sb.String()
}
