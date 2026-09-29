package vuego_test

import (
	"testing"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/tests/assert"
)

func TestNewExprEvaluator(t *testing.T) {
	t.Run("returns a usable evaluator", func(t *testing.T) {
		eval := vuego.NewExprEvaluator()
		assert.NotNil(t, eval)

		got, err := eval.Eval("1 + 1", map[string]any{})
		assert.NoError(t, err)
		assert.Equal(t, 2, got)
	})

	t.Run("each evaluator has its own cache", func(t *testing.T) {
		first := vuego.NewExprEvaluator()
		second := vuego.NewExprEvaluator()

		_, err := first.Eval("x", map[string]any{"x": 1})
		assert.NoError(t, err)

		got, err := second.Eval("x", map[string]any{"x": 2})
		assert.NoError(t, err)
		assert.Equal(t, 2, got)
	})
}

func TestExprEvaluator_Eval(t *testing.T) {
	eval := vuego.NewExprEvaluator()

	tests := []struct {
		name string
		expr string
		env  map[string]any
		want any
	}{
		// Literals.
		{name: "integer literal", expr: "42", env: map[string]any{}, want: 42},
		{name: "single-quoted string", expr: "'hello'", env: map[string]any{}, want: "hello"},
		{name: "double-quoted string", expr: `"world"`, env: map[string]any{}, want: "world"},
		{name: "true", expr: "true", env: map[string]any{}, want: true},
		{name: "false", expr: "false", env: map[string]any{}, want: false},

		// Variable and member access.
		{name: "variable", expr: "x", env: map[string]any{"x": 42}, want: 42},
		{name: "map member", expr: "obj.name", env: map[string]any{"obj": map[string]any{"name": "Alice"}}, want: "Alice"},
		{name: "slice index", expr: "items[0]", env: map[string]any{"items": []any{"first", "second"}}, want: "first"},

		// Arithmetic.
		{name: "addition", expr: "x + y", env: map[string]any{"x": 10, "y": 5}, want: 15},
		{name: "multiplication", expr: "x * y", env: map[string]any{"x": 6, "y": 7}, want: 42},

		// Comparison. == is exact equality, the same as === in JavaScript.
		{name: "equal", expr: "x == 5", env: map[string]any{"x": 5}, want: true},
		{name: "not equal", expr: "x != 5", env: map[string]any{"x": 3}, want: true},
		{name: "less than", expr: "x < 10", env: map[string]any{"x": 5}, want: true},
		{name: "greater than", expr: "x > 3", env: map[string]any{"x": 5}, want: true},
		{name: "less or equal", expr: "x <= 5", env: map[string]any{"x": 5}, want: true},
		{name: "greater or equal", expr: "x >= 5", env: map[string]any{"x": 5}, want: true},
		{name: "string equality", expr: "name == 'Alice'", env: map[string]any{"name": "Alice"}, want: true},
		{name: "string inequality", expr: "name != 'Bob'", env: map[string]any{"name": "Alice"}, want: true},
		{name: "strict equality is accepted", expr: "name === 'Alice'", env: map[string]any{"name": "Alice"}, want: true},

		// Logic.
		{name: "and", expr: "x && y", env: map[string]any{"x": true, "y": true}, want: true},
		{name: "and with a false operand", expr: "x && y", env: map[string]any{"x": true, "y": false}, want: false},
		{name: "or", expr: "x || y", env: map[string]any{"x": false, "y": true}, want: true},
		{name: "not", expr: "!x", env: map[string]any{"x": false}, want: true},
		{name: "not of true", expr: "!x", env: map[string]any{"x": true}, want: false},
		{name: "and over comparisons", expr: "x > 5 && y < 20", env: map[string]any{"x": 10, "y": 15}, want: true},
		{name: "or over comparisons", expr: "x > 100 || y < 20", env: map[string]any{"x": 10, "y": 15}, want: true},
		{name: "parenthesised and", expr: "(x > 5) && (y < 10)", env: map[string]any{"x": 7, "y": 8}, want: true},
		{name: "parenthesised or", expr: "(x > 10) || (y < 5)", env: map[string]any{"x": 7, "y": 3}, want: true},

		// Calls.
		{name: "len of a slice", expr: "len(items)", env: map[string]any{"items": []int{1, 2, 3}}, want: 3},
		{name: "len in a comparison", expr: "len(items) > 0", env: map[string]any{"items": []any{1, 2}}, want: true},
		{name: "len of an empty slice", expr: "len(items) == 0", env: map[string]any{"items": []any{}}, want: true},

		// An undefined variable is nil rather than an error, so a missing key
		// can be guarded with a type check.
		{name: "undefined variable", expr: "missing", env: map[string]any{}, want: nil},
		{name: "guarded negation of a missing key", expr: "!(type(x) == 'bool' && x)", env: map[string]any{}, want: true},
		{name: "type of a missing key", expr: "!(type(x) != 'nil')", env: map[string]any{}, want: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := eval.Eval(test.expr, test.env)
			assert.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}

	failures := []struct {
		name string
		expr string
		env  map[string]any
	}{
		{name: "a trailing operator", expr: "x ++", env: map[string]any{"x": 10}},
		{name: "an unclosed parenthesis", expr: "(x > 1", env: map[string]any{"x": 10}},
		{name: "an empty expression", expr: "", env: map[string]any{}},
	}

	for _, test := range failures {
		t.Run(test.name, func(t *testing.T) {
			_, err := eval.Eval(test.expr, test.env)
			assert.Error(t, err)
		})
	}

	t.Run("a repeated expression yields the same result", func(t *testing.T) {
		env := map[string]any{"x": 5, "y": 3}

		first, err := eval.Eval("x + y", env)
		assert.NoError(t, err)

		second, err := eval.Eval("x + y", env)
		assert.NoError(t, err)
		assert.Equal(t, first, second)
	})

	t.Run("a cached expression sees the new env", func(t *testing.T) {
		_, err := eval.Eval("x * 2", map[string]any{"x": 1})
		assert.NoError(t, err)

		got, err := eval.Eval("x * 2", map[string]any{"x": 21})
		assert.NoError(t, err)
		assert.Equal(t, 42, got)
	})
}

func TestExprEvaluator_ClearCache(t *testing.T) {
	t.Run("an expression still evaluates after a clear", func(t *testing.T) {
		eval := vuego.NewExprEvaluator()
		env := map[string]any{"x": 10, "y": 5}

		first, err := eval.Eval("x + y", env)
		assert.NoError(t, err)
		assert.Equal(t, 15, first)

		eval.ClearCache()

		second, err := eval.Eval("x + y", env)
		assert.NoError(t, err)
		assert.Equal(t, 15, second)
	})

	t.Run("clearing an empty cache is allowed", func(t *testing.T) {
		eval := vuego.NewExprEvaluator()
		eval.ClearCache()

		got, err := eval.Eval("1", map[string]any{})
		assert.NoError(t, err)
		assert.Equal(t, 1, got)
	})
}
