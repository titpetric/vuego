package vuego

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/titpetric/vuego/internal/reflect"
)

// Stack provides stack-based variable lookup and convenient typed accessors.
type Stack struct {
	stack []map[string]any // bottom..top, top is last element
}

// NewStack constructs a Stack with an optional initial root map (nil allowed).
func NewStack(root map[string]any) *Stack {
	s := &Stack{}
	if root == nil {
		root = map[string]any{}
	}
	s.stack = []map[string]any{root}
	return s
}

// Push a new map as a top-most Stack.
func (s *Stack) Push(m map[string]any) {
	if m == nil {
		m = map[string]any{}
	}
	s.stack = append(s.stack, m)
}

// Pop the top-most Stack. If only root remains it still pops to empty slice safely.
func (s *Stack) Pop() {
	if len(s.stack) == 0 {
		return
	}
	s.stack = s.stack[:len(s.stack)-1]
	if len(s.stack) == 0 {
		s.stack = append(s.stack, map[string]any{})
	}
}

// Set sets a key in the top-most Stack.
func (s *Stack) Set(key string, val any) {
	if len(s.stack) == 0 {
		s.stack = append(s.stack, map[string]any{})
	}
	s.stack[len(s.stack)-1][key] = val
}

// Lookup searches stack from top to bottom for a plain identifier (no dots).
// Returns (value, true) if found.
func (s *Stack) Lookup(name string) (any, bool) {
	for i := len(s.stack) - 1; i >= 0; i-- {
		if v, ok := s.stack[i][name]; ok {
			return v, true
		}
	}
	return nil, false
}

// Resolve resolves dotted/bracketed expression paths like:
//
//	"user.name", "items[0].title", "mapKey.sub"
//
// It returns (value, true) if resolution succeeded.
func (s *Stack) Resolve(expr string) (any, bool) {
	parts := s.splitPath(expr)
	if len(parts) == 0 {
		return nil, false
	}

	// first part must come from Stack
	cur, ok := s.Lookup(parts[0])
	if !ok {
		return nil, false
	}
	// walk the rest
	for _, p := range parts[1:] {
		if cur == nil {
			return nil, false
		}
		switch c := cur.(type) {
		case map[string]any:
			v, ok := c[p]
			if !ok {
				// also allow numeric-key lookup if p is numeric and map has stringified numeric keys
				v2, ok2 := c[p]
				if !ok2 {
					return nil, false
				}
				cur = v2
			} else {
				cur = v
			}
		case map[string]string:
			v, ok := c[p]
			if !ok {
				return nil, false
			}
			cur = v
		case []any:
			idx, err := strconv.Atoi(p)
			if err != nil || idx < 0 || idx >= len(c) {
				return nil, false
			}
			cur = c[idx]
		case []map[string]any:
			idx, err := strconv.Atoi(p)
			if err != nil || idx < 0 || idx >= len(c) {
				return nil, false
			}
			cur = c[idx]
		case []string:
			idx, err := strconv.Atoi(p)
			if err != nil || idx < 0 || idx >= len(c) {
				return nil, false
			}
			cur = c[idx]
		case []int:
			idx, err := strconv.Atoi(p)
			if err != nil || idx < 0 || idx >= len(c) {
				return nil, false
			}
			cur = c[idx]
		case []float64:
			idx, err := strconv.Atoi(p)
			if err != nil || idx < 0 || idx >= len(c) {
				return nil, false
			}
			cur = c[idx]
		default:
			// Try struct field resolution via reflection
			if v, ok := reflect.ResolveValue(cur, p); ok {
				cur = v
			} else {
				// unsupported type
				return nil, false
			}
		}
	}
	return cur, true
}

// GetString resolves and tries to return a string.
func (s *Stack) GetString(expr string) (string, bool) {
	v, ok := s.Resolve(expr)
	if !ok || v == nil {
		return "", false
	}
	switch t := v.(type) {
	case string:
		return t, true
	case fmt.Stringer:
		return t.String(), true
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", t), true
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", t), true
	case float32, float64:
		return fmt.Sprintf("%v", t), true
	case bool:
		return fmt.Sprintf("%t", t), true
	default:
		return fmt.Sprintf("%v", t), true
	}
}

// GetInt resolves and tries to return an int (best-effort).
func (s *Stack) GetInt(expr string) (int, bool) {
	v, ok := s.Resolve(expr)
	if !ok || v == nil {
		return 0, false
	}
	switch t := v.(type) {
	case int:
		return t, true
	case int8:
		return int(t), true
	case int16:
		return int(t), true
	case int32:
		return int(t), true
	case int64:
		return int(t), true
	case uint:
		return int(t), true
	case float32:
		return int(t), true
	case float64:
		return int(t), true
	case string:
		if i, err := strconv.Atoi(t); err == nil {
			return i, true
		}
	}
	return 0, false
}

// GetSlice returns a []any for supported slice kinds. Avoids reflection by only converting known types.
func (s *Stack) GetSlice(expr string) ([]any, bool) {
	v, ok := s.Resolve(expr)
	if !ok || v == nil {
		return nil, false
	}
	switch t := v.(type) {
	case []any:
		return t, true
	case []map[string]any:
		out := make([]any, len(t))
		for i := range t {
			out[i] = t[i]
		}
		return out, true
	case []string:
		out := make([]any, len(t))
		for i := range t {
			out[i] = t[i]
		}
		return out, true
	case []int:
		out := make([]any, len(t))
		for i := range t {
			out[i] = t[i]
		}
		return out, true
	case []float64:
		out := make([]any, len(t))
		for i := range t {
			out[i] = t[i]
		}
		return out, true
	default:
		return nil, false
	}
}

// GetMap returns map[string]any or converts map[string]string to map[string]any.
// Avoids reflection for other map types.
func (s *Stack) GetMap(expr string) (map[string]any, bool) {
	v, ok := s.Resolve(expr)
	if !ok || v == nil {
		return nil, false
	}
	switch t := v.(type) {
	case map[string]any:
		return t, true
	case map[string]string:
		out := make(map[string]any, len(t))
		for k, vv := range t {
			out[k] = vv
		}
		return out, true
	default:
		return nil, false
	}
}

// ForEach iterates over a collection at the given expr and calls fn(index,value).
// Supports slices/arrays and map[string]any (iteration order for maps is unspecified).
// If fn returns an error iteration is stopped and the error passed through.
func (s *Stack) ForEach(expr string, fn func(index int, value any) error) error {
	v, ok := s.Resolve(expr)
	if !ok {
		// treat missing as no-op
		return nil
	}
	switch t := v.(type) {
	case []any:
		for i := range t {
			if err := fn(i, t[i]); err != nil {
				return err
			}
		}
		return nil
	case []map[string]any:
		for i := range t {
			if err := fn(i, t[i]); err != nil {
				return err
			}
		}
		return nil
	case []string:
		for i := range t {
			if err := fn(i, t[i]); err != nil {
				return err
			}
		}
		return nil
	case []int:
		for i := range t {
			if err := fn(i, t[i]); err != nil {
				return err
			}
		}
		return nil
	case map[string]any:
		i := 0
		for _, val := range t {
			if err := fn(i, val); err != nil {
				return err
			}
			i++
		}
		return nil
	case map[string]string:
		i := 0
		for _, val := range t {
			if err := fn(i, val); err != nil {
				return err
			}
			i++
		}
		return nil
	default:
		// unsupported collection type
		return nil
	}
}

// Helpers

// splitPath turns expressions into path parts. Supports dot notation and bracket numeric/string indexes:
// examples:
//
//	"items[0].name" -> ["items","0","name"]
//	"user.name" -> ["user","name"]
//	"a['b'].c" -> ["a","b","c"]
func (s *Stack) splitPath(expr string) []string {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil
	}
	// We'll convert brackets to dot segments then split by dot
	var b strings.Builder
	i := 0
	for i < len(expr) {
		ch := expr[i]
		if ch == '[' {
			// find matching ]
			j := i + 1
			for j < len(expr) && expr[j] != ']' {
				j++
			}
			if j >= len(expr) {
				// unmatched bracket, treat literally
				b.WriteByte(ch)
				i++
				continue
			}
			inside := strings.TrimSpace(expr[i+1 : j])
			// strip quotes if present
			if len(inside) >= 2 && ((inside[0] == '\'' && inside[len(inside)-1] == '\'') || (inside[0] == '"' && inside[len(inside)-1] == '"')) {
				inside = inside[1 : len(inside)-1]
			}
			// write dot + inside (if inside is empty, skip)
			if inside != "" {
				b.WriteByte('.')
				b.WriteString(inside)
			}
			i = j + 1
			continue
		} else {
			b.WriteByte(ch)
			i++
		}
	}
	parts := strings.Split(b.String(), ".")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}
