package vuego

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

// evalTemplate processes `<template>` entries from nodes.
// If a `<template>` element has a :required attribute, the value(s) must be provided.
// Values can be a single field or comma-separated list (e.g., :required="name,title,author").
// A template element may repeat `:required` as needed. If a value is not provided,
// template evaluation fails with an error that needs to be bubbled up preventing render.
func (v *Vue) evalTemplate(ctx VueContext, nodes []*html.Node, componentData map[string]any, depth int) ([]*html.Node, error) {
	// If no nodes, return empty
	if len(nodes) == 0 {
		return nodes, nil
	}

	// Check if first node is a template tag
	if len(nodes) > 0 && nodes[0].Type == html.ElementNode && nodes[0].Data == "template" {
		node := nodes[0]

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
