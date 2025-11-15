package vuego

import (
	"fmt"
	"html"
	"io"
	"strings"
	"sync"

	"github.com/titpetric/vuego/internal/helpers"
)

// bufferPool reuses strings.Builder instances to reduce allocations.
var bufferPool = sync.Pool{
	New: func() any { return &strings.Builder{} },
}

// interpolateToWriter writes interpolated values to w, escaping for HTML safety.
// This is the core implementation that does not allocate a string result.
// For script and style tags, values are not HTML-escaped.
func (v *Vue) interpolateToWriter(ctx VueContext, w io.Writer, input string) error {
	if !strings.Contains(input, "{{") {
		_, err := io.WriteString(w, input)
		return err
	}

	// Sanity check: {{ should match }}
	if strings.Count(input, "{{") != strings.Count(input, "}}") {
		return fmt.Errorf("mismatched interpolation braces: %d {{ vs %d }}", strings.Count(input, "{{"), strings.Count(input, "}}"))
	}

	last := 0

	for {
		// Find the next {{ occurrence
		start := strings.Index(input[last:], "{{")
		if start < 0 {
			break
		}
		start += last

		// Find the closing }}
		endPos := strings.Index(input[start+2:], "}}")
		if endPos < 0 {
			break
		}
		endPos += start + 2 // Position of first }
		endPos += 2         // Skip past }}

		// Write the text before the interpolation
		if _, err := io.WriteString(w, input[last:start]); err != nil {
			return err
		}

		// Extract the expression between {{ and }}, trim whitespace
		exprStart := start + 2
		exprEnd := endPos - 2

		// Skip leading whitespace
		for exprStart < exprEnd && (input[exprStart] == ' ' || input[exprStart] == '\t' || input[exprStart] == '\n' || input[exprStart] == '\r') {
			exprStart++
		}

		// Skip trailing whitespace
		for exprEnd > exprStart && (input[exprEnd-1] == ' ' || input[exprEnd-1] == '\t' || input[exprEnd-1] == '\n' || input[exprEnd-1] == '\r') {
			exprEnd--
		}

		expr := input[exprStart:exprEnd]

		var val any
		var err error

		// Try unified pipe/expr evaluation (handles both filters and expressions)
		if strings.Contains(expr, "|") || helpers.IsFunctionCall(expr) || helpers.IsComplexExpr(expr) {
			pipe := parsePipeExpr(expr)
			val, err = v.evalPipe(ctx, pipe)
			if err != nil {
				return fmt.Errorf("in expression '{{ %s }}': %w", expr, err)
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
			// Escape value for HTML output (unless in a script/style tag)
			valStr := fmt.Sprint(val)
			parentTag := ctx.CurrentTag()
			// Skip escaping inside script and style tags, since they contain code/CSS, not HTML
			if parentTag == "script" || parentTag == "style" {
				if _, err := io.WriteString(w, valStr); err != nil {
					return err
				}
			} else {
				// Skip escaping if the string doesn't contain special characters
				// (avoids allocation in html.EscapeString for most cases)
				if !helpers.NeedsHTMLEscape(valStr) {
					if _, err := io.WriteString(w, valStr); err != nil {
						return err
					}
				} else {
					if _, err := io.WriteString(w, html.EscapeString(valStr)); err != nil {
						return err
					}
				}
			}
		}

		last = endPos
	}

	// Write remaining text
	if _, err := io.WriteString(w, input[last:]); err != nil {
		return err
	}

	return nil
}

// interpolate escapes interpolated values for HTML safety.
// Uses a buffer pool to minimize allocations.
// For script and style tags, values are not HTML-escaped.
func (v *Vue) interpolate(ctx VueContext, input string) (string, error) {
	buf := bufferPool.Get().(*strings.Builder)
	defer func() {
		buf.Reset()
		bufferPool.Put(buf)
	}()

	// Early return for no-interpolation case
	if !strings.Contains(input, "{{") {
		return input, nil
	}

	// Pre-allocate builder capacity to reduce grow operations
	// Estimate output will be ~120% of input (for escaped HTML entities)
	estimatedLen := len(input) + len(input)/5
	buf.Grow(estimatedLen)

	if err := v.interpolateToWriter(ctx, buf, input); err != nil {
		return "", err
	}

	return buf.String(), nil
}
