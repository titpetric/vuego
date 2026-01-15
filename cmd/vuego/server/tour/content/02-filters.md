# Filters

## Basic Filters

Apply filters using the pipe `|` operator: `{{ value | filterName }}`.

Common text filters include `upper`, `lower`, `title`, and `trim`.

- [Syntax - Filters and Pipes](https://github.com/titpetric/vuego/blob/main/docs/syntax.md#filters-and-pipes)

@file: basic.vuego

---

## Chained Filters

Chain multiple filters together: `{{ name | trim | lower | title }}`.

Each filter transforms the result of the previous one.

- [Syntax - Filters and Pipes](https://github.com/titpetric/vuego/blob/main/docs/syntax.md#filters-and-pipes)

@file: chained.vuego

---

## Utility Filters

Use `default` to provide fallback values: `{{ name | default("Anonymous") }}`.

Use `json` to encode values as JSON for debugging or JavaScript integration.

Use `len` to get the length of strings, arrays, or maps.

- [Built-in Functions](https://github.com/titpetric/vuego/blob/main/docs/funcmap.md#built-in-functions)

@file: utility.vuego
