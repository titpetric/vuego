package vuego

import (
	"strings"

	"golang.org/x/net/html"

	"github.com/titpetric/vuego/internal/helpers"
)

// evalSlot processes a <slot> element and inserts the appropriate content.
// If slot content was provided by the component user, use that.
// Otherwise, render the fallback content (children of the slot element).
func (v *Vue) evalSlot(ctx VueContext, node *html.Node, slotScope *SlotScope) ([]*html.Node, error) {
	slotName := helpers.GetAttr(node, "name")
	if slotName == "" {
		slotName = "default"
	}

	// Get slot props that the slot binds
	slotProps := make(map[string]any)
	for _, attr := range node.Attr {
		if attr.Key == "" || attr.Key[0] != ':' {
			continue
		}
		// Extract the binding name (without the colon)
		propName := attr.Key[1:]

		// Evaluate the binding value
		val, err := v.exprEval.Eval(attr.Val, ctx.ExprEnv())
		if err == nil && val != nil {
			slotProps[propName] = val
		}
	}

	// Try to get provided slot content (from explicit include or inherited from layout)
	if slotScope != nil {
		if slotContent := slotScope.GetSlot(slotName); slotContent != nil {
			// Found explicit slot content - evaluate it with the scoped props
			result := []*html.Node{}

			// If the slot content is a template with v-slot, evaluate it with the props
			if slotContent.TemplateNode != nil {
				// Extract scoped variable name from the template's v-slot attribute
				scopedVarName := ""
				for _, attr := range slotContent.TemplateNode.Attr {
					if attr.Key == "v-slot" {
						scopedVarName = attr.Val
					} else if strings.HasPrefix(attr.Key, "v-slot:") || (len(attr.Key) > 0 && attr.Key[0] == '#') {
						// For named slots, extract the scoped var from the attribute value
						scopedVarName = attr.Val
					}
				}

				// Push the scoped props onto the stack
				ctx.stack.Push(nil)
				defer ctx.stack.Pop()

				// If there's a scoped variable name, use it; otherwise use the props directly
				if scopedVarName != "" {
					ctx.stack.Set(scopedVarName, slotProps)
				} else {
					// Set the slot props directly in the context
					for k, v := range slotProps {
						ctx.stack.Set(k, v)
					}
				}

				// Evaluate the template content (children of the template)
				children, err := v.evaluateChildren(ctx, slotContent.TemplateNode, 0)
				if err != nil {
					return nil, err
				}
				result = append(result, children...)
			} else {
				// Use the provided content as-is
				result = append(result, slotContent.Nodes...)
			}

			return result, nil
		}
	}

	// Check for inherited slots from layout (passed via __slotScope__ in context data)
	if inheritedSlotScopeData, ok := ctx.stack.EnvMap()["__slotScope__"]; ok {
		if inheritedSlotScope, ok := inheritedSlotScopeData.(*SlotScope); ok {
			if slotContent := inheritedSlotScope.GetSlot(slotName); slotContent != nil {
				// Use the inherited slot content directly (already parsed as DOM nodes)
				return slotContent.Nodes, nil
			}
		}
	}

	// No explicit slot content - use fallback (children of the slot element)
	if node.FirstChild != nil {
		// Evaluate the fallback content
		return v.evaluateChildren(ctx, node, 0)
	}

	return []*html.Node{}, nil
}
