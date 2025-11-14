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
	result, err := v.exprEval.Eval(expr, ctx.stack.EnvMap())
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
