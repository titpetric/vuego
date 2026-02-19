# Theme Structuring Guide

This guide covers best practices for structuring themes in Vuego, including layout organization, naming conventions, and directory structures. These patterns are informed by established conventions from Hugo, Jekyll, Eleventy (11ty), and Astro.

## Table of Contents

- [Directory Structure](#directory-structure)
- [Layout System](#layout-system)
  - [Layout Hierarchy](#layout-hierarchy)
  - [Layout Chaining](#layout-chaining)
  - [Layout Resolution](#layout-resolution)
- [Naming Conventions](#naming-conventions)
  - [Layout Names](#layout-names)
  - [Component Names](#component-names)
  - [Partial Names](#partial-names)
  - [File Extensions](#file-extensions)
- [Theme Types](#theme-types)
  - [Website Themes](#website-themes)
  - [Blog Themes](#blog-themes)
  - [Documentation Themes](#documentation-themes)
  - [Dashboard Themes](#dashboard-themes)
- [Front Matter](#front-matter)
- [Best Practices](#best-practices)

---

## Directory Structure

A well-organized Vuego theme follows a consistent directory structure:

```
my-theme/
├── assets/
│   ├── css/
│   │   ├── main.css
│   │   └── themes/          # Color/style variants
│   │       ├── light.css
│   │       └── dark.css
│   ├── js/
│   │   └── main.js
│   └── images/
│       └── logo.svg
├── components/              # Reusable UI components
│   ├── alert.vuego
│   ├── button.vuego
│   ├── card.vuego
│   └── table.vuego
├── data/                    # Global data files
│   ├── menu.yml
│   └── site.yml
├── layouts/                 # Page layout templates
│   ├── base.vuego          # Root layout (HTML shell)
│   ├── default.vuego       # Default content layout
│   ├── page.vuego          # Static page layout
│   ├── post.vuego          # Blog post layout
│   └── docs.vuego          # Documentation layout
├── partials/               # Reusable template fragments
│   ├── header.vuego
│   ├── footer.vuego
│   ├── sidebar.vuego
│   ├── nav.vuego
│   └── head.vuego
├── content/                # Content pages (if bundled)
│   ├── index.vuego
│   └── about.vuego
└── theme.yml               # Theme metadata
```

### Key Directories

| Directory     | Purpose                                                                       |
|---------------|-------------------------------------------------------------------------------|
| `layouts/`    | Full-page layout templates that wrap content                                  |
| `partials/`   | Reusable template fragments (header, footer, etc.)                            |
| `components/` | Self-contained UI components with their own logic                             |
| `assets/`     | Static assets (CSS, JS, images, fonts)                                        |
| `data/`       | YAML/JSON data files for menus, config, etc. ([auto-loaded](data-loading.md)) |
| `content/`    | Page content (optional, often separate from theme)                            |

---

## Layout System

### Layout Hierarchy

Vuego uses a layout chaining system where templates can specify a parent layout using the `layout` front matter key. This creates a hierarchy of nested templates.

**Common Layout Hierarchy:**

```
base.vuego          ← Root HTML document structure
    ↓
default.vuego       ← Default page wrapper with header/footer
    ↓
page.vuego          ← Static page with sidebar
post.vuego          ← Blog post with metadata
docs.vuego          ← Documentation with TOC
```

### Layout Chaining

Each layout renders its content and passes it to the parent layout via the `content` variable:

**layouts/base.vuego** (Root layout)

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>{{ title | default("My Site") }}</title>
  <template include="partials/head.vuego"></template>
</head>
<body>
  <div v-html="content"></div>
  <template include="partials/footer.vuego"></template>
</body>
</html>
```

**layouts/default.vuego** (Extends base)

```html
---
layout: base
---
<template include="partials/header.vuego"></template>
<main class="container">
  <div v-html="content"></div>
</main>
```

**layouts/page.vuego** (Extends default)

```html
---
layout: default
---
<article class="page">
  <h1>{{ title }}</h1>
  <div v-html="content"></div>
</article>
```

**Content page using page layout:**

```html
---
layout: page
title: About Us
---
<p>Welcome to our company.</p>
```

### Layout Resolution

Vuego resolves layouts in the following order:

1. **Explicit path** with `.vuego` extension: Used as-is (relative to current file)
2. **Relative path**: Checked relative to the current template directory
3. **Layouts directory**: Falls back to `layouts/[name].vuego`

```yaml
# Explicit path
layout: ../shared/custom-layout.vuego

# Relative (looks in current dir, then layouts/)
layout: post

# Equivalent to layouts/post.vuego
layout: post
```

---

## Naming Conventions

### Layout Names

Follow these conventions for layout files, derived from established SSG practices:

| Layout Name       | Purpose                  | Use Case                     |
|-------------------|--------------------------|------------------------------|
| `base.vuego`      | Root HTML document shell | All pages (via chaining)     |
| `default.vuego`   | Default page wrapper     | General pages                |
| `page.vuego`      | Static content pages     | About, Contact, etc.         |
| `single.vuego`    | Single item detail view  | Blog post, product detail    |
| `list.vuego`      | Collection/index pages   | Blog index, category listing |
| `post.vuego`      | Blog post layout         | Blog articles                |
| `docs.vuego`      | Documentation pages      | Technical docs, guides       |
| `home.vuego`      | Homepage layout          | Landing page                 |
| `landing.vuego`   | Marketing landing page   | Campaign pages               |
| `dashboard.vuego` | Admin/dashboard layout   | Data dashboards              |
| `404.vuego`       | Error page               | Not found                    |
| `blank.vuego`     | Minimal/empty layout     | Embeds, widgets              |

**Naming Rules:**
- Use lowercase with hyphens for multi-word names: `blog-post.vuego`, `docs-page.vuego`
- Keep names short and descriptive
- Avoid abbreviations unless widely understood

### Component Names

Components follow PascalCase for filenames, converted to kebab-case for usage:

| File Name           | Tag Usage        |
|---------------------|------------------|
| `Button.vuego`      | `<button>`       |
| `AlertBox.vuego`    | `<alert-box>`    |
| `DataTable.vuego`   | `<data-table>`   |
| `NavDropdown.vuego` | `<nav-dropdown>` |

**Component Naming Patterns:**

```
components/
├── Alert.vuego           # <alert>
├── Badge.vuego           # <badge>
├── Button.vuego          # <button>
├── ButtonGroup.vuego     # <button-group>
├── Card.vuego            # <card>
├── CardHeader.vuego      # <card-header>
├── Dropdown.vuego        # <dropdown>
├── DropdownMenu.vuego    # <dropdown-menu>
├── Modal.vuego           # <modal>
├── Tabs.vuego            # <tabs>
├── TabPane.vuego         # <tab-pane>
└── Table.vuego           # <table>
```

### Partial Names

Partials use lowercase with hyphens:

```
partials/
├── head.vuego            # <head> tag contents
├── header.vuego          # Page header/nav
├── footer.vuego          # Page footer
├── sidebar.vuego         # Sidebar navigation
├── nav.vuego             # Navigation menu
├── breadcrumbs.vuego     # Breadcrumb trail
├── pagination.vuego      # Pagination controls
├── toc.vuego             # Table of contents
├── meta-tags.vuego       # SEO meta tags
├── social-share.vuego    # Social sharing buttons
├── comments.vuego        # Comments section
└── newsletter.vuego      # Newsletter signup
```

### File Extensions

| Extension        | Purpose                |
|------------------|------------------------|
| `.vuego`         | Template files         |
| `.yml` / `.yaml` | Data and configuration |
| `.css`           | Stylesheets            |
| `.js`            | JavaScript             |
| `.svg`           | Vector graphics        |

---

## Theme Types

### Website Themes

General-purpose website themes for corporate sites, portfolios, and landing pages.

**Recommended layouts:**

```
layouts/
├── base.vuego           # HTML shell
├── default.vuego        # Standard page wrapper
├── home.vuego           # Homepage with hero
├── page.vuego           # Content pages
├── landing.vuego        # Marketing landing pages
├── contact.vuego        # Contact form page
└── 404.vuego            # Error page
```

**Example hierarchy:**

```
base.vuego
├── home.vuego           # Standalone homepage
├── default.vuego
│   ├── page.vuego       # Static pages
│   ├── landing.vuego    # Landing pages
│   └── contact.vuego    # Contact page
└── 404.vuego            # Error page
```

### Blog Themes

Themes optimized for blogs and content publishing.

**Recommended layouts:**

```
layouts/
├── base.vuego           # HTML shell
├── default.vuego        # Site wrapper
├── home.vuego           # Blog homepage/feed
├── post.vuego           # Single blog post
├── list.vuego           # Post listing/archive
├── category.vuego       # Category archive
├── tag.vuego            # Tag archive
├── author.vuego         # Author profile
└── page.vuego           # Static pages (about, etc.)
```

**Post layout example:**

```html
---
layout: default
---
<article class="post">
  <header class="post-header">
    <h1>{{ title }}</h1>
    <div class="post-meta">
      <time datetime="{{ date }}">{{ date | formatTime("January 2, 2006") }}</time>
      <span v-if="author">by {{ author }}</span>
    </div>
  </header>

  <div class="post-content" v-html="content"></div>

  <footer class="post-footer">
    <div v-if="tags" class="post-tags">
      <span v-for="tag in tags" class="tag">{{ tag }}</span>
    </div>
  </footer>
</article>
```

### Documentation Themes

Themes designed for technical documentation, guides, and knowledge bases.

**Recommended layouts:**

```
layouts/
├── base.vuego           # HTML shell
├── docs.vuego           # Documentation wrapper
├── doc-page.vuego       # Documentation page
├── api.vuego            # API reference page
├── tutorial.vuego       # Step-by-step tutorial
├── guide.vuego          # Guide/how-to
└── changelog.vuego      # Version changelog
```

**Key partials:**

```
partials/
├── docs-sidebar.vuego   # Documentation navigation
├── toc.vuego            # In-page table of contents
├── breadcrumbs.vuego    # Navigation breadcrumbs
├── prev-next.vuego      # Previous/next navigation
├── search.vuego         # Search interface
└── version-select.vuego # Version selector
```

**Documentation layout example:**

```html
---
layout: base
---
<div class="docs-layout">
  <template include="partials/docs-sidebar.vuego" :menu="menu"></template>

  <main class="docs-content">
    <template include="partials/breadcrumbs.vuego"></template>

    <article>
      <h1>{{ title }}</h1>
      <div v-if="description" class="lead">{{ description }}</div>
      <div v-html="content"></div>
    </article>

    <template include="partials/prev-next.vuego"></template>
  </main>

  <aside class="docs-toc">
    <template include="partials/toc.vuego" :items="toc"></template>
  </aside>
</div>
```

### Dashboard Themes

Themes for admin panels, data dashboards, and internal tools.

**Recommended layouts:**

```
layouts/
├── base.vuego           # HTML shell
├── dashboard.vuego      # Main dashboard wrapper
├── admin.vuego          # Admin panel layout
├── settings.vuego       # Settings page
├── profile.vuego        # User profile
├── blank.vuego          # Full-width no-chrome
└── auth.vuego           # Login/auth pages
```

**Key partials:**

```
partials/
├── dashboard-header.vuego
├── dashboard-sidebar.vuego
├── stats-card.vuego
├── data-table.vuego
├── chart.vuego
├── notifications.vuego
└── user-menu.vuego
```

**Dashboard layout example:**

```html
---
layout: base
---
<div class="dashboard-layout">
  <template include="partials/dashboard-sidebar.vuego"></template>

  <div class="dashboard-main">
    <template include="partials/dashboard-header.vuego"></template>

    <main class="dashboard-content">
      <div v-html="content"></div>
    </main>
  </div>
</div>
```

---

## Front Matter

Front matter is YAML at the top of a template that defines metadata and the layout chain.

### Common Front Matter Fields

| Field         | Type    | Description              |
|---------------|---------|--------------------------|
| `layout`      | string  | Parent layout to use     |
| `title`       | string  | Page title               |
| `description` | string  | Page description/summary |
| `date`        | string  | Publication date         |
| `author`      | string  | Content author           |
| `tags`        | array   | Content tags             |
| `category`    | string  | Content category         |
| `draft`       | boolean | Draft status             |
| `weight`      | number  | Sort order               |
| `menu`        | object  | Menu configuration       |

### Example Front Matter

**Blog post:**

```yaml
---
layout: post
title: Getting Started with Vuego
description: Learn how to build your first Vuego template
date: 2024-01-15
author: Jane Doe
tags:
  - tutorial
  - getting-started
category: tutorials
---
```

**Documentation page:**

```yaml
---
layout: docs
title: Template Syntax
description: Complete reference for Vuego template syntax
weight: 10
toc:
  - id: values
    label: Values
  - id: directives
    label: Directives
  - id: components
    label: Components
---
```

**Dashboard page:**

```yaml
---
layout: dashboard
title: Analytics Overview
sidebar: false
breadcrumbs:
  - label: Home
    url: /
  - label: Analytics
    url: /analytics
---
```

---

## Best Practices

### 1. Use Layout Chaining

Build layouts hierarchically rather than duplicating code:

```
base.vuego           # HTML document structure (shared by all)
    ↓
default.vuego        # Common header/footer
    ↓
page.vuego           # Page-specific structure
```

### 2. Keep Partials Focused

Each partial should have a single responsibility:

```html
<!-- Good: focused partial -->
<template include="partials/pagination.vuego" :current="page" :total="pages"></template>

<!-- Avoid: monolithic partials that do too much -->
<template include="partials/everything.vuego"></template>
```

### 3. Use Semantic Naming

Names should describe the content or purpose:

```
layouts/
├── post.vuego              ✓ Describes content type
├── blog-single.vuego       ✓ Specific and clear
├── template1.vuego         ✗ Not descriptive
├── new-layout.vuego        ✗ Meaningless
```

### 4. Organize by Function

Group related files together:

```
components/
├── buttons/
│   ├── Button.vuego
│   ├── ButtonGroup.vuego
│   └── IconButton.vuego
├── forms/
│   ├── Input.vuego
│   ├── Select.vuego
│   └── Checkbox.vuego
└── cards/
    ├── Card.vuego
    ├── CardHeader.vuego
    └── CardBody.vuego
```

### 5. Document Layout Requirements

Add comments to explain layout expectations:

```html
---
layout: default
---
<!--
  Page Layout

  Required front matter:
    - title: Page title (string)

  Optional front matter:
    - description: Page description (string)
    - sidebar: Show sidebar (boolean, default: true)
-->
<article>
  <h1>{{ title }}</h1>
  <div v-html="content"></div>
</article>
```

### 6. Provide Sensible Defaults

Use the `default` filter for optional values:

```html
<title>{{ title | default("My Site") }}</title>
<meta name="description" content="{{ description | default("Welcome to my site") }}">
```

### 7. Support Theme Variants

Organize CSS themes for easy switching:

```
assets/css/
├── main.css              # Core styles
└── themes/
    ├── light.css         # Light theme
    ├── dark.css          # Dark theme
    └── custom.css        # Custom variant
```

### 8. Follow Platform Conventions

Adopt naming patterns from established SSGs:

| Convention           | Source   | Vuego Equivalent        |
|----------------------|----------|-------------------------|
| `_layouts/`          | Jekyll   | `layouts/`              |
| `_includes/`         | Jekyll   | `partials/`             |
| `layouts/_default/`  | Hugo     | `layouts/default.vuego` |
| `layouts/partials/`  | Hugo     | `partials/`             |
| `_includes/layouts/` | Eleventy | `layouts/`              |
| `src/layouts/`       | Astro    | `layouts/`              |

---

## See Also

- [Data Loading](data-loading.md) - Automatic config loading from `theme.yml` and `data/`
- [Template Syntax Reference](syntax.md) - Complete syntax documentation
- [Components Guide](components.md) - Component composition patterns
- [API Reference](api.md) - Go API documentation
- [FuncMap & Filters](funcmap.md) - Template functions

---

## References

- [Hugo Directory Structure](https://gohugo.io/getting-started/directory-structure/)
- [Hugo Template Lookup Order](https://gohugo.io/templates/lookup-order/)
- [Eleventy Layouts](https://www.11ty.dev/docs/layouts/)
- [Eleventy Layout Chaining](https://www.11ty.dev/docs/layout-chaining/)
- [Jekyll Layouts](https://jekyllrb.com/docs/layouts/)
- [Jekyll Directory Structure](https://jekyllrb.com/docs/structure/)
- [Astro Layouts](https://docs.astro.build/en/basics/layouts/)
