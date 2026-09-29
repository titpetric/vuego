# FuncMap: template functions

Vuego carries template functions the way Go's `text/template` and `html/template` do. An interpolation calls them through a pipe chain, and so does a `v-if` condition.

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

The `|` operator chains functions:

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

Vuego comes with several built-in utility functions:

### String Functions

- `upper` - Converts string to uppercase

  ```html
  {{ name | upper }}  <!-- "john" -> "JOHN" -->
  ```

- `lower` - Converts string to lowercase

  ```html
  {{ name | lower }}  <!-- "JOHN" -> "john" -->
  ```

- `title` - Title-cases string

  ```html
  {{ name | title }}  <!-- "john doe" -> "John Doe" -->
  ```

- `trim` - Removes leading and trailing whitespace

  ```html
  {{ text | trim }}  <!-- "  hello  " -> "hello" -->
  ```

### Utility Functions

- `default(fallback)` - Returns the fallback for an input holding nil or an empty string

  ```html
  {{ name | default("Anonymous") }}
  ```

- `len` - Returns length of string, array, or map

  ```html
  {{ items | len }}
  <div v-if="len(items)">Has items</div>
  ```

- `escape` - HTML-escapes the value. Interpolation already escapes outside `script` and `style`, so reach for this in the raw contexts where it does not

  ```html
  <script>const s = "{{ userInput | escape }}";</script>
  ```

- `type` - Returns the value's Go type, written as `%T`

  ```html
  {{ items | type }}  <!-- []string -->
  ```

- `int` - Converts value to integer

  ```html
  {{ stringValue | int }}
  <div v-if="int(count) > 0">Count is positive</div>
  ```

- `string` - Converts value to string

  ```html
  {{ count | string }}
  ```

- `json` - Converts value to JSON string

  ```html
  {{ data | json }}
  <script>var data = {{ user | json }};</script>
  ```

- `jsonPretty` - The same, indented

  ```html
  <script>const data = {{ user | jsonPretty }};</script>
  ```

### Date and time functions

Both require a layout argument and accept a `time.Time` or an RFC3339 string as the value. A value of any other type comes back unchanged.

- `formatTime(layout)` - Formats with the given Go layout. A non-string layout falls back to `2006-01-02 15:04:05`

  ```html
  {{ timestamp | formatTime("2006-01-02 15:04:05") }}
  {{ timestamp | formatTime("Jan 2, 2006") }}
  ```

- `formatDate(layout)` - The same, with a fallback of `2006-01-02`

  ```html
  {{ timestamp | formatDate("Jan 2 2006") }}
  ```

An argument that parses as a number arrives as a number, so a layout of nothing but digits never reaches the formatter and the fallback takes over: `formatDate("2006")` prints `2024-03-01`, not `2024`. Give the layout a non-digit character, as in `formatDate("Jan 2006")`.

### File functions

Each reads from the template filesystem. The argument names a file, or names a variable holding a filename.

- `file(name)` - Returns the file contents as a string

  ```html
  <div>{{ file("partials/notice.txt") }}</div>
  <div>{{ file(config.path) }}</div>
  ```

- `jsonFile(name)` - Decodes a JSON object into a `map[string]any`

  ```html
  {{ path | jsonFile }}
  ```

- `yamlFile(name)` - Decodes a YAML mapping into a `map[string]any`

  ```html
  {{ path | yamlFile }}
  ```

A call cannot sit on the left side of a pipe, so `jsonFile("config.json") | len` does not resolve. Pipe the filename into the function instead, as above, or render the value whole.

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

### Combining several custom functions

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

Vuego uses reflection to support arbitrary function signatures, matching `text/template` behavior.

A function takes any number of arguments, of any types:

```go
func(string) string                      // Single typed argument
func(int, int) int                       // Multiple typed arguments
func(*time.Time) string                  // Pointer types
func(int, string) (string, error)        // Returns value and error
func(parts ...string) string             // Variadic functions
func(v any) any                          // Generic any type
```

**Type Conversion:**
- A conversion happens where one exists, turning the string "42" into the int 42
- A literal in a template arrives as its natural type: a number, a bool or a string
- A value from the template data keeps its own type, such as *time.Time or a custom struct

**Error Handling:** When a function returns an error (second return value), template rendering fails with a descriptive error message including the template filename and function name.

## Error Messages

When function calls fail, Vuego provides detailed error messages:

```
in test.html: in expression '{{ items | double }}': double(): cannot convert argument 0 from []string to int
```

The error includes:

- Template filename
- The full expression
- Function name
- Specific error (type mismatch, division by zero, etc.)

## Notes

- A pipe chain resolves its functions left to right
- The template data supplies the first value in a pipe chain
- Each function receives the output of the previous function as its first argument
- Function call syntax carries further arguments: `fn(arg1, arg2)`
- An argument reads as a quoted string literal or as a variable reference
- A `v-if` condition calls a function outright: `v-if="len(items)"`
- Template rendering fails with an error if a function doesn't exist or has type mismatches
- The renderer HTML-escapes every interpolated value
