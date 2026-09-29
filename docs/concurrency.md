# Concurrency Safety

Goroutines may share a `*Vue` concurrently once configuration ends, and they may share a `Template` from `New` or `NewFS` as well: `Load` copies the base stack rather than writing to it.

## What each render owns

`Vue` holds the template filesystem, the loader, the funcmap, the registered node processors and the component map, none of which change after setup. The parsed-template cache alone mutates, and a `sync.RWMutex` guards it.

Every call to `Render` or `RenderFragment` builds its own `VueContext` with its own `Stack`, so no two renders share evaluation state:

```go
type VueContext struct {
	ctx   context.Context
	stack *Stack

	BaseDir       string
	CurrentDir    string
	FromFilename  string
	TemplateStack []string
	TagStack      []string
	Processors    []NodeProcessor
	SlotScope     *SlotScope
}
```

Every render calls `New` on each registered `NodeProcessor` for a fresh allocation, so a processor may keep per-render state on itself.

## Examples

### Web server

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

		if err := vue.Render(r.Context(), w, "index.vuego", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.ListenAndServe(":8080", nil)
}
```

### Parallel rendering

```go
func renderPages(ctx context.Context, tpl vuego.Template, pages []Page) []string {
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

			if err := tpl.Load("page.vuego").Fill(data).Render(ctx, &buf); err == nil {
				results[idx] = buf.String()
			}
		}(i, page)
	}

	wg.Wait()
	return results
}
```

## Safe concurrently

- `Vue.Render` and `Vue.RenderFragment`
- `Template.Load`, and the `Fill`, `Assign` and `Render` calls on the value it returns
- `Template.RenderFile`, `RenderString`, `RenderByte` and `RenderReader`
- every function in the FuncMap, so long as the function itself allows it
- reading from the template filesystem

## Not safe concurrently

Anything that changes the renderer, once a render has started:

- `Vue.Funcs`
- `Vue.RegisterNodeProcessor` and `Vue.RegisterComponent`
- `Fill` and `Assign` on the *base* template, the value `New` or `NewFS` returned

Configure first, then render:

```go
// Configured once, then shared.
vue := vuego.NewVue(fs).Funcs(myFuncs)
go vue.Render(ctx, w1, "page.vuego", data1)
go vue.Render(ctx, w2, "page.vuego", data2)

// A write during a render is a race.
vue := vuego.NewVue(fs)
go vue.Render(ctx, w, "page.vuego", data)
vue.Funcs(additionalFuncs)
```

The same holds for a `Template`: fill the base with the data every render shares, then let each goroutine call `Load` and fill its own copy.

```go
tpl := vuego.NewFS(fs).Fill(sharedData) // before any goroutine starts

go tpl.Load("a.vuego").Fill(dataA).Render(ctx, w1)
go tpl.Load("b.vuego").Fill(dataB).Render(ctx, w2)
```

## Allocation

A render allocates a `VueContext`, a `Stack` holding the flattened data, and a deep clone of the parsed DOM. Vuego caches the parse per filename and drops the entry when the file's modification time moves. `internal/helpers` pools `html.Node` values and `Stack` pools the maps `Push` and `Pop` move, so a render reuses each scope rather than allocating one.

## Testing

The concurrency tests live in `tests/`: `concurrent_test.go`, `concurrent_bench_test.go` and `template_nested_includes_test.go`.

```bash
go test -race ./...
```

The pipeline's `test` job does not pass `-race`; run it by hand when changing shared state.
