package vuego

import (
	"strings"

	"golang.org/x/net/html"

	"github.com/titpetric/vuego/internal/helpers"
)

// hasVSlot checks if a template has a v-slot directive.
func hasVSlot(node *html.Node) bool {
	for _, attr := range node.Attr {
		if attr.Key == "v-slot" {
			return true
		}
		if strings.HasPrefix(attr.Key, "v-slot:") {
			return true
		}
		if len(attr.Key) > 0 && attr.Key[0] == '#' {
			return true
		}
	}
	return false
}

// GetSlotName extracts the slot name from a v-slot directive.
// Returns "default" if no name is specified.
func GetSlotName(attr string) string {
	if attr == "" {
		return "default"
	}
	// v-slot:name or v-slot:default
	return attr
}

// parseVSlotDirective parses a v-slot directive to extract the slot name and destructuring vars.
// Supports formats like:
//   - v-slot (default slot)
//   - v-slot:name (named slot)
//   - v-slot="varName" (default slot with scoped props)
//   - v-slot:name="varName" (named slot with scoped props)
//   - v-slot="{ prop1, prop2 }" (default slot with destructured props)
//   - v-slot:name="{ prop1, prop2 }" (named slot with destructured props)
//   - #name (shorthand for v-slot:name)
func parseVSlotDirective(attrKey, attrVal string) (slotName string, scopedVar string) {
	slotName = "default"

	// Handle shorthand #name format
	if len(attrKey) > 0 && attrKey[0] == '#' {
		if len(attrKey) > 1 {
			slotName = attrKey[1:]
		}
		scopedVar = attrVal
		return
	}

	// Handle v-slot directive
	// Parse v-slot:name or v-slot
	// The slot name would be in the attribute key structure passed separately
	scopedVar = attrVal
	return
}

// extractSlotContent collects content and templates from a component element
// to be injected into the component's slot elements.
func extractSlotContent(node *html.Node) *SlotScope {
	scope := NewSlotScope()

	var defaultSlotContent []*html.Node

	// Walk through children to separate templates with v-slot from other content
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode {
			if c.Type == html.TextNode {
				// Text nodes go to default slot unless they're only whitespace
				if trimmedText := strings.TrimSpace(c.Data); trimmedText != "" {
					defaultSlotContent = append(defaultSlotContent, helpers.CloneNode(c))
				}
			}
			continue
		}

		if c.Data == "template" && hasVSlot(c) {
			// Extract slot name from v-slot directive
			slotName := "default"

			for _, attr := range c.Attr {
				// Handle shorthand #name format
				if len(attr.Key) > 0 && attr.Key[0] == '#' {
					slotName = attr.Key[1:]
					if slotName == "" {
						slotName = "default"
					}
					break
				}
				// Handle v-slot:name format
				if strings.HasPrefix(attr.Key, "v-slot:") {
					slotName = attr.Key[7:] // Remove "v-slot:" prefix
					if slotName == "" {
						slotName = "default"
					}
					break
				}
				// Handle plain v-slot with value format
				if attr.Key == "v-slot" {
					slotName, _ = parseVSlotDirective("v-slot", attr.Val)
					break
				}
			}

			if slotName == "" {
				slotName = "default"
			}

			// Clone the template's children as the slot content
			var slotNodes []*html.Node
			for cc := c.FirstChild; cc != nil; cc = cc.NextSibling {
				slotNodes = append(slotNodes, helpers.DeepCloneNode(cc))
			}

			scope.SetSlot(slotName, &SlotContent{
				Nodes:        slotNodes,
				Props:        make(map[string]any),
				TemplateNode: c,
			})
		} else {
			// Non-template element goes to default slot
			defaultSlotContent = append(defaultSlotContent, helpers.DeepCloneNode(c))
		}
	}

	// Add default slot content if any
	if len(defaultSlotContent) > 0 {
		scope.SetSlot("default", &SlotContent{
			Nodes: defaultSlotContent,
			Props: make(map[string]any),
		})
	}

	return scope
}
