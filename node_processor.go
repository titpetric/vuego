package vuego

import "golang.org/x/net/html"

// NodeProcessor is an interface for post-processing []*html.Node after template evaluation.
// Implementations can inspect and modify HTML nodes to implement custom tags and attributes.
// NodeProcessor receives the complete rendered DOM and can transform it in place.
//
// Processors are called after all Vue directives and interpolations have been evaluated.
// Multiple processors can be registered and are applied in order of registration.
//
// See docs/nodeprocessor.md for examples and best practices.
type NodeProcessor interface {
	// New ensures that we can have render-scoped allocations.
	New() NodeProcessor

	// PreProcess receives the nodes as they are decoded.
	PreProcess(nodes []*html.Node) error

	// PostProcess receives the rendered nodes and may modify them in place.
	// It should return an error if processing fails.
	PostProcess(nodes []*html.Node) error
}

// RegisterNodeProcessor adds a custom node processor to the Vue instance.
// Processors are applied in order after template evaluation completes.
func (v *Vue) RegisterNodeProcessor(processor NodeProcessor) *Vue {
	if v.nodeProcessors == nil {
		v.nodeProcessors = make([]NodeProcessor, 0)
	}
	v.nodeProcessors = append(v.nodeProcessors, processor)
	return v
}

// RegisterComponent registers a component shorthand mapping.
// The tagName (e.g., "button-primary") will be resolved to the given filename (e.g., "components/ButtonPrimary.vuego").
func (v *Vue) RegisterComponent(tagName, filename string) *Vue {
	v.componentMap[tagName] = filename
	return v
}

// GetComponentFile looks up a component file by its kebab-case tag name.
// Returns the filename and true if found, otherwise returns empty string and false.
func (v *Vue) GetComponentFile(tagName string) (string, bool) {
	filename, ok := v.componentMap[tagName]
	return filename, ok
}

// resolveComponentTags replaces component shorthand tags with vuego include directives.
// This is called before other node processors to transform component tags.
func (v *Vue) resolveComponentTags(nodes []*html.Node) error {
	for _, node := range nodes {
		if err := v.processComponentNode(node); err != nil {
			return err
		}
	}
	return nil
}

// processComponentNode recursively processes a node and its children to replace component tags.
func (v *Vue) processComponentNode(node *html.Node) error {
	if node.Type == html.ElementNode {
		// Check if this node is a registered component tag
		if filename, ok := v.GetComponentFile(node.Data); ok {
			// Replace the element with a vuego include
			if err := v.replaceWithInclude(node, filename); err != nil {
				return err
			}
			// Don't process children since we've replaced the node
			return nil
		}
	}

	// Process children
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if err := v.processComponentNode(c); err != nil {
			return err
		}
	}

	return nil
}

// replaceWithInclude transforms a component tag into a vuego include directive.
// It preserves attributes from the original tag as context variables.
func (v *Vue) replaceWithInclude(node *html.Node, filename string) error {
	// Set the include attribute on the node itself
	node.Data = "vuego"
	node.Attr = append(node.Attr, html.Attribute{
		Key: "include",
		Val: filename,
	})
	return nil
}

// preProcessNodes applies all registered node processors to the rendered nodes.
// Component tag resolution happens first, before other processors.
func (v *Vue) preProcessNodes(ctx VueContext, nodes []*html.Node) error {
	// Resolve component shorthand tags first
	if err := v.resolveComponentTags(nodes); err != nil {
		return err
	}

	// Then apply registered node processors
	for _, processor := range ctx.Processors {
		if err := processor.PreProcess(nodes); err != nil {
			return err
		}
	}
	return nil
}

// postProcessNodes applies all registered node processors to the rendered nodes.
func (v *Vue) postProcessNodes(ctx VueContext, nodes []*html.Node) error {
	for _, processor := range ctx.Processors {
		if err := processor.PostProcess(nodes); err != nil {
			return err
		}
	}
	return nil
}
