# Variable Interpolation

## Basic Variables

The simplest form of variable interpolation uses double curly braces `{{ }}` to output a variable's value.

Variables are replaced with their values from the data context. If a variable doesn't exist, it outputs an empty string.

@file: basic.vuego

---

## Nested Properties

You can access nested object properties using dot notation like `{{ object.property }}`.

This works with any depth of nesting, such as `{{ user.address.city }}`.

@file: nested.vuego

---

## Array Elements

Access array elements by index using bracket notation: `{{ array[0] }}`.

You can combine this with property access: `{{ users[0].name }}`.

@file: arrays.vuego
