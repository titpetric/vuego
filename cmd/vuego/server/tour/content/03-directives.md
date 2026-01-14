# Directives

## v-for Loops

Use `v-for="item in items"` to iterate over arrays and render elements for each item.

You can also access the index: `v-for="(index, item) in items"`.

@file: v-for.vuego

---

## v-if Conditionals

Use `v-if="condition"` to conditionally render elements based on expressions.

Combine with `v-else-if` and `v-else` for multiple branches.

@file: v-if.vuego

---

## v-show Visibility

Use `v-show="condition"` to toggle element visibility with CSS display property.

Unlike `v-if`, the element is always rendered but hidden when the condition is false.

@file: v-show.vuego

---

## Attribute Binding

Use `:attr="value"` to dynamically bind attribute values.

This is shorthand for `v-bind:attr="value"`. Common uses include `:href`, `:src`, and `:class`.

@file: binding.vuego
