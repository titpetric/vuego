package vuego

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"

	"github.com/titpetric/vuego/internal/helpers"
)

func (v *Vue) evalAttributes(ctx VueContext, n *html.Node) (map[string]any, error) {
	if n.Type != html.ElementNode {
		return nil, nil
	}

	results := map[string]any{}

	var newAttrs []html.Attribute

	// First pass: collect static attributes and evaluate bound ones
	for _, a := range n.Attr {
		key := a.Key
		val := strings.TrimSpace(a.Val)

		// Skip internal attributes that hold already-evaluated content
		// These should not be re-interpolated
		if key == "data-v-text-content" || key == "data-v-html-content" {
			newAttrs = append(newAttrs, html.Attribute{Key: key, Val: val})
			continue
		}

		boundValue := val
		boundName := key
		// literal bindings
		if strings.HasPrefix(key, ":") {
			boundName = boundName[1:]
		}
		if strings.HasPrefix(key, "v-bind:") {
			boundName = boundName[7:]
		}
		if strings.HasPrefix(key, "[") && strings.HasSuffix(key, "]") {
			key = boundName
		}

		switch {
		case boundName != key:
			boundValue, err := v.evalBoundAttribute(ctx, boundName, val)
			if err != nil {
				return nil, fmt.Errorf("error evaluating attr %s: %w", boundName, err)
			}
			if !helpers.IsTruthy(boundValue) {
				continue
			}
			results[boundName] = boundValue
		default:
			var err error
			if containsInterpolation(val) {
				boundValue, err = v.interpolate(ctx, val)
				if err != nil {
					return nil, fmt.Errorf("error evaluating attr %s: %w", boundName, err)
				}
			}
			newAttrs = append(newAttrs, html.Attribute{
				Key: boundName,
				Val: boundValue,
			})
		}
	}

	// Second pass: merge bound attributes with static ones
	for attrName, boundValue := range results {
		// Check if there's a static attribute with the same name
		staticIdx := -1
		for i, a := range newAttrs {
			if a.Key == attrName {
				staticIdx = i
				break
			}
		}

		if staticIdx >= 0 {
			// Merge with static attribute (special handling for class and style)
			if attrName == "class" {
				newAttrs[staticIdx].Val = fmt.Sprintf("%s %s", newAttrs[staticIdx].Val, boundValue)
				continue
			}
			if attrName == "style" {
				// Merge styles, with bound value taking precedence
				staticStyle := newAttrs[staticIdx].Val
				mergedStyle := v.mergeStyles(staticStyle, boundValue.(string))
				newAttrs[staticIdx].Val = mergedStyle
				continue
			}
			// For other attributes, bound value replaces static
			newAttrs[staticIdx].Val = fmt.Sprint(boundValue)
		} else {
			if helpers.IsTruthy(boundValue) {
				newAttrs = append(newAttrs, html.Attribute{
					Key: attrName,
					Val: fmt.Sprint(boundValue),
				})
				continue
			}
		}
	}

	n.Attr = newAttrs

	// Don't overwrite results with stringified attributes
	// The results map already contains the correctly typed bound values from line 39
	// Only add static attributes to results (they are naturally strings from interpolation)
	for _, v := range newAttrs {
		// Only add if not already in results (bound attributes take precedence and keep their type)
		if _, exists := results[v.Key]; !exists {
			results[v.Key] = v.Val
		}
	}

	return results, nil
}

// evalBoundAttribute evaluates a bound attribute, handling objects for class/style and function calls.
func (v *Vue) evalBoundAttribute(ctx VueContext, attrName, expr string) (any, error) {
	expr = strings.TrimSpace(expr)

	// Evaluate interpolation
	if containsInterpolation(expr) {
		interpolated, err := v.interpolate(ctx, expr)
		if err != nil {
			return "", err
		}
		return interpolated, nil
	}

	// Check if the expression is an object literal
	if strings.HasPrefix(expr, "{") && strings.HasSuffix(expr, "}") {
		return v.evalObjectBinding(ctx, attrName, expr), nil
	}

	// Check if it's a function call or pipe expression
	if strings.Contains(expr, "|") || helpers.IsFunctionCall(expr) || helpers.IsComplexExpr(expr) {
		pipe := parsePipeExpr(expr)
		val, err := v.evalPipe(ctx, pipe)
		if err != nil {
			return "", err
		}
		return val, nil
	}

	// Regular variable binding
	valResolved, ok := ctx.stack.Resolve(expr)
	if ok {
		return valResolved, nil
	}
	return "", nil
}
