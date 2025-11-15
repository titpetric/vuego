package vuego

import (
	"fmt"

	"github.com/titpetric/vuego/internal/helpers"
	"golang.org/x/net/html"
)

// evaluateChildren evaluates the children of a node without allocating a temporary slice.
func (v *Vue) evaluateChildren(ctx VueContext, node *html.Node, depth int) ([]*html.Node, error) {
	childList := make([]*html.Node, 0, helpers.CountChildren(node))
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		childList = append(childList, c)
	}
	return v.evaluate(ctx, childList, depth)
}

func (v *Vue) evaluate(ctx VueContext, nodes []*html.Node, depth int) ([]*html.Node, error) {
	var result []*html.Node

	for _, node := range nodes {

		switch node.Type {
		case html.TextNode:
			interpolated, err := v.interpolate(ctx, node.Data)
			if err != nil {
				return nil, fmt.Errorf("in %s: %w", ctx.FormatTemplateChain(), err)
			}
			newNode := helpers.CloneNode(node)
			newNode.Data = interpolated
			result = append(result, newNode)
			continue

		case html.ElementNode:
			tag := node.Data

			// Check for v-pre early - prevents all interpolation and directive processing
			if helpers.HasAttr(node, "v-pre") {
				newNode := helpers.ShallowCloneWithAttrs(node)
				// Remove the v-pre attribute from output before processing
				helpers.RemoveAttr(newNode, "v-pre")
				// Deep clone children without processing them
				var prev *html.Node
				for c := node.FirstChild; c != nil; c = c.NextSibling {
					cloned := helpers.DeepCloneNode(c)
					if prev == nil {
						newNode.FirstChild = cloned
					} else {
						prev.NextSibling = cloned
						cloned.PrevSibling = prev
					}
					prev = cloned
				}
				result = append(result, newNode)
				continue
			}

			if tag == "vuego" {
				name := helpers.GetAttr(node, "include")
				compDom, err := v.loader.LoadFragment(name)
				if err != nil {
					return nil, fmt.Errorf("error loading %s (included from %s): %w", name, ctx.FormatTemplateChain(), err)
				}

				// Push component attributes to stack for :required validation
				componentData := make(map[string]any)
				for _, attr := range node.Attr {
					if attr.Key != "include" {
						componentData[attr.Key] = attr.Val
					}
				}
				ctx.stack.Push(componentData)

				// Validate and process template tag
				processedDom, err := v.evalTemplate(ctx, compDom, componentData)
				if err != nil {
					ctx.stack.Pop()
					return nil, fmt.Errorf("error in %s (included from %s): %w", name, ctx.FormatTemplateChain(), err)
				}

				childCtx := ctx.WithTemplate(name)
				evaluated, err := v.evaluate(childCtx, processedDom, depth+1)
				ctx.stack.Pop()
				if err != nil {
					return nil, err
				}
				result = append(result, evaluated...)
				continue
			}

			if tag == "template" {
				// Omit template tags at the top level
				evaluated, err := v.evaluateChildren(ctx, node, depth+1)
				if err != nil {
					return nil, err
				}
				result = append(result, evaluated...)
				continue
			}

			if vIf := helpers.GetAttr(node, "v-if"); vIf != "" {
				ok, err := v.evalCondition(ctx, node, vIf)
				if err != nil || !ok {
					continue
				}
			}

			if vFor := helpers.GetAttr(node, "v-for"); vFor != "" {
				loopNodes, err := v.evalFor(ctx, node, vFor, depth+1)
				if err != nil {
					return nil, err
				}

				for _, n := range loopNodes {
					v.evalVHtml(ctx, n)
					v.evalAttributes(ctx, n)
				}

				result = append(result, loopNodes...)
				continue
			}

			// Check for v-html early to decide cloning strategy
			hasVHtml := helpers.GetAttr(node, "v-html") != ""
			var newNode *html.Node
			if hasVHtml {
				// Deep clone to preserve original children structure for v-html
				newNode = helpers.DeepCloneNode(node)
			} else {
				// Shallow clone with attributes - children will be re-evaluated and replaced
				newNode = helpers.ShallowCloneWithAttrs(node)
			}

			v.evalVHtml(ctx, newNode)
			v.evalAttributes(ctx, newNode)

			if !hasVHtml {
				ctx.PushTag(node.Data)
				newChildren, err := v.evaluateChildren(ctx, node, depth+1)
				ctx.PopTag()
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
			result = append(result, helpers.CloneNode(node))
		}
	}

	return result, nil
}
