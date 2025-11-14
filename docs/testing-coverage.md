# Testing Coverage

Testing criteria for a passing coverage requirement:

- Line coverage of 80%
- Cognitive complexity of 0
- Have cognitive complexity < 5, but have any coverage

Low cognitive complexity means there are few conditional branches to cover. Tests with cognitive complexity 0 would be covered by invocation.

## Packages

| Status | Package                              | Coverage | Cognitive | Lines |
|--------|--------------------------------------|----------|-----------|-------|
| ✅     | titpetric/vuego                      | 81.21%   | 506       | 1601  |
| ❌     | titpetric/vuego/cmd/vuego            | 0.00%    | 5         | 33    |
| ❌     | titpetric/vuego/cmd/vuego-playground | 0.00%    | 6         | 65    |
| ✅     | titpetric/vuego/internal/helpers     | 86.66%   | 80        | 223   |

## Functions

| Status | Package                              | Function                       | Coverage | Cognitive |
|--------|--------------------------------------|--------------------------------|----------|-----------|
| ✅     | titpetric/vuego                      | Component.Load                 | 80.00%   | 3         |
| ✅     | titpetric/vuego                      | Component.LoadFragment         | 85.70%   | 12        |
| ✅     | titpetric/vuego                      | Component.Stat                 | 0.00%    | 0         |
| ✅     | titpetric/vuego                      | DefaultFuncMap                 | 100.00%  | 0         |
| ✅     | titpetric/vuego                      | ExprEvaluator.ClearCache       | 0.00%    | 0         |
| ✅     | titpetric/vuego                      | ExprEvaluator.Eval             | 71.40%   | 2         |
| ✅     | titpetric/vuego                      | ExprEvaluator.getProgram       | 91.70%   | 2         |
| ✅     | titpetric/vuego                      | NewComponent                   | 100.00%  | 0         |
| ✅     | titpetric/vuego                      | NewExprEvaluator               | 100.00%  | 0         |
| ✅     | titpetric/vuego                      | NewStack                       | 100.00%  | 1         |
| ✅     | titpetric/vuego                      | NewVue                         | 100.00%  | 0         |
| ✅     | titpetric/vuego                      | NewVueContext                  | 100.00%  | 0         |
| ✅     | titpetric/vuego                      | Stack.ForEach                  | 93.90%   | 32        |
| ✅     | titpetric/vuego                      | Stack.GetInt                   | 73.30%   | 5         |
| ✅     | titpetric/vuego                      | Stack.GetMap                   | 100.00%  | 5         |
| ✅     | titpetric/vuego                      | Stack.GetSlice                 | 100.00%  | 11        |
| ✅     | titpetric/vuego                      | Stack.GetString                | 72.70%   | 3         |
| ✅     | titpetric/vuego                      | Stack.Lookup                   | 100.00%  | 3         |
| ✅     | titpetric/vuego                      | Stack.Pop                      | 80.00%   | 2         |
| ✅     | titpetric/vuego                      | Stack.Push                     | 100.00%  | 1         |
| ✅     | titpetric/vuego                      | Stack.Resolve                  | 81.40%   | 38        |
| ✅     | titpetric/vuego                      | Stack.Set                      | 66.70%   | 1         |
| ✅     | titpetric/vuego                      | Stack.splitPath                | 81.80%   | 25        |
| ✅     | titpetric/vuego                      | Vue.Funcs                      | 80.00%   | 2         |
| ✅     | titpetric/vuego                      | Vue.Render                     | 87.50%   | 2         |
| ✅     | titpetric/vuego                      | Vue.RenderFragment             | 75.00%   | 2         |
| ✅     | titpetric/vuego                      | Vue.callFunc                   | 92.30%   | 24        |
| ✅     | titpetric/vuego                      | Vue.evalAttributes             | 85.70%   | 13        |
| ✅     | titpetric/vuego                      | Vue.evalCondition              | 60.00%   | 2         |
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
| ❌     | titpetric/vuego                      | containsPipe                   | 0.00%    | 3         |
| ❌     | titpetric/vuego                      | convertValue                   | 55.00%   | 24        |
| ✅     | titpetric/vuego                      | defaultFunc                    | 100.00%  | 2         |
| ✅     | titpetric/vuego                      | escapeFunc                     | 100.00%  | 1         |
| ✅     | titpetric/vuego                      | formatTimeFunc                 | 62.50%   | 4         |
| ✅     | titpetric/vuego                      | getEnvMap                      | 100.00%  | 3         |
| ✅     | titpetric/vuego                      | intFunc                        | 42.90%   | 3         |
| ✅     | titpetric/vuego                      | isComplexExpr                  | 100.00%  | 3         |
| ✅     | titpetric/vuego                      | isFunctionCall                 | 100.00%  | 5         |
| ❌     | titpetric/vuego                      | isIdentifier                   | 62.50%   | 14        |
| ✅     | titpetric/vuego                      | isIdentifierChar               | 100.00%  | 4         |
| ✅     | titpetric/vuego                      | isTruthy                       | 87.50%   | 4         |
| ✅     | titpetric/vuego                      | lenFunc                        | 28.60%   | 1         |
| ✅     | titpetric/vuego                      | lowerFunc                      | 66.70%   | 1         |
| ✅     | titpetric/vuego                      | normalizeComparisonOperators   | 100.00%  | 7         |
| ✅     | titpetric/vuego                      | parseArgs                      | 100.00%  | 11        |
| ✅     | titpetric/vuego                      | parseFor                       | 86.70%   | 8         |
| ✅     | titpetric/vuego                      | parsePipeExpr                  | 92.30%   | 7         |
| ✅     | titpetric/vuego                      | renderAttrs                    | 100.00%  | 1         |
| ✅     | titpetric/vuego                      | renderNode                     | 95.00%   | 16        |
| ✅     | titpetric/vuego                      | stringFunc                     | 100.00%  | 0         |
| ✅     | titpetric/vuego                      | titleFunc                      | 85.70%   | 6         |
| ✅     | titpetric/vuego                      | trimFunc                       | 66.70%   | 1         |
| ❌     | titpetric/vuego                      | trimSpace                      | 71.40%   | 6         |
| ✅     | titpetric/vuego                      | upperFunc                      | 66.70%   | 1         |
| ❌     | titpetric/vuego/cmd/vuego            | main                           | 0.00%    | 1         |
| ❌     | titpetric/vuego/cmd/vuego            | start                          | 0.00%    | 4         |
| ❌     | titpetric/vuego/cmd/vuego-playground | handleRender                   | 0.00%    | 4         |
| ❌     | titpetric/vuego/cmd/vuego-playground | main                           | 0.00%    | 2         |
| ✅     | titpetric/vuego/internal/helpers     | CloneNode                      | 100.00%  | 0         |
| ✅     | titpetric/vuego/internal/helpers     | CompareHTML                    | 81.20%   | 8         |
| ✅     | titpetric/vuego/internal/helpers     | DeepCloneNode                  | 100.00%  | 5         |
| ✅     | titpetric/vuego/internal/helpers     | GetAttr                        | 100.00%  | 3         |
| ✅     | titpetric/vuego/internal/helpers     | RemoveAttr                     | 100.00%  | 3         |
| ✅     | titpetric/vuego/internal/helpers     | attrsEqual                     | 84.60%   | 8         |
| ❌     | titpetric/vuego/internal/helpers     | compareNodeRecursive           | 62.50%   | 17        |
| ✅     | titpetric/vuego/internal/helpers     | filteredChildren               | 100.00%  | 3         |
| ✅     | titpetric/vuego/internal/helpers     | isIgnorable                    | 80.00%   | 3         |
| ❌     | titpetric/vuego/internal/helpers     | significantChildren            | 58.30%   | 30        |
