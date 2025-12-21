package diff

import (
	"bytes"
	"fmt"
	"strings"

	"golang.org/x/net/html"
	yaml "gopkg.in/yaml.v3"
)

// DomToYAML converts HTML DOM to YAML representation for use with dyff.
func DomToYAML(content []byte) string {
	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		return fmt.Sprintf("error: %v\n", err)
	}

	// Convert to simple nested map structure
	data := nodeToMap(doc)

	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Sprintf("error: %v\n", err)
	}

	return string(yamlBytes)
}

// nodeToMap converts an HTML node to a map structure suitable for YAML.
func nodeToMap(n *html.Node) interface{} {
	if n == nil {
		return nil
	}

	switch n.Type {
	case html.DocumentNode:
		var children []interface{}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if child := nodeToMap(c); child != nil {
				children = append(children, child)
			}
		}
		if len(children) == 1 {
			return children[0]
		}
		return children

	case html.ElementNode:
		element := map[string]interface{}{
			"tag": n.Data,
		}

		// Add attributes at the same level as tag
		for _, attr := range n.Attr {
			element[attr.Key] = attr.Val
		}

		// Add children if any
		var children []interface{}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.TextNode {
				text := strings.TrimSpace(c.Data)
				if text != "" {
					children = append(children, map[string]interface{}{"text": text})
				}
			} else if child := nodeToMap(c); child != nil {
				children = append(children, child)
			}
		}

		if len(children) > 0 {
			element["children"] = children
		}

		return element

	case html.TextNode:
		text := strings.TrimSpace(n.Data)
		if text != "" {
			return map[string]interface{}{"text": text}
		}
		return nil

	case html.CommentNode:
		return map[string]interface{}{"comment": strings.TrimSpace(n.Data)}
	}

	return nil
}
