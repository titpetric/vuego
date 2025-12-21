# vuego CLI Commands

## Overview

vuego provides three main CLI commands: `fmt`, `render`, and `diff`.

## fmt - Format templates

Format vuego template files with proper indentation and whitespace handling.

```bash
vuego fmt <file.vuego>
```

**Features:**
- Formats vuego and HTML templates with proper indentation (default: 2 spaces)
- Preserves YAML frontmatter
- Handles both full documents and partial fragments
- Modifies files in-place

**Examples:**

```bash
# Format a single file
vuego fmt layout.vuego

# Format multiple files
vuego fmt page1.vuego page2.vuego
```

## render - Render templates

Render templates with JSON or YAML data.

```bash
vuego render <file.vuego> <data.json|data.yml>
```

**Features:**
- Supports both JSON and YAML data files
- Renders templates to stdout
- Supports nested includes

**Examples:**

```bash
# Render with JSON data
vuego render index.vuego data.json

# Render with YAML data
vuego render index.vuego data.yml
```

## diff - Compare HTML/vuego files

Compare two HTML or vuego files using DOM-aware comparison. This command ignores whitespace and formatting differences, comparing the actual DOM structure.

```bash
vuego diff [--output FORMAT] <file1> <file2>
```

**Output Formats:**

### simple (Human-readable)

Shows side-by-side comparison of the original file contents.

```bash
vuego diff --output simple file1.html file2.html
```

**Output Example:**

```
✗ file1.html and file2.html have different DOM trees

Expected (file1.html):
<div class="container">
  <h1>Hello</h1>
</div>

Actual (file2.html):
<div class="container">
  <h1>Goodbye</h1>
</div>
```

### unified (Line-by-line diff)

Standard unified diff format that can be piped to `colordiff` for colorized output.

```bash
vuego diff file1.html file2.html

# With colordiff for colorized output
vuego diff file1.html file2.html | colordiff
```

**Output Example:**

```
--- file1.html
+++ file2.html
  <html>
    <head>
      <title>Test</title>
    </head>
    <body>
      <h1>
-       Hello
+       Goodbye
      </h1>
    </body>
  </html>
```

### yaml (YAML representation with dyff)

Converts HTML DOM structure to YAML and uses `dyff` for visual comparison. Requires `dyff` to be installed.

```bash
vuego diff --output yaml file1.html file2.html
```

**Installation:**

If you don't have dyff installed:

```bash
go install github.com/homeport/dyff/cmd/dyff@latest
```

Or using Homebrew:

```bash
brew install homeport/tap/dyff
```

**Output Example:**

```
     _        __  __
   _| |_   _ / _|/ _|  between /tmp/test1.yaml
 / _' | | | | |_| |_       and /tmp/test2.yaml
| (_| | |_| |  _|  _|
 \__,_|\__, |_| |_|   returned one difference
        |___/

children
  - one list entry removed:     + one list entry added:
    - children:                   - children:
      - children:                   - children:
        - text: Hello                 - text: Goodbye
        tag: h1                       tag: h1
      tag: body                     tag: body
```

## Exit Codes

- `0` - Success (files match for diff command)
- `1` - Failure (files differ for diff command, or other errors)

## Usage Examples

### Verify formatting doesn't break DOM

```bash
# Format a file
vuego fmt template.html

# Compare original and formatted
vuego diff template.html.bak template.html
```

### Check template equivalence after refactoring

```bash
# Compare two refactored templates
vuego diff --output simple old-template.vuego new-template.vuego
```

### Debug DOM differences

```bash
# Get detailed YAML representation of differences
vuego diff --output yaml file1.html file2.html

# Render template and compare with expected output
vuego render template.vuego data.json > output.html
vuego diff expected.html output.html
```
