package vuego

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
)

// callFunc calls a function from the FuncMap with optional VueContext as first argument.
// If the function's first parameter is *VueContext, the context is passed automatically.
// Otherwise, all provided arguments are passed directly.
func (v *Vue) callFunc(ctx *VueContext, fn any, args ...any) (any, error) {
	fnVal := reflect.ValueOf(fn)
	fnType := fnVal.Type()

	if fnType.Kind() != reflect.Func {
		return nil, fmt.Errorf("not a function")
	}

	// Check if first parameter is *VueContext or context.Context
	hasContextParam := false
	hasStdContext := false
	numIn := fnType.NumIn()
	if numIn > 0 {
		firstParamType := fnType.In(0)
		if firstParamType == reflect.TypeOf((*VueContext)(nil)) {
			hasContextParam = true
		} else if firstParamType == reflect.TypeOf((*context.Context)(nil)).Elem() {
			hasStdContext = true
		}
	}

	// Adjust expected argument count based on context parameter
	expectedArgs := numIn
	if hasContextParam || hasStdContext {
		expectedArgs = numIn - 1 // Don't count the context parameter
	}

	// For variadic functions, adjust to required args (excluding the variadic ...T part)
	requiredArgs := expectedArgs
	if fnType.IsVariadic() {
		requiredArgs = expectedArgs - 1
	}

	// Build final argument list
	var finalArgs []any
	if hasContextParam {
		finalArgs = append(finalArgs, ctx)
		finalArgs = append(finalArgs, args...)
	} else if hasStdContext {
		finalArgs = append(finalArgs, ctx.Context())
		finalArgs = append(finalArgs, args...)
	} else {
		finalArgs = args
	}

	// Check argument count
	if fnType.IsVariadic() {
		if len(args) < requiredArgs {
			return nil, fmt.Errorf("function expects at least %d arguments, got %d", requiredArgs, len(args))
		}
	} else {
		if len(args) != expectedArgs {
			return nil, fmt.Errorf("function expects %d arguments, got %d", expectedArgs, len(args))
		}
	}

	// Prepare arguments with type conversion
	in := make([]reflect.Value, len(finalArgs))
	for i, arg := range finalArgs {
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
		switch {
		case argVal.Type().AssignableTo(argType):
			in[i] = argVal
		case argType.Kind() == reflect.String && isNumericOrBool(argVal.Kind()):
			// reflect.Value.Convert turns an integer into a one-rune string,
			// so 42 would arrive as "*". A template means the decimal form,
			// which convertValue produces.
			converted, ok := convertValue(argVal, argType)
			if !ok {
				return nil, fmt.Errorf("cannot convert argument %d from %v to %v", i, argVal.Type(), argType)
			}
			in[i] = converted
		case argVal.Type().ConvertibleTo(argType):
			in[i] = argVal.Convert(argType)
		default:
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

// isNumericOrBool reports whether k is one of the kinds convertValue formats
// into a string.
func isNumericOrBool(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Bool:
		return true
	}
	return false
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
