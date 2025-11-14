# Template Syntax Reference

Vuego implements a subset of Vue.js template syntax. This document provides a complete reference for all supported syntax features.

## Table of Contents

- [Variable Interpolation](#variable-interpolation)
- [Attribute Binding](#attribute-binding)
- [Conditional Rendering](#conditional-rendering)
- [List Rendering](#list-rendering)
- [Raw HTML](#raw-html)
- [Full Document Templates](#full-document-templates)
- [Components](#components)
- [Notes and Limitations](#notes-and-limitations)

## Variable Interpolation

Use `{{ var }}` to insert a value from the current scope. Values are automatically HTML-escaped for security.

### Basic Usage

```html
<h1>{{ title }}</h1>
<p>Hello, {{ user.name }}!</p>
```

### Nested Properties

Access nested object properties using dot notation:

```html
<p>Location: {{ user.address.city }}, {{ user.address.country }}</p>
<p>Temperature: {{ current.temperature_2m }}{{ current_units.temperature_2m }}</p>
```

### Example with Data

**Template:**

```html
<template>
  <h1>{{ title }}</h1>
  <p>Welcome, {{ user.name }}!</p>
  <p>Email: {{ user.email }}</p>
</template>
```

**Data:**

```json
{
  "title": "Dashboard",
  "user": {
    "name": "John Doe",
    "email": "john@example.com"
  }
}
```

**Output:**

```html
<h1>Dashboard</h1>
<p>Welcome, John Doe!</p>
<p>Email: john@example.com</p>
```

## Attribute Binding

Bind HTML attributes to expressions using `:attr` shorthand or `v-bind:attr` syntax.

### Shorthand Syntax (`:attr`)

```html
<a :href="url" :title="tooltip">Link</a>
<img :src="imageUrl" :alt="imageAlt">
<button :name="buttonName" :disabled="isDisabled">Click</button>
```

### Combining Static and Dynamic Attributes

```html
<a :href="url" class="link" title="{{ tooltip }}">Link</a>
```

### Example with Data

**Template:**

```html
<template>
  <a :href="link.url" :target="link.target">{{ link.text }}</a>
  <img :src="image" :alt="description">
</template>
```

**Data:**

```json
{
  "link": {
    "url": "https://example.com",
    "target": "_blank",
    "text": "Visit Example"
  },
  "image": "/logo.png",
  "description": "Company Logo"
}
```

**Output:**

```html
<a href="https://example.com" target="_blank">Visit Example</a>
<img src="/logo.png" alt="Company Logo">
```

## Conditional Rendering

Render elements only when the expression is truthy using the `v-if` directive.

**Important:** Only a single boolean parameter is supported. Use `v-if="name"` for truthy checks or `v-if="!name"` to negate.

### Basic Usage

```html
<div v-if="show">Visible when 'show' is true</div>
<div v-if="!hide">Visible when 'hide' is false</div>
<p v-if="user">User is logged in</p>
```

### Negation with `!`

```html
<p v-if="!disabled">This element is enabled</p>
<div v-if="!loading">Content loaded</div>
```

### Checking Boolean Properties

```html
<p v-if="product.inStock" class="available">In Stock</p>
<p v-if="user.isAdmin">Admin Panel</p>
<div v-if="!product.discontinued">Still available</div>
```

### Example with Data

**Template:**

```html
<template>
  <h1>Product: {{ product.name }}</h1>
  <p v-if="product.onSale">Sale Price: ${{ product.salePrice }}</p>
  <p v-if="product.inStock">Available Now</p>
  <p v-if="product.featured">Featured Item</p>
</template>
```

**Data:**

```json
{
  "product": {
    "name": "Laptop",
    "onSale": true,
    "salePrice": 999,
    "inStock": true,
    "featured": false
  }
}
```

**Output:**

```html
<h1>Product: Laptop</h1>
<p>Sale Price: $999</p>
<p>Available Now</p>
```

### Limitations

- No support for `v-else` or `v-else-if`
- No custom comparisons (e.g., `v-if="count > 5"` is not supported)
- No boolean expressions (e.g., `v-if="a && b"` is not supported)
- Only truthy/falsy evaluation is supported

## List Rendering

Iterate over arrays or slices using the `v-for` directive.

### Basic Array Iteration

```html
<ul>
  <li v-for="item in items">{{ item.name }}</li>
</ul>
```

### With Index

Access the index and value using tuple syntax:

```html
<ul>
  <li v-for="(i, item) in items">{{ i }} - {{ item.name }}</li>
</ul>
```

### Nested Properties

```html
<div v-for="user in users">
  <h2>{{ user.name }}</h2>
  <p>{{ user.email }}</p>
  <p>{{ user.address.city }}</p>
</div>
```

### Combining with Other Directives

```html
<div v-for="product in products">
  <h3>{{ product.name }}</h3>
  <p v-if="product.inStock">In Stock</p>
  <img :src="product.image" :alt="product.name">
</div>
```

### Example with Data

**Template:**

```html
<template>
  <h1>{{ store }} - Products</h1>
  <div v-for="product in products">
    <h2>{{ product.name }}</h2>
    <p>{{ product.description }}</p>
    <p class="price">${{ product.price }}</p>
    <span v-if="product.inStock">Available</span>
  </div>
</template>
```

**Data:**

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
      "description": "Wireless mouse",
      "price": 29.99,
      "inStock": false
    }
  ]
}
```

**Output:**

```html
<h1>Tech Shop - Products</h1>
<div>
  <h2>Laptop</h2>
  <p>High-performance laptop</p>
  <p class="price">$1299.99</p>
  <span>Available</span>
</div>
<div>
  <h2>Mouse</h2>
  <p>Wireless mouse</p>
  <p class="price">$29.99</p>
</div>
```

### Nested Loops

```html
<div v-for="category in categories">
  <h2>{{ category.name }}</h2>
  <ul>
    <li v-for="item in category.items">{{ item }}</li>
  </ul>
</div>
```

## Raw HTML

Insert unescaped HTML content using the `v-html` directive.

### Basic Usage

```html
<div v-html="htmlContent"></div>
```

### Example with Data

**Template:**

```html
<template>
  <div class="article">
    <h1>{{ title }}</h1>
    <div v-html="content"></div>
  </div>
</template>
```

**Data:**

```json
{
  "title": "My Article",
  "content": "<p>This is <strong>bold</strong> and <em>italic</em> text.</p>"
}
```

**Output:**

```html
<div class="article">
  <h1>My Article</h1>
  <div>
    <p>This is <strong>bold</strong> and <em>italic</em> text.</p>
  </div>
</div>
```

### Security Warning

⚠️ **Warning:** Using `v-html` with user-provided content can lead to XSS vulnerabilities. Only use `v-html` with trusted content.

## Full Document Templates

Vuego supports both complete HTML documents and fragments.

### Complete HTML Document

```html
<html>
  <head>
    <title>{{ pageTitle }}</title>
    <meta name="description" :content="description">
  </head>
  <body>
    <h1>{{ heading }}</h1>
    <ul>
      <li v-for="item in items">{{ item.name }}</li>
    </ul>
  </body>
</html>
```

### HTML Fragment

```html
<template>
  <h1>{{ title }}</h1>
  <p>{{ content }}</p>
</template>
```

Both styles are valid and work identically. Use `<template>` for fragments or components.

## Components

See [components.md](components.md) for detailed documentation on component composition.

### Basic Component Include

```html
<html>
  <head>
    <title>Vuego index example</title>
  </head>
  <body>
    <vuego include="components/Header.vuego"></vuego>
    <vuego include="IndexContent.vuego"></vuego>
    <vuego include="components/Footer.vuego"></vuego>
  </body>
</html>
```

### Component with Props

```html
<vuego include="components/Button.vuego" 
       name="submit" 
       title="Submit Form"></vuego>
```

## Notes and Limitations

### What Vuego Supports

- ✅ Variable interpolation with `{{ expr }}`
- ✅ Nested property access with dot notation
- ✅ Attribute binding with `:attr` and `v-bind:attr`
- ✅ Conditional rendering with `v-if` (truthy/falsy evaluation)
- ✅ List iteration with `v-for`
- ✅ Index access in loops with `(i, v)` syntax
- ✅ Raw HTML with `v-html`
- ✅ Component composition with `<vuego include>`
- ✅ Required component props with `:required`
- ✅ Full HTML documents and fragments

### What Vuego Does NOT Support

- ❌ Expression evaluation (e.g., `{{ 1 + 1 }}`, `{{ name.toUpperCase() }}`)
- ❌ String concatenation in templates (e.g., `{{ firstName + ' ' + lastName }}`)
- ❌ Custom comparisons in `v-if` (e.g., `v-if="count > 5"`)
- ❌ Boolean expressions in `v-if` (e.g., `v-if="a && b"`)
- ❌ `v-else` and `v-else-if` directives
- ❌ `v-show` directive
- ❌ Event handling (`@click`, `v-on`)
- ❌ Two-way binding (`v-model`)
- ❌ Computed properties
- ❌ Methods and filters
- ❌ Slots
- ❌ Custom directives

### Security

- HTML escaping is automatically applied to interpolated values using `html.EscapeString`
- This provides basic XSS protection for `{{ }}` interpolations
- `v-html` bypasses escaping - only use with trusted content

### Data Types

Vuego accepts `map[string]any` as data input. For strongly-typed data structures, you can use packages like [usepzaka/structmap](https://pkg.go.dev/github.com/usepzaka/structmap) to convert structs to maps.

**Example:**

```go
type User struct {
	Name  string
	Email string
}

user := User{Name: "John", Email: "john@example.com"}
data := structmap.Map(user) // Convert to map[string]any
```

### Whitespace

Whitespace handling in vuego follows these rules:

- Empty text nodes (whitespace-only) are omitted from output
- Whitespace differences are ignored when comparing DOM structures
- Indentation is preserved within non-empty text nodes

## See Also

- [CLI Documentation](cli.md) - Command-line usage
- [Component Documentation](components.md) - Component composition
- [Main README](../README.md) - Overview and quick start
