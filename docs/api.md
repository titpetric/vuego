# Package vuego

```go
import (
	"github.com/titpetric/vuego"
}
```

## Types

```go
// Component is a renderable .vuego template component.
type Component interface {
	// Render renders the component template to the given writer.
	Render(ctx context.Context, w io.Writer, filename string, data any) error
}
```

```go
// ExprEvaluator wraps expr for evaluating boolean and interpolated expressions.
// It caches compiled programs to avoid recompilation.
type ExprEvaluator struct {
	mu       sync.RWMutex
	programs map[string]*vm.Program
}
```

```go
// FuncMap is a map of function names to functions, similar to text/template's FuncMap.
// Functions can have any number of parameters and must return 1 or 2 values.
// If 2 values are returned, the second must be an error.
type FuncMap map[string]any
```

```go
// LessProcessor implements vuego.NodeProcessor.
// It implements the functionality to compile Less CSS to standard css.
// It processes `script` tags with a `type="text/css+less"` attribute.
type LessProcessor struct {
	// fs is optional filesystem for loading @import statements in LESS files
	fs fs.FS
}
```

```go
// LessProcessorError wraps processing errors with context.
type LessProcessorError struct {
	// Err is the underlying error from the LESS processor.
	Err error
	// Reason is a descriptive message about the LESS processing failure.
	Reason string
}
```

```go
// LoadOption is a functional option for configuring Load().
type LoadOption func(*Vue)
```

```go
// Loader loads and parses .vuego files from an fs.FS.
type Loader struct {
	// FS is the file system used to load templates.
	FS fs.FS
}
```

```go
// NodeProcessor is an interface for post-processing []*html.Node after template evaluation.
// Implementations can inspect and modify HTML nodes to implement custom tags and attributes.
// NodeProcessor receives the complete rendered DOM and can transform it in place.
//
// Processors are called after all Vue directives and interpolations have been evaluated.
// Multiple processors can be registered and are applied in order of registration.
//
// See docs/nodeprocessor.md for examples and best practices.
type NodeProcessor interface {
	// New ensures that we can have render-scoped allocations.
	New() NodeProcessor

	// PreProcess receives the nodes as they are decoded.
	PreProcess(nodes []*html.Node) error

	// PostProcess receives the rendered nodes and may modify them in place.
	// It should return an error if processing fails.
	PostProcess(nodes []*html.Node) error
}
```

```go
// Renderer renders parsed HTML nodes to output.
type Renderer interface {
	// Render renders HTML nodes to the given writer.
	Render(ctx context.Context, w io.Writer, nodes []*html.Node) error
}
```

```go
// Stack provides stack-based variable lookup and convenient typed accessors.
type Stack struct {
	stack    []map[string]any // bottom..top, top is last element
	rootData any              // original data passed to Render (for struct field fallback)
}
```

```go
// Template represents a prepared vuego template.
// It allows variable assignment and rendering with internal buffering.
type Template interface {
	TemplateConstructors
	TemplateState
	TemplateRendering
	TemplateRenderingDetail
}
```

```go
// TemplateConstructors creates new template rendering contexts.
// The results provide request scoped data allocations and should be discarded after use.
type TemplateConstructors interface {
	New() Template
	Load(filename string) Template
}
```

```go
// TemplateRendering bundles the interface for the render functions.
type TemplateRendering interface {
	Render(ctx context.Context, w io.Writer) error
	Layout(ctx context.Context, w io.Writer) error
}
```

```go
// TemplateRenderingDetail the interface for stateless render functions.
type TemplateRenderingDetail interface {
	RenderFile(ctx context.Context, w io.Writer, filename string) error
	RenderString(ctx context.Context, w io.Writer, templateStr string) error
	RenderByte(ctx context.Context, w io.Writer, templateData []byte) error
	RenderReader(ctx context.Context, w io.Writer, r io.Reader) error
}
```

```go
// TemplateState bundles the interface for template state management.
type TemplateState interface {
	Fill(vars any) Template
	Assign(key string, value any) Template
	Get(key string) string
}
```

```go
// Vue is the main template renderer for .vuego templates.
// After initialization, Vue is safe for concurrent use by multiple goroutines.
type Vue struct {
	templateFS fs.FS
	loader     *Loader
	renderer   Renderer
	funcMap    FuncMap
	exprEval   *ExprEvaluator

	// Template cache to avoid re-parsing the same template
	templateCache map[string][]*html.Node
	templateMu    sync.RWMutex

	// Custom node processors for post-processing rendered DOM
	nodeProcessors []NodeProcessor
}
```

```go
// VueContext carries template inclusion context and request-scoped state used during evaluation.
// Each render operation gets its own VueContext, making concurrent rendering safe.
type VueContext struct {
	// Variable scope and data resolution
	stack *Stack

	// BaseDir is the root directory for template inclusion chains.
	BaseDir string
	// CurrentDir is the current working directory during template processing.
	CurrentDir string
	// FromFilename is the name of the file currently being processed.
	FromFilename string
	// TemplateStack is the stack of included template files.
	TemplateStack []string

	// TagStack is the stack of HTML tags being rendered.
	TagStack []string

	// Processors are the registered template processors.
	Processors []NodeProcessor

	// v-once element tracking for deep clones
	seen map[string]bool
}
```

```go
// VueContextOptions holds configurable options for a new VueContext.
type VueContextOptions struct {
	// Stack is the resolver stack for variable lookups.
	Stack *Stack
	// Processors are the registered template processors.
	Processors []NodeProcessor
}
```

## Function symbols

- `func New (opts ...LoadOption) Template`
- `func NewExprEvaluator () *ExprEvaluator`
- `func NewFS (templateFS fs.FS, opts ...LoadOption) Template`
- `func NewLessProcessor (fsys ...fs.FS) *LessProcessor`
- `func NewLoader (fs fs.FS) *Loader`
- `func NewRenderer () Renderer`
- `func NewStack (root map[string]any) *Stack`
- `func NewStackWithData (root map[string]any, originalData any) *Stack`
- `func NewVue (templateFS fs.FS) *Vue`
- `func NewVueContext (fromFilename string, options *VueContextOptions) VueContext`
- `func WithFS (templateFS fs.FS) LoadOption`
- `func WithFuncs (funcMap FuncMap) LoadOption`
- `func WithLessProcessor () LoadOption`
- `func WithProcessor (processor NodeProcessor) LoadOption`
- `func (*ExprEvaluator) ClearCache ()`
- `func (*ExprEvaluator) Eval (expression string, env map[string]any) (any, error)`
- `func (*LessProcessor) New () NodeProcessor`
- `func (*LessProcessor) PostProcess (nodes []*html.Node) error`
- `func (*LessProcessor) PreProcess (nodes []*html.Node) error`
- `func (*LessProcessorError) Error () string`
- `func (*Loader) Load (filename string) ([]*html.Node, error)`
- `func (*Loader) LoadFragment (filename string) ([]*html.Node, error)`
- `func (*Loader) Stat (filename string) error`
- `func (*Stack) Copy () *Stack`
- `func (*Stack) EnvMap () map[string]any`
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
- `func (*Vue) DefaultFuncMap () FuncMap`
- `func (*Vue) Funcs (funcMap FuncMap) *Vue`
- `func (*Vue) RegisterNodeProcessor (processor NodeProcessor) *Vue`
- `func (*Vue) Render (w io.Writer, filename string, data any) error`
- `func (*Vue) RenderFragment (w io.Writer, filename string, data any) error`
- `func (*Vue) RenderNodes (w io.Writer, nodes []*html.Node, data any) error`
- `func (*VueContext) PopTag ()`
- `func (*VueContext) PushTag (tag string)`
- `func (VueContext) CurrentTag () string`
- `func (VueContext) FormatTemplateChain () string`
- `func (VueContext) WithTemplate (filename string) VueContext`

### New

New creates a new Template for rendering strings, bytes, or readers without a filesystem. Use this when templates are provided as strings/bytes rather than loaded from files. To render from files, use NewFS(fs) or New(WithFS(fs)) instead. The returned Template can be used for variable assignment and rendering.

```go
func New(opts ...LoadOption) Template
```

### NewExprEvaluator

NewExprEvaluator creates a new ExprEvaluator with an empty cache.

```go
func NewExprEvaluator() *ExprEvaluator
```

### NewFS

NewFS creates a new Template with access to the given filesystem and optional configurations. The returned Template can be used to render files, strings, or bytes with variable assignment.

```go
func NewFS(templateFS fs.FS, opts ...LoadOption) Template
```

### NewLessProcessor

NewLessProcessor creates a new LESS processor.

```go
func NewLessProcessor(fsys ...fs.FS) *LessProcessor
```

### NewLoader

NewLoader creates a Loader backed by fs.

```go
func NewLoader(fs fs.FS) *Loader
```

### NewRenderer

NewRenderer creates a new Renderer.

```go
func NewRenderer() Renderer
```

### NewStack

NewStack constructs a Stack with an optional initial root map (nil allowed). The originalData parameter is the original value passed to Render (for struct field fallback).

```go
func NewStack(root map[string]any) *Stack
```

### NewStackWithData

NewStackWithData constructs a Stack with both map data and original root data for struct field fallback.

```go
func NewStackWithData(root map[string]any, originalData any) *Stack
```

### NewVue

NewVue creates a new Vue backed by the given filesystem. The returned Vue is safe for concurrent use by multiple goroutines.

```go
func NewVue(templateFS fs.FS) *Vue
```

### NewVueContext

NewVueContext returns a VueContext initialized for the given template filename with initial data.

```go
func NewVueContext(fromFilename string, options *VueContextOptions) VueContext
```

### WithFS

WithFS returns a LoadOption that sets the filesystem for template loading.

```go
func WithFS(templateFS fs.FS) LoadOption
```

### WithFuncs

WithFuncs returns a LoadOption that merges custom template functions into the existing funcmap.

```go
func WithFuncs(funcMap FuncMap) LoadOption
```

### WithLessProcessor

WithLessProcessor returns a LoadOption that registers a LESS processor

```go
func WithLessProcessor() LoadOption
```

### WithProcessor

WithProcessor returns a LoadOption that registers a custom node processor

```go
func WithProcessor(processor NodeProcessor) LoadOption
```

### ClearCache

ClearCache clears the program cache (useful for testing or memory management).

```go
func (*ExprEvaluator) ClearCache()
```

### Eval

Eval evaluates an expression against the given environment (stack). It returns the result value and any error. The expression can contain:
- Variable references: item, item.title, items[0]
- Comparison: ==, !=, <, >, <=, >=
- Boolean operations: &&, ||, !
- Function calls: len(items), isActive(v)
- Literals: 42, "text", true, false

```go
func (*ExprEvaluator) Eval(expression string, env map[string]any) (any, error)
```

### New

New creates a new allocation of *LessProcessor.

```go
func (*LessProcessor) New() NodeProcessor
```

### PostProcess

PostProcess walks the DOM tree and performs compilation.

```go
func (*LessProcessor) PostProcess(nodes []*html.Node) error
```

### PreProcess

PreProcess.

```go
func (*LessProcessor) PreProcess(nodes []*html.Node) error
```

### Load

Load parses a full HTML document from the given filename.

```go
func (*Loader) Load(filename string) ([]*html.Node, error)
```

### LoadFragment

LoadFragment parses a template fragment; if the file is a full document, it falls back to Load. Front-matter is extracted and discarded; use loadFragment to access it.

```go
func (*Loader) LoadFragment(filename string) ([]*html.Node, error)
```

### Stat

Stat checks that filename exists in the loader filesystem.

```go
func (*Loader) Stat(filename string) error
```

### Copy

Copy returns a copy of the stack that can be discarded. The root data is retained as is, the envmap is a copy.

```go
func (*Stack) Copy() *Stack
```

### EnvMap

EnvMap converts the Stack to a map[string]any for expr evaluation. Includes all accessible values from stack and struct fields.

```go
func (*Stack) EnvMap() map[string]any
```

### ForEach

ForEach iterates over a collection at the given expr and calls fn(index,value). Supports slices/arrays and map[string]any (iteration order for maps is unspecified). If fn returns an error iteration is stopped and the error passed through.

```go
func (*Stack) ForEach(expr string, fn func(index int, value any) error) error
```

### GetInt

GetInt resolves and tries to return an int (best-effort).

```go
func (*Stack) GetInt(expr string) (int, bool)
```

### GetMap

GetMap returns map[string]any or converts map[string]string to map[string]any. Avoids reflection for other map types.

```go
func (*Stack) GetMap(expr string) (map[string]any, bool)
```

### GetSlice

GetSlice returns a []any for slice types.

```go
func (*Stack) GetSlice(expr string) ([]any, bool)
```

### GetString

GetString resolves and tries to return a string.

```go
func (*Stack) GetString(expr string) (string, bool)
```

### Lookup

Lookup searches stack from top to bottom for a plain identifier (no dots). If not found in the stack maps, it checks the root data struct (if any). Returns (value, true) if found.

```go
func (*Stack) Lookup(name string) (any, bool)
```

### Pop

Pop the top-most Stack. If only root remains it still pops to empty slice safely. Returns pooled maps to reduce GC pressure.

```go
func (*Stack) Pop()
```

### Push

Push a new map as a top-most Stack. If m is nil, an empty map is obtained from the pool.

```go
func (*Stack) Push(m map[string]any)
```

### Resolve

Resolve resolves dotted/bracketed expression paths like:

```
"user.name", "items[0].title", "mapKey.sub"
```

It returns (value, true) if resolution succeeded.

```go
func (*Stack) Resolve(expr string) (any, bool)
```

### Set

Set sets a key in the top-most Stack.

```go
func (*Stack) Set(key string, val any)
```

### DefaultFuncMap

DefaultFuncMap returns a FuncMap with built-in utility functions

```go
func (*Vue) DefaultFuncMap() FuncMap
```

### Funcs

Funcs merges custom template functions into the existing funcmap, overwriting any existing keys. Returns the Vue instance for chaining.

```go
func (*Vue) Funcs(funcMap FuncMap) *Vue
```

### RegisterNodeProcessor

RegisterNodeProcessor adds a custom node processor to the Vue instance. Processors are applied in order after template evaluation completes.

```go
func (*Vue) RegisterNodeProcessor(processor NodeProcessor) *Vue
```

### Render

Render processes a full-page template file and writes the output to w. Front-matter data in the template is authoritative and overrides passed data. Render is safe to call concurrently from multiple goroutines.

```go
func (*Vue) Render(w io.Writer, filename string, data any) error
```

### RenderFragment

RenderFragment processes a template fragment file and writes the output to w. Front-matter data in the template is authoritative and overrides passed data. RenderFragment is safe to call concurrently from multiple goroutines.

```go
func (*Vue) RenderFragment(w io.Writer, filename string, data any) error
```

### RenderNodes

RenderNodes evaluates and renders HTML nodes with the given data. This is the core rendering function used by all public render methods.

```go
func (*Vue) RenderNodes(w io.Writer, nodes []*html.Node, data any) error
```

### PopTag

PopTag removes the current tag from the tag stack.

```go
func (*VueContext) PopTag()
```

### PushTag

PushTag adds a tag to the tag stack.

```go
func (*VueContext) PushTag(tag string)
```

### CurrentTag

CurrentTag returns the current parent tag, or empty string if no tag is on the stack.

```go
func (VueContext) CurrentTag() string
```

### FormatTemplateChain

FormatTemplateChain returns the template inclusion chain formatted for error messages.

```go
func (VueContext) FormatTemplateChain() string
```

### WithTemplate

WithTemplate returns a copy of the context extended with filename in the inclusion chain.

```go
func (VueContext) WithTemplate(filename string) VueContext
```

### Error

```go
func (*LessProcessorError) Error() string
```
