package markdown_test

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"testing/fstest"

	"github.com/titpetric/vuego/markdown"
	"github.com/titpetric/vuego/tests/assert"
)

// Sample markdown content similar to alert.md, shared with the benchmarks in
// markdown_test.go.
var alertMD = `---
title: Alert
subtitle: Displays a callout for user attention
layout: page
---

## Usage

Use the ` + "`alert`" + ` or ` + "`alert-destructive`" + ` class name.

` + "```html" + `
<div class="alert">
  <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
    <circle cx="12" cy="12" r="10" />
    <path d="m9 12 2 2 4-4" />
  </svg>
  <h2>Success! Your changes have been saved</h2>
  <section>This is an alert with icon, title and description.</section>
</div>
` + "```" + `

The component has the following HTML structure:

Use ` + "`alert`" + ` for default styling or ` + "`alert-destructive`" + ` for error states.

- ` + "`<div class=\"alert\">`" + ` - Main container.
  - ` + "`<svg>`" + ` - optional, the icon
  - ` + "`<h2>`" + ` - the title
  - ` + "`<section>`" + ` - optional, the description

## Examples

### Destructive

This is a paragraph with *emphasis* and **strong** text and ` + "`inline code`" + `.

### No description

Another paragraph with a [link](https://example.com) and more text.

### No icon

Final paragraph here.
`

// Simple markdown for comparison.
var simpleMD = `# Hello World

This is a paragraph.
`

// Medium complexity markdown.
var mediumMD = `## Overview

The ` + "`button`" + ` component provides styled buttons.

- Item one with ` + "`code`" + `
- Item two with *emphasis*
- Item three with **strong**

### Usage

Use the button like this:

` + "```html" + `
<button class="btn">Click me</button>
` + "```" + `
`

func loadDoc(t *testing.T, content string) *markdown.Document {
	t.Helper()

	md := markdown.New(fstest.MapFS{"test.md": {Data: []byte(content)}})
	doc, err := md.Load("test.md")
	assert.NoError(t, err)
	return doc
}

func TestDocument_FrontMatter(t *testing.T) {
	t.Run("returns the parsed YAML", func(t *testing.T) {
		fm := loadDoc(t, "---\ntitle: Hello\ntags: [a, b]\n---\n# Body").FrontMatter()
		assert.Equal(t, "Hello", fm["title"])
		assert.Equal(t, []any{"a", "b"}, fm["tags"])
	})

	t.Run("returns nil when there is none", func(t *testing.T) {
		assert.Nil(t, loadDoc(t, "# Body").FrontMatter())
	})

	t.Run("returns nil for an empty block", func(t *testing.T) {
		assert.Len(t, loadDoc(t, "---\n---\n# Body").FrontMatter(), 0)
	})
}

func TestDocument_Render(t *testing.T) {
	t.Run("writes the rendered HTML", func(t *testing.T) {
		var buf bytes.Buffer
		assert.NoError(t, loadDoc(t, simpleMD).Render(&buf))
		assert.Contains(t, buf.String(), "<h1")
		assert.Contains(t, buf.String(), "Hello World")
		assert.Contains(t, buf.String(), "<p>")
	})

	t.Run("an empty document writes nothing", func(t *testing.T) {
		var buf bytes.Buffer
		assert.NoError(t, loadDoc(t, "").Render(&buf))
		assert.Equal(t, "", buf.String())
	})

	t.Run("renders the same output twice", func(t *testing.T) {
		doc := loadDoc(t, mediumMD)

		var first, second bytes.Buffer
		assert.NoError(t, doc.Render(&first))
		assert.NoError(t, doc.Render(&second))
		assert.Equal(t, first.String(), second.String())
	})

	t.Run("reports a writer error", func(t *testing.T) {
		assert.Error(t, loadDoc(t, simpleMD).Render(failingWriter{}))
	})
}

func TestDocument_RenderContext(t *testing.T) {
	t.Run("passes the context to handlers", func(t *testing.T) {
		type key struct{}

		var got any
		md := markdown.New(fstest.MapFS{"test.md": {Data: []byte("# Title")}})
		md.Register(func(ctx context.Context, _ *markdown.Node) (string, bool) {
			got = ctx.Value(key{})
			return "", true
		}, markdown.NodeHeading)

		doc, err := md.Load("test.md")
		assert.NoError(t, err)

		var buf bytes.Buffer
		assert.NoError(t, doc.RenderContext(context.WithValue(context.Background(), key{}, "value"), &buf))
		assert.Equal(t, "value", got)
	})

	t.Run("renders the same HTML as Render", func(t *testing.T) {
		var withCtx, withoutCtx bytes.Buffer
		assert.NoError(t, loadDoc(t, mediumMD).RenderContext(context.Background(), &withCtx))
		assert.NoError(t, loadDoc(t, mediumMD).Render(&withoutCtx))
		assert.Equal(t, withoutCtx.String(), withCtx.String())
	})
}

var errWriteFailed = errors.New("write failed")

// failingWriter fails every write, so a render error surfaces.
type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errWriteFailed
}

func BenchmarkLoad(b *testing.B) {
	md := markdown.New(fstest.MapFS{"test.md": {Data: []byte(alertMD)}})

	for b.Loop() {
		_, _ = md.Load("test.md")
	}
}

func BenchmarkDocumentRender(b *testing.B) {
	md := markdown.New(fstest.MapFS{"test.md": {Data: []byte(alertMD)}})
	doc, _ := md.Load("test.md")

	for b.Loop() {
		var buf bytes.Buffer
		_ = doc.Render(&buf)
	}
}

func BenchmarkLoadAndRender(b *testing.B) {
	sources := []struct {
		name string
		src  string
	}{
		{name: "simple", src: simpleMD},
		{name: "medium", src: mediumMD},
		{name: "alert", src: alertMD},
	}

	for _, source := range sources {
		b.Run(source.name, func(b *testing.B) {
			md := markdown.New(fstest.MapFS{"test.md": {Data: []byte(source.src)}})

			for b.Loop() {
				doc, _ := md.Load("test.md")
				var buf bytes.Buffer
				_ = doc.Render(&buf)
			}
		})
	}
}
