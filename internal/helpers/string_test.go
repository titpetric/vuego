package helpers_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego/internal/helpers"
)

// Tests for additional string identifier validation beyond what's in expr_test.go
func TestIsIdentifier_EdgeCases(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"_", true},
		{"_name", true},
		{"name123", true},
		{"_name_123", true},
		{"", false},
		{"1name", false},
		{"-name", false},
		{"name space", false},
		{"name$var", false},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := helpers.IsIdentifier(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestFormatAttr(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Trim whitespace
		{"  value  ", "value"},
		{"\nvalue\n", "value"},
		{"\tvalue\t", "value"},

		// Replace newlines with spaces
		{"line1\nline2", "line1 line2"},
		{"line1\r\nline2", "line1 line2"},

		// Reduce multiple spaces to single space
		{"multiple  spaces", "multiple spaces"},
		{"multiple   spaces", "multiple spaces"},
		{"multiple\t\tspaces", "multiple spaces"},
		{"multiple  \n  spaces", "multiple spaces"},

		// Complex cases
		{"  line1\n  line2  ", "line1 line2"},
		{"\n\n  value  \n\n", "value"},
		{"a  b\nc\t\td", "a b c d"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := helpers.FormatAttr(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}
