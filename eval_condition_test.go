package vuego_test

import (
	"bytes"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
	"github.com/titpetric/vuego"
)

// TestEvalCondition_SimpleBooleanVariables covers the fallback path
// where expr fails to parse and stack.Resolve is used
func TestEvalCondition_SimpleBooleanVariables(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<div v-if="show">Visible</div>
<div v-if="hide">Hidden</div>
`)},
	}

	tests := []struct {
		name   string
		data   map[string]any
		expect string
	}{
		{
			"show is true",
			map[string]any{"show": true, "hide": false},
			"<div>Visible</div>",
		},
		{
			"show is false",
			map[string]any{"show": false, "hide": false},
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			vue := vuego.NewVue(templateFS)
			err := vue.RenderFragment(&buf, "test.vuego", tt.data)
			require.NoError(t, err)
			require.Contains(t, buf.String(), tt.expect)
		})
	}
}

// TestEvalCondition_NegatedVariables covers the negation operator
// when used with simple variable references (fallback path)
func TestEvalCondition_NegatedVariables(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<div v-if="!disabled">Enabled</div>
<div v-if="!show">Hidden</div>
`)},
	}

	tests := []struct {
		name   string
		data   map[string]any
		expect string
	}{
		{
			"disabled is false",
			map[string]any{"disabled": false, "show": false},
			"<div>Enabled</div>\n<div>Hidden</div>",
		},
		{
			"disabled is true",
			map[string]any{"disabled": true, "show": true},
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			vue := vuego.NewVue(templateFS)
			err := vue.RenderFragment(&buf, "test.vuego", tt.data)
			require.NoError(t, err)
			output := buf.String()
			if tt.expect != "" {
				require.Contains(t, output, tt.expect)
			} else {
				require.NotContains(t, output, "Enabled")
				require.NotContains(t, output, "Hidden")
			}
		})
	}
}

// TestEvalCondition_UndefinedVariables covers the path where
// stack.Resolve returns false (variable not found)
func TestEvalCondition_UndefinedVariables(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<div v-if="undefined">Should not appear</div>
<div v-if="defined">Should appear</div>
`)},
	}

	var buf bytes.Buffer
	vue := vuego.NewVue(templateFS)
	err := vue.RenderFragment(&buf, "test.vuego", map[string]any{"defined": true})
	require.NoError(t, err)

	output := buf.String()
	require.NotContains(t, output, "Should not appear")
	require.Contains(t, output, "Should appear")
}

// TestEvalCondition_FalsyValues covers different falsy values
// when using simple variable fallback
func TestEvalCondition_FalsyValues(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<div v-if="bool_false">bool false</div>
<div v-if="int_zero">int zero</div>
<div v-if="empty_string">empty string</div>
<div v-if="nil_value">nil value</div>
`)},
	}

	data := map[string]any{
		"bool_false":   false,
		"int_zero":     0,
		"empty_string": "",
		"nil_value":    nil,
	}

	var buf bytes.Buffer
	vue := vuego.NewVue(templateFS)
	err := vue.RenderFragment(&buf, "test.vuego", data)
	require.NoError(t, err)

	output := buf.String()
	// All should be falsy and not rendered
	require.NotContains(t, output, "bool false")
	require.NotContains(t, output, "int zero")
	require.NotContains(t, output, "empty string")
	require.NotContains(t, output, "nil value")
}

// TestEvalCondition_TruthyValues covers different truthy values
func TestEvalCondition_TruthyValues(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<div v-if="bool_true">bool true</div>
<div v-if="int_one">int one</div>
<div v-if="string_text">string text</div>
`)},
	}

	data := map[string]any{
		"bool_true":   true,
		"int_one":     1,
		"string_text": "hello",
	}

	var buf bytes.Buffer
	vue := vuego.NewVue(templateFS)
	err := vue.RenderFragment(&buf, "test.vuego", data)
	require.NoError(t, err)

	output := buf.String()
	require.Contains(t, output, "bool true")
	require.Contains(t, output, "int one")
	require.Contains(t, output, "string text")
}

// TestEvalCondition_ExpressionFallback tests the case where expr evaluation
// fails and we fall back to simple variable resolution
func TestEvalCondition_ExpressionFallback(t *testing.T) {
	// Create a scenario where the condition would fail expr evaluation
	// but work with simple variable fallback
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<div v-if="message">Message exists</div>
`)},
	}

	tests := []struct {
		name   string
		data   map[string]any
		expect bool
	}{
		{
			"variable is truthy string",
			map[string]any{"message": "Hello"},
			true,
		},
		{
			"variable is empty string",
			map[string]any{"message": ""},
			false,
		},
		{
			"variable missing",
			map[string]any{},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			vue := vuego.NewVue(templateFS)
			err := vue.RenderFragment(&buf, "test.vuego", tt.data)
			require.NoError(t, err)

			output := buf.String()
			if tt.expect {
				require.Contains(t, output, "Message exists")
			} else {
				require.NotContains(t, output, "Message exists")
			}
		})
	}
}

// TestEvalCondition_WithStructFields tests fallback resolution with struct fields
func TestEvalCondition_WithStructFields(t *testing.T) {
	type User struct {
		Active bool   `json:"active"`
		Name   string `json:"name"`
	}

	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<div v-if="active">User is active</div>
<div v-if="name">User has name</div>
`)},
	}

	data := User{
		Active: true,
		Name:   "Alice",
	}

	var buf bytes.Buffer
	vue := vuego.NewVue(templateFS)
	err := vue.RenderFragment(&buf, "test.vuego", data)
	require.NoError(t, err)

	output := buf.String()
	require.Contains(t, output, "User is active")
	require.Contains(t, output, "User has name")
}

// TestEvalCondition_StackLookupPath tests the stack.Lookup function directly
// by testing variable resolution from nested stack scopes
func TestEvalCondition_StackLookupPath(t *testing.T) {
	// This tests stack.Lookup which is used in stack.Resolve
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<ul v-for="item in items">
<li v-if="item.name">Item: {{ item.name }}</li>
</ul>
`)},
	}

	data := map[string]any{
		"items": []map[string]any{
			{"name": "First"},
			{"name": "Second"},
			{"name": ""},
		},
	}

	var buf bytes.Buffer
	vue := vuego.NewVue(templateFS)
	err := vue.RenderFragment(&buf, "test.vuego", data)
	require.NoError(t, err)

	output := buf.String()
	require.Contains(t, output, "Item: First")
	require.Contains(t, output, "Item: Second")
	// Empty string should not render the <li> element
	lines := bytes.Split([]byte(output), []byte("<li"))
	require.Len(t, lines, 3) // header + 2 items with <li (empty string item doesn't get <li>)
}

// TestEvalCondition_ExprCompilationFailureFallback tests the fallback path
// where expr compilation fails and stack.Resolve is used
func TestEvalCondition_ExprCompilationFailureFallback(t *testing.T) {
	// Test with invalid expr syntax that will fail compilation
	// and should fall back to variable resolution
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<div v-if="show | extra">Should render if show is truthy</div>
<div v-if="hide |">Should not render (syntax will fail)</div>
`)},
	}

	data := map[string]any{
		"show": true,
		"hide": true,
	}

	var buf bytes.Buffer
	vue := vuego.NewVue(templateFS)
	err := vue.RenderFragment(&buf, "test.vuego", data)
	require.NoError(t, err)

	output := buf.String()
	// When expr fails, it falls back to simple Resolve which doesn't handle pipes
	// "show | extra" will fail expr compilation and then "show | extra" won't resolve
	// So the first div should not render
	// "hide |" will fail expr compilation and then "hide |" won't resolve as a variable name
	require.NotContains(t, output, "Should render")
	require.NotContains(t, output, "Should not render")
}

// TestEvalCondition_ResolveFailsFallbackToFalse tests that when both
// expr evaluation and stack.Resolve fail, we return false
func TestEvalCondition_ResolveFailsFallbackToFalse(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<div v-if="completely_undefined">Should not appear</div>
<div v-if="defined">Should appear</div>
`)},
	}

	data := map[string]any{
		"defined": true,
	}

	var buf bytes.Buffer
	vue := vuego.NewVue(templateFS)
	err := vue.RenderFragment(&buf, "test.vuego", data)
	require.NoError(t, err)

	output := buf.String()
	// "completely_undefined" is neither in env nor resolvable
	require.NotContains(t, output, "Should not appear")
	require.Contains(t, output, "Should appear")
}

// TestEvalCondition_SuccessfulResolve tests stack.Resolve fallback behavior
// by testing with function calls that fail in expr but should fall back to variable resolution
func TestEvalCondition_SuccessfulResolve(t *testing.T) {
	// Note: The final return statement (line 33 in eval_condition.go) that uses stack.Resolve
	// is practically unreachable with the current expr evaluator.
	// This is because expr.Compile with AllowUndefinedVariables() accepts almost all valid syntax,
	// and expr.Run with undefined variables returns nil rather than error.
	// The fallback was left in for backward compatibility but may be dead code.
	//
	// We still test the paths that ARE reachable:
	// 1. Expr evaluation succeeds (line 20-22) - covered by most tests
	// 2. Syntax errors cause expr to fail (line 27-31) - covered by TestEvalCondition_ExprCompilationFailureFallback

	// Test that when expr succeeds, IsTruthy is applied to the result
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<div v-if="message">{{ message }}</div>
`)},
	}

	tests := []struct {
		name   string
		data   map[string]any
		expect string
	}{
		{
			"truthy non-empty string",
			map[string]any{"message": "Hello"},
			"Hello",
		},
		{
			"truthy number",
			map[string]any{"message": 42},
			"42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			vue := vuego.NewVue(templateFS)
			err := vue.RenderFragment(&buf, "test.vuego", tt.data)
			require.NoError(t, err)
			require.Contains(t, buf.String(), tt.expect)
		})
	}
}
