# Testing Guide

## Test Commands

The project uses [Task](https://taskfile.dev) for common development workflows. Here are the available test-related commands:

### `task` (default)

Starts the VueGo playground web server on http://localhost:8080

### `task test`

Runs the comprehensive test suite with coverage:

- Executes tests with race detection (`-race`)
- Fails fast on first error (`-failfast`)
- Collects coverage data across all packages
- Outputs coverage profile to `pkg.cov`

**Usage:**

```bash
task test
```

### `task cover`

Generates detailed coverage reports and documentation:

- Extracts source documentation with `go-fsck`
- Generates API documentation in `docs/api.md`
- Creates JSON coverage summary in `pkg.cov.json`
- Produces coverage report in `docs/testing-coverage.md`

**Usage:**

```bash
task cover
```

### `task uncover`

Shows uncovered source lines from the last test run:

```bash
task uncover
```

### `task render`

Renders a sample template for manual testing:

```bash
task render
```

## Running Individual Tests

### Black Box Testing

Tests in this project use **black box testing** (package `vuego_test`) to test only the exported API. This allows running individual test files independently:

```bash
# Run a specific test file
go test -v stack_test.go

# Run a specific test function
go test -v -run TestStack_Push

# Run all tests in a file with race detection
go test -v -race stack_test.go
```

### All Tests

```bash
# Run all tests
go test ./...

# Run all tests with coverage
go test -cover ./...

# Run with race detection and verbose output
go test -v -race ./...
```

## Test Fixtures

Test fixtures are organized in the `testdata/` directory:

### `testdata/fixtures/`

Contains paired test files for template rendering validation:

- **`.vuego`** - Template files with directives (v-if, v-for, v-bind, etc.)
- **`.json`** - Input data for the template
- **`.html`** - Expected output after rendering

Each test case consists of three files with the same base name:

```
testdata/fixtures/
├── v-for.vuego        # Template with v-for directive
├── v-for.json         # Test data (array/object to iterate)
└── v-for.html         # Expected rendered HTML
```

The test framework automatically discovers all `.vuego` files in fixtures and validates:

1. Template renders without errors
2. Output HTML matches the expected `.html` file
3. DOM structure is equivalent (whitespace-normalized, attribute order-insensitive)

### `testdata/pages/`

Contains full page templates and components for integration testing:

- Complete page layouts
- Reusable components in `components/` subdirectory
- Component composition and nesting tests
- Error case validation (e.g., `required-error-test/`)

### `testdata/json/`

Contains complex JSON data files for realistic rendering scenarios.

## Writing Tests

### Test Naming Convention

Follow the pattern `Test[Receiver_]Function`:

- Methods: `TestVue_Render`, `TestStack_Push`
- Functions: `TestNewVue`, `TestParseFor`
- Descriptive tests: `TestRequiredAttributeError`, `TestFixtures`

### Using testify/require

All assertions use `github.com/stretchr/testify/require`:

```go
func TestVue_Render(t *testing.T) {
	vue := vuego.NewVue(os.DirFS("testdata/pages"))
	var buf bytes.Buffer
	err := vue.Render(&buf, "Index.vuego", nil)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "expected content")
}
```

### Adding Fixture Tests

To add a new fixture test:

1. Create a `.vuego` template file in `testdata/fixtures/`
2. Create a `.json` data file with the same base name
3. Create a `.html` file with the expected output

The `TestFixtures` function will automatically discover and run your test.

Example:

```bash
# Create test files
echo '<div v-if="show">Hello</div>' > testdata/fixtures/my-test.vuego
echo '{"show": true}' > testdata/fixtures/my-test.json
echo '<div>Hello</div>' > testdata/fixtures/my-test.html

# Run the test
go test -v -run TestFixtures
```
