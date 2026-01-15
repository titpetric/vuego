# Variable Interpolation

## Basic Variables

The simplest form of variable interpolation uses double curly braces `{{ }}` to output a variable's value.

Variables are replaced with their values from the data context. If a variable doesn't exist, it outputs an empty string.

- [Syntax - Variables](https://github.com/titpetric/vuego/blob/main/docs/syntax.md#values)

@file: basic.vuego

---

## Nested Properties

You can access nested object properties using dot notation like `{{ object.property }}`.

This works with any depth of nesting, such as `{{ user.address.city }}`.

**Note:** Nesting properties is recommended when using variables that might conflict with built-in function names (like `count`, `len`). Using dot notation like `vars.count` avoids conflicts with expr-lang's built-in functions. See [expr-lang issue 902](https://github.com/expr-lang/expr/issues/902) for details.

- [Syntax - Nested Properties](https://github.com/titpetric/vuego/blob/main/docs/syntax.md#nested-properties)

@file: nested.vuego

---

## Array Elements

Access array elements by index using bracket notation: `{{ array[0] }}`.

You can combine this with property access: `{{ users[0].name }}`.

- [Syntax - Array Elements](https://github.com/titpetric/vuego/blob/main/docs/syntax.md#values)

@file: arrays.vuego

---

## Expressions & Operators

Interpolations support comparison and logical operators for more complex expressions.

Use `==` and `!=` for equality, `<` `>` `<=` `>=` for numeric comparisons, and `&&` `||` `!` for logic.

Combine operators with the ternary operator `condition ? truthy : falsy` to show different values.

- [Expressions & Operators](https://github.com/titpetric/vuego/blob/main/docs/expressions.md)

@file: expressions.vuego
