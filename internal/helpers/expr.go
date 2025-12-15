package helpers

import "strings"

// IsFunctionCall checks if an expression looks like a function call.
func IsFunctionCall(expr string) bool {
	expr = strings.TrimSpace(expr)
	// Check for pattern: identifier(...)
	for i, ch := range expr {
		if ch == '(' {
			return i > 0
		}
		if !IsIdentifierChar(ch, i == 0) {
			return false
		}
	}
	return false
}

// ContainsPipe checks if an expression contains a pipe operator.
func ContainsPipe(expr string) bool {
	for _, ch := range expr {
		if ch == '|' {
			return true
		}
	}
	return false
}

// IsComplexExpr checks if an expression contains operators like ==, ===, !=, <, >, etc.
func IsComplexExpr(expr string) bool {
	operators := []string{"==", "===", "!=", "!==", "<=", ">=", "<", ">", "&&", "||"}
	for _, op := range operators {
		if strings.Contains(expr, op) {
			return true
		}
	}
	return false
}

// NormalizeComparisonOperators coalesces strict comparison operators (=== and !==) to loose operators (== and !=).
// This is needed because the underlying expr evaluator supports == and != but not === and !==.
func NormalizeComparisonOperators(expr string) string {
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
