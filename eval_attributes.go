package vuego

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

func (v *Vue) evalAttributes(n *html.Node) error {
	if n.Type != html.ElementNode {
		return nil
	}

	var newAttrs []html.Attribute
	for _, a := range n.Attr {
		key := a.Key
		val := strings.TrimSpace(a.Val)

		switch {
		case strings.HasPrefix(key, ":"):
			boundName := strings.TrimPrefix(key, ":")
			valResolved, _ := v.stack.Resolve(val)
			newAttrs = append(newAttrs, html.Attribute{
				Key: boundName,
				Val: fmt.Sprintf("%v", valResolved),
			})

		case strings.HasPrefix(key, "v-bind:"):
			boundName := strings.TrimPrefix(key, "v-bind:")
			valResolved, _ := v.stack.Resolve(val)
			newAttrs = append(newAttrs, html.Attribute{
				Key: boundName,
				Val: fmt.Sprintf("%v", valResolved),
			})

		default:
			// run interpolation on normal attributes
			interpolated, _ := v.interpolate(val)
			newAttrs = append(newAttrs, html.Attribute{
				Key: key,
				Val: interpolated,
			})
		}
	}

	n.Attr = newAttrs
	return nil
}
