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
func (v *Vue) interpolate(input string) (string, bool) {
	if !strings.Contains(input, "{{") {
		return input, false
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
		val, ok := v.stack.Resolve(expr)
		if ok && val != nil {
			// Escape value for HTML output
			out.WriteString(html.EscapeString(fmt.Sprint(val)))
		} else {
			out.WriteString("")
		}

		last = end
	}

	out.WriteString(input[last:])

	result := out.String()

	return result, result != input
}
