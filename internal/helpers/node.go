// Package helpers provides HTML node manipulation utilities for vuego.
package helpers

import (
	"strings"
	"sync"

	"golang.org/x/net/html"
)

// nodePool provides pooled html.Node allocations to reduce GC pressure.
// Nodes are reused across renders; the garbage collector will clean up any
// pooled nodes when they go out of scope at the end of Render().
var nodePool = &sync.Pool{
	New: func() any {
		return &html.Node{}
	},
}

// NewNode creates or reuses a node from the pool, clearing any previous state.
// This reduces memory allocations and GC pressure during template rendering.
func NewNode() *html.Node {
	n := nodePool.Get().(*html.Node)
	// Clear previous state
	n.Type = 0
	n.DataAtom = 0
	n.Data = ""
	n.Attr = nil
	n.FirstChild = nil
	n.LastChild = nil
	n.NextSibling = nil
	n.PrevSibling = nil
	n.Parent = nil
	n.Namespace = ""
	return n
}

// GetAttr returns the value of an attribute by key, or empty string if not found.
func GetAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

// HasAttr checks if a node has an attribute by key (regardless of its value).
func HasAttr(n *html.Node, key string) bool {
	for _, a := range n.Attr {
		if a.Key == key {
			return true
		}
	}
	return false
}

// SetAttr sets or updates an attribute on a node. If the attribute exists, its value is replaced.
// If it doesn't exist, it is added.
func SetAttr(n *html.Node, key, value string) {
	for i, a := range n.Attr {
		if a.Key == key {
			n.Attr[i].Val = value
			return
		}
	}
	// Attribute not found, append it
	n.Attr = append(n.Attr, html.Attribute{Key: key, Val: value})
}

// AppendAttr appends a value to an attribute on a node. If the attribute exists, the new value replaces it.
// If it doesn't exist, it is added.
func AppendAttr(n *html.Node, key, value string) {
	SetAttr(n, key, value)
}

// RemoveAttr removes an attribute from a node by key.
func RemoveAttr(n *html.Node, key string) {
	var attrs []html.Attribute
	for _, a := range n.Attr {
		if a.Key == key {
			continue
		}
		attrs = append(attrs, a)
	}
	n.Attr = attrs
}

// FilterAttrs returns a new attribute slice excluding the specified key.
// This is useful for creating a copy of attributes without a specific key.
func FilterAttrs(attrs []html.Attribute, excludeKey string) []html.Attribute {
	result := make([]html.Attribute, 0, len(attrs))
	for _, a := range attrs {
		if a.Key != excludeKey {
			result = append(result, a)
		}
	}
	return result
}

// CloneNode creates a shallow copy of a node without sharing children or siblings.
// Attributes are shared (not copied) to avoid unnecessary allocations.
func CloneNode(n *html.Node) *html.Node {
	newNode := NewNode()
	newNode.Type = n.Type
	newNode.DataAtom = n.DataAtom
	newNode.Data = n.Data
	newNode.Namespace = n.Namespace
	newNode.Attr = n.Attr // Share attributes, not copied
	// Children and siblings are intentionally not copied
	return newNode
}

// ShallowCloneWithAttrs creates a shallow copy of a node and copies its attributes.
// This is useful when you need a new node with the same attributes but will replace its children.
func ShallowCloneWithAttrs(n *html.Node) *html.Node {
	newNode := CloneNode(n)
	if len(n.Attr) > 0 {
		newNode.Attr = append([]html.Attribute(nil), n.Attr...)
	}
	return newNode
}

// DeepCloneNode creates a deep copy of a node including all children.
func DeepCloneNode(n *html.Node) *html.Node {
	clone := NewNode()
	clone.Type = n.Type
	clone.DataAtom = n.DataAtom
	clone.Data = n.Data
	clone.Namespace = n.Namespace
	if len(n.Attr) > 0 {
		clone.Attr = append([]html.Attribute(nil), n.Attr...)
	}

	var prev *html.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		childClone := DeepCloneNode(c)
		childClone.Parent = clone

		if prev == nil {
			clone.FirstChild = childClone
		} else {
			prev.NextSibling = childClone
			childClone.PrevSibling = prev
		}
		prev = childClone
	}

	return clone
}

// bodyNodeCache caches a pre-parsed body element for use as fragment context.
// This avoids repeated parsing of the same wrapper document.
var (
	bodyNodeCache *html.Node
	bodyNodeOnce  sync.Once
)

// GetBodyNode returns a cached body element suitable for html.ParseFragment.
func GetBodyNode() *html.Node {
	bodyNodeOnce.Do(func() {
		doc, _ := html.Parse(strings.NewReader("<html><body></body></html>"))
		var findBody func(*html.Node)
		findBody = func(node *html.Node) {
			if node.Type == html.ElementNode && node.Data == "body" {
				bodyNodeCache = node
				return
			}
			for c := node.FirstChild; c != nil; c = c.NextSibling {
				if bodyNodeCache == nil {
					findBody(c)
				}
			}
		}
		findBody(doc)
	})
	return bodyNodeCache
}

// CountChildren counts the number of child nodes of the given node.
// This is useful for preallocating slices with the correct capacity.
func CountChildren(n *html.Node) int {
	if n == nil {
		return 0
	}
	count := 0
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		count++
	}
	return count
}
