package vuego

import (
	"strings"
	"sync"
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
