package vuego

import (
	"strings"

	"golang.org/x/net/html"

	"github.com/titpetric/vuego/internal/helpers"
)

// evalVShow handles v-show="condition" directive.
// Sets display:none style when condition is falsey, removes it when truthy.
func (v *Vue) evalVShow(ctx VueContext, n *html.Node) error {
	vShowExpr := helpers.GetAttr(n, "v-show")
	if vShowExpr == "" {
		return nil
	}

	// Evaluate the expression using the same approach as v-if
	val, err := v.exprEval.Eval(vShowExpr, ctx.stack.EnvMap())
	if err != nil {
		// Fall back to stack resolution for simple variable references
		var ok bool
		val, ok = ctx.stack.Resolve(vShowExpr)
		if !ok {
			val = false
		}
	}

	// Set or remove display:none based on condition
	if !helpers.IsTruthy(val) {
		v.setStyleProperty(n, "display", "none")
	}

	return nil
}

// setStyleProperty adds or updates a CSS property in the style attribute.
// Preserves existing style properties.
func (v *Vue) setStyleProperty(n *html.Node, property, value string) {
	styleVal := helpers.GetAttr(n, "style")

	// Parse existing styles
	styleMap := parseStyleString(styleVal)
	styleMap[property] = value

	// Rebuild style string
	var styles []string
	for k, v := range styleMap {
		styles = append(styles, k+":"+v+";")
	}
	helpers.AppendAttr(n, "style", strings.Join(styles, ""))
}

// parseStyleString parses a CSS style string into a map.
// Handles styles like "color: red; display: none;"
func parseStyleString(style string) map[string]string {
	result := make(map[string]string)
	if style == "" {
		return result
	}

	// Split by semicolon to get individual properties
	parts := strings.Split(style, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split by colon to get key-value pair
		kv := strings.SplitN(part, ":", 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			val := strings.TrimSpace(kv[1])
			result[key] = val
		}
	}

	return result
}
