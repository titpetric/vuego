package vuego

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

// evalTemplate processes `<template>` entries from nodes.
// If a `<template>` element has a :required attribute, the value must be provided.
// A template element may repeat `:required` as needed. If a value is not provided,
// template evaluation fails with an error that needs to be bubbled up preventing render.
func (v *Vue) evalTemplate(ctx VueContext, nodes []*html.Node, componentData map[string]any) ([]*html.Node, error) {
	// If no nodes, return empty
	if len(nodes) == 0 {
		return nodes, nil
	}

	// Check if first node is a template tag
	if len(nodes) > 0 && nodes[0].Type == html.ElementNode && nodes[0].Data == "template" {
		templateNode := nodes[0]

		// Validate :required attributes
		var requiredAttrs []string
		for _, attr := range templateNode.Attr {
			if strings.HasPrefix(attr.Key, ":required") || strings.HasPrefix(attr.Key, ":require") {
				requiredAttrs = append(requiredAttrs, attr.Val)
			}
		}

		// Check if all required attributes are provided
		for _, required := range requiredAttrs {
			if _, exists := componentData[required]; !exists {
				return nil, fmt.Errorf("required attribute '%s' not provided", required)
			}
		}

		// Return only the children of the template tag, omitting the template tag itself
		var children []*html.Node
		for c := range templateNode.ChildNodes() {
			children = append(children, c)
		}
		return children, nil
	}

	// If no template tag, return nodes as-is
	return nodes, nil
}
