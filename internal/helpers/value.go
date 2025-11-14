package helpers

import "fmt"

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
