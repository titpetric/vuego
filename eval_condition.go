package vuego

import (
	"fmt"

	"golang.org/x/net/html"
)

func (v *Vue) evalCondition(node *html.Node, expr string) (bool, error) {
	removeAttr(node, "v-if")

	val, ok := v.stack.Lookup(expr)
	if !ok {
		return false, nil
	}
	switch b := val.(type) {
	case bool:
		return b, nil
	case string:
		return b != "", nil
	case int, int64, float64:
		return fmt.Sprintf("%v", b) != "0", nil
	default:
		return val != nil, nil
	}
}
