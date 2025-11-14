package vuego

import (
	"fmt"

	"github.com/titpetric/vuego/internal/helpers"
	"golang.org/x/net/html"
)

func (v *Vue) evalCondition(ctx VueContext, node *html.Node, expr string) (bool, error) {
	helpers.RemoveAttr(node, "v-if")

	// Handle negation with ! prefix
	negated := false
	if len(expr) > 0 && expr[0] == '!' {
		negated = true
		expr = expr[1:]
	}

	var val any
	var ok bool

	// Check if expression is a function call or contains pipes
	if isFunctionCall(expr) || containsPipe(expr) {
		pipe := parsePipeExpr(expr)
		var err error
		val, err = v.evalPipe(ctx, pipe)
		if err != nil {
			return false, fmt.Errorf("v-if='%s': %w", expr, err)
		}
		ok = true
	} else {
		val, ok = ctx.stack.Resolve(expr)
	}

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

// isFunctionCall checks if an expression looks like a function call
func isFunctionCall(expr string) bool {
	expr = trimSpace(expr)
	// Check for pattern: identifier(...)
	for i, ch := range expr {
		if ch == '(' {
			return i > 0
		}
		if !isIdentifierChar(ch) {
			return false
		}
	}
	return false
}

func containsPipe(expr string) bool {
	for _, ch := range expr {
		if ch == '|' {
			return true
		}
	}
	return false
}

func isIdentifierChar(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_'
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
