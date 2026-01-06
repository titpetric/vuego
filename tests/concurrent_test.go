package tests

import (
	"bytes"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego"
)

func TestVue_ConcurrentRender(t *testing.T) {
	fs := fstest.MapFS{
		"page.html": &fstest.MapFile{
			Data: []byte(`<html><body><h1>{{ title }}</h1><p>{{ content }}</p></body></html>`),
		},
	}

	vue := vuego.NewVue(fs)

	// Render the same template concurrently with different data
	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	errors := make(chan error, numGoroutines)
	results := make(chan string, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			data := map[string]any{
				"title":   "Page " + string(rune('A'+id%26)),
				"content": "Content for goroutine " + string(rune('0'+id%10)),
			}

			var buf bytes.Buffer
			err := vue.Render(&buf, "page.html", data)
			if err != nil {
				errors <- err
				return
			}

			results <- buf.String()
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

	// Verify all renders completed
	count := 0
	for range results {
		count++
	}
	require.Equal(t, numGoroutines, count, "Expected all goroutines to complete")
}

func TestVue_ConcurrentRenderWithFuncs(t *testing.T) {
	fs := fstest.MapFS{
		"page.html": &fstest.MapFile{
			Data: []byte(`<p>{{ name | upper }}</p>`),
		},
	}

	vue := vuego.NewVue(fs)

	const numGoroutines = 50
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			data := map[string]any{
				"name": "user",
			}

			var buf bytes.Buffer
			err := vue.Render(&buf, "page.html", data)
			require.NoError(t, err)
			require.Contains(t, buf.String(), "USER")
		}(i)
	}

	wg.Wait()
}

func TestVue_ConcurrentRenderDifferentTemplates(t *testing.T) {
	fs := fstest.MapFS{
		"template1.html": &fstest.MapFile{
			Data: []byte(`<div>{{ value1 }}</div>`),
		},
		"template2.html": &fstest.MapFile{
			Data: []byte(`<span>{{ value2 }}</span>`),
		},
		"template3.html": &fstest.MapFile{
			Data: []byte(`<p>{{ value3 }}</p>`),
		},
	}

	vue := vuego.NewVue(fs)

	const numGoroutines = 60
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			template := "template" + string(rune('1'+(id%3))) + ".html"
			dataKey := "value" + string(rune('1'+(id%3)))

			data := map[string]any{
				dataKey: "test" + string(rune('0'+(id%10))),
			}

			var buf bytes.Buffer
			err := vue.Render(&buf, template, data)
			require.NoError(t, err)
		}(i)
	}

	wg.Wait()
}
