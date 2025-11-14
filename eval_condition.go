package vuego

import (
	"strings"

	"github.com/titpetric/vuego/internal/helpers"
	"golang.org/x/net/html"
)

func (v *Vue) evalCondition(ctx VueContext, node *html.Node, expr string) (bool, error) {
	helpers.RemoveAttr(node, "v-if")

	expr = strings.TrimSpace(expr)

	// Normalize comparison operators: coalesce === to == and !== to !=
	expr = helpers.NormalizeComparisonOperators(expr)

	// Try to evaluate as expr expression first (supports ==, !=, &&, ||, !, <, >, <=, >=, and function calls)
	result, err := v.exprEval.Eval(expr, getEnvMap(ctx.stack))
	if err == nil {
		// Successfully evaluated with expr - convert to boolean
		return helpers.IsTruthy(result), nil
	}

	// Fall back to legacy behavior for simple variable references
	// This handles cases like "show" or "!show"
	val, ok := ctx.stack.Resolve(expr)
	if !ok {
		// Variable not found - return false unless negated
		return false, nil
	}

	return helpers.IsTruthy(val), nil
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
