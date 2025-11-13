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
