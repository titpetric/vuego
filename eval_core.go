package vuego

import (
	"fmt"

	"github.com/titpetric/vuego/internal/helpers"
	"golang.org/x/net/html"
)

func (v *Vue) evaluate(ctx VueContext, nodes []*html.Node, depth int) ([]*html.Node, error) {
	var result []*html.Node

	for _, node := range nodes {

		switch node.Type {
		case html.TextNode:
			interpolated, _ := v.interpolate(node.Data)
			newNode := helpers.CloneNode(node)
			newNode.Data = interpolated
			result = append(result, newNode)
			continue

		case html.ElementNode:
			tag := node.Data

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
				v.stack.Push(componentData)

				// Validate and process template tag
				processedDom, err := v.evalTemplate(compDom, componentData)
				if err != nil {
					v.stack.Pop()
					return nil, fmt.Errorf("error in %s (included from %s): %w", name, ctx.FormatTemplateChain(), err)
				}

				childCtx := ctx.WithTemplate(name)
				evaluated, err := v.evaluate(childCtx, processedDom, depth+1)
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

			if vIf := helpers.GetAttr(node, "v-if"); vIf != "" {
				ok, err := v.evalCondition(node, vIf)
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
					v.evalVHtml(n)
					v.evalAttributes(n)
				}

				result = append(result, loopNodes...)
				continue
			}

			newNode := helpers.DeepCloneNode(node)

			hasVHtml := helpers.GetAttr(newNode, "v-html") != ""
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
			result = append(result, helpers.CloneNode(node))
			}
	}

	return result, nil
}
