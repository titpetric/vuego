package vuego

import (
	"encoding/json"
	"fmt"
	stdhtml "html"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// FuncMap maps template function names to functions, as text/template does.
// A function takes any number of parameters and returns one value, or a value
// and an error. An optional first parameter of *VueContext or context.Context
// receives the execution context. See docs/funcmap.md.
type FuncMap map[string]any

func typeFunc(v any) any {
	return fmt.Sprintf("%T", v)
}

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

func formatDateFunc(v any, layout any) any {
	layoutStr := "2006-01-02"
	if s, ok := layout.(string); ok {
		layoutStr = s
	}

	return formatTimeFunc(v, layoutStr)
}

func defaultFunc(v any, def any) any {
	if v == nil || v == "" {
		return def
	}
	return v
}

func lenFunc(v any) any {
	rv := reflect.ValueOf(v)

	switch rv.Kind() {
	case reflect.String, reflect.Slice, reflect.Array, reflect.Map:
		return rv.Len()
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
		return stdhtml.EscapeString(s)
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

func jsonFunc(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}
	return string(b), nil
}

func jsonPrettyFunc(v any) (string, error) {
	b, err := json.MarshalIndent(v, "  ", "")
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}
	return string(b), nil
}
