# Expr Language Integration

Vuego now integrates [expr-lang/expr](https://github.com/expr-lang/expr) for evaluating complex expressions in `v-if` directives and interpolations. This enables support for boolean operators, comparisons, and custom functions while maintaining backward compatibility with existing syntax.

## Features

### Supported Operators in v-if and Interpolations

#### Comparison Operators
- `==` - equality
- `!=` - inequality
- `<` - less than
- `>` - greater than
- `<=` - less than or equal
- `>=` - greater than or equal

#### Logical Operators
- `&&` - logical AND
- `||` - logical OR
- `!` - logical NOT

#### Variable References
- Simple variables: `show`, `isActive`
- Nested properties: `user.name`, `item.inStock`
- Array indices: `items[0]`, `users[1].email`

#### Function Calls
- Built-in filters: `len(items)`, `upper(text)`, etc.
- Custom functions: `isActive(user)`, `hasPermission(role)`

## Examples

### Boolean Comparisons

```html
<!-- String comparison -->
<div v-if="status == 'active'">Active</div>
<div v-if="priority != 'low'">Important</div>

<!-- Numeric comparison -->
<div v-if="score >= 90">Grade: A</div>
<div v-if="count > 0">Has items</div>
```

### Logical Operations

```html
<!-- Logical AND -->
<div v-if="isAdmin && hasPermission">Admin Access</div>

<!-- Logical OR -->
<div v-if="isOwner || isAdmin">Can Edit</div>

<!-- Logical NOT -->
<div v-if="!isDeleted">Visible</div>
```

### Complex Expressions

```html
<!-- Combined conditions -->
<div v-if="(score >= 70) && (status == 'passed')">
  Passed with good score
</div>

<!-- Multiple OR conditions -->
<div v-if="role == 'admin' || role == 'moderator' || role == 'owner'">
  Can Manage
</div>

<!-- Nested property comparisons -->
<div v-if="user.profile.verified == true">
  Verified User
</div>
```

### Interpolations with Expressions

```html
<!-- Boolean result -->
<p>Active: {{ status == 'active' }}</p>

<!-- Comparison in interpolation -->
<p>In Stock: {{ item.quantity > 0 }}</p>

<!-- Custom function -->
<p>Can Edit: {{ canEditItem(item) }}</p>
```

## Backward Compatibility

All existing vuego syntax continues to work:

- Simple variable references: `{{ message }}`, `{{ user.name }}`
- Filter chains: `{{ items | len }}`, `{{ date | formatTime }}`
- Function calls: `{{ len(items) }}`
- Negation: `v-if="!show"`

## Implementation Details

### How It Works

1. **Expression Evaluation Attempt**: When evaluating a v-if expression or interpolation, vuego first attempts to parse and evaluate it using the expr language.

2. **Operator Detection**: If the expression contains comparison operators (`==`, `!=`, `<`, `>`, `<=`, `>=`) or logical operators (`&&`, `||`), it's sent to expr for evaluation.

3. **Fallback to Stack Resolution**: For simple variable names that expr cannot compile, vuego falls back to the original stack-based variable resolution.

4. **Caching**: Compiled expr programs are cached in the `ExprEvaluator` to avoid recompilation of the same expression.

### Variable Names to Avoid

Some variable names conflict with expr's built-in functions. Avoid naming variables:
- `len` - use `length` or `total` instead
- `count` - use `total` or `quantity` instead
- Other single-letter function names used by expr

## API

### ExprEvaluator

The `ExprEvaluator` type provides expression evaluation with caching:

```go
// Create an evaluator
eval := vuego.NewExprEvaluator()

// Evaluate an expression
result, err := eval.Eval("x > 5 && y < 10", map[string]any{
	"x": 7,
	"y": 8,
})

// Clear cache if needed
eval.ClearCache()
```

## Testing

New tests verify:
- Basic expr operations (arithmetic, comparisons, logical operators)
- v-if with boolean operators and property access
- Complex boolean expressions in v-if
- Interpolations with comparisons
- Backward compatibility with existing syntax

All existing tests continue to pass, ensuring no regression.

## Performance Considerations

- Expression compilation is cached per unique expression
- The cache grows with the number of unique expressions in templates
- For most applications, the overhead is negligible
- Manual cache clearing is available via `ExprEvaluator.ClearCache()`

## Limitations

- expr-lang/expr does not support the ternary operator (`? :`). For conditional logic in interpolations, use v-if with different divs or compute the value in the data.
- Variable names that are expr built-in functions should be avoided
- Not all Go functions are automatically available; only those explicitly added to the FuncMap are callable

## References

- [expr-lang/expr Documentation](https://github.com/expr-lang/expr)
- [Expr Language Spec](https://github.com/expr-lang/expr/wiki/Language-Specification)
