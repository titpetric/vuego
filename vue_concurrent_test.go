package vuego_test

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
	for err := range errors {
		t.Errorf("Concurrent render error: %v", err)
	}

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

func BenchmarkVue_ConcurrentBlogRender(b *testing.B) {
	fs := fstest.MapFS{
		"blog.html": &fstest.MapFile{
			Data: []byte(`<article>
  <h1>{{ title }}</h1>
  <p><em>By {{ author }} on {{ date }}</em></p>
  
  <p>{{ intro }}</p>

  <h2>Key Features</h2>
  <ul>
    <li v-for="feature in features">
      <strong>{{ feature.name }}</strong>: {{ feature.description }}
    </li>
  </ul>

  <h2>Getting Started</h2>
  <p>Follow these steps to begin:</p>
  <ol>
    <li v-for="step in steps">{{ step }}</li>
  </ol>

  <blockquote>
    {{ quote.text }}
    <br><strong>— {{ quote.author }}</strong>
  </blockquote>

  <h2>Code Example</h2>
  <p>Here's a simple example:</p>
  <pre><code>{{ codeExample }}</code></pre>

  <hr>
  
  <h3>Related Articles</h3>
  <ul>
    <li v-for="article in related">
      <a href="#">{{ article }}</a>
    </li>
  </ul>

  <p><small>Tags: <span v-for="tag in tags">{{ tag }} </span></small></p>
</article>`),
		},
	}

	vue := vuego.NewVue(fs)

	data := map[string]any{
		"title":  "Building Modern Web Applications with VueGo",
		"author": "Jane Developer",
		"date":   "November 14, 2025",
		"intro":  "VueGo combines the simplicity of Vue.js-style templates with the power of Go's template engine.",
		"features": []map[string]any{
			{"name": "Server-Side Rendering", "description": "Render templates on the server for better SEO"},
			{"name": "Familiar Syntax", "description": "Use Vue.js-like directives"},
			{"name": "Type Safety", "description": "Leverage Go's type system"},
		},
		"steps": []string{
			"Install VueGo package with go get",
			"Create your first template with .vuego extension",
			"Set up data as a map[string]any",
			"Render using vue.Render()",
			"Enjoy your dynamic HTML!",
		},
		"quote": map[string]any{
			"text":   "Simplicity is the ultimate sophistication.",
			"author": "Leonardo da Vinci",
		},
		"codeExample": `vue := vuego.NewVue(templateFS)\nvar buf bytes.Buffer\nerr := vue.Render(&buf, "page.vuego", data)`,
		"related": []string{
			"Understanding Template Directives",
			"Advanced Filtering Techniques",
			"Building Reusable Components",
		},
		"tags": []string{"golang", "templates", "vue", "web-development", "ssr"},
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var buf bytes.Buffer
			err := vue.Render(&buf, "blog.html", data)
			require.NoError(b, err)
		}
	})
}
