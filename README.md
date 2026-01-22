# Vuego

[![Go Reference](https://pkg.go.dev/badge/github.com/titpetric/vuego.svg)](https://pkg.go.dev/github.com/titpetric/vuego) [![Coverage](https://img.shields.io/badge/coverage-81.20%25-brightgreen.svg)](https://github.com/titpetric/vuego)

Vuego is a [Vue.js](https://vuejs.org/)-inspired template engine for Go. Render HTML templates with familiar Vue syntax - no JavaScript runtime required.

You can try Vuego:

- In your browser: The Vuego Tour [vuego.incubator.to](https://vuego.incubator.to/)
- In your terminal: The Vuego CLI [titpetric/vuego-cli](https://github.com/titpetric/vuego-cli)

The Vuego CLI is also available as a docker image and runs the tour.

## About

Vuego is a server-side template rendering engine. It provides a template
runtime that allows you to modify templates as the application is
running. It allows for composition, supports layouts and aims to be an
exclusive runtime where Go is used for front and back-end development.

The choices Vuego makes:

- A largely VueJS compatible syntax with `v-if`, `v-for` and interpolation
- Composition with the usage of the html `<template>` tag (unique syntax)
- DOM driven templating, type safety based on generics and reflection
- YAML/JSON as the default way to provide template view rendering

Especially with using a machine readable data format, vuego aims to support
the development lifecycle with the [vuego-cli](https://github.com/titpetric/vuego-cli) tool.
The tool allows you to run a development playground. The process is simple:

- start `vuego-cli serve ./templates`
- edit .json and .vuego files
- refresh to view changes

Project it in active development. Some changes in API surface are still
expected based on feedback and observations. Releases are tagged.

## Quick start

You can use vuego by importing it in your project:

```go
import "github.com/titpetric/vuego"
```

In your service you create a new vuego renderer.

```go
// load your templates from a folder.
// the argument can be swapped for an embed.FS value.
renderer := vuego.NewFS(os.DirFS("templates"))

if err := renderer.File(filename).Fill(data).Render(r.Context(), w); err != nil {
	return err
}
```

You can provide your own type safe wrappers:

```go
func (v *Views) IndexView(data IndexData) vuego.Template {
	return vuego.View[IndexData](v.renderer, "index.vuego", data)
}
```

And you don't need to chain the functions:

```go
renderer := vuego.NewFS(embedFS)

tpl := renderer.New()
tpl.Fill(data)
tpl.Assign("meta", meta)
err := tpl.Render(ctx, out)
```

If you want to render from a string and not a filesystem:

```go
renderer := vuego.New()
err = renderer.New().Fill(data).RenderString(r.Context(), w, `<li v-for="item in items">{{ item.title }}</li>`)
```

There's more detail around authoring vuego templates, check documentation:

- [API Reference - docs/api.md](./docs/api.md)
- [Go Reference - pkg.go.dev](https://pkg.go.dev/github.com/titpetric/vuego)

## Vuego CLI

With [titpetric/vuego-cli](https://github.com/titpetric/vuego-cli) you can do the following:

- `vuego-cli serve <folder>` - load a development server for templates,
- `vuego-cli tour <folder>` - load a learning tour of vuego,
- `vuego-cli docs <folder>` - load a documentation server.

In addition to those learning tools, separate commands are provided to
`fmt`, `lint` and `render` files.

## Documentation

The Template Syntax reference covers four main areas:

- **[Values](docs/syntax.md#values)** - Variable interpolation, expressions, and filters
- **[Directives](docs/syntax.md#directives)** - Complete reference of all `v-` directives
- **[Components](docs/syntax.md#components)** - `<template include>` and `<template>` tags
- **[Advanced](docs/syntax.md#advanced)** - Template functions, custom filters, and full documents

Additional resources:

- **[FuncMap & Filters](docs/funcmap.md)** - Custom template functions and built-in filters
- **[Components Guide](docs/components.md)** - Detailed component composition examples
- **[Testing](docs/testing.md)** - Running tests and interpreting results
- **[Concurrency](docs/concurrency.md)** - Thread-safety and concurrent rendering
