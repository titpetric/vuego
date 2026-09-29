package vuego

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/titpetric/vuego/internal/helpers"
)

// pipeExpr represents a parsed pipe expression like "value | fn1 | . > 5 | fn2(arg)"
type pipeExpr struct {
	initial  string
	segments []pipeSegment
}

// pipeSegment represents either a filter call or an expression in a pipe chain
type pipeSegment struct {
	typ  segmentType // "filter" or "expr"
	expr string      // The expression/filter text
	name string      // Filter name (only for filters)
	args []string    // Filter arguments (only for filters)
}

type segmentType string

const (
	segmentFilter = "filter"
	segmentExpr   = "expr"
)

// matches function calls like "fn(arg1, arg2)" or just "fn"
var filterRe = regexp.MustCompile(`^(\w+)(?:\((.*?)\))?$`)

// parsePipeExpr parses "item | double | . > 5" into segments, auto-detecting expressions vs filters
func parsePipeExpr(expr string) pipeExpr {
	// Check if this is a complex expression (contains operators like ||, &&, etc.)
	// before trying to split on pipe character
	trimmed := strings.TrimSpace(expr)
	if helpers.IsComplexExpr(trimmed) {
		return pipeExpr{
			initial: "",
			segments: []pipeSegment{{
				typ:  segmentExpr,
				expr: trimmed,
			}},
		}
	}

	if !strings.Contains(expr, "|") {
		// Check if it's a function call (including no-arg functions like "fn()")
		if matches := filterRe.FindStringSubmatch(trimmed); matches != nil && matches[1] != "" {
			return pipeExpr{
				initial: "",
				segments: []pipeSegment{{
					typ:  segmentFilter,
					expr: trimmed,
					name: matches[1],
					args: parseArgs(matches[2]),
				}},
			}
		}
		// Just a simple variable reference
		return pipeExpr{initial: trimmed}
	}

	parts := strings.Split(expr, "|")
	firstPart := strings.TrimSpace(parts[0])

	result := pipeExpr{
		initial:  firstPart,
		segments: make([]pipeSegment, 0, len(parts)-1),
	}

	for i := 1; i < len(parts); i++ {
		part := strings.TrimSpace(parts[i])
		result.segments = append(result.segments, classifySegment(part))
	}

	return result
}

// classifySegment determines if a pipe segment is a filter call or expression
func classifySegment(part string) pipeSegment {
	// Check for complex expression operators first
	if helpers.IsComplexExpr(part) {
		return pipeSegment{
			typ:  segmentExpr,
			expr: part,
		}
	}

	// Try to match as function call
	if matches := filterRe.FindStringSubmatch(part); matches != nil {
		name := matches[1]
		if helpers.IsIdentifier(name) {
			args := []string{}
			if matches[2] != "" {
				args = parseArgs(matches[2])
			}
			return pipeSegment{
				typ:  segmentFilter,
				expr: part,
				name: name,
				args: args,
			}
		}
	}

	// Otherwise treat as expression
	return pipeSegment{
		typ:  segmentExpr,
		expr: part,
	}
}

// parseArgs parses comma-separated arguments, handling quoted strings
func parseArgs(argStr string) []string {
	var args []string
	var current strings.Builder
	inQuote := false
	quoteChar := rune(0)

	for _, ch := range strings.TrimSpace(argStr) {
		switch {
		case (ch == '"' || ch == '\'') && !inQuote:
			inQuote = true
			quoteChar = ch
		case ch == quoteChar && inQuote:
			inQuote = false
			quoteChar = 0
		case ch == ',' && !inQuote:
			if current.Len() > 0 {
				args = append(args, strings.TrimSpace(current.String()))
				current.Reset()
			}
		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		args = append(args, strings.TrimSpace(current.String()))
	}

	return args
}

// evalPipe evaluates a pipe expression, handling both filters and expressions
func (v *Vue) evalPipe(ctx VueContext, expr pipeExpr) (any, error) {
	// Handle direct function calls (no initial value)
	if expr.initial == "" && len(expr.segments) > 0 {
		val, err := v.evalSegment(ctx, expr.segments[0], nil, true, false)
		if err != nil {
			return nil, err
		}
		// Apply remaining segments
		for i := 1; i < len(expr.segments); i++ {
			val, err = v.evalSegment(ctx, expr.segments[i], val, false, true)
			if err != nil {
				return nil, err
			}
		}
		return val, nil
	}

	// Resolve initial value
	var val any
	var ok bool
	val, ok = ctx.stack.Resolve(expr.initial)
	if !ok {
		if len(expr.segments) > 0 {
			val = nil // Pass nil to first segment filter
		} else {
			return nil, fmt.Errorf("variable '%s' not found", expr.initial)
		}
	}

	// Apply each segment in sequence
	for i, seg := range expr.segments {
		var err error
		val, err = v.evalSegment(ctx, seg, val, i == 0, true)
		if err != nil {
			return nil, err
		}
	}

	return val, nil
}

// evalSegment evaluates a single pipe segment (either filter or expression)
// isFirst indicates if this is the first segment
// fromInitial indicates if the input came from initial variable resolution
func (v *Vue) evalSegment(ctx VueContext, seg pipeSegment, input any, isFirst, fromInitial bool) (any, error) {
	switch seg.typ {
	case segmentFilter:
		return v.evalFilter(ctx, seg, input, isFirst, fromInitial)
	case segmentExpr:
		// Use expr library with . representing the input value
		env := ctx.ExprEnv()
		if input != nil {
			env["."] = input
			defer delete(env, ".")
		}
		result, err := v.exprEval.Eval(seg.expr, env)
		if err != nil {
			return nil, fmt.Errorf("in expression '%s': %w", seg.expr, err)
		}
		return result, nil
	}
	return nil, fmt.Errorf("unknown segment type: %v", seg.typ)
}

// evalFilter applies a filter function to input
// isFirst indicates if this is the first segment
// fromInitial indicates if input came from initial variable resolution
func (v *Vue) evalFilter(ctx VueContext, seg pipeSegment, input any, isFirst, fromInitial bool) (any, error) {
	fn, exists := v.funcMap[seg.name]
	if !exists {
		return nil, fmt.Errorf("function '%s' not found", seg.name)
	}

	// Prepend input if:
	// - Not the first segment (pipe continuation), OR
	// - First segment that came from initial variable resolution (passes input even if nil)
	args := []any{}
	if !isFirst || (isFirst && fromInitial) {
		args = append(args, input)
	}
	for _, argExpr := range seg.args {
		argVal := v.resolveArgument(ctx, argExpr)
		args = append(args, argVal)
	}

	result, err := v.callFunc(&ctx, fn, args...)
	if err != nil {
		return nil, fmt.Errorf("%s(): %w", seg.name, err)
	}
	return result, nil
}

// resolveArgument resolves a single argument (either a variable reference or literal)
func (v *Vue) resolveArgument(ctx VueContext, arg string) any {
	arg = strings.TrimSpace(arg)

	// Check if it's a quoted string literal
	if len(arg) >= 2 {
		if (arg[0] == '"' && arg[len(arg)-1] == '"') ||
			(arg[0] == '\'' && arg[len(arg)-1] == '\'') {
			return arg[1 : len(arg)-1]
		}
	}

	// Try to parse as integer
	if i, err := strconv.Atoi(arg); err == nil {
		return i
	}

	// Try to parse as float
	if f, err := strconv.ParseFloat(arg, 64); err == nil {
		return f
	}

	// Try to parse as bool
	if b, err := strconv.ParseBool(arg); err == nil {
		return b
	}

	// Try to resolve as variable
	if val, ok := ctx.stack.Resolve(arg); ok {
		return val
	}

	// Return as-is (literal string)
	return arg
}
