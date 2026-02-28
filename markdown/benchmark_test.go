package markdown_test

import (
	"bytes"
	"testing"
	"testing/fstest"

	"github.com/titpetric/vuego/markdown"
)

// Sample markdown content similar to alert.md
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

// Simple markdown for comparison
var simpleMD = `# Hello World

This is a paragraph.
`

// Medium complexity markdown
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

// BenchmarkRenderSimple tests a minimal markdown file
func BenchmarkRenderSimple(b *testing.B) {
	contentFS := fstest.MapFS{"test.md": {Data: []byte(simpleMD)}}
	md := markdown.New(contentFS)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc, _ := md.Load("test.md")
		var buf bytes.Buffer
		_ = doc.Render(&buf)
	}
}

// BenchmarkRenderMedium tests medium complexity markdown
func BenchmarkRenderMedium(b *testing.B) {
	contentFS := fstest.MapFS{"test.md": {Data: []byte(mediumMD)}}
	md := markdown.New(contentFS)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc, _ := md.Load("test.md")
		var buf bytes.Buffer
		_ = doc.Render(&buf)
	}
}

// BenchmarkRenderAlert tests the alert.md-like content
func BenchmarkRenderAlert(b *testing.B) {
	contentFS := fstest.MapFS{"test.md": {Data: []byte(alertMD)}}
	md := markdown.New(contentFS)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc, _ := md.Load("test.md")
		var buf bytes.Buffer
		_ = doc.Render(&buf)
	}
}

// BenchmarkRenderBytes tests RenderBytes directly (no Load overhead)
func BenchmarkRenderBytes(b *testing.B) {
	md := markdown.New(nil)
	src := []byte(alertMD)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_ = md.RenderBytes(&buf, src)
	}
}

// BenchmarkRenderBytesSimple tests RenderBytes with simple content
func BenchmarkRenderBytesSimple(b *testing.B) {
	md := markdown.New(nil)
	src := []byte(simpleMD)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_ = md.RenderBytes(&buf, src)
	}
}

// BenchmarkLoad tests just the Load operation
func BenchmarkLoad(b *testing.B) {
	contentFS := fstest.MapFS{"test.md": {Data: []byte(alertMD)}}
	md := markdown.New(contentFS)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = md.Load("test.md")
	}
}

// BenchmarkDocRender tests just the Render operation (after Load)
func BenchmarkDocRender(b *testing.B) {
	contentFS := fstest.MapFS{"test.md": {Data: []byte(alertMD)}}
	md := markdown.New(contentFS)
	doc, _ := md.Load("test.md")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		_ = doc.Render(&buf)
	}
}
