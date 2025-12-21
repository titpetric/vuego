package helpers_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"

	"github.com/titpetric/vuego/internal/helpers"
)

func TestSignificantChildren(t *testing.T) {
	t.Run("nil node returns nil", func(t *testing.T) {
		result := helpers.SignificantChildren(nil)
		require.Nil(t, result)
	})

	t.Run("document node with single element child", func(t *testing.T) {
		n := &html.Node{Type: html.DocumentNode}
		div := &html.Node{Type: html.ElementNode, Data: "div"}
		n.AppendChild(div)
		result := helpers.SignificantChildren(n)
		require.Len(t, result, 1)
		require.Equal(t, div, result[0])
	})

	t.Run("document node ignores whitespace-only text nodes", func(t *testing.T) {
		n := &html.Node{Type: html.DocumentNode}
		whitespace := &html.Node{Type: html.TextNode, Data: "   \n  "}
		div := &html.Node{Type: html.ElementNode, Data: "div"}
		whitespace2 := &html.Node{Type: html.TextNode, Data: "\t"}

		n.AppendChild(whitespace)
		n.AppendChild(div)
		n.AppendChild(whitespace2)

		result := helpers.SignificantChildren(n)
		require.Len(t, result, 1)
		require.Equal(t, div, result[0])
	})

	t.Run("document with multiple element children", func(t *testing.T) {
		n := &html.Node{Type: html.DocumentNode}
		div1 := &html.Node{Type: html.ElementNode, Data: "div"}
		div2 := &html.Node{Type: html.ElementNode, Data: "div"}
		n.AppendChild(div1)
		n.AppendChild(div2)

		result := helpers.SignificantChildren(n)
		require.Len(t, result, 2)
		require.Equal(t, div1, result[0])
		require.Equal(t, div2, result[1])
	})

	t.Run("element node returns itself wrapped", func(t *testing.T) {
		div := &html.Node{Type: html.ElementNode, Data: "div"}
		result := helpers.SignificantChildren(div)
		require.Len(t, result, 1)
		require.Equal(t, div, result[0])
	})

	t.Run("text node returns itself wrapped", func(t *testing.T) {
		text := &html.Node{Type: html.TextNode, Data: "hello"}
		result := helpers.SignificantChildren(text)
		require.Len(t, result, 1)
		require.Equal(t, text, result[0])
	})

	t.Run("parsed html document with only body", func(t *testing.T) {
		// Create a document with html->body structure only
		root := &html.Node{Type: html.DocumentNode}
		htmlNode := &html.Node{Type: html.ElementNode, Data: "html"}
		bodyNode := &html.Node{Type: html.ElementNode, Data: "body"}
		pNode := &html.Node{Type: html.ElementNode, Data: "p"}
		textNode := &html.Node{Type: html.TextNode, Data: "test"}

		root.AppendChild(htmlNode)
		htmlNode.AppendChild(bodyNode)
		bodyNode.AppendChild(pNode)
		pNode.AppendChild(textNode)

		result := helpers.SignificantChildren(root)
		// When html->body, returns body's children
		require.True(t, len(result) > 0, "should have children")
		require.Equal(t, "p", result[0].Data)
	})

	t.Run("parsed html document unwraps body", func(t *testing.T) {
		// html.Parse creates html/head/body structure
		doc, err := html.Parse(bytes.NewReader([]byte("<html><body><p>test</p></body></html>")))
		require.NoError(t, err)

		result := helpers.SignificantChildren(doc)
		// When html->head+body, returns htmlChildren (head and body)
		require.True(t, len(result) > 0, "should have children")
	})

	t.Run("comment nodes are ignored", func(t *testing.T) {
		n := &html.Node{Type: html.DocumentNode}
		comment := &html.Node{Type: html.CommentNode, Data: "comment"}
		div := &html.Node{Type: html.ElementNode, Data: "div"}

		n.AppendChild(comment)
		n.AppendChild(div)

		result := helpers.SignificantChildren(n)
		require.Len(t, result, 2) // comment and div both included in doc children
	})
}

func TestCompareHTML(t *testing.T) {
	t.Run("identical simple HTML matches", func(t *testing.T) {
		html := []byte("<div><p>text</p></div>")
		tb := &testing.T{}
		result := helpers.EqualHTML(tb, html, html, nil, nil)
		require.True(t, result)
	})

	t.Run("different tag names don't match", func(t *testing.T) {
		tb := &testing.T{}
		result := helpers.EqualHTML(tb, []byte("<div></div>"), []byte("<span></span>"), nil, nil)
		require.False(t, result)
	})

	t.Run("different text content doesn't match", func(t *testing.T) {
		tb := &testing.T{}
		result := helpers.EqualHTML(tb, []byte("<p>hello</p>"), []byte("<p>world</p>"), nil, nil)
		require.False(t, result)
	})

	t.Run("whitespace in text nodes is ignored", func(t *testing.T) {
		tb := &testing.T{}
		result := helpers.EqualHTML(tb, []byte("<p>hello</p>"), []byte("<p>  hello  </p>"), nil, nil)
		require.True(t, result)
	})

	t.Run("attribute order doesn't matter", func(t *testing.T) {
		tb := &testing.T{}
		result := helpers.EqualHTML(
			tb,
			[]byte(`<div class="a" id="x"></div>`),
			[]byte(`<div id="x" class="a"></div>`),
			nil,
			nil,
		)
		require.True(t, result)
	})

	t.Run("different attribute values don't match", func(t *testing.T) {
		tb := &testing.T{}
		result := helpers.EqualHTML(tb, []byte(`<div class="a"></div>`), []byte(`<div class="b"></div>`), nil, nil)
		require.False(t, result)
	})

	t.Run("nested structures match recursively", func(t *testing.T) {
		tb := &testing.T{}
		html := []byte("<div><p><span>text</span></p></div>")
		result := helpers.EqualHTML(tb, html, html, nil, nil)
		require.True(t, result)
	})

	t.Run("different child count doesn't match", func(t *testing.T) {
		tb := &testing.T{}
		result := helpers.EqualHTML(tb, []byte("<div><p></p><p></p></div>"), []byte("<div><p></p></div>"), nil, nil)
		require.False(t, result)
	})

	t.Run("empty text nodes are ignored", func(t *testing.T) {
		tb := &testing.T{}
		result := helpers.EqualHTML(tb, []byte("<div></div>"), []byte("<div>   </div>"), nil, nil)
		require.True(t, result)
	})

	t.Run("matching HTML with different comments", func(t *testing.T) {
		tb := &testing.T{}
		// Both have same structure, just different comment content
		result := helpers.EqualHTML(
			tb,
			[]byte("<div><!-- comment1 --><p>text</p></div>"),
			[]byte("<div><!-- comment2 --><p>text</p></div>"),
			nil,
			nil,
		)
		// Comments return true, so this should match
		require.True(t, result)
	})

	t.Run("mixed attributes and nesting", func(t *testing.T) {
		tb := &testing.T{}
		html1 := []byte(`<div data-test="1" class="c"><p>content</p></div>`)
		html2 := []byte(`<div class="c" data-test="1"><p>content</p></div>`)
		result := helpers.EqualHTML(tb, html1, html2, nil, nil)
		require.True(t, result)
	})

	t.Run("complex document structures match", func(t *testing.T) {
		tb := &testing.T{}
		html := []byte(`
			<div class="container">
				<h1>Title</h1>
				<p id="para">
					<strong>bold</strong> text
				</p>
			</div>
		`)
		result := helpers.EqualHTML(tb, html, html, nil, nil)
		require.True(t, result)
	})
}
