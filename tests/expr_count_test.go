package tests_test

import (
	"testing"

	"github.com/expr-lang/expr"
	"github.com/stretchr/testify/require"
)

// TestCountBuiltinConflict demonstrates the conflict between expr's count() function
// and a user-defined count variable.
//
// expr-lang provides a count(collection, predicate) function that filters and counts items.
// When a user passes count: 5 as data, expr tries to use the builtin count() function
// instead of the variable, causing a type mismatch error.
func TestCountBuiltinConflict(t *testing.T) {
	// This is what users want to do:
	env := map[string]any{
		"count": 5,
		"items": []int{1, 2, 3, 4, 5},
	}

	// Simple variable lookup works fine
	prog, err := expr.Compile("count", expr.AllowUndefinedVariables())
	require.NoError(t, err)
	result, err := expr.Run(prog, env)
	require.NoError(t, err)
	require.Equal(t, 5, result)

	// But comparison fails because expr treats "count" as the builtin function
	// Error happens at compile time, not runtime
	_, err = expr.Compile("count > 0", expr.AllowUndefinedVariables())
	// This will fail: invalid operation: > (mismatched types func(...) int and int)
	// because expr interprets "count" as the count() function, not the variable
	require.Error(t, err)
	t.Logf("Expected compile error (count is builtin function, not variable): %v", err)

	// expr's count() function works with a collection and predicate
	env2 := map[string]any{
		"items": []int{1, 2, 3, 4, 5},
	}
	prog, err = expr.Compile("count(items, # > 2)", expr.AllowUndefinedVariables())
	require.NoError(t, err)
	result, err = expr.Run(prog, env2)
	require.NoError(t, err)
	require.Equal(t, 3, result) // items > 2: [3, 4, 5]
	t.Logf("expr's count(items, # > 2) = %v", result)
}

// TestCountVariableResolution shows desired behavior:
// If we can't evaluate with expr due to builtin conflict, fall back to stack lookup.
func TestCountVariableResolution(t *testing.T) {
	env := map[string]any{
		"count": 5,
		"items": []int{1, 2, 3, 4, 5},
	}

	// What we want: "count > 0" should evaluate to true using count=5
	// This requires special handling: catch the compile error and check if "count"
	// is a user variable in env, then use that instead.

	expression := "count > 0"
	_, err := expr.Compile(expression, expr.AllowUndefinedVariables())
	// Compilation fails because expr interprets "count" as the builtin function
	require.Error(t, err)
	t.Logf("Compile error (as expected): %v", err)

	// Manual fallback: if expr fails and env["count"] exists, use it
	countValue, exists := env["count"]
	require.True(t, exists)
	require.Equal(t, 5, countValue)
	// Result should be true: 5 > 0
	t.Logf("Fallback: count from env = %v, count > 0 = true", countValue)
}

// TestLenBuiltinConflict demonstrates the same issue with len.
func TestLenBuiltinConflict(t *testing.T) {
	env := map[string]any{
		"len":   10,
		"items": []int{1, 2, 3},
	}

	// len() is also a builtin in expr
	// This will fail similarly at compile time
	_, err1 := expr.Compile("len > 5", expr.AllowUndefinedVariables())
	require.Error(t, err1)
	t.Logf("Expected compile error (len is builtin function, not variable): %v", err1)

	// But len(items) works
	prog2, err2 := expr.Compile("len(items)", expr.AllowUndefinedVariables())
	require.NoError(t, err2)
	result, err2 := expr.Run(prog2, env)
	require.NoError(t, err2)
	require.Equal(t, 3, result)
}
