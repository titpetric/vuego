package vuego_test

import (
	"bytes"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego"
)

// BlogPost represents a blog post with typed fields
type BlogPost struct {
	ID        int
	Title     string
	Content   string
	Tags      []string
	Published bool
}

func TestVue_RealWorldTypePreservation(t *testing.T) {
	// This test replicates the platform-example/blog scenario where
	// []model.Article is passed to a template via :articles="articles"
	tests := []struct {
		name     string
		parent   string
		child    string
		data     map[string]any
		expected []string
	}{
		{
			name:   "blog post slice with type checking and length filter",
			parent: `<template include="posts.vuego" :posts="posts"></template>`,
			child: `<template>
Type: {{ posts | type }}
Count: {{ posts | len }}
</template>`,
			data: map[string]any{
				"posts": []BlogPost{
					{ID: 1, Title: "Post 1", Content: "Content 1", Tags: []string{"go", "testing"}},
					{ID: 2, Title: "Post 2", Content: "Content 2", Tags: []string{"go", "types"}},
				},
			},
			expected: []string{
				"vuego_test.BlogPost", // Type should be preserved, not "string"
				"Count: 2",            // len filter should work
			},
		},
		{
			name:   "single struct preserves type",
			parent: `<template include="post.vuego" :post="featured"></template>`,
			child: `<template>
Post type: {{ post | type }}
</template>`,
			data: map[string]any{
				"featured": BlogPost{ID: 1, Title: "Featured", Content: "Featured content"},
			},
			expected: []string{
				"vuego_test.BlogPost", // Type preserved
			},
		},
		{
			name:   "string slice also preserved",
			parent: `<template include="tags.vuego" :tags="tags"></template>`,
			child: `<template>
Tags type: {{ tags | type }}
Tag count: {{ tags | len }}
</template>`,
			data: map[string]any{
				"tags": []string{"go", "template", "testing"},
			},
			expected: []string{
				"[]string", // Type preserved
				"Tag count: 3",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fs := fstest.MapFS{
				"parent.vuego": &fstest.MapFile{Data: []byte(tc.parent)},
				"posts.vuego":  &fstest.MapFile{Data: []byte(tc.child)},
				"post.vuego":   &fstest.MapFile{Data: []byte(tc.child)},
				"tags.vuego":   &fstest.MapFile{Data: []byte(tc.child)},
			}
			vue := vuego.NewVue(fs)

			var buf bytes.Buffer
			err := vue.RenderFragment(&buf, "parent.vuego", tc.data)
			require.NoError(t, err)

			output := buf.String()
			for _, exp := range tc.expected {
				require.Contains(t, output, exp, "output: %s", output)
			}
		})
	}
}
