package diff_test

import (
	"strings"
	"testing"

	"github.com/titpetric/vuego/diff"
	"github.com/titpetric/vuego/tests/assert"
)

func TestFormatToNormalizedHTML(t *testing.T) {
	t.Run("indents nested elements", func(t *testing.T) {
		out, err := diff.FormatToNormalizedHTML([]byte("<div><p>text</p></div>"))
		assert.NoError(t, err)
		assert.Contains(t, out, "<div>\n")
		assert.Contains(t, out, "  <p>\n")
		assert.Contains(t, out, "    text\n")
		assert.Contains(t, out, "  </p>\n")
		assert.Contains(t, out, "</div>\n")
	})

	t.Run("writes attributes with values", func(t *testing.T) {
		out, err := diff.FormatToNormalizedHTML([]byte(`<div class="a" id="x"></div>`))
		assert.NoError(t, err)
		assert.Contains(t, out, `<div class="a" id="x">`)
	})

	t.Run("writes valueless attributes as bare names", func(t *testing.T) {
		out, err := diff.FormatToNormalizedHTML([]byte(`<input disabled>`))
		assert.NoError(t, err)
		assert.Contains(t, out, "<input disabled>")
	})

	t.Run("drops whitespace-only text nodes", func(t *testing.T) {
		out, err := diff.FormatToNormalizedHTML([]byte("<div>\n\n   \n</div>"))
		assert.NoError(t, err)
		assert.Equal(t, "<html>\n  <head>\n  </head>\n  <body>\n    <div>\n    </div>\n  </body>\n</html>\n", out)
	})

	t.Run("keeps comments", func(t *testing.T) {
		out, err := diff.FormatToNormalizedHTML([]byte("<div><!-- note --></div>"))
		assert.NoError(t, err)
		assert.Contains(t, out, "<!-- note -->")
	})

	t.Run("wraps a fragment in the implied document", func(t *testing.T) {
		out, err := diff.FormatToNormalizedHTML([]byte("<p>a</p>"))
		assert.NoError(t, err)
		assert.Equal(t, "<html>\n  <head>\n  </head>\n  <body>\n    <p>\n      a\n    </p>\n  </body>\n</html>\n", out)
	})

	t.Run("empty input yields the implied document", func(t *testing.T) {
		out, err := diff.FormatToNormalizedHTML(nil)
		assert.NoError(t, err)
		assert.Equal(t, "<html>\n  <head>\n  </head>\n  <body>\n  </body>\n</html>\n", out)
	})
}

func TestGenerateUnifiedDiff(t *testing.T) {
	t.Run("writes the file headers", func(t *testing.T) {
		out := diff.GenerateUnifiedDiff("want.html", "got.html", "a", "a")
		assert.True(t, strings.HasPrefix(out, "--- want.html\n+++ got.html\n"), out)
	})

	t.Run("prefixes unchanged lines with two spaces", func(t *testing.T) {
		out := diff.GenerateUnifiedDiff("a", "b", "one\ntwo", "one\ntwo")
		assert.Equal(t, "--- a\n+++ b\n  one\n  two\n", out)
	})

	t.Run("writes a removal and an addition for a changed line", func(t *testing.T) {
		out := diff.GenerateUnifiedDiff("a", "b", "one\ntwo", "one\nTWO")
		assert.Equal(t, "--- a\n+++ b\n  one\n- two\n+ TWO\n", out)
	})

	t.Run("extra lines in the second file are additions", func(t *testing.T) {
		out := diff.GenerateUnifiedDiff("a", "b", "one", "one\ntwo")
		assert.Equal(t, "--- a\n+++ b\n  one\n+ two\n", out)
	})

	t.Run("extra lines in the first file are removals", func(t *testing.T) {
		out := diff.GenerateUnifiedDiff("a", "b", "one\ntwo", "one")
		assert.Equal(t, "--- a\n+++ b\n  one\n- two\n", out)
	})

	t.Run("blank lines are dropped", func(t *testing.T) {
		out := diff.GenerateUnifiedDiff("a", "b", "one\n\ntwo", "one\n\ntwo")
		assert.Equal(t, "--- a\n+++ b\n  one\n  two\n", out)
	})

	t.Run("two empty files yield headers only", func(t *testing.T) {
		out := diff.GenerateUnifiedDiff("a", "b", "", "")
		assert.Equal(t, "--- a\n+++ b\n", out)
	})
}
