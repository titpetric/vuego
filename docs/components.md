# Components in Vuego

This document covers component composition in Vuego using the `<vuego include>` tag, the `<template>` tag, and the `:required` attribute for validating component props.

## Table of Contents

- [Basic Component Composition](#basic-component-composition)
- [The Template Tag](#the-template-tag)
- [Required Attributes](#required-attributes)
- [Complete Examples](#complete-examples)

## Basic Component Composition

Vuego allows you to compose reusable components using the `<vuego include>` tag. This tag loads and renders another `.vuego` template file at the specified location.

### Syntax

```html
<vuego include="path/to/component.vuego"></vuego>
```

The path is relative to the filesystem root passed to `vuego.NewVue()`.

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
    <vuego include="components/Header.vuego"></vuego>
    <main>
      Content goes here
    </main>
  </body>
</html>
```

## The Template Tag

The `<template>` tag is a wrapper element that gets omitted from the final rendered output. It serves two purposes:

1. **Wrapping component content** - Allows multiple root elements in a component
2. **Defining component requirements** - Using the `:required` attribute

### Why Use Template?

Without a template tag, you might have multiple root elements which can complicate component structure. The template tag provides a clean wrapper that disappears during rendering.

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

Both produce the same output, but the template version is clearer about component boundaries.

## Required Attributes

The `:required` (or `:require`) attribute validates that component props are provided when the component is included. This helps catch missing data at render time.

### Syntax

```html
<template :require="propName">
  <!-- component content -->
</template>
```

You can specify multiple required attributes:

```html
<template :require="prop1" :require="prop2" :require="prop3">
  <!-- component content -->
</template>
```

### How It Works

1. When a component is included, attributes are passed as props
2. The template validates that all `:required` props are present
3. If a required prop is missing, rendering fails with an error
4. Props are available in the component scope for interpolation and binding

### Example: Button Component with Required Props

**components/Button.vuego**

```html
<template :require="name" :require="title">
  <button :name="name">{{ title }}</button>
</template>
```

**Usage:**

```html
<vuego include="components/Button.vuego" 
       name="submit-btn" 
       title="Click Me"></vuego>
```

**Rendered output:**

```html
<button name="submit-btn">Click Me</button>
```

**Error example (missing required prop):**

```html
<!-- This will fail because 'name' is missing -->
<vuego include="components/Button.vuego" title="Click Me"></vuego>
```

Error: `required attribute 'name' not provided`

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
    <vuego include="components/Header.vuego"></vuego>
    
    <main>
      <h2>Welcome!</h2>
      <p>This is the main content.</p>
    </main>
    
    <vuego include="components/Footer.vuego"></vuego>
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
    <vuego include="Button.vuego" 
           name="save" 
           title="Save"></vuego>
    <vuego include="Button.vuego" 
           name="cancel" 
           title="Cancel"></vuego>
  </div>
</template>
```

**page.vuego**

```html
<template>
  <div>
    <h2>Form Actions</h2>
    <vuego include="components/ButtonGroup.vuego"></vuego>
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
  
  <vuego include="components/WeatherCard.vuego" 
         location="{{ city }}" 
         temperature="{{ current.temp }}"></vuego>
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

3. **Use relative paths** - Keep component paths relative to your template filesystem root for portability

4. **Organize components in directories** - Use a `components/` directory structure for better organization

5. **Keep components focused** - Each component should have a single, clear responsibility

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

- [Template syntax documentation](../README.md)
- [Interpolation and data binding](../README.md)
- [Directives (v-if, v-for, v-html)](../README.md)
