package vuego

import (
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/net/html"

	"github.com/titpetric/vuego/internal/helpers"
)

// evalTemplate processes `<template>` entries from nodes.
// If a `<template>` element has a :required attribute, the value(s) must be provided.
// Values can be a single field or comma-separated list (e.g., :required="name,title,author").
// A template element may repeat `:required` as needed. If a value is not provided,
// template evaluation fails with an error that needs to be bubbled up preventing render.
// If a `<template>` element has an include attribute, it loads and processes the included file.
func (v *Vue) evalTemplate(ctx VueContext, nodes []*html.Node, componentData map[string]any, depth int) ([]*html.Node, error) {
	// If no nodes, return empty
	if len(nodes) == 0 {
		return nodes, nil
	}

	// Check if first node is a template tag
	if len(nodes) > 0 && nodes[0].Type == html.ElementNode && nodes[0].Data == "template" {
		node := nodes[0]

		// Check for include attribute - handle inclusion first
		if helpers.HasAttr(node, "include") {
			vars, err := v.evalAttributes(ctx, node)
			if err != nil {
				return nil, err
			}

			delete(vars, "include")

			// auto decode params as json, e.g. `data="{...}"` or `[...]`
			for k, v := range vars {
				if vs, ok := v.(string); ok {
					if strings.HasPrefix(vs, "{") || strings.HasPrefix(vs, "[") {
						var out any
						if err := json.Unmarshal([]byte(vs), &out); err == nil {
							vars[k] = out
						}
					}
				}
			}

			evaluated, err := v.evalInclude(ctx, node, vars, depth)
			if err != nil {
				return nil, err
			}
			return evaluated, nil
		}

		// Validate :required attributes
		var requiredAttrs []string
		for _, attr := range node.Attr {
			if attr.Key == ":require" || attr.Key == ":required" {
				// Support CSV format: :required="name,label,type"
				fields := strings.Split(attr.Val, ",")
				for _, field := range fields {
					field = strings.TrimSpace(field)
					if field != "" {
						requiredAttrs = append(requiredAttrs, field)
					}
				}
			}
		}

		// Check if all required attributes are provided
		for _, required := range requiredAttrs {
			if _, exists := componentData[required]; !exists {
				return nil, fmt.Errorf("required attribute '%s' not provided", required)
			}
		}

		// Evaluate v-html if attribute is provided
		if err := v.evalVHtml(ctx, nodes[0]); err != nil {
			return nil, err
		}

		// Check if v-html was evaluated (internal attribute set)
		hasVHtml := false
		for _, attr := range node.Attr {
			if attr.Key == "data-v-html-content" {
				hasVHtml = true
				break
			}
		}

		// If v-html was evaluated, return the template node for rendering to output its content
		if hasVHtml {
			return nodes, nil
		}

		// Evaluate attributes and set them in current scope
		// This allows templates to set variables like <template count="123"> or <template :x="5">
		for _, attr := range node.Attr {
			key := attr.Key
			val := strings.TrimSpace(attr.Val)

			// Skip directive attributes
			if strings.HasPrefix(key, "v-") {
				continue
			}

			// Handle bound attributes (: or v-bind:)
			if strings.HasPrefix(key, ":") {
				boundName := key[1:]
				// Skip special attributes like :required
				if boundName == "require" || boundName == "required" {
					continue
				}

				// Use expression evaluator for templates to support literals and expressions
				result, err := v.exprEval.Eval(val, ctx.stack.EnvMap())
				if err == nil {
					ctx.stack.Set(boundName, result)
					continue
				}

				// Fall back to variable resolution if expression evaluation fails
				valResolved, ok := ctx.stack.Resolve(val)
				if ok {
					ctx.stack.Set(boundName, valResolved)
				} else {
					ctx.stack.Set(boundName, nil)
				}
				continue
			}
			if strings.HasPrefix(key, "v-bind:") {
				boundName := key[7:]
				result, err := v.exprEval.Eval(val, ctx.stack.EnvMap())
				if err == nil {
					ctx.stack.Set(boundName, result)
					continue
				}
				valResolved, ok := ctx.stack.Resolve(val)
				if ok {
					ctx.stack.Set(boundName, valResolved)
				} else {
					ctx.stack.Set(boundName, nil)
				}
				continue
			}

			// Handle plain attributes - auto decode JSON if applicable
			if strings.HasPrefix(val, "{") || strings.HasPrefix(val, "[") {
				var out any
				if err := json.Unmarshal([]byte(val), &out); err == nil {
					ctx.stack.Set(key, out)
					continue
				}
			}
			ctx.stack.Set(key, val)
		}

		// Otherwise, evaluate children and return them (omitting the template tag)
		evaluated, err := v.evaluateChildren(ctx, node, depth+1)
		if err != nil {
			return nil, err
		}
		return evaluated, nil
	}

	// If no template tag, return nodes as-is
	return nodes, nil
}
