# Package vuego

```go
import (
	"github.com/titpetric/vuego"
)
```

## Types

<details>
<summary><code>type Component</code></summary>

```go
// Component is a renderable .vuego template component.
type Component interface {
	// Render renders the component template to the given writer.
	Render(ctx context.Context, w io.Writer, filename string, data any) error
}
```

</details>

<details>
<summary><code>type ExprEvaluator</code></summary>

```go
// ExprEvaluator wraps expr for evaluating boolean and interpolated expressions.
// It caches compiled programs to avoid recompilation.
type ExprEvaluator struct {
	mu       sync.RWMutex
	programs map[string]*vm.Program
}
```

</details>

<details>
<summary><code>type FuncMap</code></summary>

```go
// FuncMap maps template function names to functions, as text/template does.
// A function takes any number of parameters and returns one value, or a value
// and an error. An optional first parameter of *VueContext or context.Context
// receives the execution context. See docs/funcmap.md.
type FuncMap map[string]any
```

</details>

<details>
<summary><code>type LessProcessor</code></summary>

```go
// LessProcessor implements vuego.NodeProcessor.
// It implements the functionality to compile Less CSS to standard css.
// It processes `script` tags with a `type="text/css+less"` attribute.
type LessProcessor struct {
	// fs is optional filesystem for loading @import statements in LESS files
	fs fs.FS
}
```

</details>

<details>
<summary><code>type LessProcessorError</code></summary>

```go
// LessProcessorError wraps processing errors with context.
type LessProcessorError struct {
	// Err is the underlying error from the LESS processor.
	Err error
	// Reason is a descriptive message about the LESS processing failure.
	Reason string
}
```

</details>

<details>
<summary><code>type LoadOption</code></summary>

```go
// LoadOption is a functional option for configuring Load().
type LoadOption func(*Vue)
```

</details>

<details>
<summary><code>type Loader</code></summary>

```go
// Loader loads and parses .vuego files from an fs.FS.
type Loader struct {
	// FS is the file system used to load templates.
	FS fs.FS
}
```

</details>

<details>
<summary><code>type NodeProcessor</code></summary>

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

</details>

<details>
<summary><code>type OverlayFS</code></summary>

```go
// OverlayFS overlays two filesystems.
// This allows extension of the lower filesystem with modified files,
// new files and encourages composition of the contents of a `fs.FS`.
type OverlayFS struct {
	chainFS []fs.FS
}
```

</details>

<details>
<summary><code>type Renderer</code></summary>

```go
// Renderer renders parsed HTML nodes to output.
type Renderer interface {
	// Render renders HTML nodes to the given writer.
	Render(ctx context.Context, w io.Writer, nodes []*html.Node) error
}
```

</details>

<details>
<summary><code>type SlotContent</code></summary>

```go
// SlotContent holds the processed content for a slot along with any scoped props.
type SlotContent struct {
	// Nodes contains the rendered content for this slot.
	Nodes []*html.Node
	// Props contains any scoped props passed to the slot.
	Props map[string]any
	// TemplateNode holds the original template node for processing scoped slots.
	TemplateNode *html.Node
}
```

</details>

<details>
<summary><code>type SlotScope</code></summary>

```go
// SlotScope holds all slot contents indexed by name for a component instance.
type SlotScope struct {
	// Slots maps slot names to their content.
	// Empty string key is the default slot.
	Slots map[string]*SlotContent
}
```

</details>

<details>
<summary><code>type Stack</code></summary>

```go
// Stack provides stack-based variable lookup and convenient typed accessors.
type Stack struct {
	stack    []map[string]any // bottom..top, top is last element
	rootData any              // original data passed to Render (for struct field fallback)

	// envCache caches the flattened map from EnvMap().
	// Invalidated on Push/Pop/Set. Stack is request-scoped so no mutex needed.
	envCache map[string]any
}
```

</details>

<details>
<summary><code>type Template</code></summary>

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

</details>

<details>
<summary><code>type TemplateConstructors</code></summary>

```go
// TemplateConstructors creates new template rendering contexts.
// The results provide request scoped data allocations and should be discarded after use.
type TemplateConstructors interface {
	New() Template
	Load(filename string) Template
}
```

</details>

<details>
<summary><code>type TemplateRendering</code></summary>

```go
// TemplateRendering bundles the interface for the render functions.
type TemplateRendering interface {
	Render(ctx context.Context, w io.Writer) error
}
```

</details>

<details>
<summary><code>type TemplateRenderingDetail</code></summary>

```go
// TemplateRenderingDetail the interface for stateless render functions.
type TemplateRenderingDetail interface {
	RenderFile(ctx context.Context, w io.Writer, filename string) error
	RenderString(ctx context.Context, w io.Writer, templateStr string) error
	RenderByte(ctx context.Context, w io.Writer, templateData []byte) error
	RenderReader(ctx context.Context, w io.Writer, r io.Reader) error
}
```

</details>

<details>
<summary><code>type TemplateState</code></summary>

```go
// TemplateState bundles the interface for template state management.
type TemplateState interface {
	Fill(vars any) Template
	Assign(key string, value any) Template
	Get(key string) string
}
```

</details>

<details>
<summary><code>type Vue</code></summary>

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
	templateCache map[string]*templateCacheEntry
	templateMu    sync.RWMutex

	// Custom node processors for post-processing rendered DOM
	nodeProcessors []NodeProcessor

	// Component shorthand mapping: maps kebab-case tag names to component filenames
	// e.g., "button-primary" -> "components/ButtonPrimary.vuego"
	componentMap map[string]string

	// initialData is pre-loaded data from WithData() that seeds the template stack.
	// This is equivalent to calling Fill() on the base template before any New()/Load().
	initialData map[string]any
}
```

</details>

<details>
<summary><code>type VueContext</code></summary>

```go
// VueContext carries template inclusion context and request-scoped state used during evaluation.
// Each render operation gets its own VueContext, making concurrent rendering safe.
type VueContext struct {
	// Variable scope and data resolution
	ctx   context.Context
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

	// SlotScope contains slot content for the current component.
	SlotScope *SlotScope
}
```

</details>

<details>
<summary><code>type VueContextOptions</code></summary>

```go
// VueContextOptions holds configurable options for a new VueContext.
type VueContextOptions struct {
	// Stack is the resolver stack for variable lookups.
	Stack *Stack
	// Processors are the registered template processors.
	Processors []NodeProcessor
}
```

</details>

## Function symbols

- `func GetSlotName (attr string) string`
- `func New (opts ...LoadOption) Template`
- `func NewExprEvaluator () *ExprEvaluator`
- `func NewFS (templateFS fs.FS, opts ...LoadOption) Template`
- `func NewLessProcessor (fsys ...fs.FS) *LessProcessor`
- `func NewLoader (fs fs.FS) *Loader`
- `func NewOverlayFS (upper fs.FS, lower ...fs.FS) *OverlayFS`
- `func NewRenderer () Renderer`
- `func NewSlotScope () *SlotScope`
- `func NewStack (root map[string]any) *Stack`
- `func NewStackWithData (root map[string]any, originalData any) *Stack`
- `func NewVue (templateFS fs.FS) *Vue`
- `func NewVueContext (ctx context.Context, fromFilename string, options *VueContextOptions) VueContext`
- `func View (renderer Template, filename string, data V) Template`
- `func WithComponents () LoadOption`
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
- `func (*OverlayFS) Glob (pattern string) ([]string, error)`
- `func (*OverlayFS) Open (name string) (fs.File, error)`
- `func (*OverlayFS) ReadDir (name string) ([]fs.DirEntry, error)`
- `func (*SlotScope) GetSlot (name string) *SlotContent`
- `func (*SlotScope) SetSlot (name string, content *SlotContent)`
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
- `func (*Vue) GetComponentFile (tagName string) (string, bool)`
- `func (*Vue) RegisterComponent (tagName,filename string) *Vue`
- `func (*Vue) RegisterNodeProcessor (processor NodeProcessor) *Vue`
- `func (*Vue) Render (ctx context.Context, w io.Writer, filename string, data any) error`
- `func (*Vue) RenderFragment (ctx context.Context, w io.Writer, filename string, data any) error`
- `func (*VueContext) PopTag ()`
- `func (*VueContext) PushTag (tag string)`
- `func (VueContext) Context () context.Context`
- `func (VueContext) CurrentTag () string`
- `func (VueContext) ExprEnv () map[string]any`
- `func (VueContext) FormatTemplateChain () string`
- `func (VueContext) Stack () *Stack`
- `func (VueContext) WithTemplate (filename string) VueContext`

### GetSlotName

GetSlotName extracts the slot name from a v-slot directive. A directive naming none yields "default".

```go
func GetSlotName(attr string) string
```

### New

New creates a Template that renders strings, bytes or readers without a filesystem. Call NewFS(fs) or New(WithFS(fs)) to render files instead. The returned Template holds assigned variables and renders them.

```go
func New(opts ...LoadOption) Template
```

### NewExprEvaluator

NewExprEvaluator creates a new ExprEvaluator with an empty cache.

```go
func NewExprEvaluator() *ExprEvaluator
```

### NewFS

NewFS creates a Template over the given filesystem, applying each option in order. The returned Template renders files, strings or bytes and holds assigned variables.

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

### NewOverlayFS

NewOverlayFS will create a new *OverlayFS.

```go
func NewOverlayFS(upper fs.FS, lower ...fs.FS) *OverlayFS
```

### NewRenderer

NewRenderer creates a new Renderer.

```go
func NewRenderer() Renderer
```

### NewSlotScope

NewSlotScope creates a new SlotScope for a component.

```go
func NewSlotScope() *SlotScope
```

### NewStack

NewStack constructs a Stack over an optional root map; nil yields an empty one. It records no original data, so it carries no struct field fallback.

```go
func NewStack(root map[string]any) *Stack
```

### NewStackWithData

NewStackWithData constructs a Stack with both map data and original root data for struct field fallback.

```go
func NewStackWithData(root map[string]any, originalData any) *Stack
```

### NewVue

NewVue creates a Vue over the given filesystem. Goroutines may share the returned Vue concurrently once it stops changing.

```go
func NewVue(templateFS fs.FS) *Vue
```

### NewVueContext

NewVueContext returns a VueContext initialized for the given template filename with initial data.

```go
func NewVueContext(ctx context.Context, fromFilename string, options *VueContextOptions) VueContext
```

### View

View binds a template file to a data model type. Render the returned template once, then discard it.

```go
func View(renderer Template, filename string, data V) Template
```

### WithComponents

WithComponents returns a LoadOption that registers every component shorthand. It walks the components folder and its subfolders for .vuego files, turning each path into a kebab-case tag name from the directory and the filename.

```go
func WithComponents() LoadOption
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

WithLessProcessor returns a LoadOption that registers a LESS processor.

```go
func WithLessProcessor() LoadOption
```

### WithProcessor

WithProcessor returns a LoadOption that registers a custom node processor.

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
- Comparison: ==, !=, <, >, <=, >=, === (same as ==, for convenience)
- Boolean operations: &&, ||, !
- Function calls: len(items), isActive(v)
- Literals: 42, "text", true, false.

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

PreProcess does nothing. PostProcess carries the compile.

```go
func (*LessProcessor) PreProcess(nodes []*html.Node) error
```

### Error

Error returns the combined reason and underlying error message.

```go
func (*LessProcessorError) Error() string
```

### Load

Load parses a full HTML document from the given filename.

```go
func (*Loader) Load(filename string) ([]*html.Node, error)
```

### LoadFragment

LoadFragment parses a template fragment, falling back to Load for a full document. It strips the front matter and drops it; loadFragment returns it.

```go
func (*Loader) LoadFragment(filename string) ([]*html.Node, error)
```

### Stat

Stat checks that filename exists in the loader filesystem.

```go
func (*Loader) Stat(filename string) error
```

### Glob

Glob implements combined FS reading.

```go
func (*OverlayFS) Glob(pattern string) ([]string, error)
```

### Open

Open opens a file in the overlaid filesystem.

```go
func (*OverlayFS) Open(name string) (fs.File, error)
```

### ReadDir

ReadDir implements combined FS reading.

```go
func (*OverlayFS) ReadDir(name string) ([]fs.DirEntry, error)
```

### GetSlot

GetSlot retrieves slot content by name. Returns nil if the slot doesn't exist.

```go
func (*SlotScope) GetSlot(name string) *SlotContent
```

### SetSlot

SetSlot sets the content for a named slot.

```go
func (*SlotScope) SetSlot(name string, content *SlotContent)
```

### Copy

Copy returns a throwaway copy of the stack. It keeps the root data and flattens the scopes into a fresh map.

Copy flattens without touching envCache. Every goroutine calling Load on a base template reads that template's stack, so a write to the cache there races.

```go
func (*Stack) Copy() *Stack
```

### EnvMap

EnvMap converts the Stack to a map[string]any for expr evaluation, holding every value the scopes and the struct fields reach. It caches the result and hands the same map back until Push, Pop or Set invalidates it.

EnvMap serves a request-scoped stack, which one goroutine owns; the cache carries no mutex. Call Copy to read a stack another goroutine may hold.

```go
func (*Stack) EnvMap() map[string]any
```

### ForEach

ForEach walks the collection at expr and calls fn(index, value) for each element. It accepts slices, arrays and maps; a map walks in no set order. An error from fn stops the walk and reaches the caller unchanged.

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

Pop discards the top-most scope. Popping the last one leaves an empty scope behind rather than an empty slice. The discarded map returns to the pool.

```go
func (*Stack) Pop()
```

### Push

Push adds m as the top-most scope. A nil m takes an empty map from the pool instead.

```go
func (*Stack) Push(m map[string]any)
```

### Resolve

Resolve resolves dotted/bracketed expression paths like:

```go
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

DefaultFuncMap returns the built-in template functions. docs/funcmap.md documents every key.

```go
func (*Vue) DefaultFuncMap() FuncMap
```

### Funcs

Funcs merges custom template functions into the existing funcmap, overwriting any existing keys. Returns the Vue instance for chaining.

```go
func (*Vue) Funcs(funcMap FuncMap) *Vue
```

### GetComponentFile

GetComponentFile looks up a component file by its kebab-case tag name. Returns the filename and true if found, otherwise returns empty string and false.

```go
func (*Vue) GetComponentFile(tagName string) (string, bool)
```

### RegisterComponent

RegisterComponent registers a component shorthand mapping, so a tagName of "button-primary" resolves to a filename of "components/ButtonPrimary.vuego".

```go
func (*Vue) RegisterComponent(tagName, filename string) *Vue
```

### RegisterNodeProcessor

RegisterNodeProcessor adds a custom node processor to the Vue instance. Each processor runs in registration order once template evaluation ends.

```go
func (*Vue) RegisterNodeProcessor(processor NodeProcessor) *Vue
```

### Render

Render processes a full-page template file and writes the output to w. Front matter in the template wins over the data the caller hands in. Goroutines may call Render concurrently.

```go
func (*Vue) Render(ctx context.Context, w io.Writer, filename string, data any) error
```

### RenderFragment

RenderFragment processes a template fragment file and writes the output to w. Front matter in the template wins over the data the caller hands in. Goroutines may call RenderFragment concurrently.

```go
func (*Vue) RenderFragment(ctx context.Context, w io.Writer, filename string, data any) error
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

### Context

Context returns the context.Context associated with this VueContext.

```go
func (VueContext) Context() context.Context
```

### CurrentTag

CurrentTag returns the current parent tag, or the empty string when the tag stack holds none.

```go
func (VueContext) CurrentTag() string
```

### ExprEnv

ExprEnv returns the env map for expr evaluation. It wraps each function taking a context.Context first, so a template calls it without one and the wrapper supplies the render context.

```go
func (VueContext) ExprEnv() map[string]any
```

### FormatTemplateChain

FormatTemplateChain returns the template inclusion chain formatted for error messages.

```go
func (VueContext) FormatTemplateChain() string
```

### Stack

Stack returns the variable resolution stack for this context. This allows functions with *VueContext parameters to resolve variables from the execution scope.

```go
func (VueContext) Stack() *Stack
```

### WithTemplate

WithTemplate returns a copy of the context extended with filename in the inclusion chain.

```go
func (VueContext) WithTemplate(filename string) VueContext
```
