# FuncMap Advanced Examples

This document shows advanced usage of FuncMap with typed arguments and reflection.

## Working with Typed Data

### Time Formatting with *time.Time

```go
vue := vuego.NewVue(templateFS).Funcs(vuego.FuncMap{
	"formatDate": func(t *time.Time) string {
		if t == nil {
			return ""
		}
		return t.Format("January 2, 2006")
	},
})

now := time.Now()
data := map[string]any{
	"timestamp": &now,
}
```

Template:

```html
<p>{{ timestamp | formatDate }}</p>
<!-- Output: January 15, 2024 -->
```

### Type-Safe Math Functions

```go
vue := vuego.NewVue(templateFS).Funcs(vuego.FuncMap{
	"add": func(a, b int) int {
		return a + b
	},
	"multiply": func(a, b int) int {
		return a * b
	},
})
```

Template:

```html
<p>{{ add(5, 3) }}</p>          <!-- 8 -->
<p>{{ multiply(4, 7) }}</p>     <!-- 28 -->
<p>{{ price | multiply(2) }}</p> <!-- doubles price -->
```

### Automatic Type Conversion

VueGo automatically converts between compatible types:

```go
vue := vuego.NewVue(templateFS).Funcs(vuego.FuncMap{
	"repeat": func(s string, times int) string {
		return strings.Repeat(s, times)
	},
})
```

Template:

```html
<!-- String literal "5" is automatically converted to int 5 -->
<p>{{ "Hello" | repeat("5") }}</p>
<!-- Output: HelloHelloHelloHelloHello -->

<!-- Variable values use their actual types -->
<p>{{ message | repeat(count) }}</p>
```

### Variadic Functions

```go
vue := vuego.NewVue(templateFS).Funcs(vuego.FuncMap{
	"join": func(sep string, parts ...string) string {
		return strings.Join(parts, sep)
	},
	"sum": func(numbers ...int) int {
		total := 0
		for _, n := range numbers {
			total += n
		}
		return total
	},
})
```

Template:

```html
<p>{{ join(", ", "apple", "banana", "orange") }}</p>
<!-- Output: apple, banana, orange -->

<p>{{ sum(1, 2, 3, 4, 5) }}</p>
<!-- Output: 15 -->
```

### Error Handling

Functions can return errors; they're silently ignored:

```go
vue := vuego.NewVue(templateFS).Funcs(vuego.FuncMap{
	"divide": func(a, b int) (int, error) {
		if b == 0 {
			return 0, errors.New("division by zero")
		}
		return a / b, nil
	},
})
```

Template:

```html
<p>{{ divide(10, 2) }}</p>  <!-- 5 -->
<p>{{ divide(10, 0) }}</p>  <!-- empty, error silently ignored -->
```

## Custom Struct Methods

You can pass custom structs and use their methods:

```go
type User struct {
    FirstName string
    LastName  string
}

func (u User) FullName() string {
    return u.FirstName + " " + u.LastName
}

vue := vuego.NewVue(templateFS).Funcs(vuego.FuncMap{
    "fullName": func(u User) string {
        return u.FullName()
    },
})

data := map[string]any{
    "user": User{FirstName: "John", LastName: "Doe"},
}
```

Template:

```html
<p>{{ user | fullName }}</p>
<!-- Output: John Doe -->
```

## Chaining with Type Conversions

```go
vue := vuego.NewVue(templateFS).Funcs(vuego.FuncMap{
	"double": func(n int) int {
		return n * 2
	},
	"toString": func(n int) string {
		return strconv.Itoa(n)
	},
	"prefix": func(s string, p string) string {
		return p + s
	},
})
```

Template:

```html
<p>{{ 5 | double | toString | prefix("Number: ") }}</p>
<!-- Output: Number: 10 -->
```

## Context-Aware Functions

Functions can access data from the template context:

```go
vue := vuego.NewVue(templateFS).Funcs(vuego.FuncMap{
	"relativeTime": func(t time.Time, now time.Time) string {
		diff := now.Sub(t)
		hours := int(diff.Hours())
		if hours < 1 {
			return "just now"
		} else if hours < 24 {
			return fmt.Sprintf("%d hours ago", hours)
		}
		return fmt.Sprintf("%d days ago", hours/24)
	},
})

now := time.Now()
data := map[string]any{
	"publishedAt": now.Add(-3 * time.Hour),
	"now":         now,
}
```

Template:

```html
<p>{{ relativeTime(publishedAt, now) }}</p>
<!-- Output: 3 hours ago -->
```
