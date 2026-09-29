package vuego_test

import (
	"strings"
	"testing"

	"github.com/titpetric/vuego/tests/assert"
)

func TestEvalCondition(t *testing.T) {
	// A v-if condition is evaluated as an expression; the result is read for
	// truthiness.
	truthiness := []struct {
		name    string
		value   any
		visible bool
	}{
		{name: "true", value: true, visible: true},
		{name: "false", value: false},
		{name: "a non-zero number", value: 1, visible: true},
		{name: "zero", value: 0},
		{name: "a non-empty string", value: "hello", visible: true},
		{name: "an empty string", value: ""},
		{name: "nil", value: nil},
		{name: "a non-empty slice", value: []string{"a"}, visible: true},
		// A slice is truthy whatever its length; len() is how a template asks
		// about emptiness.
		{name: "an empty slice", value: []string{}, visible: true},
	}

	for _, test := range truthiness {
		t.Run(test.name+" as a condition", func(t *testing.T) {
			got := renderFragment(t, `<div v-if="value">Visible</div>`, map[string]any{"value": test.value})

			if test.visible {
				assert.Contains(t, got, "Visible")
			} else {
				assert.NotContains(t, got, "Visible")
			}
		})
	}

	t.Run("an undefined variable is false", func(t *testing.T) {
		got := renderFragment(t,
			`<div v-if="undefinedName">Hidden</div><div v-if="defined">Shown</div>`,
			map[string]any{"defined": true},
		)
		assert.NotContains(t, got, "Hidden")
		assert.Contains(t, got, "Shown")
	})

	// A struct is keyed by json tag, so a condition names the tag or the field.
	t.Run("a struct field is a condition", func(t *testing.T) {
		type user struct {
			Active bool   `json:"active"`
			Name   string `json:"name"`
		}

		got := renderFragment(t,
			`<div v-if="active">Active</div><div v-if="name">Named</div>`,
			user{Active: true, Name: "Alice"},
		)
		assert.Contains(t, got, "Active")
		assert.Contains(t, got, "Named")
	})

	t.Run("a v-for scope variable is a condition", func(t *testing.T) {
		got := renderFragment(t,
			"<ul v-for=\"item in items\">\n<li v-if=\"item.name\">Item: {{ item.name }}</li>\n</ul>",
			map[string]any{"items": []map[string]any{{"name": "First"}, {"name": "Second"}, {"name": ""}}},
		)
		assert.Contains(t, got, "Item: First")
		assert.Contains(t, got, "Item: Second")
		// The item with an empty name renders no <li> at all.
		assert.Equal(t, 2, strings.Count(got, "<li"))
	})
}

func TestEvalCondition_Negation(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		data    map[string]any
		visible bool
	}{
		{name: "not false", source: `<div v-if="!flag">Shown</div>`, data: map[string]any{"flag": false}, visible: true},
		{name: "not true", source: `<div v-if="!flag">Shown</div>`, data: map[string]any{"flag": true}},
		{name: "not an empty string", source: `<div v-if="!text">Shown</div>`, data: map[string]any{"text": ""}, visible: true},
		{name: "not a non-empty string", source: `<div v-if="!text">Shown</div>`, data: map[string]any{"text": "x"}},
		{name: "not zero", source: `<div v-if="!count">Shown</div>`, data: map[string]any{"count": 0}, visible: true},
		// A missing key negates to true, which is what makes !item.primary work
		// for items that do not carry the key.
		{name: "not a missing key", source: `<div v-if="!missing">Shown</div>`, data: map[string]any{}, visible: true},
		{name: "not a missing nested key", source: `<div v-if="!item.primary">Shown</div>`, data: map[string]any{"item": map[string]any{}}, visible: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := renderFragment(t, test.source, test.data)

			if test.visible {
				assert.Contains(t, got, "Shown")
			} else {
				assert.NotContains(t, got, "Shown")
			}
		})
	}

	t.Run("negation inside v-for", func(t *testing.T) {
		got := renderFragment(t,
			`<div v-for="item in items"><span v-if="!item.hidden">{{ item.name }}</span></div>`,
			map[string]any{"items": []map[string]any{
				{"name": "shown"},
				{"name": "gone", "hidden": true},
			}},
		)
		assert.Contains(t, got, "shown")
		assert.NotContains(t, got, "gone")
	})
}

func TestEvalCondition_Expression(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		data     any
		contains []string
		absent   []string
	}{
		{
			name:     "string equality",
			source:   `<div v-if="priority == 'high'">High</div><div v-if="priority == 'low'">Low</div>`,
			data:     map[string]any{"priority": "high"},
			contains: []string{"High"},
			absent:   []string{"Low"},
		},
		{
			name:     "numeric comparison",
			source:   `<div v-if="total > 0">Has Items</div><div v-if="total <= 0">No Items</div>`,
			data:     map[string]any{"total": 5},
			contains: []string{"Has Items"},
			absent:   []string{"No Items"},
		},
		{
			name:     "greater or equal",
			source:   `<div v-if="score >= 90">Excellent</div>`,
			data:     map[string]any{"score": 95},
			contains: []string{"Excellent"},
		},
		{
			name:     "logical and",
			source:   `<div v-if="isActive && hasPermission">Allowed</div>`,
			data:     map[string]any{"isActive": true, "hasPermission": true},
			contains: []string{"Allowed"},
		},
		{
			name:   "logical and with a false operand",
			source: `<div v-if="isActive && hasPermission">Allowed</div>`,
			data:   map[string]any{"isActive": true, "hasPermission": false},
			absent: []string{"Allowed"},
		},
		{
			name:     "logical or",
			source:   `<div v-if="isAdmin || isDeveloper">Staff</div>`,
			data:     map[string]any{"isAdmin": true, "isDeveloper": false},
			contains: []string{"Staff"},
		},
		{
			name:     "a comparison combined with and",
			source:   `<div v-if="age >= 18 && verified == true">Verified Adult</div>`,
			data:     map[string]any{"age": 21, "verified": true},
			contains: []string{"Verified Adult"},
		},
		{
			name:     "two comparisons combined with or",
			source:   `<div v-if="role == 'admin' || role == 'moderator'">Staff</div>`,
			data:     map[string]any{"role": "admin"},
			contains: []string{"Staff"},
		},
		{
			name:     "a nested path",
			source:   `<div v-if="user.active">Active</div><div v-if="item.inStock == true">In Stock</div>`,
			data:     map[string]any{"user": map[string]any{"active": true}, "item": map[string]any{"inStock": true}},
			contains: []string{"Active", "In Stock"},
		},
		{
			name:     "a nested path on a struct keyed by json tag",
			source:   `<div v-if="task.priority == 'high'">High</div>`,
			data:     map[string]any{"task": map[string]any{"priority": "high"}},
			contains: []string{"High"},
		},
		{
			name:     "a function call",
			source:   `<div v-if="len(items) > 2">Many</div>`,
			data:     map[string]any{"items": []int{1, 2, 3}},
			contains: []string{"Many"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := renderFragment(t, test.source, test.data)

			for _, want := range test.contains {
				assert.Contains(t, got, want)
			}
			for _, unwanted := range test.absent {
				assert.NotContains(t, got, unwanted)
			}
		})
	}
}

func TestEvalCondition_StrictEquality(t *testing.T) {
	// === and !== are coalesced to == and != before evaluation, because the
	// expression evaluator does not know the strict forms.
	tests := []struct {
		name     string
		source   string
		data     map[string]any
		contains string
		absent   string
	}{
		{
			name:     "=== matches",
			source:   `<div v-if="status === 'active'"><p>Active</p></div><div v-if="status !== 'active'"><p>Inactive</p></div>`,
			data:     map[string]any{"status": "active"},
			contains: "Active",
			absent:   "Inactive",
		},
		{
			name:     "!== matches",
			source:   `<div v-if="status !== 'inactive'"><p>Not Inactive</p></div>`,
			data:     map[string]any{"status": "active"},
			contains: "Not Inactive",
		},
		{
			name:     "=== between two literals",
			source:   `<div v-if="'active' === 'active'"><p>Match</p></div>`,
			data:     map[string]any{},
			contains: "Match",
		},
		{
			name:     "=== in an interpolation",
			source:   `<p>Status: {{ status === 'active' }}</p>`,
			data:     map[string]any{"status": "active"},
			contains: "Status: true",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := renderFragment(t, test.source, test.data)
			assert.Contains(t, got, test.contains)
			if test.absent != "" {
				assert.NotContains(t, got, test.absent)
			}
		})
	}
}

func TestEvalCondition_Unparsable(t *testing.T) {
	// A condition the evaluator cannot compile falls back to resolving the
	// whole text as a variable name, which does not exist, so the element is
	// dropped rather than the render failing.
	tests := []struct {
		name   string
		source string
		data   map[string]any
	}{
		{name: "a pipe in a condition", source: `<div v-if="show | extra">Shown</div>`, data: map[string]any{"show": true}},
		{name: "a trailing pipe", source: `<div v-if="hide |">Shown</div>`, data: map[string]any{"hide": true}},
		{name: "a trailing operator", source: `<div v-if="show ==">Shown</div>`, data: map[string]any{"show": true}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := renderFragmentErr(t, test.source, test.data)
			assert.NoError(t, err)
			assert.NotContains(t, got, "Shown")
		})
	}
}
