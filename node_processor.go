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

// preProcessNodes applies all registered node processors to the rendered nodes.
func (v *Vue) preProcessNodes(ctx VueContext, nodes []*html.Node) error {
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
