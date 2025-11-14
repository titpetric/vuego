package vuego

import (
	"bytes"
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/titpetric/vuego/internal/helpers"
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

		// Try unified pipe/expr evaluation (handles both filters and expressions)
		if strings.Contains(expr, "|") || helpers.IsFunctionCall(expr) || helpers.IsComplexExpr(expr) {
			pipe := parsePipeExpr(expr)
			val, err = v.evalPipe(ctx, pipe)
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
