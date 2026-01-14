# Filters

## Basic Filters

Apply filters using the pipe `|` operator: `{{ value | filterName }}`.

Common text filters include `upper`, `lower`, `title`, and `trim`.

@file: basic.vuego

---

## Chained Filters

Chain multiple filters together: `{{ name | trim | lower | title }}`.

Each filter transforms the result of the previous one.

@file: chained.vuego

---

## Utility Filters

Use `default` to provide fallback values: `{{ name | default("Anonymous") }}`.

Use `json` to encode values as JSON for debugging or JavaScript integration.

Use `len` to get the length of strings, arrays, or maps.

@file: utility.vuego
