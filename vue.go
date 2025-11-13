package vuego

import (
	"fmt"
	"io"
	"io/fs"
	"path"
	"strings"

	"golang.org/x/net/html"
)

type Vue struct {
	loader *Component
	stack  *Stack
}

type VueContext struct {
	BaseDir      string
	CurrentDir   string
	FromFilename string
}

func NewVueContext(fromFilename string) VueContext {
	return VueContext{
		CurrentDir:   path.Dir(fromFilename),
		FromFilename: fromFilename,
	}
}

func NewVue(templateFS fs.FS) *Vue {
	return &Vue{
		loader: NewComponent(templateFS),
		stack:  NewStack(nil),
	}
}

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

func (v *Vue) evaluate(ctx VueContext, nodes []*html.Node, depth int) ([]*html.Node, error) {
	var result []*html.Node

	for _, node := range nodes {

		switch node.Type {
		case html.TextNode:
			interpolated, _ := v.interpolate(node.Data)
			newNode := cloneNode(node)
			newNode.Data = interpolated
			result = append(result, newNode)
			continue

		case html.ElementNode:
			tag := node.Data

			if tag == "vuego" {
				name := getAttr(node, "include")
				compDom, err := v.loader.LoadFragment(name)
				if err != nil {
					return nil, fmt.Errorf("Error loading fragment from %s: %w", name, err)
				}

				// Push component attributes to stack for :required validation
				componentData := make(map[string]any)
				for _, attr := range node.Attr {
					if attr.Key != "include" {
						componentData[attr.Key] = attr.Val
					}
				}
				v.stack.Push(componentData)

				// Validate and process template tag
				processedDom, err := v.evalTemplate(compDom, componentData)
				if err != nil {
					v.stack.Pop()
					return nil, fmt.Errorf("Error processing template in %s: %w", name, err)
				}

				evaluated, err := v.evaluate(ctx, processedDom, depth+1)
				v.stack.Pop()
				if err != nil {
					return nil, err
				}
				result = append(result, evaluated...)
				continue
			}

			if tag == "template" {
				// Omit template tags at the top level
				var childList []*html.Node
				for c := range node.ChildNodes() {
					childList = append(childList, c)
				}
				evaluated, err := v.evaluate(ctx, childList, depth+1)
				if err != nil {
					return nil, err
				}
				result = append(result, evaluated...)
				continue
			}

			if vIf := getAttr(node, "v-if"); vIf != "" {
				ok, err := v.evalCondition(node, vIf)
				if err != nil || !ok {
					continue
				}
			}

			if vFor := getAttr(node, "v-for"); vFor != "" {
				loopNodes, err := v.evalFor(ctx, node, vFor, depth+1)
				if err != nil {
					return nil, err
				}

				for _, n := range loopNodes {
					v.evalVHtml(n)
					v.evalAttributes(n)
				}

				result = append(result, loopNodes...)
				continue
			}

			newNode := deepCloneNode(node)

			hasVHtml := getAttr(newNode, "v-html") != ""
			v.evalVHtml(newNode)
			v.evalAttributes(newNode)

			if !hasVHtml {
				var childList []*html.Node
				for c := range node.ChildNodes() {
					childList = append(childList, c)
				}

				newChildren, err := v.evaluate(ctx, childList, depth+1)
				if err != nil {
					return nil, err
				}

				newNode.FirstChild = nil
				for i, c := range newChildren {
					if i == 0 {
						newNode.FirstChild = c
					} else {
						newChildren[i-1].NextSibling = c
					}
				}
			}

			result = append(result, newNode)
		default:
			result = append(result, cloneNode(node))
		}
	}

	return result, nil
}

// Helper to get attribute
func getAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func removeAttr(n *html.Node, key string) {
	var attrs []html.Attribute
	for _, a := range n.Attr {
		if a.Key == key {
			continue
		}
		attrs = append(attrs, a)
	}
	n.Attr = attrs
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
