package vuego

import "golang.org/x/net/html"

// getAttr returns the value of an attribute by key, or empty string if not found.
func getAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

// removeAttr removes an attribute from a node by key.
func removeAttr(n *html.Node, key string) {
	var attrs []html.Attribute
	for _, a := range n.Attr {
		if a.Key == key {
			continue
		}
		attrs = append(attrs, a)
	}
	n.Attr = attrs
}

// cloneNode creates a shallow copy of a node without sharing children or siblings.
func cloneNode(n *html.Node) *html.Node {
	newNode := *n
	newNode.FirstChild = nil
	newNode.NextSibling = nil
	return &newNode
}

// deepCloneNode creates a deep copy of a node including all children.
func deepCloneNode(n *html.Node) *html.Node {
	clone := &html.Node{
		Type:     n.Type,
		DataAtom: n.DataAtom,
		Data:     n.Data,
		Attr:     append([]html.Attribute(nil), n.Attr...),
	}

	var prev *html.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		childClone := deepCloneNode(c)
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
