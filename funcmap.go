package vuego

import (
	"fmt"
	"html"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/titpetric/vuego/internal/helpers"
)

// FuncMap is a map of function names to functions, similar to text/template's FuncMap.
// Functions can have any number of parameters and must return 1 or 2 values.
// If 2 values are returned, the second must be an error.
type FuncMap map[string]any

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

var (
	// matches function calls like "fn(arg1, arg2)" or just "fn"
	filterRe = regexp.MustCompile(`^(\w+)(?:\((.*?)\))?$`)
)

// parsePipeExpr parses "item | double | . > 5" into segments, auto-detecting expressions vs filters
func parsePipeExpr(expr string) pipeExpr {
	if !strings.Contains(expr, "|") {
		trimmed := strings.TrimSpace(expr)
		// Check if it's a complex expression first
		if helpers.IsComplexExpr(trimmed) {
			return pipeExpr{
				initial: "",
				segments: []pipeSegment{{
					typ:  segmentExpr,
					expr: trimmed,
				}},
			}
		}
		// Check if it's a function call
		if matches := filterRe.FindStringSubmatch(trimmed); matches != nil && matches[2] != "" {
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
	result := pipeExpr{
		initial:  strings.TrimSpace(parts[0]),
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
		env := getEnvMap(ctx.stack)
		if input != nil {
			env["."] = input
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

	result, err := v.callFunc(fn, args...)
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

// callFunc calls a function from the FuncMap with the given arguments using reflection.
// This matches text/template's behavior and supports arbitrary function signatures.
func (v *Vue) callFunc(fn any, args ...any) (any, error) {
	fnVal := reflect.ValueOf(fn)
	fnType := fnVal.Type()

	if fnType.Kind() != reflect.Func {
		return nil, fmt.Errorf("not a function")
	}

	// Check argument count
	numIn := fnType.NumIn()
	if fnType.IsVariadic() {
		if len(args) < numIn-1 {
			return nil, fmt.Errorf("function expects at least %d arguments, got %d", numIn-1, len(args))
		}
	} else {
		if len(args) != numIn {
			return nil, fmt.Errorf("function expects %d arguments, got %d", numIn, len(args))
		}
	}

	// Prepare arguments with type conversion
	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		var argType reflect.Type
		if fnType.IsVariadic() && i >= numIn-1 {
			// Variadic argument - use element type
			argType = fnType.In(numIn - 1).Elem()
		} else {
			argType = fnType.In(i)
		}

		argVal := reflect.ValueOf(arg)

		// Handle nil
		if arg == nil {
			in[i] = reflect.Zero(argType)
			continue
		}

		// Try to convert the argument to the expected type
		if argVal.Type().AssignableTo(argType) {
			in[i] = argVal
		} else if argVal.Type().ConvertibleTo(argType) {
			in[i] = argVal.Convert(argType)
		} else {
			// Try to handle common conversions
			converted, ok := convertValue(argVal, argType)
			if !ok {
				return nil, fmt.Errorf("cannot convert argument %d from %v to %v", i, argVal.Type(), argType)
			}
			in[i] = converted
		}
	}

	// Call the function
	out := fnVal.Call(in)

	// Handle return values
	switch len(out) {
	case 0:
		return nil, nil
	case 1:
		// Single return value
		return out[0].Interface(), nil
	case 2:
		// Two return values - second should be error
		result := out[0].Interface()
		if out[1].IsNil() {
			return result, nil
		}
		if err, ok := out[1].Interface().(error); ok {
			return result, err
		}
		return result, nil
	default:
		return nil, fmt.Errorf("function returns too many values")
	}
}

// convertValue attempts common type conversions
func convertValue(val reflect.Value, targetType reflect.Type) (reflect.Value, bool) {
	// Handle string to basic types
	if val.Kind() == reflect.String {
		s := val.String()
		switch targetType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if i, err := strconv.ParseInt(s, 10, 64); err == nil {
				return reflect.ValueOf(i).Convert(targetType), true
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if u, err := strconv.ParseUint(s, 10, 64); err == nil {
				return reflect.ValueOf(u).Convert(targetType), true
			}
		case reflect.Float32, reflect.Float64:
			if f, err := strconv.ParseFloat(s, 64); err == nil {
				return reflect.ValueOf(f).Convert(targetType), true
			}
		case reflect.Bool:
			if b, err := strconv.ParseBool(s); err == nil {
				return reflect.ValueOf(b), true
			}
		}
	}

	// Handle int to string
	if val.Kind() >= reflect.Int && val.Kind() <= reflect.Int64 && targetType.Kind() == reflect.String {
		return reflect.ValueOf(fmt.Sprintf("%d", val.Int())), true
	}

	// Handle uint to string
	if val.Kind() >= reflect.Uint && val.Kind() <= reflect.Uint64 && targetType.Kind() == reflect.String {
		return reflect.ValueOf(fmt.Sprintf("%d", val.Uint())), true
	}

	// Handle float to string
	if (val.Kind() == reflect.Float32 || val.Kind() == reflect.Float64) && targetType.Kind() == reflect.String {
		return reflect.ValueOf(fmt.Sprintf("%v", val.Float())), true
	}

	// Handle bool to string
	if val.Kind() == reflect.Bool && targetType.Kind() == reflect.String {
		return reflect.ValueOf(fmt.Sprintf("%t", val.Bool())), true
	}

	return reflect.Value{}, false
}

// DefaultFuncMap returns a FuncMap with built-in utility functions
func DefaultFuncMap() FuncMap {
	return FuncMap{
		"upper":      upperFunc,
		"lower":      lowerFunc,
		"title":      titleFunc,
		"formatTime": formatTimeFunc,
		"default":    defaultFunc,
		"len":        lenFunc,
		"trim":       trimFunc,
		"escape":     escapeFunc,
		"int":        intFunc,
		"string":     stringFunc,
	}
}

// Built-in filter functions

func upperFunc(v any) any {
	if s, ok := v.(string); ok {
		return strings.ToUpper(s)
	}
	return v
}

func lowerFunc(v any) any {
	if s, ok := v.(string); ok {
		return strings.ToLower(s)
	}
	return v
}

func titleFunc(v any) any {
	if s, ok := v.(string); ok {
		words := strings.Fields(s)
		for i, word := range words {
			if len(word) > 0 {
				words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
			}
		}
		return strings.Join(words, " ")
	}
	return v
}

func formatTimeFunc(v any, layout any) any {
	layoutStr := "2006-01-02 15:04:05"
	if s, ok := layout.(string); ok {
		layoutStr = s
	}

	switch t := v.(type) {
	case time.Time:
		return t.Format(layoutStr)
	case string:
		// Try to parse as RFC3339 or Unix timestamp
		if parsed, err := time.Parse(time.RFC3339, t); err == nil {
			return parsed.Format(layoutStr)
		}
	}
	return v
}

func defaultFunc(v any, def any) any {
	if v == nil || v == "" {
		return def
	}
	return v
}

func lenFunc(v any) any {
	switch t := v.(type) {
	case string:
		return len(t)
	case []any:
		return len(t)
	case []string:
		return len(t)
	case []int:
		return len(t)
	case map[string]any:
		return len(t)
	}
	return 0
}

func trimFunc(v any) any {
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	return v
}

func escapeFunc(v any) any {
	if s, ok := v.(string); ok {
		return html.EscapeString(s)
	}
	return fmt.Sprint(v)
}

func intFunc(v any) any {
	switch t := v.(type) {
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	case string:
		if i, err := strconv.Atoi(t); err == nil {
			return i
		}
	}
	return 0
}

func stringFunc(v any) any {
	return fmt.Sprint(v)
}
