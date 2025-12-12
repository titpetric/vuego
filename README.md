# vuego

[![Go Reference](https://pkg.go.dev/badge/github.com/titpetric/vuego.svg)](https://pkg.go.dev/github.com/titpetric/vuego) [![Coverage](https://img.shields.io/badge/coverage-81.20%25-brightgreen.svg)](https://github.com/titpetric/vuego)

Vuego is a [Vue.js](https://vuejs.org/)-inspired template engine for Go. Render HTML templates with familiar Vue syntax - no JavaScript runtime required.

## Quick start

You can use vuego by importing it in your project:

```
import "github.com/titpetric/vuego"
```

In your service you create a new vuego renderer.

```go
renderer := vuego.New(
	vuego.WithFS(os.DirFS("templates")),
	vuego.WithFuncs(funcMap),
),
```

Using the fileystem is optional and you can use `vuego.New()` to omit it.

Providing and rendering a template is done in one or two ways:

- `renderer.New().Fill(data).Render(context.Context, io.Writer) error


## Features

Vuego supports the following template features:

- **Variable interpolation** - `{{ variable }}` with nested property access
- **Template functions** - Pipe-based filter chaining `{{ value | upper | trim }}`
- **Custom functions** - Define custom template functions (FuncMap)
- **Built-in filters** - String manipulation, formatting, type conversion
- **Attribute binding** - `:attr="value"` or `v-bind:attr="value"`
- **Conditional rendering** - `v-if` for showing/hiding elements with function support
- **List rendering** - `v-for` to iterate over arrays with optional index
- **Raw HTML** - `v-html` for unescaped HTML content
- **Component composition** - `<vuego include>` for reusable components
- **Required props** - `:required` attribute for component validation
- **Template wrapping** - `<template>` tag for component boundaries
- **Full documents** - Support for complete HTML documents or fragments
- **Automatic escaping** - Built-in XSS protection for interpolated values

## Playground

Try VueGo in your browser! The interactive playground lets you experiment with templates and see results instantly.

- [vuego.incubator.to](https://vuego.incubator.to/)

You can run the playground locally, and if you pass a folder as the first parameter, you have a workspace.

```bash
# go run github.com/titpetric/vuego/cmd/vuego-playground
2025/11/14 20:07:31 VueGo Playground starting on http://localhost:3000
```

## Documentation

The Template Syntax reference covers four main areas:
- **[Values](docs/syntax.md#values)** - Variable interpolation, expressions, and filters
- **[Directives](docs/syntax.md#directives)** - Complete reference of all `v-` directives
- **[Components](docs/syntax.md#components)** - `<vuego>` and `<template>` tags
- **[Advanced](docs/syntax.md#advanced)** - Template functions, custom filters, and full documents

Additional resources:
- **[FuncMap & Filters](docs/funcmap.md)** - Custom template functions and built-in filters
- **[Components Guide](docs/components.md)** - Detailed component composition examples
- **[Testing](docs/testing.md)** - Running tests and interpreting results
- **[CLI Usage](docs/cli.md)** - Command-line tool reference
- **[Concurrency](docs/concurrency.md)** - Thread-safety and concurrent rendering

## License

MIT
