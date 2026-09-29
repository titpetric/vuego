# Expressions

Vuego evaluates `v-if` conditions and interpolations with [expr-lang/expr](https://github.com/expr-lang/expr), so a condition can hold comparisons, boolean operators and function calls.

## Operators

- comparison: `==`, `!=`, `<`, `>`, `<=`, `>=`
- logical: `&&`, `||`, `!`
- arithmetic: `+`, `-`, `*`, `/`, `%`
- ternary: `? :`

Vuego coalesces `===` and `!==` to `==` and `!=` before the expression reaches the evaluator, so a template in JavaScript style works unchanged.

## Values

- simple variables: `show`, `isActive`
- nested properties: `user.name`, `item.inStock`
- slice indices: `items[0]`, `users[1].email`
- built-in filters as calls: `len(items)`, `upper(text)`
- functions registered through `Funcs`: `isActive(user)`, `hasPermission(role)`

## Examples

### Conditions

```html
<!-- String comparison -->
<div v-if="status == 'active'">Active</div>
<div v-if="priority != 'low'">Important</div>

<!-- Numeric comparison -->
<div v-if="score >= 90">Grade: A</div>
<div v-if="count > 0">Has items</div>

<!-- Boolean operators -->
<div v-if="isAdmin && hasPermission">Admin Access</div>
<div v-if="isOwner || isAdmin">Can Edit</div>
<div v-if="!isDeleted">Visible</div>

<!-- Parenthesised groups -->
<div v-if="(score >= 70) && (status == 'passed')">
  Passed with good score
</div>

<!-- Nested property -->
<div v-if="user.profile.verified == true">
  Verified User
</div>

<!-- Strict comparison, coalesced before evaluation -->
<div v-if="status === 'active'">Active</div>
```

### Interpolations

```html
<!-- Boolean result -->
<p>Active: {{ status == 'active' }}</p>

<!-- Comparison -->
<p>In Stock: {{ item.quantity > 0 }}</p>

<!-- Ternary -->
<p>{{ ok ? 'yes' : 'no' }}</p>

<!-- Registered function -->
<p>Can Edit: {{ canEditItem(item) }}</p>
```

Simple references, filter chains and negation all read the same way:

```html
{{ message }}
{{ user.name }}
{{ items | len }}
{{ date | formatTime("2006-01-02") }}
{{ len(items) }}
```

## How Vuego evaluates a condition

1. Vuego coalesces `===` and `!==` to `==` and `!=`.
2. expr compiles and runs the expression. Vuego reads the result of any type for truthiness.
3. When compilation fails and the expression starts with `!`, Vuego evaluates the inner expression and negates the result. That makes `!item.primary` work for an item carrying no `primary` key, which expr cannot negate.
4. When compilation still fails, Vuego resolves the whole text as a variable name against the stack. A name that fails to resolve counts as false, so a template with an unparsable condition renders without the element rather than failing.

The `ExprEvaluator` caches each compiled program under its expression text, so the same condition over different data reuses the program. The cache grows with the count of distinct expressions in a template set; `ExprEvaluator.ClearCache` empties it.

## Names expr already uses

A variable named after an expr built-in shadows nothing: the built-in wins and nothing reaches the variable. In practice this hits `len`, so `v-if="len > 0"` never matches. Name the variable `total` or `length` instead.

## ExprEvaluator

```go
eval := vuego.NewExprEvaluator()

result, err := eval.Eval("x > 5 && y < 10", map[string]any{
	"x": 7,
	"y": 8,
})

eval.ClearCache()
```

An undefined name evaluates to nil rather than raising an error, so a condition may name a key that never arrives.

## Limits

- A template calls a Go function when the FuncMap holds it, and reaches nothing else in the host program.
- A call cannot sit on the left side of a pipe. `jsonFile("config.json") | len` does not resolve; pipe the argument into the function instead.

## References

- [expr-lang/expr documentation](https://github.com/expr-lang/expr)
- [Expr language specification](https://github.com/expr-lang/expr/wiki/Language-Specification)
