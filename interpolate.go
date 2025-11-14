package vuego

import (
	"bytes"
	"fmt"
	"html"
	"regexp"
	"strings"
)

// {{ expr }} regex (non-greedy, no nested braces)
var interpRe = regexp.MustCompile(`\{\{\s*([^{}\s][^{}]*?)\s*\}\}`)

// interpolate escapes interpolated values for HTML safety.
func (v *Vue) interpolate(ctx VueContext, input string) (string, error) {
	if !strings.Contains(input, "{{") {
		return input, nil
	}

	var out bytes.Buffer
	last := 0
	for _, match := range interpRe.FindAllStringSubmatchIndex(input, -1) {
		start := match[0]
		end := match[1]
		exprStart := match[2]
		exprEnd := match[3]

		out.WriteString(input[last:start])

		expr := strings.TrimSpace(input[exprStart:exprEnd])

		var val any
		var err error

		// Try expr evaluation first for complex expressions
		if strings.Contains(expr, "|") || isFunctionCall(expr) || isComplexExpr(expr) {
			// Try expr first for complex expressions with operators
			if isComplexExpr(expr) {
				val, err = v.exprEval.Eval(expr, getEnvMap(ctx.stack))
				if err == nil {
					// Successfully evaluated, skip pipe fallback
				} else {
					// Fall back to pipe evaluation
					pipe := parsePipeExpr(expr)
					val, err = v.evalPipe(ctx, pipe)
				}
			} else {
				// Use pipe evaluation for filter chains
				pipe := parsePipeExpr(expr)
				val, err = v.evalPipe(ctx, pipe)
			}

			if err != nil {
				return "", fmt.Errorf("in expression '{{ %s }}': %w", expr, err)
			}
		} else {
			// Simple variable reference
			var ok bool
			val, ok = ctx.stack.Resolve(expr)
			if !ok {
				val = nil
			}
		}

		if val != nil {
			// Escape value for HTML output
			out.WriteString(html.EscapeString(fmt.Sprint(val)))
		} else {
			out.WriteString("")
		}

		last = end
	}

	out.WriteString(input[last:])

	return out.String(), nil
}

// isComplexExpr checks if an expression contains operators like ==, ===, !=, <, >, etc.
func isComplexExpr(expr string) bool {
	operators := []string{"==", "===", "!=", "!==", "<=", ">=", "<", ">", "&&", "||"}
	for _, op := range operators {
		if strings.Contains(expr, op) {
			return true
		}
	}
	return false
}
