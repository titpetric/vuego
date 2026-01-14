# Directives

## v-for Loops

Use `v-for="item in items"` to iterate over arrays and render elements for each item.

You can also access the index: `v-for="(index, item) in items"`.

[Learn more](https://github.com/titpetric/vuego/blob/main/docs/syntax.md#list-iteration-v-for)

@file: v-for.vuego

---

## v-for with Empty List

Use `v-else` after a `v-for` loop to display a message when the list is empty.

The `v-else` must be an immediate sibling element to the `v-for` element. This pattern is common for showing "no results" or "no items" messages.

[Learn more](https://github.com/titpetric/vuego/blob/main/docs/syntax.md#empty-list-with-v-for-and-v-else)

@file: v-for-empty.vuego

---

## v-if Conditionals

Use `v-if="condition"` to conditionally render elements based on expressions.

Combine with `v-else-if` and `v-else` for multiple branches.

[Learn more](https://github.com/titpetric/vuego/blob/main/docs/syntax.md#conditional-rendering-v-if-v-else-if-v-else)

@file: v-if.vuego

---

## v-show Visibility

Use `v-show="condition"` to toggle element visibility with CSS display property.

Unlike `v-if`, the element is always rendered but hidden when the condition is false.

[Learn more](https://github.com/titpetric/vuego/blob/main/docs/syntax.md#visibility-toggle-v-show)

@file: v-show.vuego

---

## Attribute Binding

Use `:attr="value"` to dynamically bind attribute values.

This is shorthand for `v-bind:attr="value"`. Common uses include `:href`, `:src`, and `:class`.

[Learn more](https://github.com/titpetric/vuego/blob/main/docs/syntax.md#attribute-binding-attr--v-bindattr)

@file: binding.vuego

---

## Class & Style Object Binding

Bind to object literals to conditionally apply CSS classes and inline styles.

For classes, object keys become class names and are applied when their values are truthy.

For styles, use camelCase property names which are automatically converted to kebab-case CSS properties.

[Learn more](https://github.com/titpetric/vuego/blob/main/docs/syntax.md#object-binding-for-class-and-style)

@file: binding-objects.vuego

---

## More Directives: v-html, v-pre, v-once

Vuego supports additional directives for special rendering scenarios.

`v-html` renders unescaped HTML content (security: only use with trusted content).

`v-pre` skips template processing and outputs content literally.

`v-once` renders an element once and skips it on subsequent renders.

[Learn more](https://github.com/titpetric/vuego/blob/main/docs/syntax.md#directives)

@file: more-directives.vuego
