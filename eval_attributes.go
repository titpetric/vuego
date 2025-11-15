package vuego

import (
	"fmt"
	"strings"

	"github.com/titpetric/vuego/internal/helpers"
	"golang.org/x/net/html"
)

func (v *Vue) evalAttributes(ctx VueContext, n *html.Node) error {
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
			valResolved, _ := ctx.stack.Resolve(val)
			// Skip falsey values for boolean attributes
			if helpers.IsTruthy(valResolved) {
				newAttrs = append(newAttrs, html.Attribute{
					Key: boundName,
					Val: fmt.Sprint(valResolved),
				})
			}

		case strings.HasPrefix(key, "v-bind:"):
			boundName := strings.TrimPrefix(key, "v-bind:")
			valResolved, _ := ctx.stack.Resolve(val)
			// Skip falsey values for boolean attributes
			if helpers.IsTruthy(valResolved) {
				newAttrs = append(newAttrs, html.Attribute{
					Key: boundName,
					Val: fmt.Sprint(valResolved),
				})
			}

		default:
			// run interpolation on normal attributes
			interpolated, err := v.interpolate(ctx, val)
			if err != nil {
				// For now, use original value on error in attributes
				interpolated = val
			}
			newAttrs = append(newAttrs, html.Attribute{
				Key: key,
				Val: interpolated,
			})
		}
	}

	n.Attr = newAttrs
	return nil
}
