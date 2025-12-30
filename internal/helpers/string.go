package helpers

import (
	"regexp"
	"strings"
)

var spacesRe = regexp.MustCompile(`\s+`)

// IsIdentifierChar reports whether ch is valid in an identifier.
// If first is true, digits are not allowed.
func IsIdentifierChar(ch rune, first bool) bool {
	if ch >= 'a' && ch <= 'z' {
		return true
	}
	if ch >= 'A' && ch <= 'Z' {
		return true
	}
	if ch == '_' {
		return true
	}
	if first {
		return false
	}
	if ch >= '0' && ch <= '9' {
		return true
	}
	return false
}

// IsIdentifier reports whether s is a valid ASCII identifier.
// An identifier must be non-empty, start with a letter (A–Z, a–z) or underscore,
// and may contain only letters, digits, or underscores thereafter.
func IsIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}

	for i, ch := range s {
		if !IsIdentifierChar(ch, i == 0) {
			return false
		}
	}

	return true
}

// NeedsHTMLEscape checks if a string contains characters that need HTML escaping.
// Returns true if the string contains &, <, >, ", or ' characters.
// This avoids calling html.EscapeString which always allocates a new string.
func NeedsHTMLEscape(s string) bool {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '&', '<', '>', '"', '\'':
			return true
		}
	}
	return false
}

// FormatAttr formats an attribute value by trimming whitespace, replacing newlines with spaces,
// and reducing repeated spaces to single spaces.
func FormatAttr(val string) string {
	var b strings.Builder

	// Trim leading and trailing whitespace
	val = strings.TrimSpace(val)

	// Replace newlines with spaces
	val = strings.ReplaceAll(val, "\n", " ")
	val = strings.ReplaceAll(val, "\r", " ")

	// Reduce multiple spaces to single space
	val = spacesRe.ReplaceAllString(val, " ")

	b.WriteString(val)
	return b.String()
}

// CamelToKebab converts a camelCase or PascalCase string to kebab-case.
// For example: "ButtonPrimary" becomes "button-primary", "myVariable" becomes "my-variable".
func CamelToKebab(s string) string {
	if len(s) == 0 {
		return s
	}

	var b strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteRune('-')
		}
		if r >= 'A' && r <= 'Z' {
			b.WriteRune(r + 32) // Convert uppercase to lowercase
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
