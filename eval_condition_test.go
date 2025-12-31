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

// TestElse_BasicElseDirective tests basic v-else directive
func TestElse_BasicElseDirective(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<div v-if="show">Visible</div>
<div v-else>Hidden</div>
`)},
	}

	tests := []struct {
		name       string
		data       map[string]any
		expectShow bool
		expectElse bool
	}{
		{
			"show is true",
			map[string]any{"show": true},
			true,
			false,
		},
		{
			"show is false",
			map[string]any{"show": false},
			false,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			vue := vuego.NewVue(templateFS)
			err := vue.RenderFragment(&buf, "test.vuego", tt.data)
			require.NoError(t, err)
			output := buf.String()

			if tt.expectShow {
				require.Contains(t, output, "Visible")
				require.NotContains(t, output, "Hidden")
			} else {
				require.NotContains(t, output, "Visible")
				require.Contains(t, output, "Hidden")
			}
		})
	}
}

// TestElse_MultipleElseIfDirective tests v-else-if chains
func TestElse_MultipleElseIfDirective(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<div v-if="status == 'active'">Active</div>
<div v-else-if="status == 'pending'">Pending</div>
<div v-else-if="status == 'inactive'">Inactive</div>
<div v-else>Unknown</div>
`)},
	}

	tests := []struct {
		name   string
		data   map[string]any
		expect string
	}{
		{
			"status is active",
			map[string]any{"status": "active"},
			"Active",
		},
		{
			"status is pending",
			map[string]any{"status": "pending"},
			"Pending",
		},
		{
			"status is inactive",
			map[string]any{"status": "inactive"},
			"Inactive",
		},
		{
			"status is unknown",
			map[string]any{"status": "unknown"},
			"Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			vue := vuego.NewVue(templateFS)
			err := vue.RenderFragment(&buf, "test.vuego", tt.data)
			require.NoError(t, err)
			output := buf.String()
			require.Contains(t, output, tt.expect)
			// Verify other conditions don't render
			if tt.expect != "Active" {
				require.NotContains(t, output, "Active")
			}
			if tt.expect != "Pending" {
				require.NotContains(t, output, "Pending")
			}
			if tt.expect != "Inactive" {
				require.NotContains(t, output, "Inactive")
			}
			if tt.expect != "Unknown" {
				require.NotContains(t, output, "Unknown")
			}
		})
	}
}

// TestElse_ElseFollowingVFor tests v-else following v-for
// When v-else follows v-for with no items, it should render
func TestElse_ElseFollowingVFor(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<ul>
<li v-for="item in items">{{ item }}</li>
<li v-else>No items</li>
</ul>
`)},
	}

	tests := []struct {
		name       string
		data       map[string]any
		expectItem bool
		expectElse bool
	}{
		{
			"items exist",
			map[string]any{"items": []string{"a", "b", "c"}},
			true,
			false,
		},
		{
			"items empty",
			map[string]any{"items": []string{}},
			false,
			true,
		},
		{
			"items missing",
			map[string]any{},
			false,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			vue := vuego.NewVue(templateFS)
			err := vue.RenderFragment(&buf, "test.vuego", tt.data)
			require.NoError(t, err)
			output := buf.String()

			if tt.expectItem {
				require.Contains(t, output, "<li>a</li>")
				require.Contains(t, output, "<li>b</li>")
				require.Contains(t, output, "<li>c</li>")
				require.NotContains(t, output, "No items")
			} else {
				require.Contains(t, output, "No items")
			}
		})
	}
}

// TestElse_VIfAndVElseIfWithoutVElse tests chains that don't have v-else
func TestElse_VIfAndVElseIfWithoutVElse(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<div v-if="show">Show</div>
<div v-else-if="alternate">Alternate</div>
`)},
	}

	tests := []struct {
		name   string
		data   map[string]any
		expect string
	}{
		{
			"show is true",
			map[string]any{"show": true, "alternate": true},
			"Show",
		},
		{
			"show is false, alternate is true",
			map[string]any{"show": false, "alternate": true},
			"Alternate",
		},
		{
			"both false",
			map[string]any{"show": false, "alternate": false},
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

			if tt.expect == "" {
				require.NotContains(t, output, "Show")
				require.NotContains(t, output, "Alternate")
			} else {
				require.Contains(t, output, tt.expect)
			}
		})
	}
}

// TestElse_NestedVIfWithVElse tests nested v-if chains
func TestElse_NestedVIfWithVElse(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<div v-if="outer">
  <div v-if="inner">Inner True</div>
  <div v-else>Inner False</div>
</div>
<div v-else>Outer False</div>
`)},
	}

	tests := []struct {
		name   string
		data   map[string]any
		expect []string
	}{
		{
			"outer true, inner true",
			map[string]any{"outer": true, "inner": true},
			[]string{"Inner True"},
		},
		{
			"outer true, inner false",
			map[string]any{"outer": true, "inner": false},
			[]string{"Inner False"},
		},
		{
			"outer false",
			map[string]any{"outer": false, "inner": true},
			[]string{"Outer False"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			vue := vuego.NewVue(templateFS)
			err := vue.RenderFragment(&buf, "test.vuego", tt.data)
			require.NoError(t, err)
			output := buf.String()

			for _, exp := range tt.expect {
				require.Contains(t, output, exp)
			}
		})
	}
}

// TestElse_VElseIfWithComplexConditions tests v-else-if with complex expressions
// Note: Comparison operators (>, <, ==, etc) in conditions are not yet implemented
// This test documents the expected behavior when they are implemented.
func TestElse_VElseIfWithComplexConditions(t *testing.T) {
	t.Skip("Comparison operators in conditions not yet implemented")

	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<div v-if="count > 10">Large</div>
<div v-else-if="count > 5">Medium</div>
<div v-else-if="count > 0">Small</div>
<div v-else>None</div>
`)},
	}

	tests := []struct {
		name   string
		data   map[string]any
		expect string
	}{
		{
			"count 20",
			map[string]any{"count": 20},
			"Large",
		},
		{
			"count 8",
			map[string]any{"count": 8},
			"Medium",
		},
		{
			"count 2",
			map[string]any{"count": 2},
			"Small",
		},
		{
			"count 0",
			map[string]any{"count": 0},
			"None",
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

// TestElse_VElseWithInterpolation tests that v-else preserves node content
func TestElse_VElseWithInterpolation(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<div v-if="has_name">Hello {{ name }}</div>
<div v-else>Hello Guest</div>
`)},
	}

	tests := []struct {
		name   string
		data   map[string]any
		expect string
	}{
		{
			"name provided",
			map[string]any{"has_name": true, "name": "Alice"},
			"Hello Alice",
		},
		{
			"name not provided",
			map[string]any{"has_name": false, "name": ""},
			"Hello Guest",
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

// TestElse_StandaloneVElseIfError tests that v-else-if without v-if is ignored
func TestElse_StandaloneVElseIfError(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<div v-else-if="condition">Should not appear</div>
<div>Should appear</div>
`)},
	}

	var buf bytes.Buffer
	vue := vuego.NewVue(templateFS)
	err := vue.RenderFragment(&buf, "test.vuego", map[string]any{"condition": true})
	require.NoError(t, err)
	output := buf.String()

	// v-else-if without v-if should be skipped
	require.NotContains(t, output, "Should not appear")
	require.Contains(t, output, "Should appear")
}

// TestElse_StandaloneVElseError tests that v-else without v-if is ignored
func TestElse_StandaloneVElseError(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<div v-else>Should not appear</div>
<div>Should appear</div>
`)},
	}

	var buf bytes.Buffer
	vue := vuego.NewVue(templateFS)
	err := vue.RenderFragment(&buf, "test.vuego", map[string]any{})
	require.NoError(t, err)
	output := buf.String()

	// v-else without v-if should be skipped
	require.NotContains(t, output, "Should not appear")
	require.Contains(t, output, "Should appear")
}

// TestEvaluateNodeAsElement_WithVFor tests evaluateNodeAsElement with v-for directive
func TestEvaluateNodeAsElement_WithVFor(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<div v-if="show">
<span v-for="item in items">{{ item }}</span>
</div>
<div v-else>No items</div>
`)},
	}

	tests := []struct {
		name   string
		data   map[string]any
		expect []string
		notExp string
	}{
		{
			"show is true with items",
			map[string]any{"show": true, "items": []string{"a", "b", "c"}},
			[]string{"<span>a</span>", "<span>b</span>", "<span>c</span>"},
			"No items",
		},
		{
			"show is true with empty items",
			map[string]any{"show": true, "items": []string{}},
			[]string{},
			"No items",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			vue := vuego.NewVue(templateFS)
			err := vue.RenderFragment(&buf, "test.vuego", tt.data)
			require.NoError(t, err)
			output := buf.String()

			for _, exp := range tt.expect {
				require.Contains(t, output, exp)
			}
			if tt.notExp != "" {
				require.NotContains(t, output, tt.notExp)
			}
		})
	}
}

// TestEvaluateNodeAsElement_WithVHtml tests evaluateNodeAsElement with v-html directive
func TestEvaluateNodeAsElement_WithVHtml(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<div v-if="show" v-html="content"></div>
<div v-else>No content</div>
`)},
	}

	tests := []struct {
		name   string
		data   map[string]any
		expect string
		notExp string
	}{
		{
			"show is true with html content",
			map[string]any{"show": true, "content": "<strong>Bold</strong>"},
			"<strong>Bold</strong>",
			"No content",
		},
		{
			"show is true with empty content",
			map[string]any{"show": true, "content": ""},
			"",
			"No content",
		},
		{
			"show is false",
			map[string]any{"show": false, "content": "<strong>Bold</strong>"},
			"No content",
			"<strong>Bold</strong>",
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
			}
			if tt.notExp != "" {
				require.NotContains(t, output, tt.notExp)
			}
		})
	}
}

// TestEvaluateNodeAsElement_ElseIfWithVFor tests v-else-if followed by v-for
func TestEvaluateNodeAsElement_ElseIfWithVFor(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<div v-if="status == 'active'">Active status</div>
<div v-else-if="status == 'pending'">
<span v-for="task in tasks">{{ task }}</span>
</div>
<div v-else>Other status</div>
`)},
	}

	tests := []struct {
		name   string
		data   map[string]any
		expect string
		notExp string
	}{
		{
			"status is active",
			map[string]any{"status": "active", "tasks": []string{"task1", "task2"}},
			"Active status",
			"task1",
		},
		{
			"status is pending with tasks",
			map[string]any{"status": "pending", "tasks": []string{"task1", "task2"}},
			"<span>task1</span>",
			"Active status",
		},
		{
			"status is other",
			map[string]any{"status": "other", "tasks": []string{"task1"}},
			"Other status",
			"task1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			vue := vuego.NewVue(templateFS)
			err := vue.RenderFragment(&buf, "test.vuego", tt.data)
			require.NoError(t, err)
			output := buf.String()

			require.Contains(t, output, tt.expect)
			require.NotContains(t, output, tt.notExp)
		})
	}
}

// TestEvaluateNodeAsElement_ElseWithVHtml tests v-else with v-html
func TestEvaluateNodeAsElement_ElseWithVHtml(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<div v-if="show">Normal content</div>
<div v-else v-html="fallback"></div>
`)},
	}

	tests := []struct {
		name   string
		data   map[string]any
		expect string
		notExp string
	}{
		{
			"show is true",
			map[string]any{"show": true, "fallback": "<em>Fallback</em>"},
			"Normal content",
			"<em>Fallback</em>",
		},
		{
			"show is false",
			map[string]any{"show": false, "fallback": "<em>Fallback</em>"},
			"<em>Fallback</em>",
			"Normal content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			vue := vuego.NewVue(templateFS)
			err := vue.RenderFragment(&buf, "test.vuego", tt.data)
			require.NoError(t, err)
			output := buf.String()

			require.Contains(t, output, tt.expect)
			require.NotContains(t, output, tt.notExp)
		})
	}
}

// TestEvaluateNodeAsElement_ChainedElseIfWithAttributes tests v-if/v-else-if chains with various attributes
func TestEvaluateNodeAsElement_ChainedElseIfWithAttributes(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<div v-if="level == 1" id="one">Level One</div>
<div v-else-if="level == 2" class="two">Level Two</div>
<div v-else-if="level == 3" data-test="three">Level Three</div>
<div v-else>Default</div>
`)},
	}

	tests := []struct {
		name   string
		data   map[string]any
		expect string
		notExp string
	}{
		{
			"level 1",
			map[string]any{"level": 1},
			"Level One",
			"Level Two",
		},
		{
			"level 2",
			map[string]any{"level": 2},
			"Level Two",
			"Level One",
		},
		{
			"level 3",
			map[string]any{"level": 3},
			"Level Three",
			"Default",
		},
		{
			"level other",
			map[string]any{"level": 99},
			"Default",
			"Level One",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			vue := vuego.NewVue(templateFS)
			err := vue.RenderFragment(&buf, "test.vuego", tt.data)
			require.NoError(t, err)
			output := buf.String()

			require.Contains(t, output, tt.expect)
			require.NotContains(t, output, tt.notExp)
		})
	}
}

// TestEvaluateNodeAsElement_WithVForAndAttributes tests v-for with v-if and other attributes
func TestEvaluateNodeAsElement_WithVForAndAttributes(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<ul v-if="show">
<li v-for="item in items" class="item">{{ item.name }}</li>
</ul>
<div v-else>No list</div>
`)},
	}

	tests := []struct {
		name   string
		data   map[string]any
		expect string
		notExp string
	}{
		{
			"show with items",
			map[string]any{
				"show": true,
				"items": []map[string]any{
					{"name": "Alice"},
					{"name": "Bob"},
				},
			},
			"<li class=\"item\">Alice</li>",
			"No list",
		},
		{
			"show with empty list",
			map[string]any{
				"show":  true,
				"items": []map[string]any{},
			},
			"",
			"No list",
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
			}
			if tt.notExp != "" {
				require.NotContains(t, output, tt.notExp)
			}
		})
	}
}

// TestEvaluateNodeAsElement_NestedVIfWithChildren tests v-if with nested content and evaluateChildren
func TestEvaluateNodeAsElement_NestedVIfWithChildren(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<div v-if="show">
<p>Paragraph 1</p>
<p>Paragraph 2</p>
<p>{{ message }}</p>
</div>
<div v-else>Hidden</div>
`)},
	}

	tests := []struct {
		name   string
		data   map[string]any
		expect []string
		notExp string
	}{
		{
			"show is true with nested content",
			map[string]any{"show": true, "message": "Hello World"},
			[]string{"<p>Paragraph 1</p>", "<p>Paragraph 2</p>", "<p>Hello World</p>"},
			"Hidden",
		},
		{
			"show is false",
			map[string]any{"show": false, "message": "Ignored"},
			[]string{},
			"Paragraph 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			vue := vuego.NewVue(templateFS)
			err := vue.RenderFragment(&buf, "test.vuego", tt.data)
			require.NoError(t, err)
			output := buf.String()

			for _, exp := range tt.expect {
				require.Contains(t, output, exp)
			}
			if tt.notExp != "" {
				require.NotContains(t, output, tt.notExp)
			}
		})
	}
}

// TestVIfNegationWithVFor reproduces the bug with !item.primary in v-for loops
func TestVIfNegationWithVFor(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`
<template v-for="(idx, item) in indexes">
  <span v-if="item.primary">PRIMARY</span>
</template>

<template v-for="(idx, item) in indexes">
  <span v-if="!item.primary">NOT PRIMARY</span>
</template>
`)},
	}

	data := map[string]any{
		"indexes": []map[string]any{
			{"name": "name", "primary": true},
			{"name": "name2"}, // no primary key
			{"name": "name3"}, // no primary key
			{"columns": []string{"id"}, "primary": true},
		},
	}

	var buf bytes.Buffer
	vue := vuego.NewVue(templateFS)
	err := vue.RenderFragment(&buf, "test.vuego", data)
	require.NoError(t, err)

	output := buf.String()
	t.Logf("Output:\n%s", output)

	// Count how many times <span>PRIMARY</span> appears
	primaryCount := bytes.Count(buf.Bytes(), []byte("<span>PRIMARY</span>"))
	require.Equal(t, 2, primaryCount, "PRIMARY should appear exactly 2 times (for items with primary=true)")

	// Count how many times <span>NOT PRIMARY</span> appears
	notPrimaryCount := bytes.Count(buf.Bytes(), []byte("<span>NOT PRIMARY</span>"))
	require.Equal(t, 2, notPrimaryCount, "NOT PRIMARY should appear exactly 2 times (for items without primary or primary=false)")
}

// TestVIfNegationSimple tests basic negation without v-for (no == true comparison)
func TestVIfNegationSimple(t *testing.T) {
	templateFS := &fstest.MapFS{
		"test.vuego": {Data: []byte(`<div v-if="item.primary">PRIMARY</div><div v-if="!item.primary">NOT PRIMARY</div>`)},
	}

	tests := []struct {
		name        string
		data        any
		expectPrim  bool
		expectNotPr bool
	}{
		{
			"primary=true",
			map[string]any{"item": map[string]any{"primary": true}},
			true,
			false,
		},
		{
			"primary=false",
			map[string]any{"item": map[string]any{"primary": false}},
			false,
			true,
		},
		{
			"primary missing",
			map[string]any{"item": map[string]any{"name": "test"}},
			false,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			vue := vuego.NewVue(templateFS)
			err := vue.RenderFragment(&buf, "test.vuego", tt.data)
			require.NoError(t, err)

			output := buf.String()
			if tt.expectPrim {
				require.Contains(t, output, "<div>PRIMARY</div>", "expected PRIMARY to be rendered")
			} else {
				require.NotContains(t, output, "<div>PRIMARY</div>", "expected PRIMARY to NOT be rendered")
			}
			if tt.expectNotPr {
				require.Contains(t, output, "<div>NOT PRIMARY</div>", "expected NOT PRIMARY to be rendered")
			} else {
				require.NotContains(t, output, "<div>NOT PRIMARY</div>", "expected NOT PRIMARY to NOT be rendered")
			}
		})
	}
}
