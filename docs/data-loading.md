# Data Loading

Vuego automatically loads YAML configuration files from the template filesystem when you create a renderer with `NewFS` or `New(WithFS(...))`. This provides global data — such as site metadata, navigation menus, and theme settings — to every template without requiring explicit `Fill()` or `Assign()` calls.

## Table of Contents

- [How It Works](#how-it-works)
- [Directory Structure](#directory-structure)
- [Load Order and Overrides](#load-order-and-overrides)
- [Accessing Loaded Data in Templates](#accessing-loaded-data-in-templates)
- [Overriding with Fill](#overriding-with-fill)
- [Examples](#examples)
  - [Site Metadata](#site-metadata)
  - [Navigation Menu](#navigation-menu)
  - [Theme Configuration](#theme-configuration)
  - [Combined Example](#combined-example)
- [Behavior by Filesystem Type](#behavior-by-filesystem-type)
- [See Also](#see-also)

---

## How It Works

When `NewFS()` is called, Vuego automatically:

1. Loads `theme.yml` from the filesystem root (if it exists)
2. Loads all `.yml` and `.yaml` files from the `data/` directory (if it exists)

The loaded data is merged into the template's initial context, making it available to all templates as top-level variables.

```go
// Config loading is implicit — no extra options needed
renderer := vuego.NewFS(os.DirFS("templates"))
```

This is equivalent to calling `Fill()` with the merged YAML data before any `Load()` or `New()` call.

## Directory Structure

```
templates/
├── theme.yml              # Root-level theme config (loaded first)
├── data/                  # Data directory (all YAML files loaded)
│   ├── menu.yml           # Navigation data
│   ├── site.yml           # Site metadata
│   └── footer.yml         # Footer configuration
├── layouts/
│   └── base.vuego
├── components/
│   └── Nav.vuego
└── index.vuego
```

Only two locations are checked:

| Location                    | What's loaded                           |
|-----------------------------|-----------------------------------------|
| `theme.yml`                 | Single file at the filesystem root      |
| `data/*.yml`, `data/*.yaml` | All YAML files in the `data/` directory |

Files that don't exist are silently skipped. Invalid YAML is also skipped silently.

## Load Order and Overrides

Files are loaded in this order:

1. `theme.yml` (root level)
2. `data/` directory files (alphabetical order)

Later files override earlier ones at the top-level key. This means `data/theme.yml` overrides keys from the root `theme.yml`:

**theme.yml** (root):

```yaml
theme:
  header:
    title: Default Title
    subtitle: Default Subtitle
```

**data/theme.yml**:

```yaml
theme:
  header:
    title: Custom Title
    subtitle: Custom Subtitle
```

The resulting context will have `theme.header.title` set to `"Custom Title"`.

> **Note:** Overrides happen at the top-level YAML key. If both files define a `theme` key, the entire `theme` value from `data/theme.yml` replaces the one from `theme.yml`.

## Accessing Loaded Data in Templates

Loaded data is available as top-level variables in every template:

**data/site.yml**:

```yaml
site:
  name: My Website
  tagline: Built with Vuego
```

**data/menu.yml**:

```yaml
menu:
  - label: Home
    url: /
  - label: About
    url: /about
  - label: Contact
    url: /contact
```

**layouts/base.vuego**:

```html
<!DOCTYPE html>
<html>
<head>
  <title>{{ site.name }}</title>
</head>
<body>
  <nav>
    <a v-for="item in menu" :href="item.url">{{ item.label }}</a>
  </nav>
  <div v-html="content"></div>
</body>
</html>
```

## Overriding with Fill

Per-request data passed via `Fill()` overrides auto-loaded config values. This allows templates to use config as defaults while accepting dynamic data:

```go
renderer := vuego.NewFS(os.DirFS("templates"))

// Fill overrides auto-loaded config values
type PageData struct {
	Theme struct {
		Header struct {
			Title    string `json:"title"`
			Subtitle string `json:"subtitle"`
		} `json:"header"`
	} `json:"theme"`
}

data := PageData{}
data.Theme.Header.Title = "Page-Specific Title"

buf := &bytes.Buffer{}
err := renderer.Load("index.vuego").Fill(data).Render(ctx, buf)
```

The precedence order (highest to lowest):

1. Front-matter in the `.vuego` file
2. `Fill()` / `Assign()` data
3. Auto-loaded config (`data/` files, then `theme.yml`)

## Examples

### Site Metadata

**data/site.yml**:

```yaml
site:
  name: Acme Corp
  url: https://acme.example.com
  description: Building the future
  language: en
```

**partials/head.vuego**:

```html
<meta charset="UTF-8">
<meta name="description" content="{{ site.description }}">
<title>{{ title | default(site.name) }}</title>
<link rel="canonical" href="{{ site.url }}">
```

### Navigation Menu

**data/menu.yml**:

```yaml
menu:
  - label: Home
    url: /
  - label: Products
    url: /products
  - label: Blog
    url: /blog
  - label: Contact
    url: /contact
```

**components/Nav.vuego**:

```html
<nav class="main-nav">
  <ul>
    <li v-for="item in menu">
      <a :href="item.url">{{ item.label }}</a>
    </li>
  </ul>
</nav>
```

### Theme Configuration

**theme.yml**:

```yaml
theme:
  header:
    title: My Site
    subtitle: Welcome
  colors:
    primary: "#3b82f6"
    secondary: "#64748b"
  footer:
    copyright: "© 2025 My Site"
```

**layouts/base.vuego**:

```html
<!DOCTYPE html>
<html>
<head>
  <title>{{ theme.header.title }}</title>
  <style>
    :root {
      --color-primary: {{ theme.colors.primary }};
      --color-secondary: {{ theme.colors.secondary }};
    }
  </style>
</head>
<body>
  <header>
    <h1>{{ theme.header.title }}</h1>
    <p>{{ theme.header.subtitle }}</p>
  </header>
  <main>
    <div v-html="content"></div>
  </main>
  <footer>{{ theme.footer.copyright }}</footer>
</body>
</html>
```

### Combined Example

A complete theme with auto-loaded config:

```
my-site/
├── theme.yml
├── data/
│   ├── menu.yml
│   └── social.yml
├── layouts/
│   └── base.vuego
├── components/
│   ├── Nav.vuego
│   └── SocialLinks.vuego
└── pages/
    └── index.vuego
```

**Go code**:

```go
renderer := vuego.NewFS(os.DirFS("my-site"), vuego.WithComponents())

// All config is auto-loaded — just render
err := renderer.Load("pages/index.vuego").Render(ctx, w)
```

No explicit config loading is needed. The renderer automatically picks up `theme.yml`, `data/menu.yml`, and `data/social.yml`.

## Behavior by Filesystem Type

| Filesystem        | Behavior                                                                                                           |
|-------------------|--------------------------------------------------------------------------------------------------------------------|
| `os.DirFS(path)`  | Config is loaded at construction time. Changes to YAML files on disk are reflected when a new renderer is created. |
| `embed.FS`        | Config is loaded once from the embedded filesystem at construction time.                                           |
| `vuego.OverlayFS` | Config is loaded from the merged filesystem, with the upper FS taking precedence.                                  |

For development servers that need live reload, create a new renderer instance on each request:

```go
// Development: new renderer per request picks up file changes
func handler(w http.ResponseWriter, r *http.Request) {
	renderer := vuego.NewFS(os.DirFS("templates"))
	err := renderer.Load("index.vuego").Fill(data).Render(r.Context(), w)
}
```

For production, create the renderer once at startup:

```go
// Production: single renderer, config loaded once
var renderer = vuego.NewFS(embedFS, vuego.WithComponents())
```

## See Also

- [Theme Structuring Guide](themes.md) — Layout organization and directory conventions
- [Components Guide](components.md) — Component composition and props
- [Template Syntax](syntax.md) — Variable interpolation and directives
- [API Reference](api.md) — Go API documentation
