package helpers

// IsIdentifierChar checks if a rune is valid in an identifier (letter, digit, or underscore).
func IsIdentifierChar(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_'
}

// IsIdentifier checks if a string is a valid identifier (starts with letter or underscore,
// followed by letters, digits, or underscores).
func IsIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, ch := range s {
		if i == 0 && !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_') {
			return false
		}
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_') {
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
