# Agent Guidelines for vuego

This file contains conventions and preferences for AI agents working on this codebase.

## Testing

### Test Framework
- Use `github.com/stretchr/testify/require` for assertions
- **Avoid** `t.Fatal()`, `t.Error()`, and `t.Errorf()` - use testify/require instead
- Use `t.Logf()` for informational output

### Test Naming Convention
Follow the pattern `Test[Receiver_]Function`:
- Methods: `TestVue_Render`, `TestStack_Push`, `TestComponent_Load`
- Functions: `TestNewVue`, `TestParseFor`
- Descriptive tests: `TestRequiredAttributeError`, `TestFixtures`

**Examples:**
```go
func TestVue_Render(t *testing.T) { /* tests (*Vue).Render */ }
func TestStack_Resolve(t *testing.T) { /* tests (*Stack).Resolve */ }
func TestNewVue(t *testing.T) { /* tests NewVue constructor */ }
```

### Assertion Examples

**Good:**
```go
require.NoError(t, err)
require.Equal(t, expected, actual)
require.Contains(t, str, substring)
require.True(t, condition, "optional message")
```

**Avoid:**
```go
if err != nil {
    t.Fatal(err)  // Don't do this
}
if got != want {
    t.Errorf("got %v, want %v", got, want)  // Don't do this
}
```

## Code Style

### Godoc Comments
- **Always** add godoc comments for exported types, functions, and methods
- Start with the name of the item being documented
- Be concise but descriptive

**Examples:**
```go
// Vue is the main template renderer.
type Vue struct { ... }

// NewVue creates a new Vue instance with the given template filesystem.
func NewVue(templateFS fs.FS) *Vue { ... }

// Render processes a template file and writes the output to w.
func (v *Vue) Render(w io.Writer, filename string, data map[string]any) error { ... }
```

### Error Messages
- Use lowercase for error messages (e.g., `"error loading %s"`)
- Use `fmt.Errorf` with `%w` to wrap errors
- Include context: filenames, template chains, variable names
- Make errors actionable and clear

**Examples:**
```go
return fmt.Errorf("error loading %s (included from %s): %w", name, ctx.FormatTemplateChain(), err)
return fmt.Errorf("required attribute '%s' not provided", attrName)
```

### File Organization
- Group related functionality in separate files (e.g., `eval_*.go` for evaluation logic)
- Place helper functions near their usage or in the most relevant file
- Keep type definitions and constructors together

## Build and Check Commands

To verify code changes:
```bash
go test ./...
go build ./...
```
