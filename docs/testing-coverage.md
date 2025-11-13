# Testing coverage

Testing criteria for a passing coverage requirement:

- Line coverage of 80%
- Cognitive complexity of 0
- Have cognitive complexity < 5, but have any coverage

Low cognitive complexity means there are few conditional branches to
cover. Tests with cognitive complexity 0 would be covered by invocation.

## Packages

| Status | Package                              | Coverage | Cognitive | Lines |
| ------ | ------------------------------------ | -------- | --------- | ----- |
| ✅      | titpetric/vuego           | 85.82%   | 315       | 988   |
| ❌      | titpetric/vuego/cmd/vuego | 0.00%    | 5         | 33    |

## Functions

| Status | Package                              | Function                       | Coverage | Cognitive |
| ------ | ------------------------------------ | ------------------------------ | -------- | --------- |
| ✅      | titpetric/vuego           | Component.Load                 | 80.00%   | 3         |
| ✅      | titpetric/vuego           | Component.LoadFragment         | 85.70%   | 12        |
| ✅      | titpetric/vuego           | Component.Stat                 | 0.00%    | 0         |
| ✅      | titpetric/vuego           | NewComponent                   | 100.00%  | 0         |
| ✅      | titpetric/vuego           | NewStack                       | 100.00%  | 1         |
| ✅      | titpetric/vuego           | NewVue                         | 100.00%  | 0         |
| ✅      | titpetric/vuego           | NewVueContext                  | 100.00%  | 0         |
| ❌      | titpetric/vuego           | Stack.ForEach                  | 60.60%   | 32        |
| ✅      | titpetric/vuego           | Stack.GetInt                   | 73.30%   | 5         |
| ✅      | titpetric/vuego           | Stack.GetMap                   | 100.00%  | 5         |
| ✅      | titpetric/vuego           | Stack.GetSlice                 | 100.00%  | 11        |
| ✅      | titpetric/vuego           | Stack.GetString                | 72.70%   | 3         |
| ✅      | titpetric/vuego           | Stack.Lookup                   | 100.00%  | 3         |
| ✅      | titpetric/vuego           | Stack.Pop                      | 80.00%   | 2         |
| ✅      | titpetric/vuego           | Stack.Push                     | 100.00%  | 1         |
| ✅      | titpetric/vuego           | Stack.Resolve                  | 81.40%   | 38        |
| ✅      | titpetric/vuego           | Stack.Set                      | 66.70%   | 1         |
| ✅      | titpetric/vuego           | Stack.splitPath                | 81.80%   | 25        |
| ✅      | titpetric/vuego           | Vue.Render                     | 90.00%   | 2         |
| ✅      | titpetric/vuego           | Vue.RenderFragment             | 80.00%   | 2         |
| ✅      | titpetric/vuego           | Vue.evalAttributes             | 76.50%   | 4         |
| ✅      | titpetric/vuego           | Vue.evalCondition              | 82.40%   | 5         |
| ✅      | titpetric/vuego           | Vue.evalFor                    | 81.80%   | 6         |
| ✅      | titpetric/vuego           | Vue.evalTemplate               | 87.50%   | 16        |
| ✅      | titpetric/vuego           | Vue.evalVHtml                  | 82.10%   | 14        |
| ✅      | titpetric/vuego           | Vue.evaluate                   | 94.40%   | 78        |
| ✅      | titpetric/vuego           | Vue.interpolate                | 100.00%  | 6         |
| ✅      | titpetric/vuego           | Vue.render                     | 75.00%   | 3         |
| ✅      | titpetric/vuego           | VueContext.FormatTemplateChain | 66.70%   | 1         |
| ✅      | titpetric/vuego           | VueContext.WithTemplate        | 100.00%  | 0         |
| ✅      | titpetric/vuego           | cloneNode                      | 100.00%  | 0         |
| ✅      | titpetric/vuego           | deepCloneNode                  | 100.00%  | 5         |
| ✅      | titpetric/vuego           | getAttr                        | 100.00%  | 3         |
| ✅      | titpetric/vuego           | parseFor                       | 86.70%   | 8         |
| ✅      | titpetric/vuego           | removeAttr                     | 100.00%  | 3         |
| ✅      | titpetric/vuego           | renderAttrs                    | 100.00%  | 1         |
| ✅      | titpetric/vuego           | renderNode                     | 90.00%   | 16        |
| ❌      | titpetric/vuego/cmd/vuego | main                           | 0.00%    | 1         |
| ❌      | titpetric/vuego/cmd/vuego | start                          | 0.00%    | 4         |

