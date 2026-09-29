package markdown_test

import (
	"io/fs"
	"testing"

	"github.com/titpetric/vuego/markdown"
	"github.com/titpetric/vuego/tests/assert"
)

func TestTemplates(t *testing.T) {
	t.Run("holds one template per rendered element", func(t *testing.T) {
		entries, err := fs.ReadDir(markdown.Templates(), "markdown")
		assert.NoError(t, err)

		var names []string
		for _, entry := range entries {
			names = append(names, entry.Name())
		}

		assert.Equal(t, []string{
			"autolink.vuego",
			"blockquote.vuego",
			"code_block.vuego",
			"code_span.vuego",
			"emphasis.vuego",
			"footnote.vuego",
			"footnote_backlink.vuego",
			"footnote_link.vuego",
			"footnote_list.vuego",
			"hard_break.vuego",
			"heading.vuego",
			"image.vuego",
			"link.vuego",
			"list.vuego",
			"list_item.vuego",
			"paragraph.vuego",
			"raw_html.vuego",
			"strikethrough.vuego",
			"table.vuego",
			"task_checkbox.vuego",
			"thematic_break.vuego",
		}, names)
	})

	t.Run("nests the templates under markdown/", func(t *testing.T) {
		_, err := fs.Stat(markdown.Templates(), "markdown/paragraph.vuego")
		assert.NoError(t, err)
	})

	t.Run("returns the same filesystem each call", func(t *testing.T) {
		assert.Equal(t, markdown.Templates(), markdown.Templates())
	})
}
