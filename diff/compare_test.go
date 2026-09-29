package diff_test

import (
	"bytes"
	"testing"

	"golang.org/x/net/html"

	"github.com/titpetric/vuego/diff"
	"github.com/titpetric/vuego/tests/assert"
)

func TestSignificantChildren(t *testing.T) {
	t.Run("nil node returns nil", func(t *testing.T) {
		result := diff.SignificantChildren(nil)
		assert.Nil(t, result)
	})

	t.Run("document node with single element child", func(t *testing.T) {
		n := &html.Node{Type: html.DocumentNode}
		div := &html.Node{Type: html.ElementNode, Data: "div"}
		n.AppendChild(div)
		result := diff.SignificantChildren(n)
		assert.Len(t, result, 1)
		assert.Equal(t, div, result[0])
	})

	t.Run("document node ignores whitespace-only text nodes", func(t *testing.T) {
		n := &html.Node{Type: html.DocumentNode}
		whitespace := &html.Node{Type: html.TextNode, Data: "   \n  "}
		div := &html.Node{Type: html.ElementNode, Data: "div"}
		whitespace2 := &html.Node{Type: html.TextNode, Data: "\t"}

		n.AppendChild(whitespace)
		n.AppendChild(div)
		n.AppendChild(whitespace2)

		result := diff.SignificantChildren(n)
		assert.Len(t, result, 1)
		assert.Equal(t, div, result[0])
	})

	t.Run("document with multiple element children", func(t *testing.T) {
		n := &html.Node{Type: html.DocumentNode}
		div1 := &html.Node{Type: html.ElementNode, Data: "div"}
		div2 := &html.Node{Type: html.ElementNode, Data: "div"}
		n.AppendChild(div1)
		n.AppendChild(div2)

		result := diff.SignificantChildren(n)
		assert.Len(t, result, 2)
		assert.Equal(t, div1, result[0])
		assert.Equal(t, div2, result[1])
	})

	t.Run("element node returns itself wrapped", func(t *testing.T) {
		div := &html.Node{Type: html.ElementNode, Data: "div"}
		result := diff.SignificantChildren(div)
		assert.Len(t, result, 1)
		assert.Equal(t, div, result[0])
	})

	t.Run("text node returns itself wrapped", func(t *testing.T) {
		text := &html.Node{Type: html.TextNode, Data: "hello"}
		result := diff.SignificantChildren(text)
		assert.Len(t, result, 1)
		assert.Equal(t, text, result[0])
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

		result := diff.SignificantChildren(root)
		// When html->body, returns body's children
		assert.True(t, len(result) > 0, "should have children")
		assert.Equal(t, "p", result[0].Data)
	})

	t.Run("parsed html document unwraps body", func(t *testing.T) {
		// html.Parse creates html/head/body structure
		doc, err := html.Parse(bytes.NewReader([]byte("<html><body><p>test</p></body></html>")))
		assert.NoError(t, err)

		result := diff.SignificantChildren(doc)
		// When html->head+body, returns htmlChildren (head and body)
		assert.True(t, len(result) > 0, "should have children")
	})

	t.Run("comment nodes are ignored", func(t *testing.T) {
		n := &html.Node{Type: html.DocumentNode}
		comment := &html.Node{Type: html.CommentNode, Data: "comment"}
		div := &html.Node{Type: html.ElementNode, Data: "div"}

		n.AppendChild(comment)
		n.AppendChild(div)

		result := diff.SignificantChildren(n)
		assert.Len(t, result, 2) // comment and div both included in doc children
	})
}

func TestCompareHTML(t *testing.T) {
	tests := []struct {
		name string
		want string
		got  string
		same bool
	}{
		{
			name: "identical simple HTML",
			want: "<div><p>text</p></div>",
			got:  "<div><p>text</p></div>",
			same: true,
		},
		{
			name: "whitespace in text nodes is ignored",
			want: "<p>hello</p>",
			got:  "<p>  hello  </p>",
			same: true,
		},
		{
			name: "attribute order is ignored",
			want: `<div class="a" id="x"></div>`,
			got:  `<div id="x" class="a"></div>`,
			same: true,
		},
		{
			name: "class order is ignored",
			want: `<div class="a b c"></div>`,
			got:  `<div class="c a b"></div>`,
			same: true,
		},
		{
			name: "style property order is ignored",
			want: `<div style="color: red; margin: 0"></div>`,
			got:  `<div style="margin: 0; color: red"></div>`,
			same: true,
		},
		{
			name: "nested structures match recursively",
			want: "<div><p><span>text</span></p></div>",
			got:  "<div><p><span>text</span></p></div>",
			same: true,
		},
		{
			name: "whitespace-only text nodes are ignored",
			want: "<div></div>",
			got:  "<div>   </div>",
			same: true,
		},
		{
			name: "comment content is ignored",
			want: "<div><!-- comment1 --><p>text</p></div>",
			got:  "<div><!-- comment2 --><p>text</p></div>",
			same: true,
		},
		{
			name: "full document compares against fragment",
			want: "<html><body><p>text</p></body></html>",
			got:  "<p>text</p>",
			same: true,
		},
		{
			name: "different tag name",
			want: "<div>text</div>",
			got:  "<span>text</span>",
			same: false,
		},
		{
			name: "different text content",
			want: "<p>hello</p>",
			got:  "<p>goodbye</p>",
			same: false,
		},
		{
			name: "different attribute value",
			want: `<div id="a"></div>`,
			got:  `<div id="b"></div>`,
			same: false,
		},
		{
			name: "extra attribute",
			want: `<div id="a"></div>`,
			got:  `<div id="a" class="c"></div>`,
			same: false,
		},
		{
			name: "missing attribute key",
			want: `<div id="a"></div>`,
			got:  `<div class="a"></div>`,
			same: false,
		},
		{
			name: "different class count",
			want: `<div class="a b"></div>`,
			got:  `<div class="a"></div>`,
			same: false,
		},
		{
			name: "different class names",
			want: `<div class="a b"></div>`,
			got:  `<div class="a c"></div>`,
			same: false,
		},
		{
			name: "different style value",
			want: `<div style="color: red"></div>`,
			got:  `<div style="color: blue"></div>`,
			same: false,
		},
		{
			name: "different style property count",
			want: `<div style="color: red; margin: 0"></div>`,
			got:  `<div style="color: red"></div>`,
			same: false,
		},
		{
			name: "extra child element",
			want: "<div><p>a</p></div>",
			got:  "<div><p>a</p><p>b</p></div>",
			same: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.same, diff.CompareHTML([]byte(test.want), []byte(test.got)))
		})
	}
}

func TestCountChildren(t *testing.T) {
	t.Run("nil node counts zero", func(t *testing.T) {
		assert.Equal(t, 0, diff.CountChildren(nil))
	})

	t.Run("childless node counts zero", func(t *testing.T) {
		assert.Equal(t, 0, diff.CountChildren(&html.Node{Type: html.ElementNode, Data: "div"}))
	})

	t.Run("counts every child including whitespace", func(t *testing.T) {
		n := &html.Node{Type: html.ElementNode, Data: "div"}
		n.AppendChild(&html.Node{Type: html.TextNode, Data: "  "})
		n.AppendChild(&html.Node{Type: html.ElementNode, Data: "p"})
		n.AppendChild(&html.Node{Type: html.CommentNode, Data: "c"})

		assert.Equal(t, 3, diff.CountChildren(n))
	})

	t.Run("counts direct children only", func(t *testing.T) {
		doc, err := html.Parse(bytes.NewReader([]byte("<html><body><div><p><span>a</span></p><p>b</p></div></body></html>")))
		assert.NoError(t, err)

		// html.Parse always inserts a head, so SignificantChildren returns
		// head and body rather than unwrapping body.
		children := diff.SignificantChildren(doc)
		assert.Len(t, children, 2)
		assert.Equal(t, "body", children[1].Data)

		div := children[1].FirstChild
		assert.Equal(t, "div", div.Data)
		assert.Equal(t, 2, diff.CountChildren(div))
	})
}
