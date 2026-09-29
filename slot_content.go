package vuego

import (
	"golang.org/x/net/html"
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
