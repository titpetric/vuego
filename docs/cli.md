# Vuego CLI Documentation

The `vuego` command-line tool renders `.vuego` templates with JSON or YAML data and outputs HTML to stdout.

## Installation

```bash
go install github.com/titpetric/vuego/cmd/vuego@latest
```

Or build from source:

```bash
git clone https://github.com/titpetric/vuego
cd vuego
go build ./cmd/vuego
```

## Usage

```bash
vuego render <template.vuego> <data.json/yml>
```

### Arguments

- `<template.vuego>` - Path to the `.vuego` template file
- `<data.json/yml>` - Path to the JSON or YAML file containing template data

### Output

The rendered HTML is written to stdout. You can redirect it to a file:

```bash
vuego render template.vuego data.json > output.html
vuego render template.vuego data.yaml > output.html
```

## Basic Example

### Template File: `hello.vuego`

```html
<template>
  <h1>Hello, {{ name }}!</h1>
  <p>Welcome to {{ app }}.</p>
</template>
```

### Data File: `hello.json` (or `hello.yaml`)

JSON format:

```json
{
  "name": "World",
  "app": "Vuego"
}
```

YAML format:

```yaml
name: World
app: Vuego
```

### Command

```bash
vuego render hello.vuego hello.json
# or with YAML:
vuego render hello.vuego hello.yaml
```

### Output

```html
<h1>Hello, World!</h1>
<p>Welcome to Vuego.</p>
```

## Complete Example: Weather Forecast

This example demonstrates using the vuego CLI to render a weather forecast template with real API data.

### Template File: `forecast.vuego`

```html
<template>
	<h1>Weather telemetry for {{ latitude }} lat {{ longitude }}</h1>

	<p>Current weather is {{current.temperature_2m}}{{current_units.temperature_2m}}.</p>

	<p>The data is updated every {{current.interval}} {{current_units.interval}}.</p>
</template>
```

### Data File: `forecast.json`

```json
{
  "latitude": 45.22,
  "longitude": 13.599998,
  "generationtime_ms": 0.043272972106933594,
  "utc_offset_seconds": 0,
  "timezone": "GMT",
  "timezone_abbreviation": "GMT",
  "elevation": 0.0,
  "current_units": {
    "time": "iso8601",
    "interval": "seconds",
    "temperature_2m": "°C",
    "weather_code": "wmo code"
  },
  "current": {
    "time": "2025-11-13T20:30",
    "interval": 900,
    "temperature_2m": 13.1,
    "weather_code": 45
  }
}
```

### Command

```bash
vuego render forecast.vuego forecast.json
```

### Output

```html
<h1>Weather telemetry for 45.22 lat 13.599998</h1>
<p>Current weather is 13.1°C.</p>
<p>The data is updated every 900 seconds.</p>
```

## Advanced Example: Page with Components

When using components with `<vuego include>`, the CLI resolves component paths relative to the template file's directory.

### Directory Structure

```
pages/
├── index.vuego
├── components/
│   ├── Header.vuego
│   └── Footer.vuego
└── data/
    └── page.json
```

### Template File: `pages/index.vuego`

```html
<html>
  <head>
    <title>{{ title }}</title>
  </head>
  <body>
    <vuego include="components/Header.vuego"></vuego>
    
    <main>
      <h1>{{ heading }}</h1>
      <p>{{ content }}</p>
    </main>
    
    <vuego include="components/Footer.vuego"></vuego>
  </body>
</html>
```

### Component: `pages/components/Header.vuego`

```html
<template>
  <header>
    <nav>Site Navigation</nav>
  </header>
</template>
```

### Component: `pages/components/Footer.vuego`

```html
<template>
  <footer>
    <p>© 2024 My Site</p>
  </footer>
</template>
```

### Data File: `pages/data/page.json`

```json
{
  "title": "Home Page",
  "heading": "Welcome!",
  "content": "This is the main content of the page."
}
```

### Command

```bash
vuego render pages/index.vuego pages/data/page.json > index.html
```

### Output: `index.html`

```html
<html>
  <head>
    <title>Home Page</title>
  </head>
  <body>
    <header>
      <nav>Site Navigation</nav>
    </header>
    
    <main>
      <h1>Welcome!</h1>
      <p>This is the main content of the page.</p>
    </main>
    
    <footer>
      <p>© 2024 My Site</p>
    </footer>
  </body>
</html>
```

## Working with Loops and Conditionals

### Template: `products.vuego`

```html
<template>
  <h1>{{ store }} - Product Catalog</h1>
  
  <div v-for="item in products">
    <div class="product">
      <h2>{{ item.name }}</h2>
      <p>{{ item.description }}</p>
      <p class="price">${{ item.price }}</p>
      <p v-if="item.inStock" class="stock">In Stock</p>
      <p v-if="!item.inStock" class="stock">Out of Stock</p>
    </div>
  </div>
</template>
```

### Data: `products.json`

```json
{
  "store": "Tech Shop",
  "products": [
    {
      "name": "Laptop",
      "description": "High-performance laptop",
      "price": 1299.99,
      "inStock": true
    },
    {
      "name": "Mouse",
      "description": "Wireless optical mouse",
      "price": 29.99,
      "inStock": false
    },
    {
      "name": "Keyboard",
      "description": "Mechanical keyboard",
      "price": 149.99,
      "inStock": true
    }
  ]
}
```

### Command

```bash
vuego render products.vuego products.json
```

## Error Handling

### Missing Arguments

```bash
$ vuego render
Error: render: requires exactly 2 arguments
Usage: vuego render <file.vuego> <data.json/yml>
```

### Template Not Found

```bash
$ vuego render missing.vuego data.json
Error: opening template file: open missing.vuego: no such file or directory
```

### Data File Not Found

```bash
$ vuego render template.vuego missing.json
Error: reading data file: open missing.json: no such file or directory
```

### Invalid Data Format

```bash
$ vuego render template.vuego invalid.json
Error: parsing data file: invalid character '}' looking for beginning of value
```

### Missing Required Component Prop

```bash
$ vuego render page.vuego data.json
Error: rendering template: required attribute 'name' not provided
```

## Integration Examples

### Shell Script

```bash
#!/bin/bash
# generate-pages.sh

for template in templates/*.vuego; do
  basename=$(basename "$template" .vuego)
  vuego "$template" "data/${basename}.json" > "output/${basename}.html"
  echo "Generated output/${basename}.html"
done
```

### Makefile

```makefile
.PHONY: build
build: output/index.html output/about.html output/contact.html

output/%.html: templates/%.vuego data/%.json
	@mkdir -p output
	vuego $^ > $@
	@echo "Generated $@"
```

### Watch and Rebuild

Using `entr` for live reloading:

```bash
# Watch for changes and rebuild
ls templates/*.vuego data/*.json | entr -c vuego templates/index.vuego data/index.json
```

## Programmatic Usage

While this document focuses on CLI usage, vuego can also be used as a Go library:

```go
package main

import (
	"os"
	"path/filepath"

	"github.com/titpetric/vuego"
)

func main() {
	// Create filesystem from template directory
	templateFS := os.DirFS("templates")

	// Create Vue instance
	vue := vuego.NewVue(templateFS)

	// Prepare data
	data := map[string]any{
		"title": "Hello",
		"name":  "World",
	}

	// Render to stdout
	vue.Render(os.Stdout, "page.vuego", data)
}
```

See [components.md](components.md) for more details on component composition and the [main README](../README.md) for full template syntax documentation.

## Tips and Best Practices

1. **Use stdout redirection** - Redirect output to files for static site generation
2. **Validate data files first** - Use `jq` for JSON or `yamllint` for YAML before passing to vuego
3. **Version control templates** - Keep templates and example data in version control
4. **Organize by feature** - Group related templates and data files together
5. **Use Make or Task** - Automate builds with task runners for complex projects
6. **Test with real data** - Use actual API responses as test data files

## See Also

- [Component Documentation](components.md) - Learn about component composition
- [Template Syntax](../README.md) - Full template syntax reference
- [Go Package Documentation](https://pkg.go.dev/github.com/titpetric/vuego) - API reference
