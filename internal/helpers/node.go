// Package helpers provides HTML node manipulation utilities for vuego.
package helpers

import "golang.org/x/net/html"

// GetAttr returns the value of an attribute by key, or empty string if not found.
func GetAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
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

// CloneNode creates a shallow copy of a node without sharing children or siblings.
func CloneNode(n *html.Node) *html.Node {
	newNode := *n
	newNode.FirstChild = nil
	newNode.NextSibling = nil
	return &newNode
}

// DeepCloneNode creates a deep copy of a node including all children.
func DeepCloneNode(n *html.Node) *html.Node {
	clone := &html.Node{
		Type:     n.Type,
		DataAtom: n.DataAtom,
		Data:     n.Data,
		Attr:     append([]html.Attribute(nil), n.Attr...),
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
