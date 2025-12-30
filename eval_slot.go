package vuego

import (
	"strings"

	"golang.org/x/net/html"

	"github.com/titpetric/vuego/internal/helpers"
)

// SlotContent holds the processed content for a slot along with any scoped props.
type SlotContent struct {
	// Nodes contains the rendered content for this slot.
	Nodes []*html.Node
	// Props contains any scoped props passed to the slot.
	Props map[string]any
	// TemplateNode holds the original template node for processing scoped slots.
	TemplateNode *html.Node
}

// SlotScope holds all slot contents indexed by name for a component instance.
type SlotScope struct {
	// Slots maps slot names to their content.
	// Empty string key is the default slot.
	Slots map[string]*SlotContent
}

// NewSlotScope creates a new SlotScope for a component.
func NewSlotScope() *SlotScope {
	return &SlotScope{
		Slots: make(map[string]*SlotContent),
	}
}

// GetSlot retrieves slot content by name.
// Returns nil if the slot doesn't exist.
func (ss *SlotScope) GetSlot(name string) *SlotContent {
	return ss.Slots[name]
}

// SetSlot sets the content for a named slot.
func (ss *SlotScope) SetSlot(name string, content *SlotContent) {
	ss.Slots[name] = content
}

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
		val, err := v.exprEval.Eval(attr.Val, ctx.stack.EnvMap())
		if err == nil && val != nil {
			slotProps[propName] = val
		}
	}

	// Try to get provided slot content
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

	// No explicit slot content - use fallback (children of the slot element)
	if node.FirstChild != nil {
		// Evaluate the fallback content
		return v.evaluateChildren(ctx, node, 0)
	}

	return []*html.Node{}, nil
}
