# Components

## Include Components

Use `<vuego include="file">` to include another template file.

This is the basic way to compose templates from reusable parts.

@file: include.vuego
@file: header.vuego

---

## Template Require

Use `<template :required="varName">` to require that specific variables are provided.

This helps catch missing data early with clear error messages.

@file: require.vuego

---

## Slots

Use `<slot>` to create insertion points where parent content can be injected.

Named slots allow multiple content areas: `<slot name="header">`.

@file: slots.vuego

---

## Scoped Slots

Child components can pass data to slots via attribute bindings.

Use `v-slot` or the `#` shorthand to receive slot props.

@file: scoped.vuego
@file: list.vuego

---

## Layouts

Use layouts to wrap page content with a common structure.

Set `layout: layout.vuego` in the front-matter to wrap the page content.

The layout uses `<slot>` to insert the page content.

@file: layout.vuego
@file: base.vuego
