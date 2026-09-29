package vuego_test

import (
	"testing"

	"github.com/titpetric/vuego/tests/assert"
)

// chainCase drives the v-if/v-else-if/v-else tables below: one template, one
// set of data, the text that must appear and the text that must not.
type chainCase struct {
	name     string
	data     map[string]any
	contains []string
	absent   []string
}

func runChainCases(t *testing.T, source string, cases []chainCase) {
	t.Helper()

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got := renderFragment(t, source, test.data)

			for _, want := range test.contains {
				assert.Contains(t, got, want)
			}
			for _, unwanted := range test.absent {
				assert.NotContains(t, got, unwanted)
			}
		})
	}
}

func TestEvalElseIfChain_Else(t *testing.T) {
	runChainCases(t, "<div v-if=\"show\">Visible</div>\n<div v-else>Hidden</div>", []chainCase{
		{name: "the condition holds", data: map[string]any{"show": true}, contains: []string{"Visible"}, absent: []string{"Hidden"}},
		{name: "the condition fails", data: map[string]any{"show": false}, contains: []string{"Hidden"}, absent: []string{"Visible"}},
		{name: "the condition names a missing key", data: map[string]any{}, contains: []string{"Hidden"}, absent: []string{"Visible"}},
	})
}

func TestEvalElseIfChain_ElseIf(t *testing.T) {
	const source = `
<div v-if="status == 'active'">Active</div>
<div v-else-if="status == 'pending'">Pending</div>
<div v-else-if="status == 'inactive'">Inactive</div>
<div v-else>Unknown</div>
`

	// Exactly one branch renders, whichever matches first.
	runChainCases(t, source, []chainCase{
		{name: "first branch", data: map[string]any{"status": "active"}, contains: []string{"Active"}, absent: []string{"Pending", "Inactive", "Unknown"}},
		{name: "second branch", data: map[string]any{"status": "pending"}, contains: []string{"Pending"}, absent: []string{"Active", "Inactive", "Unknown"}},
		{name: "third branch", data: map[string]any{"status": "inactive"}, contains: []string{"Inactive"}, absent: []string{"Active", "Pending", "Unknown"}},
		{name: "the else branch", data: map[string]any{"status": "unknown"}, contains: []string{"Unknown"}, absent: []string{"Active", "Pending", "Inactive"}},
	})
}

func TestEvalElseIfChain_WithoutElse(t *testing.T) {
	const source = `
<div v-if="show">Show</div>
<div v-else-if="alternate">Alternate</div>
`

	runChainCases(t, source, []chainCase{
		{name: "the v-if wins", data: map[string]any{"show": true, "alternate": true}, contains: []string{"Show"}, absent: []string{"Alternate"}},
		{name: "the v-else-if wins", data: map[string]any{"show": false, "alternate": true}, contains: []string{"Alternate"}, absent: []string{"Show"}},
		{name: "neither matches and nothing renders", data: map[string]any{"show": false, "alternate": false}, absent: []string{"Show", "Alternate"}},
	})
}

func TestEvalElseIfChain_ComplexConditions(t *testing.T) {
	const source = `
<div v-if="count > 10">Large</div>
<div v-else-if="count > 5">Medium</div>
<div v-else-if="count > 0">Small</div>
<div v-else>None</div>
`

	runChainCases(t, source, []chainCase{
		{name: "count above ten", data: map[string]any{"count": 20}, contains: []string{"Large"}, absent: []string{"Medium", "Small", "None"}},
		{name: "count above five", data: map[string]any{"count": 8}, contains: []string{"Medium"}, absent: []string{"Large", "Small", "None"}},
		{name: "count above zero", data: map[string]any{"count": 2}, contains: []string{"Small"}, absent: []string{"Large", "Medium", "None"}},
		{name: "count at zero", data: map[string]any{"count": 0}, contains: []string{"None"}, absent: []string{"Large", "Medium", "Small"}},
	})
}

func TestEvalElseIfChain_Nested(t *testing.T) {
	const source = `
<div v-if="outer">
  <div v-if="inner">Inner True</div>
  <div v-else>Inner False</div>
</div>
<div v-else>Outer False</div>
`

	runChainCases(t, source, []chainCase{
		{name: "both hold", data: map[string]any{"outer": true, "inner": true}, contains: []string{"Inner True"}, absent: []string{"Inner False", "Outer False"}},
		{name: "the inner fails", data: map[string]any{"outer": true, "inner": false}, contains: []string{"Inner False"}, absent: []string{"Inner True", "Outer False"}},
		{name: "the outer fails", data: map[string]any{"outer": false, "inner": true}, contains: []string{"Outer False"}, absent: []string{"Inner True", "Inner False"}},
	})
}

func TestEvalElseIfChain_AfterVFor(t *testing.T) {
	const source = `
<ul>
<li v-for="item in items">{{ item }}</li>
<li v-else>No items</li>
</ul>
`

	runChainCases(t, source, []chainCase{
		{
			name:     "the loop has items",
			data:     map[string]any{"items": []string{"a", "b", "c"}},
			contains: []string{"<li>a</li>", "<li>b</li>", "<li>c</li>"},
			absent:   []string{"No items"},
		},
		{name: "the loop is empty", data: map[string]any{"items": []string{}}, contains: []string{"No items"}},
		{name: "the collection is missing", data: map[string]any{}, contains: []string{"No items"}},
	})
}

func TestEvalElseIfChain_Standalone(t *testing.T) {
	// A v-else-if or v-else with no v-if before it has no chain to join, so it
	// is dropped rather than rendered.
	tests := []struct {
		name   string
		source string
		data   map[string]any
	}{
		{
			name:   "v-else-if alone",
			source: "<div v-else-if=\"condition\">Dropped</div>\n<div>Kept</div>",
			data:   map[string]any{"condition": true},
		},
		{
			name:   "v-else alone",
			source: "<div v-else>Dropped</div>\n<div>Kept</div>",
			data:   map[string]any{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := renderFragment(t, test.source, test.data)
			assert.NotContains(t, got, "Dropped")
			assert.Contains(t, got, "Kept")
		})
	}
}

func TestEvaluateNodeAsElement_VFor(t *testing.T) {
	runChainCases(t, "<div v-if=\"show\">\n<span v-for=\"item in items\">{{ item }}</span>\n</div>\n<div v-else>No items</div>", []chainCase{
		{
			name:     "the branch renders its loop",
			data:     map[string]any{"show": true, "items": []string{"a", "b", "c"}},
			contains: []string{"<span>a</span>", "<span>b</span>", "<span>c</span>"},
			absent:   []string{"No items"},
		},
		{
			name:   "an empty loop still takes the branch",
			data:   map[string]any{"show": true, "items": []string{}},
			absent: []string{"No items", "<span>"},
		},
	})

	runChainCases(t, "<ul v-if=\"show\">\n<li v-for=\"item in items\" class=\"item\">{{ item.name }}</li>\n</ul>\n<div v-else>No list</div>", []chainCase{
		{
			name:     "loop items keep their attributes",
			data:     map[string]any{"show": true, "items": []map[string]any{{"name": "Alice"}, {"name": "Bob"}}},
			contains: []string{`<li class="item">Alice</li>`, `<li class="item">Bob</li>`},
			absent:   []string{"No list"},
		},
		{
			name:   "an empty loop renders no items",
			data:   map[string]any{"show": true, "items": []map[string]any{}},
			absent: []string{"No list", "<li"},
		},
	})
}

func TestEvaluateNodeAsElement_VHtml(t *testing.T) {
	runChainCases(t, "<div v-if=\"show\" v-html=\"content\"></div>\n<div v-else>No content</div>", []chainCase{
		{
			name:     "the branch inserts the markup",
			data:     map[string]any{"show": true, "content": "<strong>Bold</strong>"},
			contains: []string{"<strong>Bold</strong>"},
			absent:   []string{"No content"},
		},
		{
			name:   "empty content still takes the branch",
			data:   map[string]any{"show": true, "content": ""},
			absent: []string{"No content"},
		},
		{
			name:     "the else branch is taken",
			data:     map[string]any{"show": false, "content": "<strong>Bold</strong>"},
			contains: []string{"No content"},
			absent:   []string{"<strong>Bold</strong>"},
		},
	})

	runChainCases(t, "<div v-if=\"show\">Normal content</div>\n<div v-else v-html=\"fallback\"></div>", []chainCase{
		{
			name:     "v-html on the v-if branch is not evaluated",
			data:     map[string]any{"show": true, "fallback": "<em>Fallback</em>"},
			contains: []string{"Normal content"},
			absent:   []string{"<em>Fallback</em>"},
		},
		{
			name:     "v-html on the v-else branch is evaluated",
			data:     map[string]any{"show": false, "fallback": "<em>Fallback</em>"},
			contains: []string{"<em>Fallback</em>"},
			absent:   []string{"Normal content"},
		},
	})
}

func TestEvaluateNodeAsElement_Attributes(t *testing.T) {
	const source = `
<div v-if="level == 1" id="one">Level One</div>
<div v-else-if="level == 2" class="two">Level Two</div>
<div v-else-if="level == 3" data-test="three">Level Three</div>
<div v-else>Default</div>
`

	// The matched branch keeps its own attributes and the directive is dropped.
	runChainCases(t, source, []chainCase{
		{name: "first branch", data: map[string]any{"level": 1}, contains: []string{`id="one"`, "Level One"}, absent: []string{"Level Two", "v-if"}},
		{name: "second branch", data: map[string]any{"level": 2}, contains: []string{`class="two"`, "Level Two"}, absent: []string{"Level One", "v-else-if"}},
		{name: "third branch", data: map[string]any{"level": 3}, contains: []string{`data-test="three"`, "Level Three"}, absent: []string{"Default", "v-else-if"}},
		{name: "the else branch", data: map[string]any{"level": 99}, contains: []string{"Default"}, absent: []string{"Level One", "v-else"}},
	})
}

func TestEvaluateNodeAsElement_Children(t *testing.T) {
	const source = `
<div v-if="show">
<p>Paragraph 1</p>
<p>Paragraph 2</p>
<p>{{ message }}</p>
</div>
<div v-else>Hidden</div>
`

	runChainCases(t, source, []chainCase{
		{
			name:     "children are evaluated and interpolated",
			data:     map[string]any{"show": true, "message": "Hello World"},
			contains: []string{"<p>Paragraph 1</p>", "<p>Paragraph 2</p>", "<p>Hello World</p>"},
			absent:   []string{"Hidden"},
		},
		{
			name:     "a dropped branch drops its children",
			data:     map[string]any{"show": false, "message": "Ignored"},
			contains: []string{"Hidden"},
			absent:   []string{"Paragraph 1", "Ignored"},
		},
	})
}

func TestEvaluateNodeAsElement_Interpolation(t *testing.T) {
	const source = `
<div v-if="hasName">Hello {{ name }}</div>
<div v-else>Hello Guest</div>
`

	runChainCases(t, source, []chainCase{
		{name: "the v-if branch interpolates", data: map[string]any{"hasName": true, "name": "Alice"}, contains: []string{"Hello Alice"}, absent: []string{"Hello Guest"}},
		{name: "the v-else branch renders its own text", data: map[string]any{"hasName": false, "name": ""}, contains: []string{"Hello Guest"}, absent: []string{"Hello Alice"}},
	})
}

func TestEvaluateNodeAsElement_Template(t *testing.T) {
	// A template tag with a bound attribute sets the name in the current scope
	// and renders its children without the tag.
	t.Run("a bound attribute becomes a scope variable", func(t *testing.T) {
		got := renderFragment(t,
			`<template v-if="show" :label="title"><p>{{ label }}</p></template>`,
			map[string]any{"show": true, "title": "Heading"},
		)
		assert.Contains(t, got, "<p>Heading</p>")
		assert.NotContains(t, got, "<template")
	})

	t.Run("a bound literal becomes a scope variable", func(t *testing.T) {
		got := renderFragment(t,
			`<template v-if="show" :label="'Literal'"><p>{{ label }}</p></template>`,
			map[string]any{"show": true},
		)
		assert.Contains(t, got, "<p>Literal</p>")
	})

	t.Run("a bound name with no value resolves to nothing", func(t *testing.T) {
		got := renderFragment(t,
			`<template v-if="show" :label="missing"><p>[{{ label }}]</p></template>`,
			map[string]any{"show": true},
		)
		assert.Contains(t, got, "[]")
	})
}
