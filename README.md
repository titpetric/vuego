# Vuego - A Go based VueJS syntax template engine

[![Go Reference](https://pkg.go.dev/badge/github.com/titpetric/vuego.svg)](https://pkg.go.dev/github.com/titpetric/vuego) [![Coverage](https://img.shields.io/badge/coverage-92.20%25-brightgreen.svg)](https://github.com/titpetric/vuego)

Vuego brings a [Vue.js](https://vuejs.org/)-inspired template engine to Go. Render HTML templates in familiar Vue syntax, with no JavaScript runtime.

You can try Vuego:

- In your browser: The Vuego Tour [vuego.incubator.to](https://vuego.incubator.to/)
- In your terminal: The Vuego CLI [titpetric/vuego-cli](https://github.com/titpetric/vuego-cli)

The Vuego CLI also ships as a docker image, which runs the tour.

## About

Vuego renders templates on the server. Its runtime picks up template edits while the application keeps running. It carries composition and layouts, and aims to serve as the single runtime for a codebase that writes Go on both the front end and the back end.

The choices Vuego makes:

- A VueJS-compatible syntax covering `v-if`, `v-for` and interpolation
- Composition with the usage of the html `<template>` tag (unique syntax)
- DOM driven templating, type safety based on generics and reflection
- YAML/JSON as the default way to provide template view rendering

A machine readable data format lets vuego support the development lifecycle through the [vuego-cli](https://github.com/titpetric/vuego-cli) tool, which runs a development playground. Three steps:

- start `vuego-cli serve ./templates`
- edit .json and .vuego files
- refresh to view changes

The project remains under active development. The API surface may still shift as feedback and observations arrive. Each release carries a tag.

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

For more detail on authoring vuego templates, check the documentation:

- [API Reference - docs/api.md](./docs/api.md)
- [Go Reference - pkg.go.dev](https://pkg.go.dev/github.com/titpetric/vuego)

## Vuego CLI

With [titpetric/vuego-cli](https://github.com/titpetric/vuego-cli) you can do the following:

- `vuego-cli serve <folder>` - load a development server for templates,
- `vuego-cli tour <folder>` - load a learning tour of vuego,
- `vuego-cli docs <folder>` - load a documentation server.

Beside those learning tools, separate commands `fmt`, `lint` and `render` files.

## Documentation

The Template Syntax reference covers four main areas:

- **[Values](docs/syntax.md#values)** - Variable interpolation, expressions, and filters
- **[Directives](docs/syntax.md#directives)** - Complete reference of all `v-` directives
- **[Components](docs/syntax.md#components)** - `<template include>` and `<template>` tags
- **[Advanced](docs/syntax.md#advanced)** - Template functions, custom filters, and full documents

Further resources:

- **[FuncMap & Filters](docs/funcmap.md)** - Custom template functions and built-in filters
- **[Components Guide](docs/components.md)** - Detailed component composition examples
- **[Data Loading](docs/data-loading.md)** - Automatic config loading from `theme.yml` and `data/`
- **[Theme Structuring Guide](docs/themes.md)** - Layout organization, naming conventions, and theme types
- **[Testing](docs/testing.md)** - Running tests and interpreting results
- **[Concurrency](docs/concurrency.md)** - Thread-safety and concurrent rendering
