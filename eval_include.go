package vuego

import (
	"fmt"

	"golang.org/x/net/html"

	"github.com/titpetric/vuego/internal/helpers"
	"github.com/titpetric/vuego/internal/parser"
)

// evalInclude processes a <vuego include="..."> tag with the given vars map.
// Handles stack push/pop properly using defer to ensure cleanup even on error.
func (v *Vue) evalInclude(ctx VueContext, node *html.Node, vars map[string]any, depth int) ([]*html.Node, error) {
	ctx.stack.Push(vars)
	defer ctx.stack.Pop()

	// Extract slot content from the component tag if not already processed
	if ctx.SlotScope == nil {
		ctx.SlotScope = extractSlotContent(node)
	}

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
	return v.evaluate(childCtx, processedDom, depth+1)
}
