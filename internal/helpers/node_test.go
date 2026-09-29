package helpers_test

import (
	"strings"
	"testing"

	"golang.org/x/net/html"

	"github.com/titpetric/vuego/internal/helpers"
	"github.com/titpetric/vuego/tests/assert"
)

func TestNewNode(t *testing.T) {
	t.Run("returns a zeroed node", func(t *testing.T) {
		n := helpers.NewNode()
		assert.Equal(t, html.NodeType(0), n.Type)
		assert.Equal(t, "", n.Data)
		assert.Equal(t, "", n.Namespace)
		assert.Nil(t, n.Attr)
		assert.Nil(t, n.FirstChild)
		assert.Nil(t, n.LastChild)
		assert.Nil(t, n.NextSibling)
		assert.Nil(t, n.PrevSibling)
		assert.Nil(t, n.Parent)
	})

	t.Run("clears state carried over from the pool", func(t *testing.T) {
		// Fill a node, drop the reference and allocate enough nodes that a
		// pooled one is handed back at least once.
		dirty := helpers.NewNode()
		dirty.Type = html.ElementNode
		dirty.Data = "div"
		dirty.Namespace = "svg"
		dirty.Attr = []html.Attribute{{Key: "class", Val: "container"}}
		dirty.FirstChild = &html.Node{Type: html.TextNode, Data: "text"}

		for i := 0; i < 32; i++ {
			n := helpers.NewNode()
			assert.Equal(t, html.NodeType(0), n.Type)
			assert.Equal(t, "", n.Data)
			assert.Nil(t, n.Attr)
			assert.Nil(t, n.FirstChild)
		}
	})
}

func TestGetAttr(t *testing.T) {
	t.Run("existing attribute", func(t *testing.T) {
		n := &html.Node{
			Type: html.ElementNode,
			Data: "div",
			Attr: []html.Attribute{
				{Key: "class", Val: "container"},
				{Key: "id", Val: "main"},
			},
		}
		assert.Equal(t, "container", helpers.GetAttr(n, "class"))
		assert.Equal(t, "main", helpers.GetAttr(n, "id"))
	})

	t.Run("missing attribute", func(t *testing.T) {
		n := &html.Node{
			Type: html.ElementNode,
			Data: "div",
			Attr: []html.Attribute{{Key: "class", Val: "container"}},
		}
		assert.Equal(t, "", helpers.GetAttr(n, "id"))
	})

	t.Run("empty attributes", func(t *testing.T) {
		n := &html.Node{Type: html.ElementNode, Data: "div"}
		assert.Equal(t, "", helpers.GetAttr(n, "class"))
	})
}

func TestHasAttr(t *testing.T) {
	n := &html.Node{
		Type: html.ElementNode,
		Data: "div",
		Attr: []html.Attribute{
			{Key: "class", Val: "container"},
			{Key: "v-keep"},
		},
	}

	t.Run("attribute with a value", func(t *testing.T) {
		assert.True(t, helpers.HasAttr(n, "class"))
	})

	t.Run("attribute with an empty value", func(t *testing.T) {
		assert.True(t, helpers.HasAttr(n, "v-keep"))
	})

	t.Run("missing attribute", func(t *testing.T) {
		assert.False(t, helpers.HasAttr(n, "id"))
	})

	t.Run("no attributes", func(t *testing.T) {
		assert.False(t, helpers.HasAttr(&html.Node{Type: html.ElementNode, Data: "div"}, "class"))
	})
}

func TestSetAttr(t *testing.T) {
	t.Run("replaces an existing value", func(t *testing.T) {
		n := &html.Node{
			Type: html.ElementNode,
			Data: "div",
			Attr: []html.Attribute{
				{Key: "class", Val: "old"},
				{Key: "id", Val: "main"},
			},
		}
		helpers.SetAttr(n, "class", "new")
		assert.Len(t, n.Attr, 2)
		assert.Equal(t, "new", helpers.GetAttr(n, "class"))
		assert.Equal(t, "main", helpers.GetAttr(n, "id"))
	})

	t.Run("appends a missing attribute", func(t *testing.T) {
		n := &html.Node{
			Type: html.ElementNode,
			Data: "div",
			Attr: []html.Attribute{{Key: "class", Val: "container"}},
		}
		helpers.SetAttr(n, "id", "main")
		assert.Len(t, n.Attr, 2)
		assert.Equal(t, "id", n.Attr[1].Key)
		assert.Equal(t, "main", n.Attr[1].Val)
	})

	t.Run("appends to a node with no attributes", func(t *testing.T) {
		n := &html.Node{Type: html.ElementNode, Data: "div"}
		helpers.SetAttr(n, "class", "container")
		assert.Len(t, n.Attr, 1)
		assert.Equal(t, "container", helpers.GetAttr(n, "class"))
	})

	t.Run("sets an empty value", func(t *testing.T) {
		n := &html.Node{
			Type: html.ElementNode,
			Data: "div",
			Attr: []html.Attribute{{Key: "class", Val: "container"}},
		}
		helpers.SetAttr(n, "class", "")
		assert.Len(t, n.Attr, 1)
		assert.Equal(t, "", helpers.GetAttr(n, "class"))
	})
}

func TestAppendAttr(t *testing.T) {
	// AppendAttr delegates to SetAttr: an existing value is replaced, not
	// extended.
	t.Run("replaces an existing value", func(t *testing.T) {
		n := &html.Node{
			Type: html.ElementNode,
			Data: "div",
			Attr: []html.Attribute{{Key: "class", Val: "one"}},
		}
		helpers.AppendAttr(n, "class", "two")
		assert.Len(t, n.Attr, 1)
		assert.Equal(t, "two", helpers.GetAttr(n, "class"))
	})

	t.Run("appends a missing attribute", func(t *testing.T) {
		n := &html.Node{Type: html.ElementNode, Data: "div"}
		helpers.AppendAttr(n, "class", "one")
		assert.Len(t, n.Attr, 1)
		assert.Equal(t, "one", helpers.GetAttr(n, "class"))
	})
}

func TestRemoveAttr(t *testing.T) {
	t.Run("removes existing attribute", func(t *testing.T) {
		n := &html.Node{
			Type: html.ElementNode,
			Data: "div",
			Attr: []html.Attribute{
				{Key: "class", Val: "container"},
				{Key: "id", Val: "main"},
			},
		}
		helpers.RemoveAttr(n, "class")
		assert.Len(t, n.Attr, 1)
		assert.Equal(t, "id", n.Attr[0].Key)
		assert.Equal(t, "main", n.Attr[0].Val)
	})

	t.Run("does nothing for missing attribute", func(t *testing.T) {
		n := &html.Node{
			Type: html.ElementNode,
			Data: "div",
			Attr: []html.Attribute{{Key: "class", Val: "container"}},
		}
		helpers.RemoveAttr(n, "id")
		assert.Len(t, n.Attr, 1)
		assert.Equal(t, "class", n.Attr[0].Key)
	})

	t.Run("handles empty attributes", func(t *testing.T) {
		n := &html.Node{Type: html.ElementNode, Data: "div"}
		helpers.RemoveAttr(n, "class")
		assert.Len(t, n.Attr, 0)
	})
}

func TestFilterAttrs(t *testing.T) {
	t.Run("returns new slice excluding key", func(t *testing.T) {
		attrs := []html.Attribute{
			{Key: "class", Val: "container"},
			{Key: "id", Val: "main"},
			{Key: "data-test", Val: "value"},
		}
		filtered := helpers.FilterAttrs(attrs, "id")
		assert.Len(t, filtered, 2)
		assert.Equal(t, "class", filtered[0].Key)
		assert.Equal(t, "data-test", filtered[1].Key)
	})

	t.Run("does not modify original slice", func(t *testing.T) {
		attrs := []html.Attribute{
			{Key: "class", Val: "container"},
			{Key: "id", Val: "main"},
		}
		filtered := helpers.FilterAttrs(attrs, "class")
		assert.Len(t, attrs, 2)
		assert.Len(t, filtered, 1)
		assert.Equal(t, "class", attrs[0].Key)
		assert.Equal(t, "id", filtered[0].Key)
	})

	t.Run("handles missing key", func(t *testing.T) {
		attrs := []html.Attribute{
			{Key: "class", Val: "container"},
		}
		filtered := helpers.FilterAttrs(attrs, "id")
		assert.Len(t, filtered, 1)
		assert.Equal(t, "class", filtered[0].Key)
	})

	t.Run("handles empty attributes", func(t *testing.T) {
		filtered := helpers.FilterAttrs(nil, "class")
		assert.Len(t, filtered, 0)
	})
}

func TestCloneNode(t *testing.T) {
	t.Run("creates shallow copy", func(t *testing.T) {
		n := &html.Node{
			Type: html.ElementNode,
			Data: "div",
			Attr: []html.Attribute{{Key: "class", Val: "container"}},
		}
		child := &html.Node{Type: html.TextNode, Data: "text"}
		sibling := &html.Node{Type: html.TextNode, Data: "sibling"}
		n.FirstChild = child
		n.NextSibling = sibling

		clone := helpers.CloneNode(n)
		assert.Equal(t, n.Type, clone.Type)
		assert.Equal(t, n.Data, clone.Data)
		assert.Equal(t, n.Attr, clone.Attr)
		assert.Nil(t, clone.FirstChild)
		assert.Nil(t, clone.NextSibling)
	})

	t.Run("modifying clone attrs does not affect original", func(t *testing.T) {
		n := &html.Node{
			Type: html.ElementNode,
			Data: "div",
			Attr: []html.Attribute{{Key: "class", Val: "container"}},
		}
		clone := helpers.CloneNode(n)
		clone.Attr = append(clone.Attr, html.Attribute{Key: "id", Val: "main"})
		assert.Len(t, n.Attr, 1)
		assert.Len(t, clone.Attr, 2)
	})
}

func TestShallowCloneWithAttrs(t *testing.T) {
	t.Run("copies the attribute slice", func(t *testing.T) {
		n := &html.Node{
			Type: html.ElementNode,
			Data: "div",
			Attr: []html.Attribute{{Key: "class", Val: "container"}},
		}

		clone := helpers.ShallowCloneWithAttrs(n)
		clone.Attr[0].Val = "changed"

		assert.Equal(t, "container", helpers.GetAttr(n, "class"))
		assert.Equal(t, "changed", helpers.GetAttr(clone, "class"))
	})

	t.Run("does not copy children or siblings", func(t *testing.T) {
		n := &html.Node{Type: html.ElementNode, Data: "div"}
		n.AppendChild(&html.Node{Type: html.TextNode, Data: "text"})
		n.NextSibling = &html.Node{Type: html.TextNode, Data: "sibling"}

		clone := helpers.ShallowCloneWithAttrs(n)
		assert.Equal(t, n.Type, clone.Type)
		assert.Equal(t, n.Data, clone.Data)
		assert.Nil(t, clone.FirstChild)
		assert.Nil(t, clone.NextSibling)
	})

	t.Run("leaves attrs nil when the source has none", func(t *testing.T) {
		clone := helpers.ShallowCloneWithAttrs(&html.Node{Type: html.ElementNode, Data: "div"})
		assert.Len(t, clone.Attr, 0)
	})
}

func TestGetBodyNode(t *testing.T) {
	t.Run("returns a body element", func(t *testing.T) {
		body := helpers.GetBodyNode()
		assert.NotNil(t, body)
		assert.Equal(t, html.ElementNode, body.Type)
		assert.Equal(t, "body", body.Data)
	})

	t.Run("returns the same cached node", func(t *testing.T) {
		assert.Equal(t, helpers.GetBodyNode(), helpers.GetBodyNode())
	})

	t.Run("is usable as ParseFragment context", func(t *testing.T) {
		nodes, err := html.ParseFragment(strings.NewReader("<p>a</p><p>b</p>"), helpers.GetBodyNode())
		assert.NoError(t, err)
		assert.Len(t, nodes, 2)
		assert.Equal(t, "p", nodes[0].Data)
	})
}

func TestDeepCloneNode(t *testing.T) {
	t.Run("clones node with children", func(t *testing.T) {
		n := &html.Node{
			Type: html.ElementNode,
			Data: "div",
			Attr: []html.Attribute{{Key: "class", Val: "container"}},
		}
		child1 := &html.Node{Type: html.TextNode, Data: "text1"}
		child2 := &html.Node{Type: html.ElementNode, Data: "span"}
		grandchild := &html.Node{Type: html.TextNode, Data: "nested"}

		n.AppendChild(child1)
		n.AppendChild(child2)
		child2.AppendChild(grandchild)

		clone := helpers.DeepCloneNode(n)

		assert.Equal(t, n.Type, clone.Type)
		assert.Equal(t, n.Data, clone.Data)
		assert.NotNil(t, clone.FirstChild)

		cloneChild1 := clone.FirstChild
		assert.Equal(t, child1.Data, cloneChild1.Data)
		assert.Equal(t, clone, cloneChild1.Parent)

		cloneChild2 := cloneChild1.NextSibling
		assert.NotNil(t, cloneChild2)
		assert.Equal(t, child2.Data, cloneChild2.Data)
		assert.Equal(t, clone, cloneChild2.Parent)

		cloneGrandchild := cloneChild2.FirstChild
		assert.NotNil(t, cloneGrandchild)
		assert.Equal(t, grandchild.Data, cloneGrandchild.Data)
		assert.Equal(t, cloneChild2, cloneGrandchild.Parent)
	})

	t.Run("modifying clone does not affect original", func(t *testing.T) {
		n := &html.Node{
			Type: html.ElementNode,
			Data: "div",
		}
		child := &html.Node{Type: html.TextNode, Data: "original"}
		n.AppendChild(child)

		clone := helpers.DeepCloneNode(n)
		clone.FirstChild.Data = "modified"

		assert.Equal(t, "original", n.FirstChild.Data)
		assert.Equal(t, "modified", clone.FirstChild.Data)
	})

	t.Run("preserves sibling relationships", func(t *testing.T) {
		n := &html.Node{Type: html.ElementNode, Data: "div"}
		child1 := &html.Node{Type: html.TextNode, Data: "first"}
		child2 := &html.Node{Type: html.TextNode, Data: "second"}
		child3 := &html.Node{Type: html.TextNode, Data: "third"}

		n.AppendChild(child1)
		n.AppendChild(child2)
		n.AppendChild(child3)

		clone := helpers.DeepCloneNode(n)

		c1 := clone.FirstChild
		assert.NotNil(t, c1)
		assert.Nil(t, c1.PrevSibling)
		assert.Equal(t, "first", c1.Data)

		c2 := c1.NextSibling
		assert.NotNil(t, c2)
		assert.Equal(t, c1, c2.PrevSibling)
		assert.Equal(t, "second", c2.Data)

		c3 := c2.NextSibling
		assert.NotNil(t, c3)
		assert.Equal(t, c2, c3.PrevSibling)
		assert.Nil(t, c3.NextSibling)
		assert.Equal(t, "third", c3.Data)
	})
}
