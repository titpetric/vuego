# Components in Vuego

This document covers component composition in Vuego using the `<template include>` tag, component shorthands, and the `:required` attribute for validating component props.

## Table of Contents

- [Component Shorthands](#component-shorthands)
- [Basic Component Composition](#basic-component-composition)
- [Slots](#slots)
- [The Template Tag](#the-template-tag)
- [Required Attributes](#required-attributes)
- [YAML Front-Matter for Single File Components](#yaml-front-matter-for-single-file-components)
- [Complete Examples](#complete-examples)

## Component Shorthands

Vuego carries a shorthand for using a component without writing the full `<template include>` tag. It lets a custom tag stand for a component file, which the loader maps for you.

### Enabling Component Shorthands

To use component shorthands, enable them when creating your template renderer:

```go
package main

import (
	"github.com/titpetric/vuego"
	"os"
)

func main() {
	root := os.DirFS("templates")

	// Enable component shorthands with the default pattern
	tpl := vuego.NewFS(root, vuego.WithComponents())

	// Or specify custom patterns
	tpl := vuego.NewFS(root, vuego.WithComponents("components/*.vuego", "ui/*.vuego"))
}
```

### Default Behavior

Calling `WithComponents()` with no arguments loads every component under `components/*.vuego`, turning each filename into a kebab-case tag name.

### Shorthand Syntax

Component shorthands use custom HTML tags that match your component names (converted to kebab-case):

**File: components/ButtonPrimary.vuego**

```html
<button class="btn-primary">
  <slot>Click me</slot>
</button>
```

**Usage with shorthand:**

```html
<button-primary>Submit</button-primary>
```

**Rendered output:**

```html
<button class="btn-primary">Submit</button>
```

**Usage without content (uses fallback):**

```html
<button-primary></button-primary>
```

**Rendered output with fallback:**

```html
<button class="btn-primary">Click me</button>
```

**The same thing written in full:**

```html
<template include="components/ButtonPrimary.vuego"></template>
```

### How It Works

1. A glob such as `components/*.vuego` collects the component files
2. The loader turns the filename, minus its extension, from PascalCase into kebab-case:
   - `ButtonPrimary.vuego` becomes `<button-primary>`
   - `AlertBox.vuego` becomes `<alert-box>`
   - `MyComponent.vuego` becomes `<my-component>`
3. A shorthand tag the renderer meets becomes a `<template include>` directive
4. Each attribute on the tag reaches the component as a context variable

### Passing Attributes

Each attribute on a component shorthand tag reaches the component as a context variable:

**components/Badge.vuego:**

```html
<template :require="type" :require="text">
  <span class="badge badge-{{ type }}">{{ text }}</span>
</template>
```

**Usage:**

```html
<badge type="success" text="Approved"></badge>
```

**Rendered output:**

```html
<span class="badge badge-success">Approved</span>
```

### Custom Glob Patterns

You can use custom glob patterns to load components from different directories:

```go
// Load components from multiple directories
tpl := vuego.NewFS(root, vuego.WithComponents(
	"components/*.vuego",
	"ui/buttons/*.vuego",
	"ui/modals/*.vuego",
))
```

With this setup:
- Files in `components/` use the format `<component-name>`
- Files in `ui/buttons/` use the format `<button-name>`
- Files in `ui/modals/` use the format `<modal-name>`

## Basic Component Composition

Vuego allows you to compose reusable components using the `<template include>` tag. This tag loads and renders another `.vuego` template file at the specified location.

### Syntax

```html
<template include="path/to/component.vuego"></template>
```

The path sits relative to the filesystem root handed to `vuego.NewVue()`.

### Example: Including a Header Component

**components/Header.vuego**

```html
<template>
  <header>
    This is the header
  </header>
</template>
```

**index.vuego**

```html
<html>
  <head>
    <title>My Page</title>
  </head>
  <body>
    <template include="components/Header.vuego"></template>
    <main>
      Content goes here
    </main>
  </body>
</html>
```

## Slots

A slot marks a hole in a component template that the including template fills. Vuego carries default slots, named slots and scoped slots.

### Default Slots

The simplest form of slot - parent content goes into a single unnamed slot:

**components/Card.vuego**

```html
<div class="card">
  <slot></slot>
</div>
```

**Usage:**

```html
<card>
  <h2>My Card Title</h2>
  <p>Card content here</p>
</card>
```

**Rendered output:**

```html
<div class="card">
  <h2>My Card Title</h2>
  <p>Card content here</p>
</div>
```

### Named Slots

Give each slot in a component a name to use more than one:

**components/Modal.vuego**

```html
<div class="modal">
  <div class="modal-header">
    <slot name="header"></slot>
  </div>
  <div class="modal-body">
    <slot></slot>
  </div>
  <div class="modal-footer">
    <slot name="footer"></slot>
  </div>
</div>
```

**Usage with v-slot:name syntax:**

```html
<modal>
  <template v-slot:header>Confirm Action</template>
  <template v-slot:footer>
    <button>Cancel</button>
    <button>Confirm</button>
  </template>
  Are you sure you want to proceed?
</modal>
```

**Shorthand with #name:**

```html
<modal>
  <template #header>Confirm Action</template>
  <template #footer>
    <button>Cancel</button>
    <button>Confirm</button>
  </template>
  Are you sure you want to proceed?
</modal>
```

### Scoped Slots

Pass data from the child component to the parent's slot template using `:prop="expression"` bindings:

**components/List.vuego**

```html
<ul>
  <li v-for="(index, item) in items">
    <slot :item="item" :index="index"></slot>
  </li>
</ul>
```

**Parent receives props:**

```html
<list :items="products">
  <template v-slot="slotProps">
    <strong>{{ slotProps.item.name }}</strong> ({{ slotProps.index }})
  </template>
</list>
```

**With destructuring:**

```html
<list :items="products">
  <template v-slot="{ item, index }">
    <strong>{{ item.name }}</strong> ({{ index }})
  </template>
</list>
```

### Named Scoped Slots

Combine named slots with scoped props:

**components/DataTable.vuego**

```html
<table>
  <thead>
    <slot name="header" :columns="columns"></slot>
  </thead>
  <tbody>
    <tr v-for="row in rows">
      <slot name="row" :row="row"></slot>
    </tr>
  </tbody>
</table>
```

**Parent component:**

```html
<data-table :columns="cols" :rows="data">
  <template #header="{ columns }">
    <tr>
      <th v-for="col in columns">{{ col }}</th>
    </tr>
  </template>
  <template #row="{ row }">
    <td>{{ row.name }}</td>
    <td>{{ row.value }}</td>
  </template>
</data-table>
```

### Fallback Content

Put fallback content inside a `<slot>`; it renders when the caller supplies none:

**components/Button.vuego**

```html
<button class="btn">
  <slot>Click me</slot>
</button>
```

**Usage:**

```html
<!-- With content -->
<button>Submit</button>
<!-- Renders: <button class="btn">Submit</button> -->

<!-- Without content -->
<button></button>
<!-- Renders: <button class="btn">Click me</button> (uses fallback) -->
```

## The Template Tag

The `<template>` tag wraps content and leaves no trace in the rendered output. It does two jobs:

1. **Wrapping component content** - lets a component carry more than one root element
2. **Defining component requirements** - Using the `:required` attribute

### Why Use Template?

Without a template tag a component can end up with several root elements, which muddles its structure. The template tag wraps them and disappears during rendering.

**Without template (valid but less organized):**

```html
<h1>Title</h1>
<p>Description</p>
```

**With template (recommended):**

```html
<template>
  <h1>Title</h1>
  <p>Description</p>
</template>
```

Both produce the same output; the template version states the component boundary more plainly.

## Required Attributes

The `:required` attribute, spelled `:require` too, checks that the caller supplies each prop when it includes the component. It catches missing data at render time.

### Syntax

```html
<template :require="propName">
  <!-- component content -->
</template>
```

Name several required attributes at once:

```html
<template :require="prop1" :require="prop2" :require="prop3">
  <!-- component content -->
</template>
```

### How It Works

1. Including a component hands its attributes over as props
2. The template checks that every `:required` prop arrived
3. A missing required prop fails the render with an error
4. The component scope holds each prop for interpolation and binding

### Example: Button Component with Required Props

**components/Button.vuego**

```html
<template :require="name" :require="title">
  <button :name="name">{{ title }}</button>
</template>
```

**Usage:**

```html
<template include="components/Button.vuego" 
       name="submit-btn" 
       title="Click Me"></template>
```

**Rendered output:**

```html
<button name="submit-btn">Click Me</button>
```

**Error example (missing required prop):**

```html
<!-- This will fail because 'name' is missing -->
<template include="components/Button.vuego" title="Click Me"></template>
```

Error: `required attribute 'name' not provided`

## YAML Front-Matter for Single File Components

Vuego reads YAML front matter at the head of a `.vuego` file, which lets the template file carry its own component data, the way a single-file component does elsewhere.

### Syntax

Place YAML between `---` delimiters at the start of your template file:

```html
---
title: Component Title
count: 42
items:
  - apple
  - banana
---
<template>
  <div>
    <h1>{{ title }}</h1>
    <p>Count: {{ count }}</p>
  </div>
</template>
```

### How It Works

1. The loader strips the front matter and parses it as YAML
2. Each variable joins the template context
3. Front matter holds **the final say**: it overrides whatever `Render()` or `RenderFragment()` received
4. Each included component can have its own front-matter

### Use Cases

Reach for front matter to carry:
- Defining default/static component data
- Storing component metadata
- Simplifying single-purpose components that don't need external data
- Creating self-contained components with built-in defaults

### Example: Static Component with Front-Matter

**components/Card.vuego**

```html
---
backgroundColor: "#f5f5f5"
borderColor: "#ddd"
---
<template>
  <div style="background: {{ backgroundColor }}; border: 1px solid {{ borderColor }};">
    <slot></slot>
  </div>
</template>
```

### Example: Component with Default Data

**components/Hero.vuego**

```html
---
title: "Welcome to Our Site"
subtitle: "Your tagline here"
cta-text: "Get Started"
background-image: "/images/hero.jpg"
---
<template>
  <section class="hero" style="background-image: url({{ background-image }})">
    <h1>{{ title }}</h1>
    <p>{{ subtitle }}</p>
    <button>{{ cta-text }}</button>
  </section>
</template>
```

YAML keys with hyphens work directly in templates (e.g., `background-image`, `data-value`, `aria-label`).

### Front-Matter Overrides Passed Data

Front matter wins over the attributes an include carries:

**components/Config.vuego**

```html
---
environment: "production"
debug: false
---
<div>Environment: {{ environment }}</div>
```

**Usage with conflicting data:**

```html
<template include="components/Config.vuego" environment="development" debug="true"></template>
```

**Output:**

```html
<div>Environment: production</div>
```

The front-matter values (`production`, `false`) override the passed attributes (`development`, `true`).

## Complete Examples

### Example 1: Page Layout with Components

**components/Header.vuego**

```html
<template>
  <header>
    <h1>My Website</h1>
    <nav>Navigation here</nav>
  </header>
</template>
```

**components/Footer.vuego**

```html
<template>
  <footer>
    <p>© 2024 My Website</p>
  </footer>
</template>
```

**index.vuego**

```html
<html>
  <head>
    <title>Home Page</title>
  </head>
  <body>
    <template include="components/Header.vuego"></template>
    
    <main>
      <h2>Welcome!</h2>
      <p>This is the main content.</p>
    </main>
    
    <template include="components/Footer.vuego"></template>
  </body>
</html>
```

### Example 2: Nested Components with Props

**components/Button.vuego**

```html
<template :require="name" :require="title">
  <button :name="name">{{ title }}</button>
</template>
```

**components/ButtonGroup.vuego**

```html
<template>
  <div class="button-group">
    <template include="Button.vuego" 
           name="save" 
           title="Save"></template>
    <template include="Button.vuego" 
           name="cancel" 
           title="Cancel"></template>
  </div>
</template>
```

**page.vuego**

```html
<template>
  <div>
    <h2>Form Actions</h2>
    <template include="components/ButtonGroup.vuego"></template>
  </div>
</template>
```

### Example 3: Component with Dynamic Data

**components/WeatherCard.vuego**

```html
<template :require="temperature" :require="location">
  <div class="weather-card">
    <h3>{{ location }}</h3>
    <p class="temp">{{ temperature }}°C</p>
  </div>
</template>
```

**forecast.vuego**

```html
<template>
  <h1>Weather Forecast</h1>
  
  <template include="components/WeatherCard.vuego" 
         location="{{ city }}" 
         temperature="{{ current.temp }}"></template>
</template>
```

With JSON data:

```json
{
  "city": "San Francisco",
  "current": {
    "temp": "18.5"
  }
}
```

**Rendered output:**

```html
<h1>Weather Forecast</h1>

<div class="weather-card">
  <h3>San Francisco</h3>
  <p class="temp">18.5°C</p>
</div>
```

## Best Practices

1. **Always wrap components with `<template>`** - Even if you have a single root element, it clarifies component boundaries

2. **Use `:require` for essential props** - If your component cannot function without certain data, mark it as required

3. **Use front-matter for static data** - For components with built-in defaults or static data, use YAML front-matter to keep them self-contained

4. **Use relative paths** - Keep component paths relative to your template filesystem root for portability

5. **Organize components in directories** - Use a `components/` directory structure for better organization

6. **Keep components focused** - Each component should have a single, clear responsibility

7. **Front matter holds the final say** - it overrides the attributes an include carries, so keep data a caller may want to change out of it

## Component File Structure Example

```
pages/
├── index.vuego
├── forecast.vuego
└── components/
    ├── Header.vuego
    ├── Footer.vuego
    ├── Button.vuego
    └── WeatherCard.vuego
```

## See Also

- [Data Loading](data-loading.md) - Automatic config loading from `theme.yml` and `data/`
- [Theme Structuring Guide](themes.md) - Layout organization and directory conventions
- [Template Syntax](syntax.md) - Variable interpolation and directives
- [FuncMap & Filters](funcmap.md) - Template functions and custom filters
