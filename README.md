# vuego

[![Go Reference](https://pkg.go.dev/badge/github.com/titpetric/vuego.svg)](https://pkg.go.dev/github.com/titpetric/vuego) [![Coverage](https://img.shields.io/badge/coverage-81.20%25-brightgreen.svg)](https://github.com/titpetric/vuego)

Vuego is a [Vue.js](https://vuejs.org/)-inspired template engine for Go. Render HTML templates with familiar Vue syntax - no JavaScript runtime required.

## Status

Project it in active development. Some changes in API surface are still expected based on feedback and observations.

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
)
```

With this you have a `vuego.Template`. From here you can:

1. Construct new template rendering contexts with renderer.New and renderer.Load
2. Manage state with `Fill`, `Set` and `Get` functions
3. Finalize the rendering context with `Render` function
4. Additional renderers are provided to render from non-FS sources

An example of a rendering invocation would be:

```go
err = renderer.File(filename).Fill(data).Render(r.Context(), w)
```

`Render` automatically applies layouts when a front-matter `layout` hint is present. It will load layouts from `layouts/%s.vuego`. If no hint is provided, `layouts/base.vuego` is automatically used and passed the rendered content.

For rendering fragments (from variables or otherwise):

- `RenderFile(ctx context.Context, w io.Writer, filename string) error`
- `RenderString(ctx context.Context, w io.Writer, templateStr string) error`
- `RenderByte(ctx context.Context, w io.Writer, templateData []byte) error`
- `RenderReader(ctx context.Context, w io.Writer, r io.Reader) error`

An example of a rendering invocation is:

```go
err = renderer.New().Fill(data).RenderString(r.Context(), w, `<li v-for="item in items">{{ item.title }}</li>`)
```

There's more detail around authoring vuego templates, check documentation:

- [API Reference - docs/api.md](./docs/api.md)
- [Go Reference - pkg.go.dev](https://pkg.go.dev/github.com/titpetric/vuego)

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

Try Vuego in your browser! The interactive playground lets you experiment with templates and see results instantly.

- [vuego.incubator.to](https://vuego.incubator.to/)

You can run the playground locally, and if you pass a folder as the first parameter, you have a workspace.

```bash
# go run github.com/titpetric/vuego/cmd/vuego-playground
2025/11/14 20:07:31 Vuego Playground starting on http://localhost:3000
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
