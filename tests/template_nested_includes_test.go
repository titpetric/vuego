package tests

import (
	"bytes"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego"
)

// TestTemplate_NestedIncludesStateIsolation reproduces the issue where shared vuego.Template
// instances lose article data on subsequent renders when using nested component includes.
//
// The blog module uses: tpl.Load(filename).Fill(data).Render(ctx, w)
// where tpl is shared across requests. With nested includes like:
//   - blog.vuego includes article-list.vuego
//   - article-list.vuego iterates over articles with v-for
//
// The first request renders correctly, but subsequent requests lose the articles data.
//
// Root cause investigation: The issue appears to be related to how the Stack is managed
// across multiple Load().Fill().Render() calls on the same template instance.
// Even though Copy() creates a new stack, something in the shared Vue instance may be
// retaining state that affects subsequent renders.
func TestTemplate_NestedIncludesStateIsolation(t *testing.T) {
	fs := fstest.MapFS{
		"pages/blog.vuego": &fstest.MapFile{
			Data: []byte(`<div class="blog">
  <template include="components/article-list.vuego" :articles="articles"></template>
</div>`),
		},
		"components/article-list.vuego": &fstest.MapFile{
			Data: []byte(`<ul>
  <li v-for="article in articles">{{ article.Title }}</li>
</ul>`),
		},
		"layouts/base.vuego": &fstest.MapFile{
			Data: []byte(`<html><body v-html="content"></body></html>`),
		},
	}

	tpl := vuego.NewFS(fs)

	// First render
	data1 := map[string]any{
		"articles": []map[string]any{
			{"Title": "Article 1"},
			{"Title": "Article 2"},
		},
	}

	var buf1 bytes.Buffer
	err := tpl.Load("pages/blog.vuego").Fill(data1).Render(t.Context(), &buf1)
	require.NoError(t, err)
	output1 := buf1.String()
	require.Contains(t, output1, "Article 1")
	require.Contains(t, output1, "Article 2")

	// Second render with different data using same shared tpl instance
	data2 := map[string]any{
		"articles": []map[string]any{
			{"Title": "Article 3"},
			{"Title": "Article 4"},
		},
	}

	var buf2 bytes.Buffer
	err = tpl.Load("pages/blog.vuego").Fill(data2).Render(t.Context(), &buf2)
	require.NoError(t, err)
	output2 := buf2.String()

	// Verify second render has new articles and not old ones
	require.Contains(t, output2, "Article 3", "Second render should have Article 3")
	require.Contains(t, output2, "Article 4", "Second render lost articles data")
	require.NotContains(t, output2, "Article 1", "Old data should not appear")
	require.NotContains(t, output2, "Article 2", "Old data should not appear")
}

// TestTemplate_ConcurrentNestedIncludesRender tests concurrent renders with nested includes
// using the Template.New().Load().Fill() pattern.
func TestTemplate_ConcurrentNestedIncludesRender(t *testing.T) {
	fs := fstest.MapFS{
		"pages/blog.vuego": &fstest.MapFile{
			Data: []byte(`<div class="blog">
  <template include="components/article-list.vuego" :articles="articles"></template>
</div>`),
		},
		"components/article-list.vuego": &fstest.MapFile{
			Data: []byte(`<ul>
  <li v-for="article in articles">{{ article.Title }}</li>
</ul>`),
		},
		"layouts/base.vuego": &fstest.MapFile{
			Data: []byte(`<html><body v-html="content"></body></html>`),
		},
	}

	sharedTpl := vuego.NewFS(fs)

	const numGoroutines = 50
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	errors := make(chan error, numGoroutines)
	results := make(chan string, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			// Each goroutine renders with different article data
			articles := []map[string]any{
				{"Title": "Article_" + string(rune('A'+id%26))},
				{"Title": "Content_" + string(rune('0'+id%10))},
			}

			data := map[string]any{
				"articles": articles,
			}

			var buf bytes.Buffer
			err := sharedTpl.Load("pages/blog.vuego").Fill(data).Render(t.Context(), &buf)
			if err != nil {
				errors <- err
				return
			}

			output := buf.String()
			results <- output
		}(i)
	}

	wg.Wait()
	close(errors)
	close(results)

	// Check for errors
	var renderErrors []error
	for err := range errors {
		renderErrors = append(renderErrors, err)
	}
	require.Empty(t, renderErrors)

	// Verify all renders completed and contain expected content
	count := 0
	for output := range results {
		count++
		// Each output should contain article list items
		require.Contains(t, output, "<li>", "Output should contain list items")
		require.Contains(t, output, "Article_", "Output should contain article data")
	}
	require.Equal(t, numGoroutines, count, "Expected all goroutines to complete")
}
