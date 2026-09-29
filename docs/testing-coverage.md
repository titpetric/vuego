# Testing Coverage

A function passes coverage when it meets one of:

- Line coverage of 80%
- Cognitive complexity of 0
- Cognitive complexity under 5 with any coverage at all

Low cognitive complexity means the function holds little branching to cover. One call reaches every line of a function with cognitive complexity 0.

## Packages

| Status | Package          | Coverage | Cognitive | Lines |
|--------|------------------|----------|-----------|-------|
| ✅     | .                | 91.42%   | 1257      | 3827  |
| ✅     | diff             | 92.44%   | 155       | 435   |
| ✅     | formatter        | 97.83%   | 119       | 517   |
| ✅     | internal/helpers | 100.00%  | 84        | 277   |
| ✅     | internal/parser  | 84.62%   | 6         | 22    |
| ✅     | internal/reflect | 100.00%  | 44        | 162   |
| ✅     | internal/ulid    | 100.00%  | 0         | 15    |
| ✅     | markdown         | 92.54%   | 59        | 463   |
| ❌     | tests/assert     | 73.57%   | 65        | 266   |

## Functions

| Status | Package          | Function                          | Coverage | Cognitive |
|--------|------------------|-----------------------------------|----------|-----------|
| ✅     | .                | ExprEvaluator.ClearCache          | 100.00%  | 0         |
| ✅     |                  | ExprEvaluator.Eval                | 100.00%  | 2         |
| ✅     |                  | ExprEvaluator.getProgram          | 100.00%  | 2         |
| ✅     |                  | GetSlotName                       | 100.00%  | 1         |
| ✅     |                  | LessProcessor.New                 | 100.00%  | 0         |
| ✅     |                  | LessProcessor.PostProcess         | 75.00%   | 3         |
| ✅     |                  | LessProcessor.PreProcess          | 100.00%  | 0         |
| ✅     |                  | LessProcessor.compileLessTag      | 90.00%   | 7         |
| ✅     |                  | LessProcessor.isLessStyleTag      | 100.00%  | 4         |
| ❌     |                  | LessProcessor.processNode         | 70.00%   | 11        |
| ✅     |                  | LessProcessor.replaceWithStyleTag | 97.22%   | 3         |
| ✅     |                  | LessProcessorError.Error          | 100.00%  | 0         |
| ✅     |                  | Loader.Load                       | 100.00%  | 0         |
| ✅     |                  | Loader.LoadFragment               | 100.00%  | 1         |
| ✅     |                  | Loader.Stat                       | 100.00%  | 0         |
| ✅     |                  | Loader.loadFragment               | 100.00%  | 3         |
| ✅     |                  | New                               | 100.00%  | 0         |
| ✅     |                  | NewExprEvaluator                  | 100.00%  | 0         |
| ✅     |                  | NewFS                             | 100.00%  | 2         |
| ✅     |                  | NewLessProcessor                  | 100.00%  | 1         |
| ✅     |                  | NewLoader                         | 100.00%  | 0         |
| ✅     |                  | NewOverlayFS                      | 100.00%  | 0         |
| ✅     |                  | NewRenderer                       | 100.00%  | 0         |
| ✅     |                  | NewSlotScope                      | 100.00%  | 0         |
| ✅     |                  | NewStack                          | 100.00%  | 0         |
| ✅     |                  | NewStackWithData                  | 100.00%  | 1         |
| ✅     |                  | NewVue                            | 100.00%  | 0         |
| ✅     |                  | NewVueContext                     | 100.00%  | 1         |
| ✅     |                  | OverlayFS.Glob                    | 100.00%  | 6         |
| ✅     |                  | OverlayFS.Open                    | 100.00%  | 5         |
| ✅     |                  | OverlayFS.ReadDir                 | 100.00%  | 16        |
| ✅     |                  | SlotScope.GetSlot                 | 100.00%  | 0         |
| ✅     |                  | SlotScope.SetSlot                 | 100.00%  | 0         |
| ✅     |                  | Stack.Copy                        | 100.00%  | 0         |
| ✅     |                  | Stack.EnvMap                      | 100.00%  | 1         |
| ✅     |                  | Stack.ForEach                     | 100.00%  | 12        |
| ✅     |                  | Stack.GetInt                      | 100.00%  | 5         |
| ✅     |                  | Stack.GetMap                      | 100.00%  | 5         |
| ✅     |                  | Stack.GetSlice                    | 100.00%  | 3         |
| ✅     |                  | Stack.GetString                   | 90.91%   | 3         |
| ✅     |                  | Stack.Lookup                      | 100.00%  | 6         |
| ✅     |                  | Stack.Pop                         | 93.33%   | 6         |
| ✅     |                  | Stack.Push                        | 100.00%  | 1         |
| ✅     |                  | Stack.Resolve                     | 100.00%  | 7         |
| ✅     |                  | Stack.Set                         | 75.00%   | 1         |
| ✅     |                  | Stack.flatten                     | 100.00%  | 2         |
| ✅     |                  | Stack.resolveStep                 | 100.00%  | 8         |
| ✅     |                  | View                              | 100.00%  | 0         |
| ✅     |                  | Vue.DefaultFuncMap                | 100.00%  | 0         |
| ✅     |                  | Vue.Funcs                         | 100.00%  | 1         |
| ✅     |                  | Vue.GetComponentFile              | 100.00%  | 0         |
| ✅     |                  | Vue.RegisterComponent             | 100.00%  | 0         |
| ✅     |                  | Vue.RegisterNodeProcessor         | 100.00%  | 1         |
| ✅     |                  | Vue.Render                        | 100.00%  | 3         |
| ✅     |                  | Vue.RenderFragment                | 80.00%   | 4         |
| ✅     |                  | Vue.buildClassString              | 95.00%   | 7         |
| ✅     |                  | Vue.buildStyleString              | 92.31%   | 10        |
| ✅     |                  | Vue.callFunc                      | 88.73%   | 37        |
| ✅     |                  | Vue.evalAttributes                | 95.65%   | 47        |
| ✅     |                  | Vue.evalBoundAttribute            | 90.00%   | 10        |
| ✅     |                  | Vue.evalCondition                 | 100.00%  | 0         |
| ✅     |                  | Vue.evalConditionExpr             | 82.76%   | 7         |
| ✅     |                  | Vue.evalElseIfChain               | 93.02%   | 34        |
| ✅     |                  | Vue.evalFilter                    | 100.00%  | 6         |
| ✅     |                  | Vue.evalFor                       | 84.62%   | 6         |
| ✅     |                  | Vue.evalInclude                   | 92.31%   | 15        |
| ✅     |                  | Vue.evalObjectBinding             | 93.75%   | 6         |
| ✅     |                  | Vue.evalPipe                      | 86.36%   | 16        |
| ✅     |                  | Vue.evalSegment                   | 72.73%   | 5         |
| ✅     |                  | Vue.evalSlot                      | 93.75%   | 47        |
| ❌     |                  | Vue.evalTemplate                  | 72.45%   | 111       |
| ✅     |                  | Vue.evalVFor                      | 86.21%   | 29        |
| ✅     |                  | Vue.evalVHtml                     | 96.55%   | 9         |
| ✅     |                  | Vue.evalVShow                     | 66.67%   | 5         |
| ✅     |                  | Vue.evalVText                     | 96.88%   | 9         |
| ✅     |                  | Vue.evaluate                      | 91.89%   | 109       |
| ✅     |                  | Vue.evaluateChildren              | 100.00%  | 1         |
| ❌     |                  | Vue.evaluateNodeAsElement         | 66.67%   | 41        |
| ✅     |                  | Vue.interpolate                   | 100.00%  | 2         |
| ✅     |                  | Vue.interpolateToWriter           | 88.73%   | 48        |
| ✅     |                  | Vue.loadCachedWithFrontMatter     | 96.77%   | 8         |
| ✅     |                  | Vue.mergeStyles                   | 100.00%  | 2         |
| ✅     |                  | Vue.parseObjectPairs              | 93.75%   | 10        |
| ✅     |                  | Vue.postProcessNodes              | 100.00%  | 3         |
| ✅     |                  | Vue.preProcessNodes               | 83.33%   | 4         |
| ❌     |                  | Vue.processComponentNode          | 77.78%   | 9         |
| ✅     |                  | Vue.propagateTemplateAttributes   | 89.47%   | 14        |
| ✅     |                  | Vue.render                        | 75.00%   | 3         |
| ✅     |                  | Vue.renderNodesWithContext        | 100.00%  | 4         |
| ✅     |                  | Vue.replaceWithInclude            | 100.00%  | 0         |
| ✅     |                  | Vue.resolveArgument               | 93.33%   | 10        |
| ✅     |                  | Vue.resolveComponentTags          | 75.00%   | 3         |
| ✅     |                  | Vue.setStyleProperty              | 100.00%  | 1         |
| ✅     |                  | Vue.splitObjectItems              | 85.71%   | 10        |
| ✅     |                  | VueContext.Context                | 100.00%  | 0         |
| ✅     |                  | VueContext.CurrentTag             | 100.00%  | 1         |
| ✅     |                  | VueContext.ExprEnv                | 100.00%  | 0         |
| ✅     |                  | VueContext.FormatTemplateChain    | 100.00%  | 1         |
| ✅     |                  | VueContext.PopTag                 | 100.00%  | 1         |
| ✅     |                  | VueContext.PushTag                | 100.00%  | 0         |
| ✅     |                  | VueContext.Stack                  | 100.00%  | 0         |
| ✅     |                  | VueContext.WithTemplate           | 100.00%  | 0         |
| ✅     |                  | VueContext.bindContextFuncs       | 100.00%  | 7         |
| ✅     |                  | VueContext.nextSeenID             | 100.00%  | 0         |
| ✅     |                  | VueContext.wrapContextFunc        | 92.86%   | 2         |
| ✅     |                  | WithComponents                    | 100.00%  | 14        |
| ✅     |                  | WithFS                            | 100.00%  | 0         |
| ✅     |                  | WithFuncs                         | 100.00%  | 0         |
| ✅     |                  | WithLessProcessor                 | 100.00%  | 0         |
| ✅     |                  | WithProcessor                     | 100.00%  | 0         |
| ✅     |                  | assignSeenAttrs                   | 100.00%  | 5         |
| ✅     |                  | camelToKebab                      | 100.00%  | 5         |
| ❌     |                  | classifySegment                   | 80.00%   | 7         |
| ✅     |                  | containsInterpolation             | 100.00%  | 1         |
| ❌     |                  | convertValue                      | 70.00%   | 24        |
| ✅     |                  | defaultFunc                       | 100.00%  | 2         |
| ✅     |                  | escapeAttrValue                   | 100.00%  | 4         |
| ✅     |                  | escapeFunc                        | 66.67%   | 1         |
| ✅     |                  | extractFrontMatter                | 100.00%  | 5         |
| ✅     |                  | extractSlotContent                | 92.86%   | 47        |
| ✅     |                  | extractSlotsFromDOM               | 100.00%  | 25        |
| ✅     |                  | fileFunc                          | 100.00%  | 2         |
| ✅     |                  | formatDateFunc                    | 100.00%  | 1         |
| ✅     |                  | formatTimeFunc                    | 87.50%   | 4         |
| ✅     |                  | getCachedPath                     | 100.00%  | 2         |
| ✅     |                  | getIndent                         | 66.67%   | 1         |
| ✅     |                  | hasVSlot                          | 100.00%  | 8         |
| ✅     |                  | htmlRenderer.Render               | 75.00%   | 3         |
| ✅     |                  | init                              | 100.00%  | 1         |
| ✅     |                  | intFunc                           | 71.43%   | 3         |
| ✅     |                  | isLiteralAttr                     | 100.00%  | 2         |
| ✅     |                  | isNumericOrBool                   | 66.67%   | 1         |
| ✅     |                  | jsonFileFunc                      | 100.00%  | 4         |
| ✅     |                  | jsonFunc                          | 75.00%   | 1         |
| ✅     |                  | jsonPrettyFunc                    | 75.00%   | 1         |
| ✅     |                  | lenFunc                           | 100.00%  | 1         |
| ✅     |                  | loadConfig                        | 96.15%   | 15        |
| ✅     |                  | lowerFunc                         | 66.67%   | 1         |
| ✅     |                  | parseArgs                         | 100.00%  | 11        |
| ✅     |                  | parseFor                          | 89.47%   | 8         |
| ✅     |                  | parsePipeExpr                     | 100.00%  | 6         |
| ✅     |                  | parseStyleMap                     | 92.86%   | 6         |
| ✅     |                  | parseStyleString                  | 100.00%  | 6         |
| ✅     |                  | parseVSlotDirective               | 60.00%   | 4         |
| ❌     |                  | parseValue                        | 69.23%   | 7         |
| ✅     |                  | readTemplateFile                  | 77.78%   | 5         |
| ✅     |                  | renderAttrs                       | 100.00%  | 6         |
| ✅     |                  | renderNode                        | 100.00%  | 0         |
| ✅     |                  | renderNodeWithContext             | 95.61%   | 67        |
| ✅     |                  | shouldEscapeTextNode              | 100.00%  | 2         |
| ✅     |                  | shouldIgnoreAttr                  | 100.00%  | 2         |
| ✅     |                  | splitPathImpl                     | 95.83%   | 31        |
| ✅     |                  | stringFunc                        | 100.00%  | 0         |
| ✅     |                  | template.Assign                   | 100.00%  | 0         |
| ✅     |                  | template.Err                      | 100.00%  | 0         |
| ✅     |                  | template.Fill                     | 100.00%  | 5         |
| ✅     |                  | template.Get                      | 100.00%  | 4         |
| ✅     |                  | template.Load                     | 100.00%  | 1         |
| ✅     |                  | template.New                      | 100.00%  | 0         |
| ✅     |                  | template.Render                   | 100.00%  | 4         |
| ✅     |                  | template.RenderByte               | 100.00%  | 0         |
| ✅     |                  | template.RenderFile               | 100.00%  | 0         |
| ✅     |                  | template.RenderReader             | 88.24%   | 4         |
| ✅     |                  | template.RenderString             | 100.00%  | 0         |
| ✅     |                  | template.SetErr                   | 0.00%    | 0         |
| ✅     |                  | template.layout                   | 90.32%   | 19        |
| ✅     |                  | template.new                      | 100.00%  | 0         |
| ✅     |                  | template.renderWithoutLayout      | 66.67%   | 1         |
| ✅     |                  | template.resolveLayoutPath        | 100.00%  | 4         |
| ✅     |                  | titleFunc                         | 85.71%   | 6         |
| ✅     |                  | toMapData                         | 85.71%   | 3         |
| ✅     |                  | trimFunc                          | 66.67%   | 1         |
| ✅     |                  | typeFunc                          | 100.00%  | 0         |
| ✅     |                  | upperFunc                         | 100.00%  | 1         |
| ✅     |                  | yamlFileFunc                      | 100.00%  | 4         |
| ✅     | diff             | CompareHTML                       | 82.35%   | 6         |
| ✅     |                  | CountChildren                     | 100.00%  | 2         |
| ✅     |                  | DomToYAML                         | 81.82%   | 2         |
| ✅     |                  | FormatToNormalizedHTML            | 83.33%   | 1         |
| ✅     |                  | GenerateUnifiedDiff               | 100.00%  | 15        |
| ✅     |                  | SignificantChildren               | 95.83%   | 30        |
| ✅     |                  | attrsEqual                        | 100.00%  | 15        |
| ✅     |                  | classEqual                        | 100.00%  | 4         |
| ✅     |                  | compareNodeRecursive              | 87.50%   | 17        |
| ✅     |                  | filteredChildren                  | 100.00%  | 3         |
| ✅     |                  | isIgnorable                       | 80.00%   | 3         |
| ❌     |                  | nodeToMap                         | 78.12%   | 27        |
| ✅     |                  | parseClasses                      | 88.89%   | 4         |
| ✅     |                  | parseStyleProperties              | 86.67%   | 6         |
| ✅     |                  | renderNodeForDiff                 | 97.50%   | 15        |
| ✅     |                  | styleEqual                        | 100.00%  | 5         |
| ✅     | formatter        | DefaultFormatterOptions           | 100.00%  | 0         |
| ✅     |                  | FormatString                      | 100.00%  | 0         |
| ✅     |                  | Formatter.Format                  | 100.00%  | 2         |
| ❌     |                  | Formatter.allChildrenAreInline    | 80.00%   | 12        |
| ✅     |                  | Formatter.collectChildren         | 100.00%  | 4         |
| ✅     |                  | Formatter.formatFragment          | 96.97%   | 3         |
| ✅     |                  | Formatter.formatFullDocument      | 98.04%   | 7         |
| ✅     |                  | Formatter.formatNode              | 96.23%   | 18        |
| ✅     |                  | Formatter.formatNodeChildren      | 100.00%  | 1         |
| ✅     |                  | Formatter.formatRawTextElement    | 100.00%  | 5         |
| ✅     |                  | Formatter.isIgnorableWhitespace   | 66.67%   | 1         |
| ✅     |                  | Formatter.renderCloseTag          | 100.00%  | 0         |
| ✅     |                  | Formatter.renderInlineChildren    | 100.00%  | 6         |
| ✅     |                  | Formatter.renderOpenTag           | 100.00%  | 3         |
| ✅     |                  | Formatter.renderPreContent        | 100.00%  | 6         |
| ✅     |                  | Formatter.shouldKeepInline        | 100.00%  | 6         |
| ✅     |                  | Formatter.splitFrontmatter        | 100.00%  | 5         |
| ✅     |                  | IndentString                      | 100.00%  | 3         |
| ✅     |                  | NewFormatter                      | 100.00%  | 0         |
| ✅     |                  | NewFormatterWithOptions           | 100.00%  | 0         |
| ✅     |                  | escapeText                        | 100.00%  | 9         |
| ✅     |                  | fragmentContext                   | 100.00%  | 7         |
| ✅     |                  | isInlineAtom                      | 100.00%  | 3         |
| ✅     |                  | isPhrasingContainer               | 100.00%  | 3         |
| ✅     |                  | isVoidElement                     | 100.00%  | 3         |
| ✅     |                  | normalizeInlineText               | 94.12%   | 7         |
| ✅     |                  | trimRawContent                    | 100.00%  | 5         |
| ✅     | internal/helpers | AppendAttr                        | 100.00%  | 0         |
| ✅     |                  | CamelToKebab                      | 100.00%  | 9         |
| ✅     |                  | CloneNode                         | 100.00%  | 0         |
| ✅     |                  | ContainsPipe                      | 100.00%  | 3         |
| ✅     |                  | DeepCloneNode                     | 100.00%  | 6         |
| ✅     |                  | FilterAttrs                       | 100.00%  | 3         |
| ✅     |                  | FormatAttr                        | 100.00%  | 0         |
| ✅     |                  | GetAttr                           | 100.00%  | 3         |
| ✅     |                  | GetBodyNode                       | 100.00%  | 11        |
| ✅     |                  | HasAttr                           | 100.00%  | 3         |
| ✅     |                  | IsComplexExpr                     | 100.00%  | 8         |
| ✅     |                  | IsFunctionCall                    | 100.00%  | 5         |
| ✅     |                  | IsIdentifier                      | 100.00%  | 4         |
| ✅     |                  | IsIdentifierChar                  | 100.00%  | 8         |
| ✅     |                  | IsTruthy                          | 100.00%  | 4         |
| ✅     |                  | NeedsHTMLEscape                   | 100.00%  | 3         |
| ✅     |                  | NewNode                           | 100.00%  | 0         |
| ✅     |                  | NormalizeComparisonOperators      | 100.00%  | 7         |
| ✅     |                  | RemoveAttr                        | 100.00%  | 3         |
| ✅     |                  | SetAttr                           | 100.00%  | 3         |
| ✅     |                  | ShallowCloneWithAttrs             | 100.00%  | 1         |
| ✅     | internal/parser  | ParseTemplateBytes                | 84.62%   | 6         |
| ✅     | internal/reflect | CanDescend                        | 100.00%  | 5         |
| ✅     |                  | IsSlice                           | 100.00%  | 2         |
| ✅     |                  | ResolveValue                      | 100.00%  | 2         |
| ✅     |                  | SliceToAny                        | 100.00%  | 4         |
| ✅     |                  | StructToMap                       | 100.00%  | 18        |
| ✅     |                  | resolveMap                        | 100.00%  | 1         |
| ✅     |                  | resolveSliceIndex                 | 100.00%  | 2         |
| ✅     |                  | resolveStruct                     | 100.00%  | 6         |
| ✅     |                  | resolveValueRecursive             | 100.00%  | 4         |
| ✅     | internal/ulid    | Parse                             | 100.00%  | 0         |
| ✅     |                  | String                            | 100.00%  | 0         |
| ✅     |                  | ULID                              | 100.00%  | 0         |
| ✅     |                  | Valid                             | 100.00%  | 0         |
| ✅     | markdown         | Document.FrontMatter              | 100.00%  | 0         |
| ✅     |                  | Document.Render                   | 100.00%  | 0         |
| ✅     |                  | Document.RenderContext            | 100.00%  | 0         |
| ✅     |                  | Markdown.Load                     | 100.00%  | 2         |
| ✅     |                  | Markdown.PostProcess              | 100.00%  | 0         |
| ✅     |                  | Markdown.Register                 | 100.00%  | 0         |
| ✅     |                  | Markdown.RenderBytes              | 100.00%  | 0         |
| ✅     |                  | Markdown.RenderBytesContext       | 100.00%  | 0         |
| ✅     |                  | Markdown.collectCells             | 100.00%  | 3         |
| ✅     |                  | Markdown.inlineContent            | 100.00%  | 0         |
| ✅     |                  | Markdown.renderBlockquote         | 75.00%   | 1         |
| ✅     |                  | Markdown.renderChildren           | 100.00%  | 3         |
| ✅     |                  | Markdown.renderCodeBlock          | 71.43%   | 1         |
| ✅     |                  | Markdown.renderFencedCodeBlock    | 100.00%  | 1         |
| ✅     |                  | Markdown.renderFootnote           | 75.00%   | 1         |
| ✅     |                  | Markdown.renderFootnoteList       | 75.00%   | 1         |
| ✅     |                  | Markdown.renderHTMLBlock          | 80.00%   | 3         |
| ✅     |                  | Markdown.renderHeading            | 100.00%  | 1         |
| ✅     |                  | Markdown.renderInlineChildren     | 75.00%   | 3         |
| ✅     |                  | Markdown.renderInlineNode         | 89.19%   | 10        |
| ✅     |                  | Markdown.renderList               | 75.00%   | 1         |
| ✅     |                  | Markdown.renderListItem           | 75.00%   | 1         |
| ✅     |                  | Markdown.renderNode               | 82.35%   | 3         |
| ✅     |                  | Markdown.renderParagraph          | 100.00%  | 1         |
| ✅     |                  | Markdown.renderTable              | 100.00%  | 3         |
| ✅     |                  | Markdown.renderTemplate           | 88.89%   | 2         |
| ✅     |                  | Markdown.tryHandlers              | 100.00%  | 3         |
| ✅     |                  | New                               | 100.00%  | 2         |
| ✅     |                  | Templates                         | 100.00%  | 0         |
| ✅     |                  | alignString                       | 100.00%  | 1         |
| ✅     |                  | codeBlockContent                  | 100.00%  | 1         |
| ✅     |                  | codeSpanContent                   | 100.00%  | 3         |
| ✅     |                  | headingID                         | 100.00%  | 0         |
| ✅     |                  | inlineText                        | 100.00%  | 5         |
| ✅     |                  | splitFrontMatter                  | 100.00%  | 3         |
| ✅     | tests/assert     | Assert                            | 16.67%   | 1         |
| ✅     |                  | CheckEquals                       | 20.00%   | 1         |
| ✅     |                  | Contains                          | 50.00%   | 1         |
| ✅     |                  | Empty                             | 50.00%   | 1         |
| ✅     |                  | Equal                             | 50.00%   | 1         |
| ✅     |                  | EqualHTML                         | 100.00%  | 0         |
| ✅     |                  | EqualValues                       | 50.00%   | 1         |
| ✅     |                  | Error                             | 50.00%   | 1         |
| ✅     |                  | ErrorIs                           | 50.00%   | 1         |
| ✅     |                  | Errorf                            | 100.00%  | 0         |
| ✅     |                  | Fail                              | 100.00%  | 1         |
| ✅     |                  | False                             | 50.00%   | 1         |
| ✅     |                  | Greater                           | 50.00%   | 1         |
| ✅     |                  | IsIncreasing                      | 50.00%   | 5         |
| ✅     |                  | Len                               | 50.00%   | 2         |
| ✅     |                  | Nil                               | 50.00%   | 1         |
| ✅     |                  | NoError                           | 50.00%   | 1         |
| ✅     |                  | NotContains                       | 50.00%   | 1         |
| ✅     |                  | NotEmpty                          | 50.00%   | 1         |
| ✅     |                  | NotEqual                          | 50.00%   | 1         |
| ✅     |                  | NotNil                            | 50.00%   | 1         |
| ✅     |                  | ObjectsAreEqualValues             | 100.00%  | 0         |
| ✅     |                  | Run                               | 100.00%  | 10        |
| ✅     |                  | TestSuite.SetT                    | 100.00%  | 0         |
| ✅     |                  | TestSuite.T                       | 100.00%  | 0         |
| ✅     |                  | True                              | 50.00%   | 1         |
| ✅     |                  | compare                           | 60.00%   | 2         |
| ✅     |                  | containsElement                   | 91.67%   | 9         |
| ✅     |                  | hasLength                         | 100.00%  | 1         |
| ✅     |                  | isEmpty                           | 100.00%  | 2         |
| ✅     |                  | isNil                             | 83.33%   | 2         |
| ✅     |                  | orderedCompare                    | 85.71%   | 14        |
