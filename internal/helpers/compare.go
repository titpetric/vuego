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
		var out []*html.Node
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
			var htmlChildren []*html.Node
			for c := out[0].FirstChild; c != nil; c = c.NextSibling {
				if isIgnorable(c) {
					continue
				}
				htmlChildren = append(htmlChildren, c)
			}
			// if the single child is <body>, return its children (so body content is comparable)
			if len(htmlChildren) == 1 && htmlChildren[0].Type == html.ElementNode && htmlChildren[0].Data == "body" {
				var bodyChildren []*html.Node
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
	var out []*html.Node
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
func attrsEqual(a, b []html.Attribute) bool {
	if len(a) != len(b) {
		// quick fail, but still need to account for duplicate attribute keys (rare)
		// we'll still do full map compare
	}
	ma := make(map[string]string, len(a))
	for _, at := range a {
		ma[at.Key] = at.Val
	}
	mb := make(map[string]string, len(b))
	for _, at := range b {
		mb[at.Key] = at.Val
	}
	if len(ma) != len(mb) {
		return false
	}
	for k, v := range ma {
		if vb, ok := mb[k]; !ok || vb != v {
			return false
		}
	}
	return true
}
