package tests

import (
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"

	"github.com/titpetric/vuego"
)

func getBlogData() map[string]any {
	return map[string]any{
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
}

func getBlogTemplate() []byte {
	return []byte(`<article>
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
</article>`)
}

func BenchmarkVue_Render_Detailed(b *testing.B) {
	fs := fstest.MapFS{
		"blog.vuego": &fstest.MapFile{
			Data: getBlogTemplate(),
		},
	}

	vue := vuego.NewVue(fs)
	data := getBlogData()

	// Warm up cache
	var buf strings.Builder
	err := vue.Render(&buf, "blog.vuego", data)
	require.NoError(b, err)

	// Measure total time
	var totalDuration time.Duration

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		start := time.Now()
		var buf strings.Builder
		err := vue.Render(&buf, "blog.vuego", data)
		require.NoError(b, err)
		totalDuration += time.Since(start)
	}

	// Calculate average times
	avgTotal := totalDuration / time.Duration(b.N)
	b.ReportMetric(float64(avgTotal.Microseconds()), "µs/op")
}

func TestDetailedRenderTiming(t *testing.T) {
	fs := fstest.MapFS{
		"blog.vuego": &fstest.MapFile{
			Data: getBlogTemplate(),
		},
	}

	vue := vuego.NewVue(fs)
	data := getBlogData()

	iterations := 1000

	// Measure overall render time with warm cache
	var totalTime time.Duration
	for i := 0; i < iterations; i++ {
		start := time.Now()
		var buf strings.Builder
		err := vue.Render(&buf, "blog.vuego", data)
		require.NoError(t, err)
		totalTime += time.Since(start)
	}

	avgTotal := totalTime / time.Duration(iterations)
	t.Logf("Average Render() time: %d µs", avgTotal.Microseconds())
}

func TestDetailedRenderBreakdown(t *testing.T) {
	fs := fstest.MapFS{
		"blog.vuego": &fstest.MapFile{
			Data: getBlogTemplate(),
		},
	}

	vue := vuego.NewVue(fs)
	data := getBlogData()

	// Warm up the cache and compile
	var warmupBuf strings.Builder
	_ = vue.Render(&warmupBuf, "blog.vuego", data)

	// Now measure component loading (parse step - will be cached after first call)
	var parseTime time.Duration
	parseIterations := 100
	for i := 0; i < parseIterations; i++ {
		// Create a fresh Vue instance to measure parse time
		freshVue := vuego.NewVue(fs)
		start := time.Now()
		_ = freshVue.Render(&strings.Builder{}, "blog.vuego", data)
		parseTime += time.Since(start)
	}
	avgParseIncluded := parseTime / time.Duration(parseIterations)

	// Measure cached render (evaluate + render steps)
	var cachedRenderTime time.Duration
	cachedIterations := 10000
	for i := 0; i < cachedIterations; i++ {
		start := time.Now()
		var buf strings.Builder
		_ = vue.Render(&buf, "blog.vuego", data)
		cachedRenderTime += time.Since(start)
	}
	avgCachedRender := cachedRenderTime / time.Duration(cachedIterations)

	// Estimate parse time (first render with fresh Vue minus cached render)
	estimatedParse := avgParseIncluded - avgCachedRender

	t.Logf("\n=== RENDER TIMING BREAKDOWN ===")
	t.Logf("Parse (loading + initial compilation): ~%d µs (estimated from fresh instance)", estimatedParse.Microseconds())
	t.Logf("Evaluate + Render (cached):            %d µs", avgCachedRender.Microseconds())
	t.Logf("Total with cold cache:                 ~%d µs", avgParseIncluded.Microseconds())
	t.Logf("================================\n")

	// Show percentage breakdown
	if avgParseIncluded > 0 {
		parsePercent := float64(estimatedParse) / float64(avgParseIncluded) * 100
		evalRenderPercent := float64(avgCachedRender) / float64(avgParseIncluded) * 100
		t.Logf("Parse: %.1f%% | Evaluate+Render: %.1f%%", parsePercent, evalRenderPercent)
	}
}

func BenchmarkHtmlParse(b *testing.B) {
	templateData := getBlogTemplate()
	reader := strings.NewReader(string(templateData))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader.Reset(string(templateData))
		_, _ = html.Parse(reader)
	}
}

func TestHtmlParseVsVueGoRender(t *testing.T) {
	templateData := getBlogTemplate()
	data := getBlogData()

	// Benchmark html.Parse
	htmlParseIterations := 10000
	var htmlParseTime time.Duration
	for i := 0; i < htmlParseIterations; i++ {
		reader := strings.NewReader(string(templateData))
		start := time.Now()
		_, _ = html.Parse(reader)
		htmlParseTime += time.Since(start)
	}
	avgHtmlParse := htmlParseTime / time.Duration(htmlParseIterations)

	// Benchmark vuego full render
	fs := fstest.MapFS{
		"blog.vuego": &fstest.MapFile{
			Data: templateData,
		},
	}
	vue := vuego.NewVue(fs)

	vuegoIterations := 10000
	var vuegoTime time.Duration
	for i := 0; i < vuegoIterations; i++ {
		start := time.Now()
		var buf strings.Builder
		_ = vue.Render(&buf, "blog.vuego", data)
		vuegoTime += time.Since(start)
	}
	avgVuegoRender := vuegoTime / time.Duration(vuegoIterations)

	t.Logf("\n=== PARSE VS FULL RENDER ===")
	t.Logf("html.Parse() alone:              %d µs", avgHtmlParse.Microseconds())
	t.Logf("vuego Render() (full pipeline):  %d µs", avgVuegoRender.Microseconds())
	t.Logf("VueGo overhead (eval+render):    %d µs", (avgVuegoRender - avgHtmlParse).Microseconds())
	t.Logf("Ratio (VueGo / Parse):           %.1fx", float64(avgVuegoRender)/float64(avgHtmlParse))
	t.Logf("============================\n")
}
