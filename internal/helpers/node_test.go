package helpers_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"

	"github.com/titpetric/vuego/internal/helpers"
)

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
		require.Equal(t, "container", helpers.GetAttr(n, "class"))
		require.Equal(t, "main", helpers.GetAttr(n, "id"))
	})

	t.Run("missing attribute", func(t *testing.T) {
		n := &html.Node{
			Type: html.ElementNode,
			Data: "div",
			Attr: []html.Attribute{{Key: "class", Val: "container"}},
		}
		require.Equal(t, "", helpers.GetAttr(n, "id"))
	})

	t.Run("empty attributes", func(t *testing.T) {
		n := &html.Node{Type: html.ElementNode, Data: "div"}
		require.Equal(t, "", helpers.GetAttr(n, "class"))
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
		require.Len(t, n.Attr, 1)
		require.Equal(t, "id", n.Attr[0].Key)
		require.Equal(t, "main", n.Attr[0].Val)
	})

	t.Run("does nothing for missing attribute", func(t *testing.T) {
		n := &html.Node{
			Type: html.ElementNode,
			Data: "div",
			Attr: []html.Attribute{{Key: "class", Val: "container"}},
		}
		helpers.RemoveAttr(n, "id")
		require.Len(t, n.Attr, 1)
		require.Equal(t, "class", n.Attr[0].Key)
	})

	t.Run("handles empty attributes", func(t *testing.T) {
		n := &html.Node{Type: html.ElementNode, Data: "div"}
		helpers.RemoveAttr(n, "class")
		require.Len(t, n.Attr, 0)
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
		require.Equal(t, n.Type, clone.Type)
		require.Equal(t, n.Data, clone.Data)
		require.Equal(t, n.Attr, clone.Attr)
		require.Nil(t, clone.FirstChild)
		require.Nil(t, clone.NextSibling)
	})

	t.Run("modifying clone attrs does not affect original", func(t *testing.T) {
		n := &html.Node{
			Type: html.ElementNode,
			Data: "div",
			Attr: []html.Attribute{{Key: "class", Val: "container"}},
		}
		clone := helpers.CloneNode(n)
		clone.Attr = append(clone.Attr, html.Attribute{Key: "id", Val: "main"})
		require.Len(t, n.Attr, 1)
		require.Len(t, clone.Attr, 2)
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

		require.Equal(t, n.Type, clone.Type)
		require.Equal(t, n.Data, clone.Data)
		require.NotNil(t, clone.FirstChild)

		cloneChild1 := clone.FirstChild
		require.Equal(t, child1.Data, cloneChild1.Data)
		require.Equal(t, clone, cloneChild1.Parent)

		cloneChild2 := cloneChild1.NextSibling
		require.NotNil(t, cloneChild2)
		require.Equal(t, child2.Data, cloneChild2.Data)
		require.Equal(t, clone, cloneChild2.Parent)

		cloneGrandchild := cloneChild2.FirstChild
		require.NotNil(t, cloneGrandchild)
		require.Equal(t, grandchild.Data, cloneGrandchild.Data)
		require.Equal(t, cloneChild2, cloneGrandchild.Parent)
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

		require.Equal(t, "original", n.FirstChild.Data)
		require.Equal(t, "modified", clone.FirstChild.Data)
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
		require.NotNil(t, c1)
		require.Nil(t, c1.PrevSibling)
		require.Equal(t, "first", c1.Data)

		c2 := c1.NextSibling
		require.NotNil(t, c2)
		require.Equal(t, c1, c2.PrevSibling)
		require.Equal(t, "second", c2.Data)

		c3 := c2.NextSibling
		require.NotNil(t, c3)
		require.Equal(t, c2, c3.PrevSibling)
		require.Nil(t, c3.NextSibling)
		require.Equal(t, "third", c3.Data)
	})
}
