package vuego

import (
	"fmt"

	"golang.org/x/net/html"
)

func (v *Vue) evalCondition(node *html.Node, expr string) (bool, error) {
	removeAttr(node, "v-if")

	// Handle negation with ! prefix
	negated := false
	if len(expr) > 0 && expr[0] == '!' {
		negated = true
		expr = expr[1:]
	}

	val, ok := v.stack.Resolve(expr)
	if !ok {
		return negated, nil
	}

	var result bool
	switch b := val.(type) {
	case bool:
		result = b
	case string:
		result = b != ""
	case int, int64, float64:
		result = fmt.Sprintf("%v", b) != "0"
	default:
		result = val != nil
	}

	if negated {
		return !result, nil
	}
	return result, nil
}
