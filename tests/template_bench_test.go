package vuego_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego"
)

// getTOCData returns a read-only table of contents data structure
// with nested headings (h2-h4) for benchmarking.
func getTOCData() map[string]any {
	return map[string]any{
		"chapters": []map[string]any{
			{
				"title": "Getting Started",
				"id":    "chapter-1",
				"sections": []map[string]any{
					{
						"title": "Installation",
						"id":    "section-1-1",
						"subsections": []map[string]any{
							{"title": "Using Go Modules", "id": "subsec-1-1-1"},
							{"title": "Setting Up Your Project", "id": "subsec-1-1-2"},
						},
					},
					{
						"title": "Basic Concepts",
						"id":    "section-1-2",
						"subsections": []map[string]any{
							{"title": "Templates and Data", "id": "subsec-1-2-1"},
							{"title": "Rendering a Template", "id": "subsec-1-2-2"},
							{"title": "Working with Variables", "id": "subsec-1-2-3"},
						},
					},
					{
						"title": "Quick Tutorial",
						"id":    "section-1-3",
						"subsections": []map[string]any{
							{"title": "Creating Your First Template", "id": "subsec-1-3-1"},
							{"title": "Adding Dynamic Content", "id": "subsec-1-3-2"},
						},
					},
				},
			},
			{
				"title": "Core Features",
				"id":    "chapter-2",
				"sections": []map[string]any{
					{
						"title": "Template Syntax",
						"id":    "section-2-1",
						"subsections": []map[string]any{
							{"title": "Interpolation", "id": "subsec-2-1-1"},
							{"title": "Directives", "id": "subsec-2-1-2"},
							{"title": "Expressions", "id": "subsec-2-1-3"},
							{"title": "Filters and Pipes", "id": "subsec-2-1-4"},
						},
					},
					{
						"title": "Control Flow",
						"id":    "section-2-2",
						"subsections": []map[string]any{
							{"title": "Conditional Rendering", "id": "subsec-2-2-1"},
							{"title": "Loops and Iterations", "id": "subsec-2-2-2"},
						},
					},
					{
						"title": "Data Binding",
						"id":    "section-2-3",
						"subsections": []map[string]any{
							{"title": "Binding to Variables", "id": "subsec-2-3-1"},
							{"title": "Two-Way Binding", "id": "subsec-2-3-2"},
							{"title": "Computed Properties", "id": "subsec-2-3-3"},
						},
					},
					{
						"title": "Event Handling",
						"id":    "section-2-4",
						"subsections": []map[string]any{
							{"title": "Event Listeners", "id": "subsec-2-4-1"},
							{"title": "Event Modifiers", "id": "subsec-2-4-2"},
						},
					},
				},
			},
			{
				"title": "Advanced Techniques",
				"id":    "chapter-3",
				"sections": []map[string]any{
					{
						"title": "Component Architecture",
						"id":    "section-3-1",
						"subsections": []map[string]any{
							{"title": "Building Reusable Components", "id": "subsec-3-1-1"},
							{"title": "Component Composition", "id": "subsec-3-1-2"},
							{"title": "Props and Slots", "id": "subsec-3-1-3"},
						},
					},
					{
						"title": "State Management",
						"id":    "section-3-2",
						"subsections": []map[string]any{
							{"title": "Global State", "id": "subsec-3-2-1"},
							{"title": "Local Component State", "id": "subsec-3-2-2"},
						},
					},
					{
						"title": "Performance Optimization",
						"id":    "section-3-3",
						"subsections": []map[string]any{
							{"title": "Caching Strategies", "id": "subsec-3-3-1"},
							{"title": "Lazy Loading", "id": "subsec-3-3-2"},
							{"title": "Bundle Optimization", "id": "subsec-3-3-3"},
						},
					},
				},
			},
			{
				"title": "Best Practices",
				"id":    "chapter-4",
				"sections": []map[string]any{
					{
						"title": "Code Organization",
						"id":    "section-4-1",
						"subsections": []map[string]any{
							{"title": "Project Structure", "id": "subsec-4-1-1"},
							{"title": "Naming Conventions", "id": "subsec-4-1-2"},
						},
					},
					{
						"title": "Testing and Debugging",
						"id":    "section-4-2",
						"subsections": []map[string]any{
							{"title": "Unit Testing", "id": "subsec-4-2-1"},
							{"title": "Debugging Templates", "id": "subsec-4-2-2"},
						},
					},
				},
			},
		},
	}
}

// getTOCTemplate returns the table of contents template
func getTOCTemplate() []byte {
	return []byte(`<nav class="toc">
  <ol>
    <li v-for="chapter in chapters">
      <a href="#{{ chapter.id }}"><strong>{{ chapter.title }}</strong></a>
      <ol>
        <li v-for="section in chapter.sections">
          <a href="#{{ section.id }}">{{ section.title }}</a>
          <ol>
            <li v-for="subsection in section.subsections">
              <a href="#{{ subsection.id }}">{{ subsection.title }}</a>
            </li>
          </ol>
        </li>
      </ol>
    </li>
  </ol>
</nav>`)
}

// TestTemplate_RenderOutput_Equivalence verifies that all rendering methods
// produce identical output for the same template and data.
func TestTemplate_RenderOutput_Equivalence(t *testing.T) {
	fs := fstest.MapFS{
		"toc.html": &fstest.MapFile{
			Data: getTOCTemplate(),
		},
	}

	data := getTOCData()

	// Render using Render (file-based)
	tmpl1 := vuego.NewFS(fs).Fill(data)
	buf1 := &bytes.Buffer{}
	err := tmpl1.Load("toc.html").Render(context.Background(), buf1)
	require.NoError(t, err)
	output1 := buf1.String()

	// Render using RenderString
	tmpl2 := vuego.NewFS(fs).Fill(data)
	buf2 := &bytes.Buffer{}
	err = tmpl2.RenderString(context.Background(), buf2, string(getTOCTemplate()))
	require.NoError(t, err)
	output2 := buf2.String()

	// Render using RenderByte
	tmpl3 := vuego.NewFS(fs).Fill(data)
	buf3 := &bytes.Buffer{}
	err = tmpl3.RenderByte(context.Background(), buf3, getTOCTemplate())
	require.NoError(t, err)
	output3 := buf3.String()

	// Render using RenderReader
	tmpl4 := vuego.NewFS(fs).Fill(data)
	buf4 := &bytes.Buffer{}
	err = tmpl4.RenderReader(context.Background(), buf4, bytes.NewReader(getTOCTemplate()))
	require.NoError(t, err)
	output4 := buf4.String()

	// Verify all outputs are identical
	require.Equal(t, output1, output2, "RenderString should produce same output as Render")
	require.Equal(t, output1, output3, "RenderByte should produce same output as Render")
	require.Equal(t, output1, output4, "RenderReader should produce same output as Render")

	// Verify output is not empty and contains expected content
	require.Greater(t, len(output1), 0, "output should not be empty")
	require.Contains(t, output1, "Getting Started", "output should contain chapter title")
	require.Contains(t, output1, "Installation", "output should contain section")
	require.Contains(t, output1, "Using Go Modules", "output should contain subsection")
}

// BenchmarkTemplate_Render benchmarks rendering from a file.
func BenchmarkTemplate_Render(b *testing.B) {
	fs := fstest.MapFS{
		"toc.html": &fstest.MapFile{
			Data: getTOCTemplate(),
		},
	}

	data := getTOCData()
	tmpl := vuego.NewFS(fs).Fill(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := &strings.Builder{}
		_ = tmpl.Load("toc.html").Render(context.Background(), buf)
	}
}

// BenchmarkTemplate_RenderString benchmarks rendering from a string.
func BenchmarkTemplate_RenderString(b *testing.B) {
	fs := fstest.MapFS{}

	data := getTOCData()
	templateStr := string(getTOCTemplate())
	tmpl := vuego.NewFS(fs).Fill(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := &strings.Builder{}
		_ = tmpl.RenderString(context.Background(), buf, templateStr)
	}
}

// BenchmarkTemplate_RenderByte benchmarks rendering from bytes.
func BenchmarkTemplate_RenderByte(b *testing.B) {
	fs := fstest.MapFS{}

	data := getTOCData()
	templateBytes := getTOCTemplate()
	tmpl := vuego.NewFS(fs).Fill(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := &strings.Builder{}
		_ = tmpl.RenderByte(context.Background(), buf, templateBytes)
	}
}

// BenchmarkTemplate_RenderReader benchmarks rendering from a reader.
func BenchmarkTemplate_RenderReader(b *testing.B) {
	fs := fstest.MapFS{}

	data := getTOCData()
	templateBytes := getTOCTemplate()
	tmpl := vuego.NewFS(fs).Fill(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := &strings.Builder{}
		_ = tmpl.RenderReader(context.Background(), buf, bytes.NewReader(templateBytes))
	}
}

// BenchmarkTemplate_RenderAllocations measures memory allocations for each rendering method.
func BenchmarkTemplate_RenderAllocations(b *testing.B) {
	fs := fstest.MapFS{
		"toc.html": &fstest.MapFile{
			Data: getTOCTemplate(),
		},
	}

	data := getTOCData()

	b.Run("Render", func(b *testing.B) {
		tmpl := vuego.NewFS(fs).Fill(data)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := &strings.Builder{}
			_ = tmpl.Load("toc.html").Render(context.Background(), buf)
		}
	})

	b.Run("RenderString", func(b *testing.B) {
		templateStr := string(getTOCTemplate())
		tmpl := vuego.NewFS(fs).Fill(data)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := &strings.Builder{}
			_ = tmpl.RenderString(context.Background(), buf, templateStr)
		}
	})

	b.Run("RenderByte", func(b *testing.B) {
		templateBytes := getTOCTemplate()
		tmpl := vuego.NewFS(fs).Fill(data)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := &strings.Builder{}
			_ = tmpl.RenderByte(context.Background(), buf, templateBytes)
		}
	})

	b.Run("RenderReader", func(b *testing.B) {
		templateBytes := getTOCTemplate()
		tmpl := vuego.NewFS(fs).Fill(data)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := &strings.Builder{}
			_ = tmpl.RenderReader(context.Background(), buf, bytes.NewReader(templateBytes))
		}
	})
}
