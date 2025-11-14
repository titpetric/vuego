package vuego

import (
	"fmt"

	"github.com/titpetric/vuego/internal/helpers"
	"golang.org/x/net/html"
)

func (v *Vue) evalCondition(ctx VueContext, node *html.Node, expr string) (bool, error) {
	helpers.RemoveAttr(node, "v-if")

	expr = trimSpace(expr)

	// Normalize comparison operators: coalesce === to == and !== to !=
	expr = normalizeComparisonOperators(expr)

	// Try to evaluate as expr expression first (supports ==, !=, &&, ||, !, <, >, <=, >=, and function calls)
	result, err := v.exprEval.Eval(expr, getEnvMap(ctx.stack))
	if err == nil {
		// Successfully evaluated with expr - convert to boolean
		return isTruthy(result), nil
	}

	// Fall back to legacy behavior for simple variable references
	// This handles cases like "show" or "!show"
	val, ok := ctx.stack.Resolve(expr)
	if !ok {
		// Variable not found - return false unless negated
		return false, nil
	}

	return isTruthy(val), nil
}

// isTruthy converts a value to boolean following Vue semantics.
// For bound attributes, false values should not render the attribute.
func isTruthy(val any) bool {
	switch b := val.(type) {
	case bool:
		return b
	case string:
		// Empty string and "false" are falsey
		if b == "" || b == "false" {
			return false
		}
		return true
	case int, int64, float64:
		return fmt.Sprintf("%v", b) != "0"
	case nil:
		return false
	default:
		return true
	}
}

// getEnvMap converts a Stack to a map[string]any for expr evaluation.
// This flattens the stack into a single environment map.
func getEnvMap(s *Stack) map[string]any {
	result := make(map[string]any)
	// Iterate through stack from bottom to top, with top overriding bottom
	for i := 0; i < len(s.stack); i++ {
		for k, v := range s.stack[i] {
			result[k] = v
		}
	}
	return result
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

// normalizeComparisonOperators coalesces strict comparison operators (=== and !==) to loose operators (== and !=).
// This is needed because the underlying expr evaluator supports == and != but not === and !==.
func normalizeComparisonOperators(expr string) string {
	result := make([]byte, 0, len(expr))
	for i := 0; i < len(expr); i++ {
		if i+3 <= len(expr) && expr[i:i+3] == "===" {
			result = append(result, '=', '=')
			i += 2
		} else if i+3 <= len(expr) && expr[i:i+3] == "!==" {
			result = append(result, '!', '=')
			i += 2
		} else {
			result = append(result, expr[i])
		}
	}
	return string(result)
}
