# Agent Guidelines for vuego

This file contains conventions and preferences for AI agents working on this codebase.

## Testing

### Black Box Testing Philosophy
- **Prefer black box tests** using `package vuego_test` instead of `package vuego`
- Tests should only interact with exported APIs, not internal implementation
- This allows running individual test files: `go test -v stack_test.go`
- Avoid testing internal/private functions and types directly

**Why black box?**
- Forces good API design by testing from user's perspective
- Tests are more resilient to refactoring
- Can run individual test files independently
- Better documentation through examples

### Test Framework
- Use `github.com/stretchr/testify/require` for assertions
- **Avoid** `t.Fatal()`, `t.Error()`, and `t.Errorf()` - use testify/require instead
- Use `t.Logf()` for informational output

### Test File Organization
- **Each `.go` file must have a corresponding `_test.go` file** in the same package
- Example: `vue.go` → `vue_test.go`, `stack.go` → `stack_test.go`
- Group tests for related functionality in the same test file as the implementation
- This prevents test duplication and makes it easy to find tests for a specific module
- **Standalone/integration tests** should be placed in the `tests/` subfolder
  - These test scenarios across multiple components (e.g., concurrency, integration scenarios)
  - Example: `tests/concurrent_test.go`, `tests/integration_test.go`
  - Run with: `go test -v tests/concurrent_test.go`
- **Benchmarks** should be in the `tests/` subfolder with the `_bench_test.go` suffix
  - Example: `tests/concurrent_bench_test.go`
  - Run with: `go test -bench -v tests/concurrent_bench_test.go`

### Test Naming Convention

Follow the pattern `Test[Receiver_]Function`:
- Methods: `TestVue_Render`, `TestStack_Push`, `TestComponent_Load`
- Functions: `TestNewVue`, `TestParseFor`
- Descriptive tests: `TestRequiredAttributeError`, `TestFixtures`

**Examples:**

```go
func TestVue_Render(t *testing.T)    { /* tests (*Vue).Render */ }
func TestStack_Resolve(t *testing.T) { /* tests (*Stack).Resolve */ }
func TestNewVue(t *testing.T)        { /* tests NewVue constructor */ }
```

### Avoid Test Duplication
- **Don't test the same code path in multiple test files**
- Test each exported function/method once in its corresponding `_test.go` file
- If functionality is tested in one place, don't re-test it elsewhere
- For integration tests that combine multiple components, create a dedicated test in the main test file (e.g., `vue_concurrent_test.go` for concurrency-specific tests)
- Review existing tests before adding new ones to avoid redundant coverage

### Assertion Examples

**Good:**

```go
require.NoError(t, err)
require.Equal(t, expected, actual)
require.Contains(t, str, substring)
require.True(t, condition, "optional message")
```

**Error message assertions:**

```go
// Good: Assert on complete error message
require.Equal(t, "in test.html: in expression '{{ items | double }}': double(): cannot convert argument 0 from []string to int", err.Error())

// Avoid: Multiple partial assertions on error
require.Contains(t, err.Error(), "double()") // Don't decompose errors
require.Contains(t, err.Error(), "cannot convert")
```

**Avoid:**

```go
if err != nil {
	t.Fatal(err) // Don't do this
}
if got != want {
	t.Errorf("got %v, want %v", got, want) // Don't do this
}
```

### Running Tests

See [docs/testing.md](docs/testing.md) for comprehensive testing documentation.

**Quick commands:**

```bash
# Run all tests
go test ./...

# Run specific test file (black box tests only)
go test -v stack_test.go

# Run specific test function
go test -v -run TestStack_Push

# Full test suite with coverage
task test
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
