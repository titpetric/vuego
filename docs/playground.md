# VueGo Playground Guide

The VueGo Playground is an interactive web application for testing and experimenting with VueGo templates in real-time.

## Features

### 1. **Split-Panel Interface**

The playground uses a 4-panel layout:
- **Template Editor** (top-left): Write your VueGo template
- **Data Editor** (top-right): Provide JSON data
- **Preview** (bottom): See rendered HTML output
- **Controls** (bottom bar): Render button and status

### 2. **Live Rendering**

Press **Render** or use `Ctrl+Enter` / `Cmd+Enter` to render your template instantly.

### 3. **Built-in Examples**

Click example buttons to load common patterns:
- **Basic**: Simple interpolation and loops
- **Filters**: Using pipe-based filters
- **Conditional**: v-if with function calls

### 4. **Error Display**

Errors are shown with full context:

```
in template.vuego: in expression '{{ items | double }}':
  double(): cannot convert argument 0 from []string to int
```

## Example Workflows

### Testing a New Filter

1. Write your template:

```html
<p>{{ price | currency }}</p>
```

2. Add data:

```json
{
  "price": 19.99
}
```

3. Click Render to see if your custom filter works

### Debugging Template Issues

1. Paste your problematic template
2. Paste your actual data
3. Click Render
4. Review error message with line numbers and context

### Prototyping Components

1. Design your component structure:

```html
<div class="card">
  <h2>{{ title }}</h2>
  <div v-if="len(items)">
    <ul>
      <li v-for="item in items">{{ item.name }}</li>
    </ul>
  </div>
</div>
```

2. Test with sample data
3. Iterate quickly without restarting your app

## Keyboard Shortcuts

- `Ctrl+Enter` / `Cmd+Enter`: Render template
- `Tab` in textarea: Insert tab (doesn't switch focus)

## API Integration

You can integrate the playground API into your own tools:

```bash
curl -X POST http://localhost:8080/api/render \
  -H "Content-Type: application/json" \
  -d '{
    "template": "<p>{{ message }}</p>",
    "data": "{\"message\": \"Hello!\"}"
  }'
```

Response:

```json
{
  "html": "<p>Hello!</p>"
}
```

## Common Use Cases

### 1. Learning VueGo

New to VueGo? Use the playground to:

- Try examples from documentation
- Experiment with syntax
- Understand filter chaining
- Test v-if and v-for logic

### 2. Debugging Production Issues

Reproduce issues by:

1. Copy production template
2. Copy production data (sanitized)
3. Identify the problem
4. Test the fix

### 3. Template Development

Develop templates faster:

1. Write template in playground
2. Test with realistic data
3. Copy to your project when ready

## Limitations

1. **No Custom Functions**: Built-in filters only (can't add custom FuncMap)
2. **No File Includes**: `<vuego include>` not supported
3. **Fragment Rendering**: Uses `RenderFragment()` (no `<html>` wrapper)
4. **No State**: Each render is independent

We could add support for the local filesystem which could serve as a document root for vuego templates. This would enable vuego components on the playground. It's also possible to construct a virtual filesystem and manage a package of json+vuego templates and any additional assets like CSS and JS.

## See Also

- [FuncMap & Filters](funcmap.md) - Available filters
- [Template Syntax](syntax.md) - Full syntax reference
