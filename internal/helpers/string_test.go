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
