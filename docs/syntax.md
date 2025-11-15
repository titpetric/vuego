# Template Syntax Reference

Vuego implements a subset of Vue.js template syntax. This document provides a complete reference for all supported syntax features.

## Table of Contents

- [Values](#values)
- [Directives](#directives)
- [Components](#components)
- [Advanced](#advanced)
- [Notes and Limitations](#notes-and-limitations)

## Values

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

### Expressions and Operators

Interpolation supports comparison operators, logical operators, and ternary expressions:

```html
<!-- Comparisons -->
<span>{{ status == "active" ? "Online" : "Offline" }}</span>
<span>{{ age >= 18 ? "Adult" : "Minor" }}</span>

<!-- Logical operators -->
<span>{{ isAdmin && hasPermission ? "Allowed" : "Denied" }}</span>
<span>{{ isEmpty || isNull ? "No data" : "Has data" }}</span>

<!-- Negation -->
<span>{{ !disabled ? "Enabled" : "Disabled" }}</span>
```

### Filters and Pipes

Transform values using filters with the pipe syntax:

```html
<!-- Single filter -->
<p>{{ name | upper }}</p>

<!-- Chained filters -->
<p>{{ name | upper | title }}</p>

<!-- Filter with arguments -->
<p>{{ items | default("No items") }}</p>

<!-- Complex expressions in pipes -->
<p>{{ price | . > 100 ? "Expensive" : "Affordable" }}</p>
```

**Available Filters:**
- `upper` - Convert to uppercase
- `lower` - Convert to lowercase
- `title` - Capitalize first letter of each word
- `len` - Get length of string, array, or map
- `trim` - Remove leading/trailing whitespace
- `escape` - HTML escape the value
- `default(fallback)` - Use fallback if value is nil or empty
- `formatTime([layout])` - Format time value
- `int` - Convert to integer
- `string` - Convert to string
- `json` - Marshal value to JSON (unescaped in `<script>` tags, escaped elsewhere)

#### JSON Filter

The `json` filter marshals a value to JSON format. Its behavior differs depending on the context:

**In `<script>` tags:** JSON is output unescaped, allowing safe embedding of data in JavaScript:

```html
<script>
  const user = {{ user | json }};
  const items = {{ items | json }};
</script>
```

**In other contexts** (`<pre>`, `<code>`, fragments, etc.): JSON is HTML-escaped to prevent XSS:

```html
<pre><code>{{ data | json }}</code></pre>
```

This automatic escaping prevents malicious content in the JSON from breaking out of HTML context, while allowing safe direct use in JavaScript where JSON has its own syntax safety.

## Directives

Directives are special attributes that apply dynamic behavior to elements.

| Directive                | Description                                        |
|--------------------------|----------------------------------------------------|
| `:attr` or `v-bind:attr` | Bind HTML attributes to expressions                |
| `v-if`                   | Conditionally render elements based on expressions |
| `v-else-if`              | Alternative condition for `v-if`                   |
| `v-else`                 | Fallback render when all previous conditions fail  |
| `v-for`                  | Iterate over arrays with optional index            |
| `v-html`                 | Render unescaped HTML content                      |
| `v-show`                 | Toggle element visibility with CSS display         |
| `v-pre`                  | Skip template processing for element and children  |

### Attribute Binding (`:attr` / `v-bind:attr`)

Bind HTML attributes to dynamic values:

```html
<a :href="url" :title="tooltip">Link</a>
<img :src="imageUrl" :alt="imageAlt">
<button :name="buttonName" :disabled="isDisabled">Click</button>
```

Shorthand `:attr` is equivalent to `v-bind:attr`.

#### Object Binding for `class` and `style`

Bind to object literals to conditionally apply classes or styles. This is useful for dynamic styling based on data conditions.

**Class Object Binding:**

Object keys become class names, and are included only when their values are truthy:

```html
<!-- Bootstrap-style example: apply btn-primary only when isPrimary is true -->
<button :class="{'btn': true, 'btn-primary': isPrimary, 'btn-large': isLarge}">
  Click me
</button>

<!-- Combine with static classes -->
<div class="card" :class="{'card-active': isActive, 'card-disabled': isDisabled}"></div>

<!-- Boolean flags -->
<div :class="{show: isVisible, hidden: isHidden, error: hasError}"></div>
```

Falsey values (including `0`, `false`, `""`, `nil`) exclude the class. Any other value includes it:

```html
<div :class="{alert: itemCount}"></div>
<!-- With itemCount=0: no class applied -->
<!-- With itemCount=5: class="alert" applied -->
```

**Style Object Binding:**

Object keys use camelCase and are automatically converted to kebab-case CSS properties. Values are applied as-is:

```html
<!-- camelCase properties are converted to kebab-case -->
<div :style="{color: textColor, fontSize: size, backgroundColor: bgColor}"></div>
<!-- Output: style="color:textColor;font-size:size;background-color:bgColor;" -->

<!-- Combine with static styles (dynamic style overwrites conflicting static style) -->
<div style="padding: 10px;" :style="{color: errorColor, fontWeight: 'bold'}"></div>

<!-- Static string values -->
<div :style="{display: 'block', color: 'red', marginTop: '20px'}"></div>
```

**camelCase to kebab-case conversion:**

CSS property names must use kebab-case (e.g., `font-size`, `background-color`). To match JavaScript convention, use camelCase in the object binding and they will be automatically converted:

| camelCase         | converts to        | CSS property     |
|-------------------|--------------------|------------------|
| `fontSize`        | `font-size`        | font-size        |
| `backgroundColor` | `background-color` | background-color |
| `fontWeight`      | `font-weight`      | font-weight      |
| `marginTop`       | `margin-top`       | margin-top       |
| `borderRadius`    | `border-radius`    | border-radius    |

Properties that already contain hyphens (e.g., custom CSS properties like `--my-color`) are left unchanged.

### Conditional Rendering (`v-if`, `v-else-if`, `v-else`)

Render elements only when conditions are met. Use `v-else-if` and `v-else` to provide alternative branches:

#### Basic `v-if`

```html
<div v-if="show">Visible when true</div>
<div v-if="!hide">Visible when false</div>
<p v-if="status == 'active'">Active</p>
<p v-if="age >= 18">Adult</p>
<p v-if="isAdmin && hasPermission">Admin access</p>
```

#### Chained Conditions with `v-else-if` and `v-else`

Use `v-else-if` to test multiple conditions and `v-else` as the default fallback:

```html
<div v-if="status == 'pending'">
  <p class="warning">Your request is pending</p>
</div>
<div v-else-if="status == 'approved'">
  <p class="success">Your request was approved</p>
</div>
<div v-else-if="status == 'rejected'">
  <p class="error">Your request was rejected</p>
</div>
<div v-else>
  <p class="info">Status unknown</p>
</div>
```

#### Empty List with `v-for` and `v-else`

Use `v-else` after a `v-for` loop to display a "no results" message when the list is empty. Note that `v-else` must be an immediate sibling to the `v-for` element:

```html
<ul>
  <li v-for="item in items">{{ item.name }}</li>
  <li v-else>No results found</li>
</ul>
```

### List Rendering (`v-for`)

Iterate over arrays or slices:

```html
<!-- Basic iteration -->
<li v-for="item in items">{{ item.name }}</li>

<!-- With index -->
<li v-for="(i, item) in items">{{ i }} - {{ item.name }}</li>

<!-- Nested loops -->
<div v-for="category in categories">
  <h2>{{ category.name }}</h2>
  <ul>
    <li v-for="item in category.items">{{ item }}</li>
  </ul>
</div>
```

### Raw HTML (`v-html`)

Insert unescaped HTML content:

```html
<div v-html="htmlContent"></div>
```

**⚠️ Warning:** Only use with trusted content; user-provided content can lead to XSS vulnerabilities.

### Visibility Control (`v-show`)

Control element visibility with CSS without removing from DOM:

```html
<!-- v-show: toggles display property based on condition -->
<div v-show="isVisible">Visible when true</div>
<div v-show="user.isAdmin">Admin content</div>
```

The `v-show` directive uses CSS `display` property for visibility control, keeping elements in the DOM. Use this instead of `v-if` when you want to preserve component state or avoid repeated rendering.

### Skip Template Processing (`v-pre`)

Prevent template processing for an element and its children:

```html
<div v-pre>
  {{ variableName }} and {{ expression }}
</div>
```

Useful for displaying template syntax as literal text in documentation or code examples.

## Components

Vuego supports component composition using `<vuego>` and `<template>` tags. See [Components Guide](components.md) for detailed examples.

### `<vuego>` Tag (Include/Embed Components)

Include reusable component files:

```html
<vuego include="components/Header.vuego"></vuego>
<vuego include="components/Button.vuego" name="submit" title="Submit Form"></vuego>
```

The `include` attribute specifies the template file path. Additional attributes are passed as props to the component.

### `<template>` Tag

Wrap component content or fragments:

```html
<template>
  <h1>{{ title }}</h1>
  <p>{{ content }}</p>
</template>
```

For component validation, use the `:required` attribute on the `<template>` tag to ensure props are provided:

```html
<template :required="name,title">
  <h1>{{ title }}</h1>
  <p>Name: {{ name }}</p>
</template>
```

Props listed in `:required` must be provided when the component is included, or rendering will fail with a validation error.

## Advanced

### Template Functions and Filters

VueGo supports custom template functions and pipe-based filter chaining. See [FuncMap & Filters](funcmap.md) for detailed documentation on:
- Setting custom functions with `FuncMap`
- Built-in utility functions
- Function signatures and type conversion
- Error handling in template functions

### Full Document Templates

Vuego supports both complete HTML documents and fragments:

```html
<!-- Complete document -->
<html>
  <head>
    <title>{{ pageTitle }}</title>
  </head>
  <body>
    <h1>{{ heading }}</h1>
  </body>
</html>

<!-- Fragment with template tag -->
<template>
  <h1>{{ title }}</h1>
  <p>{{ content }}</p>
</template>
```

Both styles are valid and work identically.

## Notes and Limitations

### What Vuego Supports

- ✅ Variable interpolation with `{{ expr }}`
- ✅ Nested property access with dot notation
- ✅ Expressions (comparisons, logical operators, ternary)
- ✅ Attribute binding with `:attr` and `v-bind:attr`
- ✅ Conditional rendering with `v-if`, `v-else-if`, and `v-else`
- ✅ Visibility control with `v-show`
- ✅ List iteration with `v-for` (with optional index)
- ✅ Raw HTML with `v-html`
- ✅ Skip template processing with `v-pre`
- ✅ Component composition with `<vuego include>`
- ✅ Component prop validation with `:required`
- ✅ Full HTML documents and fragments
- ✅ Custom template functions and filters

### What Vuego Does NOT Support

- ❌ Event handling (`@click`, `v-on`)
- ❌ Two-way binding (`v-model`)
- ❌ Computed properties
- ❌ Slots
- ❌ Custom directives

### Security

- HTML escaping is automatically applied to interpolated values using `html.EscapeString`
- This provides basic XSS protection for `{{ }}` interpolations
- `v-html` bypasses escaping - only use with trusted content

### Data Types

Vuego accepts `map[string]any` as data input. For strongly-typed data structures, convert them to maps before rendering.

### Whitespace

- Empty text nodes (whitespace-only) are omitted from output
- Whitespace differences are ignored when comparing DOM structures
- Indentation is preserved within non-empty text nodes

## See Also

- [CLI Documentation](cli.md) - Command-line usage
- [Components](components.md) - Detailed component composition guide
- [FuncMap & Filters](funcmap.md) - Template functions and custom filters
- [Main README](../README.md) - Overview and quick start
