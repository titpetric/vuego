package vuego

import (
	"strings"

	"golang.org/x/net/html"

	"github.com/titpetric/vuego/internal/helpers"
)

// evalCondition evaluates a v-if condition without modifying any node attributes.
// Attributes should be removed later during rendering.
func (v *Vue) evalCondition(ctx VueContext, expr string) (bool, error) {
	return v.evalConditionExpr(ctx, expr)
}

// evalConditionExpr evaluates a condition expression without modifying any node attributes.
// Used for v-if, v-else-if and other condition evaluations.
func (v *Vue) evalConditionExpr(ctx VueContext, expr string) (bool, error) {
	expr = strings.TrimSpace(expr)

	// Normalize comparison operators: coalesce === to == and !== to !=
	expr = helpers.NormalizeComparisonOperators(expr)

	// Try to evaluate as expr expression first (supports ==, !=, &&, ||, !, <, >, <=, >=, and function calls)
	result, err := v.exprEval.Eval(expr, ctx.stack.EnvMap())
	if err == nil {
		// Successfully evaluated with expr - convert to boolean
		return helpers.IsTruthy(result), nil
	}

	// Fall back to legacy behavior for simple variable references
	// This handles cases like "show" or "!show"
	val, ok := ctx.stack.Resolve(expr)
	if !ok {
		// Variable not found - return false unless negated
		return false, nil
	}

	return helpers.IsTruthy(val), nil
}

// evalElseIfChain evaluates a v-if, v-else-if, v-else chain starting at the given node.
// It returns the result nodes, the number of nodes to skip (including v-else-if/v-else),
// and an error if evaluation fails.
// The skipCount includes all nodes consumed by the chain (up to and including the matched node or the end of the chain).
func (v *Vue) evalElseIfChain(ctx VueContext, node *html.Node, nodes []*html.Node, depth int) ([]*html.Node, int, error) {
	var result []*html.Node
	lastChainNodeIdx := 0 // Track the last node in the chain for skipCount

	// Check for v-if
	if vIf := helpers.GetAttr(node, "v-if"); vIf != "" {
		ok, err := v.evalCondition(ctx, vIf)
		if err != nil {
			return nil, 0, err
		}
		if ok {
			// v-if condition is true - evaluate node (don't remove attribute here, filter during rendering)
			// Find the last node in the chain to determine skipCount
			// We need to skip past any v-else-if and v-else nodes that follow
			for idx := 1; idx < len(nodes); idx++ {
				nextNode := nodes[idx]
				if nextNode.Type != html.ElementNode {
					continue
				}
				if !helpers.HasAttr(nextNode, "v-else-if") && !helpers.HasAttr(nextNode, "v-else") {
					break
				}
				lastChainNodeIdx = idx
			}
			// Evaluate the node (evaluateNodeAsElement handles cloning internally)
			evaluated, err := v.evaluateNodeAsElement(ctx, node, depth)
			return evaluated, lastChainNodeIdx, err
		}
		// v-if condition is false - check next nodes for v-else-if or v-else
	} else {
		// No v-if attribute found - shouldn't happen, but handle gracefully
		return result, lastChainNodeIdx, nil
	}

	// Check for v-else-if and v-else in following nodes
	for idx := 1; idx < len(nodes); idx++ {
		nextNode := nodes[idx]

		// Skip text nodes (whitespace)
		if nextNode.Type != html.ElementNode {
			continue
		}

		// Only consider nodes with v-else-if or v-else
		if !helpers.HasAttr(nextNode, "v-else-if") && !helpers.HasAttr(nextNode, "v-else") {
			// No more else directives in the chain
			break
		}

		lastChainNodeIdx = idx // Update the last chain node index

		if vElseIf := helpers.GetAttr(nextNode, "v-else-if"); vElseIf != "" {
			ok, err := v.evalConditionExpr(ctx, vElseIf)
			if err != nil {
				return nil, 0, err
			}
			if ok {
				// v-else-if condition is true - evaluate and return this node (don't remove attribute, filter during rendering)
				// Evaluate the node (evaluateNodeAsElement handles cloning internally)
				evaluated, err := v.evaluateNodeAsElement(ctx, nextNode, depth)
				return evaluated, idx, err
			}
			// v-else-if condition is false - continue to next
			continue
		}

		// Check for v-else
		if helpers.HasAttr(nextNode, "v-else") {
			// v-else always matches - evaluate and return this node (don't remove attribute, filter during rendering)
			// Evaluate the node (evaluateNodeAsElement handles cloning internally)
			evaluated, err := v.evaluateNodeAsElement(ctx, nextNode, depth)
			return evaluated, idx, err
		}
	}

	// No condition in the chain was true - skip all chain nodes anyway
	return result, lastChainNodeIdx, nil
}

// evaluateNodeAsElement evaluates a single element node with its v-for and other directives.
// This is used internally by the else-if chain handler.
func (v *Vue) evaluateNodeAsElement(ctx VueContext, node *html.Node, depth int) ([]*html.Node, error) {
	var result []*html.Node

	// Handle v-for if present
	if vFor := helpers.GetAttr(node, "v-for"); vFor != "" {
		loopNodes, err := v.evalFor(ctx, node, vFor, depth+1)
		if err != nil {
			return nil, err
		}

		for _, n := range loopNodes {
			if err := v.evalVHtml(ctx, n); err != nil {
				return nil, err
			}
			if _, err := v.evalAttributes(ctx, n); err != nil {
				return nil, err
			}
		}

		result = append(result, loopNodes...)
		return result, nil
	}

	// Regular element node processing (no v-for)
	hasVHtml := helpers.GetAttr(node, "v-html") != ""
	var newNode *html.Node
	if hasVHtml {
		newNode = helpers.DeepCloneNode(node)
	} else {
		newNode = helpers.ShallowCloneWithAttrs(node)
	}

	if err := v.evalVHtml(ctx, newNode); err != nil {
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
	return result, nil
}
