# Template Syntax Reference

Vuego implements a subset of Vue.js template syntax. This document provides a complete reference for all supported syntax features.

## Table of Contents

- [Values](#values)
- [Directives](#directives)
  - [Slots](#slots)
- [Components](#components)
- [Advanced](#advanced)
- [Notes and Limitations](#notes-and-limitations)

## Values

Use `{{ var }}` to insert a value from the current scope. The renderer HTML-escapes every value it writes.

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
- `default(fallback)` - Falls back when the value holds nil or an empty string
- `formatTime([layout])` - Format time value
- `int` - Convert to integer
- `string` - Convert to string
- `json` - Marshal value to JSON (unescaped in `<script>` tags, escaped elsewhere)

#### JSON Filter

The `json` filter marshals a value to JSON format. Its behavior differs depending on the context:

**In `<script>` tags:** the renderer writes JSON raw, so the data lands as JavaScript:

```html
<script>
  const user = {{ user | json }};
  const items = {{ items | json }};
</script>
```

**In other contexts** (`<pre>`, `<code>`, fragments): the renderer HTML-escapes the JSON, which blocks XSS:

```html
<pre><code>{{ data | json }}</code></pre>
```

This automatic escaping prevents malicious content in the JSON from breaking out of HTML context, while allowing safe direct use in JavaScript where JSON has its own syntax safety.

## Directives

A directive reads as an attribute and gives the element dynamic behavior.

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
| `v-once`                 | Render element once and skip it on later renders   |
| `v-slot` or `#`          | Define slot content and scoped slot props          |

### Attribute Binding (`:attr` / `v-bind:attr`)

Bind HTML attributes to dynamic values:

```html
<a :href="url" :title="tooltip">Link</a>
<img :src="imageUrl" :alt="imageAlt">
<button :name="buttonName" :disabled="isDisabled">Click</button>
```

The shorthand `:attr` means the same as `v-bind:attr`.

#### Object Binding for `class` and `style`

Bind to an object literal to apply classes or styles under a condition, which lets the data drive the styling.

**Class Object Binding:**

Object keys become class names, and the renderer writes a key when its value holds something truthy:

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

Object keys use camelCase, which the renderer converts to kebab-case CSS properties. It writes each value unchanged:

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

CSS property names must use kebab-case, as in `font-size` and `background-color`. Write camelCase in the object binding to match the JavaScript convention, and the renderer converts each one:

| camelCase         | converts to        | CSS property     |
|-------------------|--------------------|------------------|
| `fontSize`        | `font-size`        | font-size        |
| `backgroundColor` | `background-color` | background-color |
| `fontWeight`      | `font-weight`      | font-weight      |
| `marginTop`       | `margin-top`       | margin-top       |
| `borderRadius`    | `border-radius`    | border-radius    |

A property name that already holds a hyphen, such as the custom property `--my-color`, passes through unchanged.

### Conditional Rendering (`v-if`, `v-else-if`, `v-else`)

Render an element when its condition holds. Use `v-else-if` and `v-else` for the other branches:

#### Basic `v-if`

```html
<div v-if="show">Visible when true</div>
<div v-if="!hide">Visible when false</div>
<p v-if="status == 'active'">Active</p>
<p v-if="age >= 18">Adult</p>
<p v-if="isAdmin && hasPermission">Admin access</p>
```

#### Bracket Syntax (`[v-if]`)

Wrap a directive in brackets to write it out as a literal attribute, skipping interpolation. Reach for this to leave the directive in the rendered HTML instead of letting Vuego act on it:

```html
<!-- Processed by Vuego (removed after evaluation) -->
<div v-if="show">Content</div>

<!-- Output as literal attribute (not processed, passed to client) -->
<div [v-if]="show">Content</div>
<!-- Renders as: <div v-if="show">Content</div> -->
```

An attribute in bracket syntax carries **no interpolation**. The renderer takes the value as a literal string:

```html
<!-- Without brackets: {{ count }} is evaluated and replaced -->
<p [v-if]="{{count}}">Text</p>  <!-- Output: [v-if]="123" -->

<!-- This is the correct syntax for bracket notation -->
<p [v-if]="count">Text</p>  <!-- Output: [v-if]="count" (no interpolation) -->
```

#### Chained Conditions with `v-else-if` and `v-else`

Use `v-else-if` to test further conditions and `v-else` as the fallback:

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

Use `v-else` after a `v-for` loop to show a "no results" message for an empty list. The `v-else` element must sit as an immediate sibling of the `v-for` element:

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

**Warning:** reach for this with trusted content alone. Content from a user opens an XSS hole.

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

### Single Render (`v-once`)

Render an element once and skip it on later evaluations. This suits a `<script>` or `<style>` tag that belongs on the page a single time:

```html
<script v-once src="/assets/js/component.js"></script>

<script v-once>
  (function () {
    let appearance = localStorage.getItem("appearance");
    let theme = localStorage.getItem("theme");
    appearance && document.documentElement.setAttribute("data-appearance", appearance);
    theme && document.body.setAttribute("data-theme", theme);
  })();
</script>

<style v-once>
  .component { color: blue; }
</style>
```

The `v-once` directive also holds inside a `v-for` loop and across reusable components, so a component the page renders more than once contributes the content a single time.

### Slots

Slots allow parent components to provide content to child components. This enables flexible component composition and content distribution.

#### Default Slots

A child component marks where the content belongs with a `<slot>` element:

**Child Component (components/Card.vuego):**

```html
<div class="card">
  <slot></slot>
</div>
```

**Parent Component Usage:**

```html
<card>
  <h2>Card Title</h2>
  <p>Card content goes here.</p>
</card>
```

**Rendered Output:**

```html
<div class="card">
  <h2>Card Title</h2>
  <p>Card content goes here.</p>
</div>
```

#### Fallback Content

Put fallback content inside the `<slot>` tags; it renders when the caller supplies none:

**Child Component:**

```html
<button class="btn">
  <slot>Click me</slot>
</button>
```

**Usage without content:**

```html
<button></button>
```

**Rendered output uses fallback:**

```html
<button class="btn">Click me</button>
```

#### Named Slots

Give each slot a `name` attribute to send content to a particular place:

**Child Component (components/Modal.vuego):**

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

**Parent Component with Named Slots:**

```html
<modal>
  <template v-slot:header>Modal Title</template>
  <template v-slot:footer>Action Buttons</template>
  Modal content goes here.
</modal>
```

The `#name` shorthand means the same as `v-slot:name`:

```html
<modal>
  <template #header>Modal Title</template>
  <template #footer>Action Buttons</template>
  Modal content goes here.
</modal>
```

#### Scoped Slots

Child components can pass data to slots via attribute bindings (`:prop="value"`), which become available in the parent's template context:

**Child Component (components/List.vuego):**

```html
<ul>
  <li v-for="(index, item) in items">
    <slot :item="item" :index="index"></slot>
  </li>
</ul>
```

**Parent Component Receives Props:**

```html
<list :items="products">
  <template v-slot="slotProps">
    {{ slotProps.item.name }} (#{{ slotProps.index }})
  </template>
</list>
```

**Shorthand with destructuring:**

```html
<list :items="products">
  <template v-slot="{ item, index }">
    {{ item.name }} (#{{ index }})
  </template>
</list>
```

#### Named Scoped Slots

Combine named slots with scoped props:

**Child Component:**

```html
<div>
  <slot name="header" :title="headerTitle"></slot>
  <slot name="content" :count="items.length"></slot>
</div>
```

**Parent Component:**

```html
<my-component>
  <template #header="props">
    Title: {{ props.title }}
  </template>
  <template #content="props">
    Items: {{ props.count }}
  </template>
</my-component>
```

## Components

Vuego supports component composition using the `<template>` tag. See [Components Guide](components.md) for detailed examples.

### `<template include>` Tag (Include/Embed Components)

Include reusable component files:

```html
<template include="components/Header.vuego"></template>
<template include="components/Button.vuego" name="submit" title="Submit Form"></template>
```

The `include` attribute names the template file. Every other attribute reaches the component as a prop.

### `<template>` Tag (Fragment Wrapper)

Wrap component content or fragments:

```html
<template>
  <h1>{{ title }}</h1>
  <p>{{ content }}</p>
</template>
```

Use the `:required` attribute on the `<template>` tag to make the caller supply a prop. Name several required props in a comma-separated list:

```html
<template :required="name,title">
  <h1>{{ title }}</h1>
  <p>Name: {{ name }}</p>
</template>
```

A caller including the component must supply every prop `:required` lists, or the render fails with a validation error.

## Advanced

### Template Functions and Filters

Vuego supports custom template functions and pipe-based filter chaining. See [FuncMap & Filters](funcmap.md) for detailed documentation on:
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

Both styles hold and render the same output.

## Notes and Limitations

### What Vuego Supports

- ✅ Variable interpolation with `{{ expr }}`
- ✅ Nested property access with dot notation
- ✅ Expressions (comparisons, logical operators, ternary)
- ✅ Attribute binding with `:attr` and `v-bind:attr`
- ✅ Bracket syntax for literal attributes: `[directive]="value"` (no interpolation)
- ✅ Conditional rendering with `v-if`, `v-else-if`, and `v-else`
- ✅ Visibility control with `v-show`
- ✅ List iteration with `v-for` (with optional index)
- ✅ Raw HTML with `v-html`
- ✅ Skip template processing with `v-pre`
- ✅ Single render deduplication with `v-once`
- ✅ Slots with `<slot>`, `v-slot`, and `#` (default, named, scoped)
- ✅ Component composition with `<template include>`
- ✅ Component prop validation with `:required`
- ✅ Full HTML documents and fragments
- ✅ Custom template functions and filters

### What Vuego Does NOT Support

- ❌ Event handling (`@click`, `v-on`)
- ❌ Two-way binding (`v-model`)
- ❌ Computed properties
- ❌ Custom directives

### Security

- `html.EscapeString` escapes every interpolated value
- This provides basic XSS protection for `{{ }}` interpolations
- `v-html` bypasses escaping; reach for it with trusted content alone

### Data Types

Vuego accepts `map[string]any` as data input. For strongly-typed data structures, convert them to maps before rendering.

The package supplies `vuego.View[V any](renderer Template, filename string, data V) Template` for this.

### Whitespace

- The renderer drops a text node holding nothing but whitespace
- A DOM comparison disregards whitespace differences
- A text node holding content keeps its indentation

## See Also

- [Components](components.md) - Detailed component composition guide
- [FuncMap & Filters](funcmap.md) - Template functions and custom filters
- [Main README](../README.md) - Overview and quick start
