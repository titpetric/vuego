# Testing Guide

## Pipeline

[atkins](https://github.com/titpetric/atkins) drives the repository. `atkins.yml` defines the jobs:

| Job       | What it runs                                                                                   |
|-----------|------------------------------------------------------------------------------------------------|
| `default` | fmt, lint, test and cover in order                                                             |
| `fmt`     | `splint fix ./...` and `go mod tidy`                                                           |
| `lint`    | `splint --linters godoc,pairing,grouping,filecheck,coverage ./...`                             |
| `test`    | `go test -count 1 -cover -coverpkg github.com/titpetric/vuego/...`, replayed through gotestsum |
| `cover`   | regenerates `docs/api.md` and `docs/testing-coverage.md`                                       |

Run a job by name:

```bash
atkins            # the default job
atkins test
atkins cover
```

`Taskfile.yml` carries two extra targets: `task uncover` lists source the last run never reached, and `task bench` runs the benchmarks at 1, 2 and 4 CPUs.

The test job writes two gitignored files: `pkg.cov`, the coverage profile, and `test.json`, the `go test -json` stream gotestsum replays.

## Running tests directly

```bash
go test ./...
go test -cover ./...
go test -race ./...
go test -run TestStack_Push ./
```

`TestTemplate_ConcurrentNestedIncludesRender` in `tests/` reports a data race under `-race`. The pipeline does not pass `-race`.

## Layout

Tests run black box: every test file sits in the external test package, such as `vuego_test` or `markdown_test`, and reaches the exported API and nothing more. Each source file has a `_test.go` file beside it, and `splint --linters pairing` fails when one does not.

Shared helpers live in the test file paired with the source they exercise:

- `vue_test.go` holds `renderFragment`, `renderPage`, `renderFuncs`, `renderCtx`, `renderInclude` and `assertRenderHTML`
- `renderer_test.go` holds `parseNodes` and `renderNodes`
- `markdown/markdown_test.go` holds `render`

Assertions come from `tests/assert`, an internal package with the subset of the testify API this repository uses.

```go
func TestVue_Render(t *testing.T) {
	vue := vuego.NewVue(os.DirFS("testdata/pages"))

	var buf bytes.Buffer
	assert.NoError(t, vue.Render(t.Context(), &buf, "Index.vuego", nil))
	assert.Contains(t, buf.String(), "expected content")
}
```

## Naming

A test takes the name of the symbol it covers, `Test[Receiver_]Function`:

- methods: `TestVue_Render`, `TestStack_Push`
- functions: `TestNewVue`, `TestParseFor`
- unexported functions, named without a receiver: `TestEvalCondition`, `TestConvertValue`

`splint --linters coverage` reports an exported symbol that no test references and no test takes its name from.

## Fixtures

### `testdata/fixtures/`

Three files share a base name:

```
testdata/fixtures/
├── v-for.vuego        # template with the directive under test
├── v-for.json         # data the template is filled with
└── v-for.html         # expected output
```

`TestFixtures` in `vue_test.go` globs every `*.vuego` under the directory, fills it from the matching `.json` and compares the result against the `.html` with `assert.EqualHTML`, which ignores whitespace in text nodes and attribute order.

To add a case, write the three files:

```bash
echo '<div v-if="show">Hello</div>' > testdata/fixtures/my-test.vuego
echo '{"show": true}' > testdata/fixtures/my-test.json
echo '<div>Hello</div>' > testdata/fixtures/my-test.html

go test -run TestFixtures ./
```

### `testdata/pages/`

Full page templates with a `components/` directory, `theme.yml` and LESS sources, used by the layout and component tests.

### Other directories

`testdata/LoadConfig/` covers the `theme.yml` plus `data/` loading order, `testdata/nodeprocessor/` covers the LESS processor, and `testdata/bugs/`, `testdata/required-error-test/` and `testdata/root-template-required/` cover error cases.

## Coverage

`atkins cover` writes `docs/testing-coverage.md` from the profile. A function passes when it has 80% line coverage, or has cognitive complexity 0, or has complexity under 5 and any coverage at all.

`uncover pkg.cov` prints the source no test reached. Pass a function name to narrow it:

```bash
uncover pkg.cov
uncover pkg.cov Stack.Resolve
```
