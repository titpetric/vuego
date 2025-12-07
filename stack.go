package vuego

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	ireflect "github.com/titpetric/vuego/internal/reflect"
)

// pathCache caches parsed paths to avoid re-parsing frequently-used expressions.
var pathCache = &struct {
	sync.RWMutex
	m map[string][]string
}{
	m: make(map[string][]string),
}

const pathCacheLimit = 256

// getCachedPath returns a cached path or computes and caches it.
func getCachedPath(expr string) []string {
	pathCache.RLock()
	if parts, ok := pathCache.m[expr]; ok {
		pathCache.RUnlock()
		return parts
	}
	pathCache.RUnlock()

	// Not in cache, compute it
	parts := splitPathImpl(expr)

	// Cache if under limit
	if len(pathCache.m) < pathCacheLimit {
		pathCache.Lock()
		pathCache.m[expr] = parts
		pathCache.Unlock()
	}

	return parts
}

// Stack provides stack-based variable lookup and convenient typed accessors.
type Stack struct {
	stack    []map[string]any // bottom..top, top is last element
	rootData any              // original data passed to Render (for struct field fallback)
}

// NewStack constructs a Stack with an optional initial root map (nil allowed).
// The originalData parameter is the original value passed to Render (for struct field fallback).
func NewStack(root map[string]any) *Stack {
	return NewStackWithData(root, nil)
}

// NewStackWithData constructs a Stack with both map data and original root data for struct field fallback.
func NewStackWithData(root map[string]any, originalData any) *Stack {
	s := &Stack{}
	if root == nil {
		root = map[string]any{}
	}
	s.stack = []map[string]any{root}
	s.rootData = originalData
	return s
}

// mapPool caches map[string]any allocations to reduce GC pressure.
var mapPool = sync.Pool{
	New: func() any {
		return make(map[string]any, 0)
	},
}

// Copy returns a copy of the stack that can be discarded.
// The root data is retained as is, the envmap is a copy.
func (s *Stack) Copy() *Stack {
	return NewStackWithData(s.EnvMap(), s.rootData)
}

// Push a new map as a top-most Stack.
// If m is nil, an empty map is obtained from the pool.
func (s *Stack) Push(m map[string]any) {
	if m == nil {
		m = mapPool.Get().(map[string]any)
	}
	s.stack = append(s.stack, m)
}

// Pop the top-most Stack. If only root remains it still pops to empty slice safely.
// Returns pooled maps to reduce GC pressure.
func (s *Stack) Pop() {
	if len(s.stack) == 0 {
		return
	}
	// Return the top map to the pool before removing it
	topIdx := len(s.stack) - 1
	topMap := s.stack[topIdx]
	// Clear the map and return it to pool if it's not the root
	if topIdx > 0 && len(topMap) > 0 {
		for k := range topMap {
			delete(topMap, k)
		}
		mapPool.Put(topMap)
	}
	s.stack = s.stack[:topIdx]
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
// If not found in the stack maps, it checks the root data struct (if any).
// Returns (value, true) if found.
func (s *Stack) Lookup(name string) (any, bool) {
	for i := len(s.stack) - 1; i >= 0; i-- {
		if v, ok := s.stack[i][name]; ok {
			return v, true
		}
	}
	// Fallback: check root data struct fields
	if s.rootData != nil {
		if v, ok := ireflect.ResolveValue(s.rootData, name); ok {
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
	// Fast path: if no dots or brackets, do direct lookup
	if !strings.ContainsAny(expr, ".[") {
		return s.Lookup(expr)
	}

	// Parse once (with caching)
	parts := getCachedPath(expr)
	if len(parts) == 0 {
		return nil, false
	}

	// first part must come from Stack
	cur, ok := s.Lookup(parts[0])
	if cur == nil || !ok {
		return nil, false
	}
	// walk the rest
	for _, p := range parts[1:] {
		cur = s.resolveStep(cur, p)
		if cur == nil {
			return nil, false
		}
	}
	return cur, true
}

// resolveStep resolves a single step in a path, returning nil if resolution fails.
func (s *Stack) resolveStep(cur any, p string) any {
	// Try maps first
	switch c := cur.(type) {
	case map[string]any:
		return c[p]
	case map[string]string:
		return c[p]
	}

	// Try numeric index for slices and arrays
	idx, err := strconv.Atoi(p)
	if err == nil && idx >= 0 {
		v := reflect.ValueOf(cur)
		if (v.Kind() == reflect.Slice || v.Kind() == reflect.Array) && idx < v.Len() {
			return v.Index(idx).Interface()
		}
	}

	// Fall back to struct field resolution
	if v, ok := ireflect.ResolveValue(cur, p); ok {
		return v
	}
	return nil
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

// GetSlice returns a []any for slice types.
func (s *Stack) GetSlice(expr string) ([]any, bool) {
	v, ok := s.Resolve(expr)
	if !ok || v == nil {
		return nil, false
	}
	if ireflect.IsSlice(v) {
		return ireflect.SliceToAny(v), true
	}
	return nil, false
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

// EnvMap converts the Stack to a map[string]any for expr evaluation.
// Includes all accessible values from stack and struct fields.
func (s *Stack) EnvMap() map[string]any {
	result := make(map[string]any)
	// Iterate through stack from bottom to top, with top overriding bottom
	for i := 0; i < len(s.stack); i++ {
		for k, v := range s.stack[i] {
			result[k] = v
		}
	}

	// Also include struct fields from rootData (if available)
	if s.rootData != nil {
		ireflect.PopulateStructFields(result, s.rootData)
	}
	return result
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

	rv := reflect.ValueOf(v)

	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		// []V
		for i := 0; i < rv.Len(); i++ {
			if err := fn(i, rv.Index(i).Interface()); err != nil {
				return err
			}
		}
		return nil
	case reflect.Map:
		keys := rv.MapKeys()
		for i, key := range keys {
			if err := fn(i, rv.MapIndex(key).Interface()); err != nil {
				return err
			}
		}
		return nil
	}

	return nil
	// return fmt.Errorf("unsupported collection type: %T, expr: %s", v, expr)
}

// Helpers

// splitPathImpl is the actual implementation of path splitting.
// Called by getCachedPath which caches the results.
func splitPathImpl(expr string) []string {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil
	}

	// Fast path: if no brackets, just split by dots
	if !strings.Contains(expr, "[") {
		parts := strings.Split(expr, ".")
		// Sanitize in-place to avoid extra allocation
		out := parts[:0]
		for _, p := range parts {
			if p = strings.TrimSpace(p); p != "" {
				out = append(out, p)
			}
		}
		return out
	}

	// Full parsing with bracket support
	var b strings.Builder
	b.Grow(len(expr) + 8)
	i := 0
	for i < len(expr) {
		ch := expr[i]
		if ch == '[' {
			j := i + 1
			for j < len(expr) && expr[j] != ']' {
				j++
			}
			if j >= len(expr) {
				b.WriteByte(ch)
				i++
				continue
			}
			inside := strings.TrimSpace(expr[i+1 : j])
			if len(inside) >= 2 && ((inside[0] == '\'' && inside[len(inside)-1] == '\'') || (inside[0] == '"' && inside[len(inside)-1] == '"')) {
				inside = inside[1 : len(inside)-1]
			}
			if inside != "" {
				b.WriteByte('.')
				b.WriteString(inside)
			}
			i = j + 1
		} else {
			b.WriteByte(ch)
			i++
		}
	}

	builtStr := b.String()
	parts := strings.Split(builtStr, ".")
	// Sanitize in-place to avoid extra allocation
	out := parts[:0]
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}
