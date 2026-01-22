# Usage Examples

This guide provides practical examples of using Vuego in your applications.

## Quick Start Examples

### Using the Vue Renderer (Standard API)

```bash
go get github.com/titpetric/vuego
```

```go
package main

import (
	"github.com/titpetric/vuego"
	"os"
)

func main() {
	// Create filesystem from template directory
	templateFS := os.DirFS("templates")

	// Create Vue instance
	vue := vuego.NewVue(templateFS)

	// Render template with data
	data := map[string]any{
		"title": "Hello World",
		"user": map[string]any{
			"name": "John Doe",
		},
	}

	if err := vue.Render(os.Stdout, "page.vuego", data); err != nil {
		panic(err)
	}
}
```

### Using the Template API (Stateful)

The Template API provides a stateful alternative to the Vue renderer, useful for workflows involving front-matter data and layout composition.

```go
package main

import (
	"github.com/titpetric/vuego"
	"os"
)

func main() {
	// Load a template with front-matter
	templateFS := os.DirFS("templates")

	tmpl, err := vuego.Load(templateFS, "page.vuego")
	if err != nil {
		panic(err)
	}

	// Access front-matter variables (e.g., layout, title)
	layout := tmpl.GetVar("layout")
	title := tmpl.GetVar("title")

	// Assign additional variables with method chaining
	tmpl.Assign("author", "John Doe").Assign("date", "2024-01-01")

	// Render with buffering (output writer unmodified on error)
	if err := tmpl.Render(os.Stdout); err != nil {
		panic(err)
	}
}
```

**Template file example** (`page.vuego`):

```yaml
---
layout: layouts/base.html
title: My Page
---
<article>
  <h1>{{ title }}</h1>
  <p>By {{ author }}</p>
  <div>{{ content }}</div>
</article>
```

The Template API is particularly useful for:
- Accessing layout information from front-matter
- Building composite layouts (render page, then wrap in layout template)
- Stateful template workflows where you load once and render multiple times with different data
- Safe output buffering that guarantees the writer is unmodified on rendering errors

### Using Typed Values (Structs)

Vuego supports passing typed values directly to `Render` and `RenderFragment`. Struct fields are accessible directly by their field names or JSON tags, without requiring the `data.` prefix.

```go
type AppConfig struct {
	Title          string `json:"title"`
	Version        string `json:"version"`
	MaxConnections int    `json:"max_connections"`
}

func main() {
	templateFS := os.DirFS("templates")
	vue := vuego.NewVue(templateFS)

	// Pass struct directly - accessible as Title, Version, MaxConnections, etc.
	config := &AppConfig{
		Title:          "My App",
		Version:        "1.0.0",
		MaxConnections: 100,
	}

	if err := vue.Render(os.Stdout, "config.vuego", config); err != nil {
		panic(err)
	}
}
```

In the template, access fields using both Go field names and JSON tags:

```html
<h1>{{ Title }}</h1>
<p>Version: {{ version }}</p>
<!-- Both MaxConnections and max_connections work -->
<p>Max: {{ MaxConnections }}</p>
```

### As a CLI Tool

```bash
go install github.com/titpetric/vuego/cmd/vuego@latest
```

```bash
vuego template.vuego data.json > output.html
```

## Template Examples

### Variable Interpolation

```html
<h1>{{ title }}</h1>
<p>Hello, {{ user.name }}!</p>
```

### List Rendering

```html
<ul>
  <li v-for="item in items">{{ item.name }}</li>
</ul>
```

### Component Composition

```html
<html>
  <body>
    <template include="components/Header.vuego"></template>
    <main>{{ content }}</main>
    <template include="components/Footer.vuego"></template>
  </body>
</html>
```

For complete syntax reference and more advanced patterns, see the [syntax documentation](syntax.md).
