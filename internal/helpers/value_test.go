package helpers_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego/internal/helpers"
)

func TestIsTruthy(t *testing.T) {
	tests := []struct {
		value    any
		expected bool
	}{
		// Booleans
		{true, true},
		{false, false},

		// Strings
		{"hello", true},
		{"", false},
		{"false", false},
		{"0", true},
		{"true", true},

		// Numbers
		{0, false},
		{1, true},
		{42, true},
		{-1, true},
		{int64(0), false},
		{int64(100), true},
		{0.0, false},
		{1.5, true},
		{-0.5, true},

		// Nil
		{nil, false},

		// Slices, maps, etc (default to true)
		{[]int{}, true},
		{[]int{1}, true},
		{map[string]int{}, true},
		{struct{}{}, true},
	}

	for _, tc := range tests {
		t.Run("", func(t *testing.T) {
			result := helpers.IsTruthy(tc.value)
			require.Equal(t, tc.expected, result)
		})
	}
}
