package vuego

import (
	"strings"

	"github.com/titpetric/vuego/internal/helpers"
)

// evalCondition evaluates a v-if condition without modifying any node attributes.
// Attributes should be removed later during rendering.
func (v *Vue) evalCondition(ctx VueContext, expr string) (bool, error) {
	return v.evalConditionExpr(ctx, expr)
}

// evalConditionExpr evaluates a condition expression without modifying any node attributes.
// Used for v-if, v-else-if and other condition evaluations.
func (v *Vue) evalConditionExpr(ctx VueContext, expr string) (bool, error) {
	expr = strings.TrimSpace(expr)

	// Normalize comparison operators: coalesce === to == and !== to !=
	expr = helpers.NormalizeComparisonOperators(expr)

	// Try to evaluate as expr expression first (supports ==, !=, &&, ||, !, <, >, <=, >=, and function calls)
	result, err := v.exprEval.Eval(expr, ctx.ExprEnv())
	if err == nil {
		// Successfully evaluated with expr - convert to boolean
		return helpers.IsTruthy(result), nil
	}

	// If expr evaluation failed and expression starts with !, handle nil negation manually.
	// expr library fails when trying to negate nil (e.g., "!item.primary" where primary key doesn't exist).
	// Workaround: evaluate the inner expression and negate the boolean conversion.
	if strings.HasPrefix(expr, "!") {
		innerExpr := strings.TrimSpace(expr[1:])
		// Try to evaluate inner expression (may return nil)
		innerResult, innerErr := v.exprEval.Eval(innerExpr, ctx.ExprEnv())
		if innerErr == nil {
			// Successfully evaluated - convert nil to bool and negate
			return !helpers.IsTruthy(innerResult), nil
		}
		// Fall back to stack resolution if expr evaluation fails
		val, ok := ctx.stack.Resolve(innerExpr)
		if ok {
			return !helpers.IsTruthy(val), nil
		}
		// Undefined value: !undefined = true
		return true, nil
	}

	// Fall back to legacy behavior for simple variable references
	val, ok := ctx.stack.Resolve(expr)
	if !ok {
		// Variable not found - return false
		return false, nil
	}

	return helpers.IsTruthy(val), nil
}
