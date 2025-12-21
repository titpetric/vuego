package formatter_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego/formatter"
)

func TestNewFormatter(t *testing.T) {
	f := formatter.NewFormatter()
	require.NotNil(t, f)
}

func TestFormatter_Format_FrontmatterPreserved(t *testing.T) {
	f := formatter.NewFormatter()

	content := `---
layout: base
title: "Test Page"
---

<div class="container">
  <h1>Hello</h1>
</div>`

	formatted, err := f.Format(content)
	require.NoError(t, err)
	require.Contains(t, formatted, "layout: base")
	require.Contains(t, formatted, "title: \"Test Page\"")
}

func TestFormatter_Format_SimpleHTML(t *testing.T) {
	f := formatter.NewFormatter()

	content := `<div><h1>Title</h1></div>`

	formatted, err := f.Format(content)
	require.NoError(t, err)
	require.NotEmpty(t, formatted)
}

func TestFormatter_Format_VuegoTemplate(t *testing.T) {
	f := formatter.NewFormatter()

	content := `<template v-for="item in items">
  <div>{{ item.name }}</div>
</template>`

	formatted, err := f.Format(content)
	require.NoError(t, err)
	require.NotEmpty(t, formatted)
}

func TestFormatter_Format_InsertsFinalNewline(t *testing.T) {
	f := formatter.NewFormatter()

	content := `<div>Content</div>`

	formatted, err := f.Format(content)
	require.NoError(t, err)
	require.True(t, len(formatted) > 0 && formatted[len(formatted)-1] == '\n',
		"formatted content should end with newline")
}

func TestFormatter_Format_NoFinalNewline(t *testing.T) {
	opts := formatter.FormatterOptions{
		IndentWidth: 2,
		InsertFinal: false,
	}
	f := formatter.NewFormatterWithOptions(opts)

	content := `<div>Content</div>`

	formatted, err := f.Format(content)
	require.NoError(t, err)
	require.False(t, len(formatted) > 0 && formatted[len(formatted)-1] == '\n',
		"formatted content should not end with newline when InsertFinal is false")
}

func TestFormatString(t *testing.T) {
	content := `<div><span>Test</span></div>`

	formatted, err := formatter.FormatString(content)
	require.NoError(t, err)
	require.NotEmpty(t, formatted)
}

func TestDefaultFormatterOptions(t *testing.T) {
	opts := formatter.DefaultFormatterOptions()
	require.Equal(t, 2, opts.IndentWidth)
	require.True(t, opts.InsertFinal)
}

func TestIndentString(t *testing.T) {
	text := "line1\nline2\nline3"
	indented := formatter.IndentString(text, 2, 2)

	expected := "    line1\n    line2\n    line3"

	require.Equal(t, expected, indented)
}

func TestIndentString_WithEmptyLines(t *testing.T) {
	text := "line1\n\nline3"
	indented := formatter.IndentString(text, 1, 2)
	require.Contains(t, indented, "line1")
	require.Contains(t, indented, "line3")
}

// Tests for formatter handling of partial fragments
func TestFormatter_Format_PartialFragment_SingleElement(t *testing.T) {
	f := formatter.NewFormatter()

	content := `<div class="card">
  <h1>Title</h1>
  <p>Content</p>
</div>`

	formatted, err := f.Format(content)
	require.NoError(t, err)
	// Should not wrap in extra divs
	require.True(t, len(formatted) > 0)
	require.Contains(t, formatted, "card")
	// Should not have outer wrapper
	require.NotContains(t, formatted, "<div>\n  <div class=\"card\">")
}

func TestFormatter_Format_PartialFragment_MultipleElements(t *testing.T) {
	f := formatter.NewFormatter()

	content := `<li v-for="item in items" :key="item.id">
  <span>{{ item.name }}</span>
</li>
<li v-for="item in items" :key="item.id">
  <span>{{ item.name }}</span>
</li>`

	formatted, err := f.Format(content)
	require.NoError(t, err)
	require.Contains(t, formatted, "<li")
	require.Contains(t, formatted, "v-for")
	lines := 0
	for i := 0; i < len(formatted); i++ {
		if formatted[i] == '\n' {
			lines++
		}
	}
	require.Greater(t, lines, 1, "should format multiple list items")
}

// Tests for formatter handling of full documents
func TestFormatter_Format_FullDocument_PreservesDoctype(t *testing.T) {
	f := formatter.NewFormatter()

	content := `<!DOCTYPE html>
<html lang="en">
<head>
  <title>Test</title>
</head>
<body>
  <h1>Hello</h1>
</body>
</html>`

	formatted, err := f.Format(content)
	require.NoError(t, err)
	require.True(t, len(formatted) > 0)
	require.Equal(t, "<!DOCTYPE html>", formatted[:15], "should preserve DOCTYPE")
	require.Contains(t, formatted, "<html")
}

func TestFormatter_Format_FullDocument_WithHead(t *testing.T) {
	f := formatter.NewFormatter()

	content := `<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>Test Page</title>
</head>
<body>
<h1>Content</h1>
</body>
</html>`

	formatted, err := f.Format(content)
	require.NoError(t, err)
	require.Contains(t, formatted, "<meta")
	require.Contains(t, formatted, "<title")
	require.Contains(t, formatted, "<!DOCTYPE")
}
