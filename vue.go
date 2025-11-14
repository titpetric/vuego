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
type Vue struct {
	loader *Component
	stack  *Stack
}

// VueContext carries template inclusion context used during evaluation and for error reporting.
type VueContext struct {
	BaseDir       string
	CurrentDir    string
	FromFilename  string
	TemplateStack []string
}

// NewVueContext returns a VueContext initialized for the given template filename.
func NewVueContext(fromFilename string) VueContext {
	return VueContext{
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
func NewVue(templateFS fs.FS) *Vue {
	return &Vue{
		loader: NewComponent(templateFS),
		stack:  NewStack(nil),
	}
}

// Render processes a full-page template file and writes the output to w.
func (v *Vue) Render(w io.Writer, filename string, data map[string]any) error {
	dom, err := v.loader.Load(filename)
	if err != nil {
		return err
	}

	v.stack.Push(data)
	defer v.stack.Pop()

	ctx := NewVueContext(filename)

	result, err := v.evaluate(ctx, dom, 0)
	if err != nil {
		return err
	}

	return v.render(w, result)
}

// RenderFragment processes a template fragment file and writes the output to w.
func (v *Vue) RenderFragment(w io.Writer, filename string, data map[string]any) error {
	dom, err := v.loader.LoadFragment(filename)
	if err != nil {
		return err
	}

	v.stack.Push(data)
	defer v.stack.Pop()

	ctx := NewVueContext(filename)

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
