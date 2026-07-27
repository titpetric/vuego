package assert

import (
	"testing"

	"github.com/titpetric/vuego/diff"
)

// EqualHTML parses two HTML strings and returns true if their DOMs match.
// It ignores pure-whitespace text nodes and compares attributes order-insensitively.
// If DOMs don't match, it logs template and data context for debugging.
func EqualHTML(tb testing.TB, want, got, template, data []byte) {
	isEqual := diff.CompareHTML(want, got)
	True(tb, isEqual, "\n--- template:\n%s\n--- json:\n%s\n--- expected:\n%s\n--- actual:\n%s\n", string(template), string(data), string(want), string(got))
}
