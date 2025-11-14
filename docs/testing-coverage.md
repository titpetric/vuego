# Testing Coverage

Testing criteria for a passing coverage requirement:

- Line coverage of 80%
- Cognitive complexity of 0
- Have cognitive complexity < 5, but have any coverage

Low cognitive complexity means there are few conditional branches to cover. Tests with cognitive complexity 0 would be covered by invocation.

## Packages

| Status | Package                              | Coverage | Cognitive | Lines |
|--------|--------------------------------------|----------|-----------|-------|
| ✅     | titpetric/vuego                      | 82.09%   | 470       | 1547  |
| ❌     | titpetric/vuego/cmd/vuego            | 0.00%    | 5         | 33    |
| ❌     | titpetric/vuego/cmd/vuego-playground | 0.00%    | 6         | 65    |
| ✅     | titpetric/vuego/internal/helpers     | 96.29%   | 120       | 313   |
| ✅     | titpetric/vuego/internal/reflect     | 92.65%   | 55        | 198   |

## Functions

| Status | Package                              | Function                       | Coverage | Cognitive |
|--------|--------------------------------------|--------------------------------|----------|-----------|
| ✅     | titpetric/vuego                      | Component.Load                 | 80.00%   | 3         |
| ✅     | titpetric/vuego                      | Component.LoadFragment         | 85.70%   | 12        |
| ✅     | titpetric/vuego                      | Component.Stat                 | 0.00%    | 0         |
| ✅     | titpetric/vuego                      | DefaultFuncMap                 | 100.00%  | 0         |
| ✅     | titpetric/vuego                      | ExprEvaluator.ClearCache       | 0.00%    | 0         |
| ✅     | titpetric/vuego                      | ExprEvaluator.Eval             | 85.70%   | 2         |
| ✅     | titpetric/vuego                      | ExprEvaluator.getProgram       | 100.00%  | 2         |
| ✅     | titpetric/vuego                      | NewComponent                   | 100.00%  | 0         |
| ✅     | titpetric/vuego                      | NewExprEvaluator               | 100.00%  | 0         |
| ✅     | titpetric/vuego                      | NewStack                       | 100.00%  | 0         |
| ✅     | titpetric/vuego                      | NewStackWithData               | 100.00%  | 1         |
| ✅     | titpetric/vuego                      | NewVue                         | 100.00%  | 0         |
| ✅     | titpetric/vuego                      | NewVueContext                  | 0.00%    | 0         |
| ✅     | titpetric/vuego                      | NewVueContextWithData          | 100.00%  | 0         |
| ✅     | titpetric/vuego                      | Stack.EnvMap                   | 100.00%  | 4         |
| ✅     | titpetric/vuego                      | Stack.ForEach                  | 93.90%   | 32        |
| ✅     | titpetric/vuego                      | Stack.GetInt                   | 73.30%   | 5         |
| ✅     | titpetric/vuego                      | Stack.GetMap                   | 100.00%  | 5         |
| ✅     | titpetric/vuego                      | Stack.GetSlice                 | 100.00%  | 11        |
| ✅     | titpetric/vuego                      | Stack.GetString                | 72.70%   | 3         |
| ✅     | titpetric/vuego                      | Stack.Lookup                   | 100.00%  | 6         |
| ✅     | titpetric/vuego                      | Stack.Pop                      | 80.00%   | 2         |
| ✅     | titpetric/vuego                      | Stack.Push                     | 100.00%  | 1         |
| ✅     | titpetric/vuego                      | Stack.Resolve                  | 84.40%   | 42        |
| ✅     | titpetric/vuego                      | Stack.Set                      | 66.70%   | 1         |
| ✅     | titpetric/vuego                      | Stack.splitPath                | 81.80%   | 25        |
| ✅     | titpetric/vuego                      | Vue.Funcs                      | 80.00%   | 2         |
| ✅     | titpetric/vuego                      | Vue.Render                     | 88.90%   | 2         |
| ✅     | titpetric/vuego                      | Vue.RenderFragment             | 77.80%   | 2         |
| ✅     | titpetric/vuego                      | Vue.callFunc                   | 92.30%   | 24        |
| ✅     | titpetric/vuego                      | Vue.evalAttributes             | 85.70%   | 13        |
| ✅     | titpetric/vuego                      | Vue.evalCondition              | 90.00%   | 2         |
| ✅     | titpetric/vuego                      | Vue.evalFilter                 | 100.00%  | 6         |
| ✅     | titpetric/vuego                      | Vue.evalFor                    | 81.80%   | 6         |
| ✅     | titpetric/vuego                      | Vue.evalPipe                   | 81.80%   | 16        |
| ✅     | titpetric/vuego                      | Vue.evalSegment                | 70.00%   | 5         |
| ✅     | titpetric/vuego                      | Vue.evalTemplate               | 87.50%   | 16        |
| ✅     | titpetric/vuego                      | Vue.evalVHtml                  | 82.10%   | 14        |
| ✅     | titpetric/vuego                      | Vue.evaluate                   | 94.50%   | 81        |
| ✅     | titpetric/vuego                      | Vue.interpolate                | 100.00%  | 14        |
| ✅     | titpetric/vuego                      | Vue.render                     | 75.00%   | 3         |
| ✅     | titpetric/vuego                      | Vue.resolveArgument            | 92.30%   | 10        |
| ✅     | titpetric/vuego                      | VueContext.FormatTemplateChain | 66.70%   | 1         |
| ✅     | titpetric/vuego                      | VueContext.WithTemplate        | 100.00%  | 0         |
| ❌     | titpetric/vuego                      | classifySegment                | 80.00%   | 7         |
| ❌     | titpetric/vuego                      | convertValue                   | 65.00%   | 24        |
| ✅     | titpetric/vuego                      | defaultFunc                    | 100.00%  | 2         |
| ✅     | titpetric/vuego                      | escapeFunc                     | 100.00%  | 1         |
| ✅     | titpetric/vuego                      | formatTimeFunc                 | 62.50%   | 4         |
| ✅     | titpetric/vuego                      | intFunc                        | 42.90%   | 3         |
| ✅     | titpetric/vuego                      | lenFunc                        | 28.60%   | 1         |
| ✅     | titpetric/vuego                      | lowerFunc                      | 66.70%   | 1         |
| ✅     | titpetric/vuego                      | parseArgs                      | 100.00%  | 11        |
| ✅     | titpetric/vuego                      | parseFor                       | 86.70%   | 8         |
| ✅     | titpetric/vuego                      | parsePipeExpr                  | 92.30%   | 7         |
| ✅     | titpetric/vuego                      | renderAttrs                    | 100.00%  | 1         |
| ✅     | titpetric/vuego                      | renderNode                     | 95.00%   | 16        |
| ✅     | titpetric/vuego                      | stringFunc                     | 100.00%  | 0         |
| ✅     | titpetric/vuego                      | titleFunc                      | 85.70%   | 6         |
| ✅     | titpetric/vuego                      | toMapData                      | 100.00%  | 2         |
| ✅     | titpetric/vuego                      | trimFunc                       | 66.70%   | 1         |
| ✅     | titpetric/vuego                      | upperFunc                      | 66.70%   | 1         |
| ❌     | titpetric/vuego/cmd/vuego            | main                           | 0.00%    | 1         |
| ❌     | titpetric/vuego/cmd/vuego            | start                          | 0.00%    | 4         |
| ❌     | titpetric/vuego/cmd/vuego-playground | handleRender                   | 0.00%    | 4         |
| ❌     | titpetric/vuego/cmd/vuego-playground | main                           | 0.00%    | 2         |
| ✅     | titpetric/vuego/internal/helpers     | CloneNode                      | 100.00%  | 0         |
| ✅     | titpetric/vuego/internal/helpers     | CompareHTML                    | 93.80%   | 8         |
| ✅     | titpetric/vuego/internal/helpers     | ContainsPipe                   | 100.00%  | 3         |
| ✅     | titpetric/vuego/internal/helpers     | DeepCloneNode                  | 100.00%  | 5         |
| ✅     | titpetric/vuego/internal/helpers     | GetAttr                        | 100.00%  | 3         |
| ✅     | titpetric/vuego/internal/helpers     | IsComplexExpr                  | 100.00%  | 3         |
| ✅     | titpetric/vuego/internal/helpers     | IsFunctionCall                 | 100.00%  | 5         |
| ✅     | titpetric/vuego/internal/helpers     | IsIdentifier                   | 100.00%  | 14        |
| ✅     | titpetric/vuego/internal/helpers     | IsIdentifierChar               | 100.00%  | 4         |
| ✅     | titpetric/vuego/internal/helpers     | IsTruthy                       | 87.50%   | 4         |
| ✅     | titpetric/vuego/internal/helpers     | NormalizeComparisonOperators   | 100.00%  | 7         |
| ✅     | titpetric/vuego/internal/helpers     | RemoveAttr                     | 100.00%  | 3         |
| ✅     | titpetric/vuego/internal/helpers     | SignificantChildren            | 95.80%   | 30        |
| ✅     | titpetric/vuego/internal/helpers     | attrsEqual                     | 92.30%   | 8         |
| ✅     | titpetric/vuego/internal/helpers     | compareNodeRecursive           | 87.50%   | 17        |
| ✅     | titpetric/vuego/internal/helpers     | filteredChildren               | 100.00%  | 3         |
| ✅     | titpetric/vuego/internal/helpers     | isIgnorable                    | 80.00%   | 3         |
| ✅     | titpetric/vuego/internal/reflect     | CanDescend                     | 100.00%  | 5         |
| ❌     | titpetric/vuego/internal/reflect     | PopulateStructFields           | 79.20%   | 17        |
| ✅     | titpetric/vuego/internal/reflect     | ResolveValue                   | 100.00%  | 2         |
| ❌     | titpetric/vuego/internal/reflect     | StructToMap                    | 73.10%   | 18        |
| ✅     | titpetric/vuego/internal/reflect     | resolveMap                     | 100.00%  | 1         |
| ✅     | titpetric/vuego/internal/reflect     | resolveSliceIndex              | 100.00%  | 2         |
| ✅     | titpetric/vuego/internal/reflect     | resolveStruct                  | 100.00%  | 6         |
| ✅     | titpetric/vuego/internal/reflect     | resolveValueRecursive          | 88.90%   | 4         |
