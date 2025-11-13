package vuego

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

// parseLoops parses v-for strings like:
//
//	"item in items"
//	"(i, v) in items"
func parseFor(s string) ([]string, string, error) {
	s = strings.TrimSpace(s)
	parts := strings.SplitN(s, " in ", 2)
	if len(parts) != 2 {
		return nil, "", fmt.Errorf("invalid v-for expression: %q", s)
	}
	left := strings.TrimSpace(parts[0])
	right := strings.TrimSpace(parts[1])

	var vars []string
	if strings.HasPrefix(left, "(") && strings.HasSuffix(left, ")") {
		inside := strings.TrimSpace(left[1 : len(left)-1])
		for _, p := range strings.Split(inside, ",") {
			vars = append(vars, strings.TrimSpace(p))
		}
	} else {
		vars = []string{left}
	}
	if len(vars) == 0 || vars[0] == "" {
		return nil, "", fmt.Errorf("no iteration variables in v-for: %q", s)
	}
	return vars, right, nil
}

func (v *Vue) evalFor(ctx VueContext, node *html.Node, expr string, depth int) ([]*html.Node, error) {
	vars, collectionName, err := parseFor(expr)
	if err != nil {
		return nil, err
	}

	var result []*html.Node

	err = v.stack.ForEach(collectionName, func(index int, value any) error {
		iterNode := deepCloneNode(node)
		removeAttr(iterNode, "v-for")

		v.stack.Push(nil)
		defer v.stack.Pop()

		switch len(vars) {
		case 1:
			v.stack.Set(vars[0], value)
		case 2:
			v.stack.Set(vars[0], index)
			v.stack.Set(vars[1], value)
		default:
			return fmt.Errorf("v-for variables must be 1 or 2, got %d", len(vars))
		}

		evaluated, err := v.evaluate(ctx, []*html.Node{iterNode}, depth) // Use iterNode instead of node█
		if err != nil {
			return err
		}
		result = append(result, evaluated...)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// clone a node without sharing children/next
func cloneNode(n *html.Node) *html.Node {
	newNode := *n
	newNode.FirstChild = nil
	newNode.NextSibling = nil
	return &newNode
}

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
