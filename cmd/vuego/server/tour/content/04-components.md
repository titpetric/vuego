# Components

## Include Components

Use `<vuego include="file">` to include another template file.

This is the basic way to compose templates from reusable parts.

- [Basic Component Composition](https://github.com/titpetric/vuego/blob/main/docs/components.md#basic-component-composition)

@file: include.vuego
@file: header.vuego

---

## Template Require

Use `<template :required="varName">` to require that specific variables are provided.

This helps catch missing data early with clear error messages.

- [Required Attributes](https://github.com/titpetric/vuego/blob/main/docs/components.md#required-attributes)

@file: require.vuego

---

## Slots

Use `<slot>` to create insertion points where parent content can be injected.

Named slots allow multiple content areas: `<slot name="header">`.

- [Slots](https://github.com/titpetric/vuego/blob/main/docs/components.md#slots)

@file: slots.vuego

---

## Scoped Slots

Child components can pass data to slots via attribute bindings.

Use `v-slot` or the `#` shorthand to receive slot props.

- [Scoped Slots](https://github.com/titpetric/vuego/blob/main/docs/components.md#scoped-slots)

@file: scoped.vuego
@file: list.vuego

---

## Component Shorthands

When component shorthands are enabled, you can use simple tag names instead of `<vuego include>`.

Component filenames are converted from PascalCase to kebab-case: `ButtonPrimary.vuego` becomes `<button-primary>`.

Attributes on shorthand tags are passed as props to the component. Slot content works the same as with `<vuego include>`.

- [Component Shorthands](https://github.com/titpetric/vuego/blob/main/docs/components.md#component-shorthands)

@file: shorthands.vuego
@file: button-component.vuego
@file: alert-box.vuego

---

## Layouts

Use layouts to wrap page content with a common structure.

Set `layout: layout.vuego` in the front-matter to wrap the page content.

The layout uses `<slot>` to insert the page content.

- [Components](https://github.com/titpetric/vuego/blob/main/docs/syntax.md#components)

@file: layout.vuego
@file: base.vuego
