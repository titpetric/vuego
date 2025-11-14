package vuego

import (
	"fmt"
	"sync"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

// ExprEvaluator wraps expr for evaluating boolean and interpolated expressions.
// It caches compiled programs to avoid recompilation.
type ExprEvaluator struct {
	mu       sync.RWMutex
	programs map[string]*vm.Program
}

// NewExprEvaluator creates a new ExprEvaluator with an empty cache.
func NewExprEvaluator() *ExprEvaluator {
	return &ExprEvaluator{
		programs: make(map[string]*vm.Program),
	}
}

// Eval evaluates an expression against the given environment (stack).
// It returns the result value and any error.
// The expression can contain:
//   - Variable references: item, item.title, items[0]
//   - Comparison: ==, !=, <, >, <=, >=
//   - Boolean operations: &&, ||, !
//   - Function calls: len(items), isActive(v)
//   - Literals: 42, "text", true, false
func (e *ExprEvaluator) Eval(expression string, env map[string]any) (any, error) {
	// Get or compile the program
	prog, err := e.getProgram(expression)
	if err != nil {
		return nil, err
	}

	// Run the compiled program
	result, err := expr.Run(prog, env)
	if err != nil {
		return nil, fmt.Errorf("eval error: %w", err)
	}
	return result, nil
}

// getProgram returns a cached compiled program or compiles a new one.
func (e *ExprEvaluator) getProgram(expression string) (*vm.Program, error) {
	e.mu.RLock()
	if prog, ok := e.programs[expression]; ok {
		e.mu.RUnlock()
		return prog, nil
	}
	e.mu.RUnlock()

	// Compile the expression
	prog, err := expr.Compile(expression, expr.AllowUndefinedVariables())
	if err != nil {
		return nil, fmt.Errorf("compile error: %w", err)
	}

	// Cache it
	e.mu.Lock()
	e.programs[expression] = prog
	e.mu.Unlock()

	return prog, nil
}

// ClearCache clears the program cache (useful for testing or memory management).
func (e *ExprEvaluator) ClearCache() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.programs = make(map[string]*vm.Program)
}
