# Concurrency Safety

Vuego is designed to be **safe for concurrent use** after initialization.

## Thread-Safe Design

The `*Vue` instance is **immutable after setup**, making it naturally thread-safe:

```go
vue := vuego.NewVue(templateFS).Funcs(vuego.FuncMap{
	"formatDate": func(t *time.Time) string {
		return t.Format("2006-01-02")
	},
})

// Safe to call from multiple goroutines
go vue.Render(w1, "page.html", data1)
go vue.Render(w2, "page.html", data2)
go vue.Render(w3, "other.html", data3)
```

## Request-Scoped State

Each call to `Render()` or `RenderFragment()` creates its own `VueContext` with a dedicated stack:

```go
type Vue struct {
	templateFS fs.FS      // Read-only after initialization
	loader     *Component // Read-only after initialization
	funcMap    FuncMap    // Read-only after Funcs()
}

type VueContext struct {
	stack         *Stack // Request-scoped!
	BaseDir       string
	CurrentDir    string
	FromFilename  string
	TemplateStack []string
}
```

**Each render gets its own isolated stack** → No shared mutable state → No race conditions!

## Usage Examples

### Web Server

```go
func main() {
	vue := vuego.NewVue(os.DirFS("templates")).Funcs(vuego.FuncMap{
		"formatTime": formatTime,
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := map[string]any{
			"user":      getCurrentUser(r),
			"timestamp": time.Now(),
		}

		// Safe - each request gets its own context
		vue.Render(w, "index.html", data)
	})

	http.ListenAndServe(":8080", nil)
}
```

### Parallel Rendering

```go
func renderMultiplePages(vue *vuego.Vue, pages []Page) []string {
	results := make([]string, len(pages))
	var wg sync.WaitGroup

	for i, page := range pages {
		wg.Add(1)
		go func(idx int, p Page) {
			defer wg.Done()

			var buf bytes.Buffer
			data := map[string]any{
				"title":   p.Title,
				"content": p.Content,
			}

			// Safe - each goroutine gets its own context
			vue.Render(&buf, "page.html", data)
			results[idx] = buf.String()
		}(i, page)
	}

	wg.Wait()
	return results
}
```

## What's Thread-Safe

✅ **Safe for concurrent use:**

- `vue.Render()`
- `vue.RenderFragment()`
- All template functions in `FuncMap`
- Reading from `templateFS`

## What's NOT Thread-Safe

❌ **Not safe after first render:**

- `vue.Funcs()` - Call this during initialization only

**Best Practice:** Set up your Vue instance completely before using it concurrently:

```go
// ✅ Good: Setup then use
vue := vuego.NewVue(fs).Funcs(myFuncs)
// Now safe for concurrent rendering

// ❌ Bad: Modifying during use
vue := vuego.NewVue(fs)
go vue.Render(...)           // Started rendering
vue.Funcs(additionalFuncs)   // RACE CONDITION!
```

## Performance

Request-scoped allocation is **extremely cheap** in Go:

- Each `VueContext` is stack-allocated
- Garbage collector handles cleanup efficiently
- No mutex contention
- Scales linearly with CPU cores

## Testing

The test suite includes race detection:

```bash
go test ./... -race
```

All tests pass with the race detector, ensuring thread safety.
