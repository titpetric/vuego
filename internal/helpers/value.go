package helpers

import "fmt"

// SliceToAny converts a typed slice to []any.
func SliceToAny[T any](s []T) []any {
	out := make([]any, len(s))
	for i := range s {
		out[i] = s[i]
	}
	return out
}

// IsTruthy converts a value to boolean following Vue semantics.
// For bound attributes, false values should not render the attribute.
func IsTruthy(val any) bool {
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
