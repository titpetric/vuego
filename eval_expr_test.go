package vuego_test

import (
	"bytes"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
	"github.com/titpetric/vuego"
)

func TestExprEvaluator_SimpleExpressions(t *testing.T) {
	evaluator := vuego.NewExprEvaluator()

	tests := []struct {
		name   string
		expr   string
		env    map[string]any
		expect any
	}{
		// Literal values
		{"integer literal", "42", map[string]any{}, 42},
		{"string literal", "'hello'", map[string]any{}, "hello"},
		{"double-quoted string", `"world"`, map[string]any{}, "world"},
		{"boolean true", "true", map[string]any{}, true},
		{"boolean false", "false", map[string]any{}, false},

		// Variable references
		{"simple variable", "x", map[string]any{"x": 42}, 42},
		{"nested property", "obj.name", map[string]any{"obj": map[string]any{"name": "Alice"}}, "Alice"},
		{"array index", "items[0]", map[string]any{"items": []any{"first", "second"}}, "first"},

		// Comparison operators
		// Note: expr uses == for equality (equivalent to === in JavaScript)
		{"equality ==", "x == 5", map[string]any{"x": 5}, true},
		{"inequality !=", "x != 5", map[string]any{"x": 3}, true},
		{"less than <", "x < 10", map[string]any{"x": 5}, true},
		{"greater than >", "x > 3", map[string]any{"x": 5}, true},
		{"less than or equal <=", "x <= 5", map[string]any{"x": 5}, true},
		{"greater than or equal >=", "x >= 5", map[string]any{"x": 5}, true},

		// Logical operators
		{"logical AND", "x && y", map[string]any{"x": true, "y": true}, true},
		{"logical AND false", "x && y", map[string]any{"x": true, "y": false}, false},
		{"logical OR", "x || y", map[string]any{"x": false, "y": true}, true},
		{"logical NOT", "!x", map[string]any{"x": false}, true},
		{"logical NOT true", "!x", map[string]any{"x": true}, false},

		// String comparisons
		{"string equality", "name == 'Alice'", map[string]any{"name": "Alice"}, true},
		{"string inequality", "name != 'Bob'", map[string]any{"name": "Alice"}, true},

		// Complex expressions
		{"compound AND", "(x > 5) && (y < 10)", map[string]any{"x": 7, "y": 8}, true},
		{"compound OR", "(x > 10) || (y < 5)", map[string]any{"x": 7, "y": 3}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.Eval(tt.expr, tt.env)
			require.NoError(t, err)
			require.Equal(t, tt.expect, result)
		})
	}
}

func TestExprEvaluator_FunctionCalls(t *testing.T) {
	evaluator := vuego.NewExprEvaluator()

	tests := []struct {
		name      string
		expr      string
		env       map[string]any
		expectErr bool
	}{
		// Built-in functions in expr
		{"len function", "len(items) > 0", map[string]any{"items": []any{1, 2}}, false},
		{"empty slice", "len(items) == 0", map[string]any{"items": []any{}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := evaluator.Eval(tt.expr, tt.env)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestExprEvaluator_UndefinedVariables(t *testing.T) {
	evaluator := vuego.NewExprEvaluator()

	// expr allows undefined variables with AllowUndefinedVariables()
	// They evaluate to nil
	result, err := evaluator.Eval("missing", map[string]any{})
	require.NoError(t, err)
	require.Nil(t, result)
}

func TestExprEvaluator_Caching(t *testing.T) {
	evaluator := vuego.NewExprEvaluator()

	expr := "x + y"
	env := map[string]any{"x": 5, "y": 3}

	// First evaluation
	result1, err1 := evaluator.Eval(expr, env)
	require.NoError(t, err1)

	// Second evaluation should use cache
	result2, err2 := evaluator.Eval(expr, env)
	require.NoError(t, err2)

	require.Equal(t, result1, result2)
}

func TestVue_VIfWithExprComparison(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<div v-if="priority == 'high'">High Priority</div>
<div v-if="priority == 'low'">Low Priority</div>
<div v-if="total > 0">Has Items</div>
<div v-if="total <= 0">No Items</div>
`)},
	}

	tests := []struct {
		name   string
		data   map[string]any
		expect string
	}{
		{
			"priority high",
			map[string]any{"priority": "high", "total": 5},
			"<div>High Priority</div>\n<div>Has Items</div>\n",
		},
		{
			"priority low",
			map[string]any{"priority": "low", "total": 0},
			"<div>Low Priority</div>\n<div>No Items</div>\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			vue := vuego.NewVue(templateFS)
			err := vue.RenderFragment(&buf, "test.vuego", tt.data)
			require.NoError(t, err)
			require.Equal(t, tt.expect, buf.String())
		})
	}
}

func TestVue_VIfWithLogicalOperators(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<div v-if="isActive && hasPermission">Active with Permission</div>
<div v-if="isAdmin || isDeveloper">Admin or Developer</div>
`)},
	}

	tests := []struct {
		name   string
		data   map[string]any
		expect string
	}{
		{
			"both true",
			map[string]any{"isActive": true, "hasPermission": true, "isAdmin": false, "isDeveloper": false},
			"<div>Active with Permission</div>\n",
		},
		{
			"admin role",
			map[string]any{"isActive": false, "hasPermission": false, "isAdmin": true, "isDeveloper": false},
			"<div>Admin or Developer</div>\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			vue := vuego.NewVue(templateFS)
			err := vue.RenderFragment(&buf, "test.vuego", tt.data)
			require.NoError(t, err)
			require.Equal(t, tt.expect, buf.String())
		})
	}
}

func TestVue_VIfWithNestedProperties(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<div v-if="user.active">User is Active</div>
<div v-if="item.inStock == true">In Stock</div>
`)},
	}

	data := map[string]any{
		"user": map[string]any{"active": true},
		"item": map[string]any{"inStock": true},
	}

	var buf bytes.Buffer
	vue := vuego.NewVue(templateFS)
	err := vue.RenderFragment(&buf, "test.vuego", data)
	require.NoError(t, err)

	expected := "<div>User is Active</div>\n<div>In Stock</div>\n"
	require.Equal(t, expected, buf.String())
}

func TestVue_InterpolationWithComparison(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<p>Status: {{ status == 'active' }}</p>
`)},
	}

	tests := []struct {
		name   string
		status string
		expect string
	}{
		{"active status", "active", "<p>Status: true</p>\n"},
		{"inactive status", "inactive", "<p>Status: false</p>\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			vue := vuego.NewVue(templateFS)
			err := vue.RenderFragment(&buf, "test.vuego", map[string]any{"status": tt.status})
			require.NoError(t, err)
			require.Equal(t, tt.expect, buf.String())
		})
	}
}

func TestVue_BackwardCompatibility(t *testing.T) {
	// Ensure existing syntax still works
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<div v-if="show">Visible</div>
<div v-if="!hide">Not Hidden</div>
<p>{{ message }}</p>
<p>{{ user.name }}</p>
`)},
	}

	data := map[string]any{
		"show": true,
		"hide": false,
		"message": "Hello",
		"user": map[string]any{"name": "Alice"},
	}

	var buf bytes.Buffer
	vue := vuego.NewVue(templateFS)
	err := vue.RenderFragment(&buf, "test.vuego", data)
	require.NoError(t, err)

	expected := "<div>Visible</div>\n<div>Not Hidden</div>\n<p>Hello</p>\n<p>Alice</p>\n"
	require.Equal(t, expected, buf.String())
}

// Test the exact examples from the original request

func TestRequestExample_TaskPriority(t *testing.T) {
	// Test: v-if="task.priority == 'high'" as requested
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<div v-if="task.priority == 'high'">High Priority Task</div>
<div v-if="task.priority == 'medium'">Medium Priority Task</div>
<div v-if="task.priority == 'low'">Low Priority Task</div>
`)},
	}

	tests := []struct {
		name     string
		priority string
		expect   string
	}{
		{"high priority", "high", "High Priority Task"},
		{"medium priority", "medium", "Medium Priority Task"},
		{"low priority", "low", "Low Priority Task"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			vue := vuego.NewVue(templateFS)
			err := vue.RenderFragment(&buf, "test.vuego", map[string]any{
				"task": map[string]any{"priority": tt.priority},
			})
			require.NoError(t, err)
			require.Contains(t, buf.String(), tt.expect)
		})
	}
}

func TestRequestExample_BooleanOperations(t *testing.T) {
	// Test various boolean operators as requested
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<div v-if="score >= 90">Excellent</div>
<div v-if="age >= 18 && verified == true">Verified Adult</div>
<div v-if="role == 'admin' || role == 'moderator'">Staff</div>
`)},
	}

	var buf bytes.Buffer
	vue := vuego.NewVue(templateFS)
	err := vue.RenderFragment(&buf, "test.vuego", map[string]any{
		"score": 95,
		"age": 21,
		"verified": true,
		"role": "admin",
	})
	require.NoError(t, err)

	require.Contains(t, buf.String(), "Excellent")
	require.Contains(t, buf.String(), "Verified Adult")
	require.Contains(t, buf.String(), "Staff")
}

func TestRequestExample_StackCompatibility(t *testing.T) {
	// Test that stack map[string]any is compatible with expr
	evaluator := vuego.NewExprEvaluator()

	env := map[string]any{
		"item": map[string]any{
			"title": "Hello",
			"id": 42,
		},
		"items": []any{
			map[string]any{"name": "first"},
			map[string]any{"name": "second"},
		},
	}

	// Test property access like "item.title" as mentioned in request
	result, err := evaluator.Eval("item.title", env)
	require.NoError(t, err)
	require.Equal(t, "Hello", result)

	// Test comparison "item.id == 42"
	result, err = evaluator.Eval("item.id == 42", env)
	require.NoError(t, err)
	require.Equal(t, true, result)

	// Test nested access "items[0].name"
	result, err = evaluator.Eval("items[0].name", env)
	require.NoError(t, err)
	require.Equal(t, "first", result)
}
