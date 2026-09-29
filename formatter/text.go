package formatter

import (
	"strings"
	"unicode"
)

// escapeText escapes HTML-significant characters (&, <, >) in text content.
// Content inside {{ }} template expressions is preserved as-is to avoid
// breaking template syntax like {{ a < b }}.
func escapeText(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	i := 0
	for i < len(s) {
		if i+1 < len(s) && s[i] == '{' && s[i+1] == '{' {
			end := strings.Index(s[i+2:], "}}")
			if end != -1 {
				b.WriteString(s[i : i+2+end+2])
				i += 2 + end + 2
				continue
			}
		}
		switch s[i] {
		case '&':
			b.WriteString("&amp;")
		case '<':
			b.WriteString("&lt;")
		case '>':
			b.WriteString("&gt;")
		default:
			b.WriteByte(s[i])
		}
		i++
	}
	return b.String()
}

// trimRawContent trims leading and trailing blank lines from raw content
// while preserving internal indentation structure.
func trimRawContent(s string) string {
	lines := strings.Split(s, "\n")
	start := 0
	for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	end := len(lines)
	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	if start >= end {
		return ""
	}
	return strings.Join(lines[start:end], "\n")
}

// normalizeInlineText collapses whitespace in inline text while preserving
// boundary spaces needed between inline elements and text.
func normalizeInlineText(s string) string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		// Whitespace-only text between inline elements: preserve as single space
		if len(s) > 0 {
			return " "
		}
		return ""
	}

	// Collapse internal whitespace runs to single spaces
	fields := strings.Fields(trimmed)
	out := strings.Join(fields, " ")

	// Preserve leading space if original had one (boundary between elements)
	runes := []rune(s)
	if len(runes) > 0 && unicode.IsSpace(runes[0]) {
		out = " " + out
	}

	// Preserve trailing space if original had one
	if len(runes) > 0 && unicode.IsSpace(runes[len(runes)-1]) {
		out = out + " "
	}

	return out
}
