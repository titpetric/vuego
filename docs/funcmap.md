# FuncMap - Template Functions

VueGo supports template functions similar to Go's `text/template` and `html/template` packages. Functions can be used in interpolations with pipe-based chaining and in `v-if` conditions.

## Basic Usage

### Setting Custom Functions

```go
vue := vuego.NewVue(templateFS)

// Add custom functions
vue.Funcs(vuego.FuncMap{
	"greet": func(v string) string {
		return "Hello, " + v
	},
	"multiply": func(v int, factor int) int {
		return v * factor
	},
})
```

### Using Functions in Templates

#### Pipe Syntax in Interpolation

Functions can be chained using the `|` (pipe) operator:

```html
<!-- Single filter -->
<p>{{ name | upper }}</p>

<!-- Multiple filters chained -->
<p>{{ name | lower | title }}</p>

<!-- Filters with arguments -->
<p>{{ value | default("fallback") }}</p>

<!-- Complex chains -->
<p>{{ text | trim | lower | title }}</p>

<!-- Nested data with filters -->
<p>{{ user.name | upper }}</p>

<!-- Multiple filters with arguments -->
<p>{{ timestamp | formatTime("2006-01-02") | upper }}</p>
```

#### Function Calls in v-if

```html
<!-- Direct function call -->
<div v-if="len(items)">
    <p>Has items</p>
</div>

<!-- Using custom functions -->
<div v-if="isValid(data)">
    <p>Data is valid</p>
</div>
```

## Built-in Functions

VueGo comes with several built-in utility functions:

### String Functions

- **`upper`** - Converts string to uppercase

  ```html
  {{ name | upper }}  <!-- "john" -> "JOHN" -->
  ```

- **`lower`** - Converts string to lowercase

  ```html
  {{ name | lower }}  <!-- "JOHN" -> "john" -->
  ```

- **`title`** - Title-cases string

  ```html
  {{ name | title }}  <!-- "john doe" -> "John Doe" -->
  ```

- **`trim`** - Removes leading and trailing whitespace

  ```html
  {{ text | trim }}  <!-- "  hello  " -> "hello" -->
  ```

### Utility Functions

- **`default(fallback)`** - Returns fallback value if input is nil or empty

  ```html
  {{ name | default("Anonymous") }}
  ```

- **`len`** - Returns length of string, array, or map

  ```html
  {{ items | len }}
  <div v-if="len(items)">Has items</div>
  ```

- **`escape`** - HTML-escapes the value

  ```html
  {{ userInput | escape }}
  ```

### Date/Time Functions

- **`formatTime(layout)`** - Formats time.Time using Go layout format

  ```html
  {{ timestamp | formatTime("2006-01-02 15:04:05") }}
  {{ timestamp | formatTime("Jan 2, 2006") }}
  ```

## Custom Function Examples

### Simple String Transformation

```go
vue.Funcs(vuego.FuncMap{
	"reverse": func(s string) string {
		runes := []rune(s)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes)
	},
})
```

```html
<p>{{ word | reverse }}</p>
```

### Function with Arguments

```go
vue.Funcs(vuego.FuncMap{
	"truncate": func(s string, max int) string {
		if len(s) > max {
			return s[:max] + "..."
		}
		return s
	},
})
```

```html
<p>{{ description | truncate(50) }}</p>
```

### Currency Formatting

```go
vue.Funcs(vuego.FuncMap{
	"currency": func(amount float64) string {
		return fmt.Sprintf("$%.2f", amount)
	},
})
```

```html
<p>{{ price | currency }}</p>
```

### Combining Multiple Custom Functions

```go
vue.Funcs(vuego.FuncMap{
	"slugify": func(v string) string {
		s := strings.ToLower(v)
		s = strings.ReplaceAll(s, " ", "-")
		return s
	},
	"prefix": func(v string, pre string) string {
		return pre + v
	},
})
```

```html
<a href="{{ title | slugify | prefix('/blog/') }}">{{ title }}</a>
<!-- "My Blog Post" -> "/blog/my-blog-post" -->
```

## Function Signatures

VueGo uses reflection to support arbitrary function signatures, matching `text/template` behavior.

Functions can accept any number and type of arguments:

```go
func(string) string                      // Single typed argument
func(int, int) int                       // Multiple typed arguments
func(*time.Time) string                  // Pointer types
func(int, string) (string, error)        // Returns value and error
func(parts ...string) string             // Variadic functions
func(v any) any                          // Generic any type
```

**Type Conversion:**
- Arguments are automatically converted when possible (e.g., string "42" to int 42)
- String literals in templates are parsed as their natural type (numbers, bools, strings)
- Values from template data retain their original types (*time.Time, custom structs, etc.)

**Error Handling:** When a function returns an error (second return value), template rendering fails with a descriptive error message including the template filename and function name.

## Error Messages

When function calls fail, VueGo provides detailed error messages:

```
in test.html: in expression '{{ items | double }}': double(): cannot convert argument 0 from []string to int
```

The error includes:

- Template filename
- The full expression
- Function name
- Specific error (type mismatch, division by zero, etc.)

## Notes

- Functions are resolved from left to right in a pipe chain
- The first value in a pipe chain is resolved from template data
- Each function receives the output of the previous function as its first argument
- Additional arguments can be passed using function call syntax: `fn(arg1, arg2)`
- Arguments can be string literals (quoted) or variable references
- In `v-if` conditions, functions can be called directly: `v-if="len(items)"`
- **Template rendering fails with an error if a function doesn't exist or has type mismatches**
- All interpolated values are HTML-escaped by default for security
