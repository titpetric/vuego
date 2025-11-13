# vuego

[![Go Reference](https://pkg.go.dev/badge/github.com/titpetric/vuego.svg)](https://pkg.go.dev/github.com/titpetric/vuego)
[![Coverage](https://img.shields.io/badge/coverage-85.4%25-brightgreen.svg)](https://github.com/titpetric/vuego)

Vuego is a lightweight [Vue.js](https://vuejs.org/)-inspired template engine for Go. Render HTML templates with familiar Vue syntax - no JavaScript runtime required.

The engine uses [golang.org/x/net/html](https://pkg.go.dev/golang.org/x/net/html) for HTML parsing and rendering.

## Features

Vuego supports the following template features:

- **Variable interpolation** - `{{ variable }}` with nested property access
- **Attribute binding** - `:attr="value"` or `v-bind:attr="value"`
- **Conditional rendering** - `v-if` for showing/hiding elements
- **List rendering** - `v-for` to iterate over arrays with optional index
- **Raw HTML** - `v-html` for unescaped HTML content
- **Component composition** - `<vuego include>` for reusable components
- **Required props** - `:required` attribute for component validation
- **Template wrapping** - `<template>` tag for component boundaries
- **Full documents** - Support for complete HTML documents or fragments
- **Automatic escaping** - Built-in XSS protection for interpolated values

## Quick Start

### As a Go Package

```bash
go get github.com/titpetric/vuego
```

```go
package main

import (
    "os"
    "github.com/titpetric/vuego"
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

### As a CLI Tool

```bash
go install github.com/titpetric/vuego/cmd/vuego@latest
```

```bash
vuego template.vuego data.json > output.html
```

## Documentation

- **[CLI Usage](docs/cli.md)** - Command-line tool documentation with examples
- **[Template Syntax](docs/syntax.md)** - Complete syntax reference for all features
- **[Components](docs/components.md)** - Component composition and the `:required` attribute

## Quick Examples

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

See the [syntax documentation](docs/syntax.md) for complete examples and usage patterns.

## License

MIT
