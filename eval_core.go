package vuego

import (
	"fmt"

	"golang.org/x/net/html"

	"github.com/titpetric/vuego/internal/helpers"
	"github.com/titpetric/vuego/internal/parser"
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

	for i := 0; i < len(nodes); i++ {
		node := nodes[i]

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

			// Check for v-once early - skip if already rendered
			if helpers.HasAttr(node, "v-once") {
				vSeenID := helpers.GetAttr(node, "v-once-id")
				if ctx.seen[vSeenID] {
					// This v-once element has already been rendered, skip it
					continue
				}
				// Mark this v-once element as rendered
				ctx.seen[vSeenID] = true
			}

			// Check for v-pre early - prevents all interpolation and directive processing
			if helpers.HasAttr(node, "v-pre") {
				newNode := helpers.ShallowCloneWithAttrs(node)
				// Deep clone children without processing them (v-pre will be filtered during rendering)
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
				vars, err := v.evalAttributes(ctx, node)
				if err != nil {
					return nil, err
				}

				delete(vars, "include")
				ctx.stack.Push(vars)

				name := helpers.GetAttr(node, "include")
				frontMatter, templateBytes, err := v.loader.loadFragment(name)
				if err != nil {
					return nil, fmt.Errorf("error loading %s (included from %s): %w", name, ctx.FormatTemplateChain(), err)
				}

				// Merge front-matter data (authoritative - overrides passed data)
				for k, v := range frontMatter {
					ctx.stack.Set(k, v)
				}

				// Parse template bytes into DOM
				compDom, err := parser.ParseTemplateBytes(templateBytes)
				if err != nil {
					return nil, fmt.Errorf("error parsing %s (included from %s): %w", name, ctx.FormatTemplateChain(), err)
				}

				// Validate and process template tag
				processedDom, err := v.evalTemplate(ctx, compDom, ctx.stack.EnvMap(), depth+1)
				if err != nil {
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

			if helpers.HasAttr(node, "v-for") {
				chainResult, skipCount, err := v.evalVFor(ctx, node, nodes[i:], depth)
				if err != nil {
					return nil, err
				}
				result = append(result, chainResult...)
				i += skipCount
				continue
			}

			if tag == "template" {
				evaluated, err := v.evalTemplate(ctx, []*html.Node{node}, ctx.stack.EnvMap(), depth+1)
				if err != nil {
					return nil, err
				}
				result = append(result, evaluated...)
				continue
			}

			// Handle v-if chains (v-if, v-else-if, v-else)
			if helpers.HasAttr(node, "v-if") {
				chainResult, skipCount, err := v.evalElseIfChain(ctx, node, nodes[i:], depth)
				if err != nil {
					return nil, err
				}
				result = append(result, chainResult...)
				// Skip past the v-else-if and v-else nodes that were part of this chain
				i += skipCount
				continue
			}

			// Skip v-else-if and v-else if they appear without v-if
			// (they should be handled as part of a chain)
			if helpers.HasAttr(node, "v-else-if") || helpers.HasAttr(node, "v-else") {
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

			if err := v.evalVHtml(ctx, newNode); err != nil {
				return nil, err
			}
			if err := v.evalVShow(ctx, newNode); err != nil {
				return nil, err
			}
			if _, err := v.evalAttributes(ctx, newNode); err != nil {
				return nil, err
			}

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
