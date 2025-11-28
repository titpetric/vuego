package vuego_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego"
)

func TestNewExprEvaluator(t *testing.T) {
	eval := vuego.NewExprEvaluator()
	require.NotNil(t, eval)
}

func TestExprEvaluator_Eval(t *testing.T) {
	eval := vuego.NewExprEvaluator()

	tests := []struct {
		name        string
		expression  string
		env         map[string]any
		expected    any
		expectError bool
	}{
		{
			name:       "simple variable",
			expression: "x",
			env:        map[string]any{"x": 42},
			expected:   42,
		},
		{
			name:       "addition",
			expression: "x + y",
			env:        map[string]any{"x": 10, "y": 5},
			expected:   15,
		},
		{
			name:       "comparison",
			expression: "x > 5",
			env:        map[string]any{"x": 10},
			expected:   true,
		},
		{
			name:       "boolean AND",
			expression: "x > 5 && y < 20",
			env:        map[string]any{"x": 10, "y": 15},
			expected:   true,
		},
		{
			name:       "boolean OR",
			expression: "x > 100 || y < 20",
			env:        map[string]any{"x": 10, "y": 15},
			expected:   true,
		},
		{
			name:       "len function",
			expression: "len(items)",
			env:        map[string]any{"items": []int{1, 2, 3}},
			expected:   3,
		},
		{
			name:       "string literal",
			expression: "'hello'",
			env:        map[string]any{},
			expected:   "hello",
		},
		{
			name:       "true literal",
			expression: "true",
			env:        map[string]any{},
			expected:   true,
		},
		{
			name:       "undefined variable",
			expression: "undefined",
			env:        map[string]any{},
			expected:   nil,
		},
		{
			name:        "invalid expression",
			expression:  "x ++",
			env:         map[string]any{"x": 10},
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := eval.Eval(tc.expression, tc.env)
			if tc.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestExprEvaluator_ClearCache(t *testing.T) {
	eval := vuego.NewExprEvaluator()

	expr := "x + y"
	env := map[string]any{"x": 10, "y": 5}

	result1, err := eval.Eval(expr, env)
	require.NoError(t, err)
	require.Equal(t, 15, result1)

	eval.ClearCache()

	// Should still work after cache clear
	result2, err := eval.Eval(expr, env)
	require.NoError(t, err)
	require.Equal(t, 15, result2)
}
