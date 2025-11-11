# vuego

Vuego (pronounced "fuego") is a template engine based on the document
object model. It is inspired by [vue.js](https://vuejs.org/) syntax and
tries to support a subset of their template engine syntax.

The engine is implemented in pure go, without the use of javascript
virtual machines like goja. The project was a quick attempt at achieving
a functional subset of compatible template engine syntax but have it's
own runtime in Go, without the complexity of virtual machines.

It's made possible thanks to [golang.org/x/net/html](https://pkg.go.dev/golang.org/x/net/html).

## Usage

To use vuego as a package:

```bash
go get github.com/titpetric/vuego
```

```go
package views

import (
	"github.com/titpetric/vuego"
)

func RenderTemplate(w io.Writer, template string, data map[string]any) error {
	return vuego.Render(w, template, data)
}
```

Invoke `vuego.Render(io.Writer, template string, data map[string]any) error` to render a template.

You can also use vuego as a command line tool:

```bash
go install github.com/titpetric/vuego/cmd/vuego@latest
```

Usage: `vuego file.tpl file.json` will render HTML to stdout.

## Supported syntax

Vuego implements a subset of VueJS syntax:

- Loops with `v-for`
- Conditions with `v-if`
- Attributes `:attr="value"`, `v-bind`
- Set inner HTML: `v-html`
- Variable interpolation: `<p>Hello, {{user.name}}</p>`

There's no support for other VueJS syntax.

## Examples

### Variable interpolation

Use `{{ var }}` to insert a value from the current scope.

```html
<h1>{{ title }}</h1>
<p>Hello, {{ user.name }}!</p>
```

### Attribute binding (`:attr` / `v-bind:`)

Bind attributes to expressions.

```html
<a :href="url" title="{{ tooltip }}">Link</a>
```

### Conditional rendering (`v-if`)

Render elements only when the expression is truthy.

```html
<div v-if="show">Visible</div>
<div v-if="hide">Hidden</div>
```

### List rendering (`v-for`)

Iterate over arrays or slices.

```html
<ul>
  <li v-for="item in items">{{ item.name }}</li>
</ul>
```

With index:

```html
<li v-for="(i, v) in list">{{ i }} - {{ v }}</li>
```

### Raw HTML (`v-html`)

Insert unescaped HTML.

```html
<div v-html="content"></div>
```

### Full document template

Supports complete HTML documents or fragments.

```html
<html>
  <head><title>{{ title }}</title></head>
  <body>
    <ul>
      <li v-for="item in items">{{ item.name }}</li>
    </ul>
  </body>
</html>
```

## Notes

- Interpolation uses `{{ expr }}` (no `{expr}` support).  
- `v-if`, `v-for`, `v-html`, and `v-bind` / `:attr` follow Vue-like semantics.  
- Whitespace differences are ignored when testing or comparing DOM.

Other things that this doesn't support:

- expression evaluation, math, string concatenation of values
- custom comparisons and boolean expressions in `v-if`

Basic XSS security is provided with a `html.EscapeString` (quoting).

This doesn't support strong typed arguments, however, one could use
[usepzaka/structmap](https://pkg.go.dev/github.com/usepzaka/structmap)
or other similar packages that map structs to a `map[string]any`.
