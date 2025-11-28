package helpers

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

// CompareHTML parses two HTML strings and returns true if their DOMs match.
// It ignores pure-whitespace text nodes and compares attributes order-insensitively.
func CompareHTML(tb testing.TB, want, got, template, data []byte) bool {
	an, err := html.Parse(bytes.NewReader(want))
	require.NoError(tb, err)

	bn, err := html.Parse(bytes.NewReader(got))
	require.NoError(tb, err)

	// build significant children lists for each root
	ac := SignificantChildren(an)
	bc := SignificantChildren(bn)

	isEqual := func() bool {
		if len(ac) != len(bc) {
			return false
		}

		for i := range ac {
			if !compareNodeRecursive(ac[i], bc[i]) {
				return false
			}
		}
		return true
	}()

	if !isEqual {
		tb.Logf("\n--- template:\n%s\n--- json:\n%s\n--- expected:\n%s\n--- actual:\n%s\n", string(template), string(data), string(want), string(got))
	}
	return isEqual
}

// SignificantChildren returns a slice of child nodes of root that are not pure-whitespace text nodes.
// For document nodes it returns children, for element nodes it returns the node itself wrapped as single-item slice.
func SignificantChildren(root *html.Node) []*html.Node {
	if root == nil {
		return nil
	}
	// If root is a DocumentNode (the parsed doc), return its non-whitespace children.
	if root.Type == html.DocumentNode {
		out := make([]*html.Node, 0, CountChildren(root))
		for c := root.FirstChild; c != nil; c = c.NextSibling {
			if isIgnorable(c) {
				continue
			}
			out = append(out, c)
		}
		// If there's exactly one non-whitespace child and that child is <html>,
		// prefer returning that child's significant children (so full-doc vs fragment compares).
		if len(out) == 1 && out[0].Type == html.ElementNode && out[0].Data == "html" {
			// descend into html node
			htmlChildren := make([]*html.Node, 0, CountChildren(out[0]))
			for c := out[0].FirstChild; c != nil; c = c.NextSibling {
				if isIgnorable(c) {
					continue
				}
				htmlChildren = append(htmlChildren, c)
			}
			// if the single child is <body>, return its children (so body content is comparable)
			if len(htmlChildren) == 1 && htmlChildren[0].Type == html.ElementNode && htmlChildren[0].Data == "body" {
				bodyChildren := make([]*html.Node, 0, CountChildren(htmlChildren[0]))
				for c := htmlChildren[0].FirstChild; c != nil; c = c.NextSibling {
					if isIgnorable(c) {
						continue
					}
					bodyChildren = append(bodyChildren, c)
				}
				return bodyChildren
			}
			// otherwise return the htmlChildren (head/body elements)
			return htmlChildren
		}
		return out
	}

	// For non-document roots, return the node itself (so caller compares element vs element)
	return []*html.Node{root}
}

// compareNodeRecursive compares two nodes recursively.
// Returns true if they match (tag names, attributes, and child nodes / text content).
func compareNodeRecursive(a, b *html.Node) bool {
	if a == nil || b == nil {
		return a == b
	}
	// Skip ignorable nodes (shouldn't be present in significantChildren)
	if isIgnorable(a) && isIgnorable(b) {
		return true
	}
	// Types must match
	if a.Type != b.Type {
		return false
	}
	switch a.Type {
	case html.ElementNode:
		// tag name
		if a.Data != b.Data {
			return false
		}
		// attributes (order-insensitive)
		if !attrsEqual(a.Attr, b.Attr) {
			return false
		}
	case html.TextNode:
		// compare trimmed text
		at := strings.TrimSpace(a.Data)
		bt := strings.TrimSpace(b.Data)
		if at != bt {
			return false
		}
	case html.CommentNode:
		// ignore comments (optional: compare them by trimmed content instead)
		return true
	}

	// Compare children (ignoring whitespace-only text nodes)
	ac := filteredChildren(a)
	bc := filteredChildren(b)
	if len(ac) != len(bc) {
		return false
	}
	for i := range ac {
		if !compareNodeRecursive(ac[i], bc[i]) {
			return false
		}
	}
	return true
}

// filteredChildren returns non-ignorable children of n.
func filteredChildren(n *html.Node) []*html.Node {
	out := make([]*html.Node, 0, CountChildren(n))
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if isIgnorable(c) {
			continue
		}
		out = append(out, c)
	}
	return out
}

// isIgnorable returns true for text nodes that are pure whitespace.
func isIgnorable(n *html.Node) bool {
	if n == nil {
		return true
	}
	if n.Type == html.TextNode && strings.TrimSpace(n.Data) == "" {
		return true
	}
	return false
}

// attrsEqual compares two attribute slices as unordered sets (key->value).
// For style attributes, compares CSS properties order-insensitively.
// For class attributes, compares classes order-insensitively.
func attrsEqual(a, b []html.Attribute) bool {
	if len(a) != len(b) {
		return false
	}
	ma := make(map[string]string, len(a))
	for _, at := range a {
		ma[at.Key] = at.Val
	}
	mb := make(map[string]string, len(b))
	for _, at := range b {
		mb[at.Key] = at.Val
	}
	for k, v := range ma {
		if vb, ok := mb[k]; !ok {
			return false
		} else if k == "style" {
			// Compare style attributes by parsed properties (order-insensitive)
			if !styleEqual(v, vb) {
				return false
			}
		} else if k == "class" {
			// Compare class attributes by class names (order-insensitive)
			if !classEqual(v, vb) {
				return false
			}
		} else if vb != v {
			return false
		}
	}
	return true
}

// styleEqual compares two CSS style strings, ignoring property order.
func styleEqual(a, b string) bool {
	propsA := parseStyleProperties(a)
	propsB := parseStyleProperties(b)

	if len(propsA) != len(propsB) {
		return false
	}
	for k, v := range propsA {
		if vb, ok := propsB[k]; !ok || v != vb {
			return false
		}
	}
	return true
}

// parseStyleProperties parses a CSS style string into a map of property:value pairs.
func parseStyleProperties(style string) map[string]string {
	props := make(map[string]string)
	if style == "" {
		return props
	}

	// Split by semicolon
	parts := strings.Split(style, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split by colon
		idx := strings.Index(part, ":")
		if idx == -1 {
			continue
		}

		key := strings.TrimSpace(part[:idx])
		val := strings.TrimSpace(part[idx+1:])
		props[key] = val
	}
	return props
}

// classEqual compares two class attribute strings, ignoring order.
func classEqual(a, b string) bool {
	classesA := parseClasses(a)
	classesB := parseClasses(b)

	if len(classesA) != len(classesB) {
		return false
	}
	for class := range classesA {
		if !classesB[class] {
			return false
		}
	}
	return true
}

// parseClasses parses a space-separated class string into a set (map of bools).
func parseClasses(classes string) map[string]bool {
	result := make(map[string]bool)
	if classes == "" {
		return result
	}

	// Split by whitespace
	parts := strings.Fields(classes)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result[part] = true
		}
	}
	return result
}
