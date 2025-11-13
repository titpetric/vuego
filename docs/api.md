# Package vuego

```go
import (
	"github.com/titpetric/vuego"
}
```

## Types

```go
// Component loads and parses .vuego component files from an fs.FS.
type Component struct {
	FS fs.FS
}
```

```go
// Stack provides stack-based variable lookup and convenient typed accessors.
type Stack struct {
	stack []map[string]any	// bottom..top, top is last element
}
```

```go
// Vue is the main template renderer for .vuego templates.
type Vue struct {
	loader	*Component
	stack	*Stack
}
```

```go
// VueContext carries template inclusion context used during evaluation and for error reporting.
type VueContext struct {
	BaseDir		string
	CurrentDir	string
	FromFilename	string
	TemplateStack	[]string
}
```

## Function symbols

- `func NewComponent (fs fs.FS) *Component`
- `func NewStack (root map[string]any) *Stack`
- `func NewVue (templateFS fs.FS) *Vue`
- `func NewVueContext (fromFilename string) VueContext`
- `func (*Stack) ForEach (expr string, fn func(index int, value any) error) error`
- `func (*Stack) GetInt (expr string) (int, bool)`
- `func (*Stack) GetMap (expr string) (map[string]any, bool)`
- `func (*Stack) GetSlice (expr string) ([]any, bool)`
- `func (*Stack) GetString (expr string) (string, bool)`
- `func (*Stack) Lookup (name string) (any, bool)`
- `func (*Stack) Pop ()`
- `func (*Stack) Push (m map[string]any)`
- `func (*Stack) Resolve (expr string) (any, bool)`
- `func (*Stack) Set (key string, val any)`
- `func (*Vue) Render (w io.Writer, filename string, data map[string]any) error`
- `func (*Vue) RenderFragment (w io.Writer, filename string, data map[string]any) error`
- `func (Component) Load (filename string) ([]*html.Node, error)`
- `func (Component) LoadFragment (filename string) ([]*html.Node, error)`
- `func (Component) Stat (filename string) error`
- `func (VueContext) FormatTemplateChain () string`
- `func (VueContext) WithTemplate (filename string) VueContext`

### NewComponent

NewComponent creates a Component backed by fs.

```go
func NewComponent (fs fs.FS) *Component
```

### NewStack

NewStack constructs a Stack with an optional initial root map (nil allowed).

```go
func NewStack (root map[string]any) *Stack
```

### NewVue

NewVue creates a new Vue backed by the given filesystem.

```go
func NewVue (templateFS fs.FS) *Vue
```

### NewVueContext

NewVueContext returns a VueContext initialized for the given template filename.

```go
func NewVueContext (fromFilename string) VueContext
```

### ForEach

ForEach iterates over a collection at the given expr and calls fn(index,value).
Supports slices/arrays and map[string]any (iteration order for maps is unspecified).
If fn returns an error iteration is stopped and the error passed through.

```go
func (*Stack) ForEach (expr string, fn func(index int, value any) error) error
```

### GetInt

GetInt resolves and tries to return an int (best-effort).

```go
func (*Stack) GetInt (expr string) (int, bool)
```

### GetMap

GetMap returns map[string]any or converts map[string]string to map[string]any.
Avoids reflection for other map types.

```go
func (*Stack) GetMap (expr string) (map[string]any, bool)
```

### GetSlice

GetSlice returns a []any for supported slice kinds. Avoids reflection by only converting known types.

```go
func (*Stack) GetSlice (expr string) ([]any, bool)
```

### GetString

GetString resolves and tries to return a string.

```go
func (*Stack) GetString (expr string) (string, bool)
```

### Lookup

Lookup searches stack from top to bottom for a plain identifier (no dots).
Returns (value, true) if found.

```go
func (*Stack) Lookup (name string) (any, bool)
```

### Pop

Pop the top-most Stack. If only root remains it still pops to empty slice safely.

```go
func (*Stack) Pop ()
```

### Push

Push a new map as a top-most Stack.

```go
func (*Stack) Push (m map[string]any)
```

### Resolve

Resolve resolves dotted/bracketed expression paths like:

	"user.name", "items[0].title", "mapKey.sub"

It returns (value, true) if resolution succeeded.

```go
func (*Stack) Resolve (expr string) (any, bool)
```

### Set

Set sets a key in the top-most Stack.

```go
func (*Stack) Set (key string, val any)
```

### Render

Render processes a full-page template file and writes the output to w.

```go
func (*Vue) Render (w io.Writer, filename string, data map[string]any) error
```

### RenderFragment

RenderFragment processes a template fragment file and writes the output to w.

```go
func (*Vue) RenderFragment (w io.Writer, filename string, data map[string]any) error
```

### Load

Load parses a full HTML document from the given filename.

```go
func (Component) Load (filename string) ([]*html.Node, error)
```

### LoadFragment

LoadFragment parses a template fragment; if the file is a full document, it falls back to Load.

```go
func (Component) LoadFragment (filename string) ([]*html.Node, error)
```

### Stat

Stat checks that filename exists in the component filesystem.

```go
func (Component) Stat (filename string) error
```

### FormatTemplateChain

FormatTemplateChain returns the template inclusion chain formatted for error messages.

```go
func (VueContext) FormatTemplateChain () string
```

### WithTemplate

WithTemplate returns a copy of the context extended with filename in the inclusion chain.

```go
func (VueContext) WithTemplate (filename string) VueContext
```


