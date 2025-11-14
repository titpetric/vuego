# Usage Examples

This guide provides practical examples of using Vuego in your applications.

## Quick Start Examples

### As a Go Package

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

### Using Typed Values (Structs)

Vuego supports passing typed values directly to `Render` and `RenderFragment`. These are accessible via the `data.` prefix in templates.

```go
type AppConfig struct {
	Title          string `json:"title"`
	Version        string `json:"version"`
	MaxConnections int    `json:"max_connections"`
}

func main() {
	templateFS := os.DirFS("templates")
	vue := vuego.NewVue(templateFS)

	// Pass struct directly - accessible as data.Title, data.Version, etc.
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
<h1>{{ data.Title }}</h1>
<p>Version: {{ data.version }}</p>
<!-- Both data.MaxConnections and data.max_connections work -->
<p>Max: {{ data.MaxConnections }}</p>
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
    <vuego include="components/Header.vuego"></vuego>
    <main>{{ content }}</main>
    <vuego include="components/Footer.vuego"></vuego>
  </body>
</html>
```

For complete syntax reference and more advanced patterns, see the [syntax documentation](syntax.md).
