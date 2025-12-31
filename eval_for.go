package vuego

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"

	"github.com/titpetric/vuego/internal/helpers"
)

// parseLoops parses v-for strings like:
//
//	"item in items"
//	"(i, v) in items"
func parseFor(s string) ([]string, string, error) {
	s = strings.TrimSpace(s)
	parts := strings.SplitN(s, " in ", 2)
	if len(parts) != 2 {
		return nil, "", fmt.Errorf("invalid v-for expression: %q", s)
	}
	left := strings.TrimSpace(parts[0])
	right := strings.TrimSpace(parts[1])

	var vars []string
	if strings.HasPrefix(left, "(") && strings.HasSuffix(left, ")") {
		inside := strings.TrimSpace(left[1 : len(left)-1])
		for _, p := range strings.Split(inside, ",") {
			vars = append(vars, strings.TrimSpace(p))
		}
	} else {
		vars = []string{left}
	}
	if len(vars) == 0 || vars[0] == "" {
		return nil, "", fmt.Errorf("no iteration variables in v-for: %q", s)
	}
	return vars, right, nil
}

func (v *Vue) evalVFor(ctx VueContext, node *html.Node, nodes []*html.Node, depth int) ([]*html.Node, int, error) {
	skipCount := 0
	result := []*html.Node{}
	if vFor := helpers.GetAttr(node, "v-for"); vFor != "" {
		loopNodes, err := v.evalFor(ctx, node, vFor, depth+1)
		if err != nil {
			return result, skipCount, err
		}

		for _, n := range loopNodes {
			if err := v.evalVHtml(ctx, n); err != nil {
				return result, skipCount, err
			}
			if _, err := v.evalAttributes(ctx, n); err != nil {
				return result, skipCount, err
			}
		}

		result = append(result, loopNodes...)

		// If v-for produced no results, check for v-else on the next sibling
		if len(loopNodes) == 0 {
			for j := 1; j < len(nodes); j++ {
				nextNode := nodes[j]
				// Skip text nodes (whitespace)
				if nextNode.Type != html.ElementNode {
					continue
				}
				// Found the next element - check if it's v-else
				if helpers.HasAttr(nextNode, "v-else") {
					// Evaluate the v-else node without cloning yet - let evaluateNodeAsElement handle it
					vElseResult, err := v.evaluateNodeAsElement(ctx, nextNode, depth)
					if err != nil {
						return result, skipCount, err
					}
					result = append(result, vElseResult...)
					skipCount = j
				}
				// Stop looking after the first element node (v-else or not)
				break
			}
		}
	}
	return result, skipCount, nil
}

func (v *Vue) evalFor(ctx VueContext, node *html.Node, expr string, depth int) ([]*html.Node, error) {
	vars, collectionName, err := parseFor(expr)
	if err != nil {
		return nil, err
	}

	var result []*html.Node

	err = ctx.stack.ForEach(collectionName, func(index int, value any) error {
		iterNode := helpers.DeepCloneNode(node)
		helpers.RemoveAttr(iterNode, "v-for")

		ctx.stack.Push(nil)

		switch len(vars) {
		case 1:
			ctx.stack.Set(vars[0], value)
		case 2:
			ctx.stack.Set(vars[0], index)
			ctx.stack.Set(vars[1], value)
		default:
			ctx.stack.Pop()
			return fmt.Errorf("v-for variables must be 1 or 2, got %d", len(vars))
		}

		evaluated, err := v.evaluate(ctx, []*html.Node{iterNode}, depth)
		if err != nil {
			ctx.stack.Pop()
			return err
		}

		// After evaluation, propagate bound attributes from template back to parent scope
		// This allows stateful attributes like :printed="printed+1" to persist across iterations
		if iterNode.Type == html.ElementNode && iterNode.Data == "template" {
			for _, attr := range iterNode.Attr {
				boundName := attr.Key
				if strings.HasPrefix(boundName, ":") {
					boundName = boundName[1:]
				} else if strings.HasPrefix(boundName, "v-bind:") {
					boundName = boundName[7:]
				} else {
					continue
				}
				// Get the value that was set in the current scope and propagate to parent
				if val, ok := ctx.stack.Lookup(boundName); ok {
					// Pop current scope, set in parent, push back
					topIdx := len(ctx.stack.stack) - 1
					topMap := ctx.stack.stack[topIdx]
					ctx.stack.stack = ctx.stack.stack[:topIdx]
					ctx.stack.Set(boundName, val)
					ctx.stack.stack = append(ctx.stack.stack, topMap)
				}
			}
		}

		ctx.stack.Pop()
		result = append(result, evaluated...)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}
