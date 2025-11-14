// Package reflect provides utilities for traversing struct fields and values
// during template evaluation, supporting both field names and JSON tags.
package reflect

import (
	"reflect"
	"strconv"
	"strings"
)

// ResolveValue traverses a value by field access (supporting nested structs, maps, slices).
// Field access can use either the struct field name or its JSON tag (if present).
// Returns (value, true) if resolution succeeds, (nil, false) otherwise.
func ResolveValue(v any, fieldName string) (any, bool) {
	if v == nil || fieldName == "" {
		return nil, false
	}

	rv := reflect.ValueOf(v)
	return resolveValueRecursive(rv, fieldName)
}

func resolveValueRecursive(rv reflect.Value, fieldName string) (any, bool) {
	// Dereference pointers
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, false
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Struct:
		return resolveStruct(rv, fieldName)
	case reflect.Map:
		return resolveMap(rv, fieldName)
	case reflect.Slice, reflect.Array:
		return resolveSliceIndex(rv, fieldName)
	default:
		return nil, false
	}
}

// resolveStruct looks up a field by name or JSON tag.
func resolveStruct(rv reflect.Value, fieldName string) (any, bool) {
	rt := rv.Type()

	// Try field name first
	if f, ok := rt.FieldByName(fieldName); ok {
		fv := rv.FieldByIndex(f.Index)
		return fv.Interface(), true
	}

	// Try JSON tag
	for i := range rt.NumField() {
		f := rt.Field(i)
		tag := f.Tag.Get("json")
		if tag == "" {
			continue
		}

		// Parse the JSON tag, stripping options (e.g., "user_id,omitempty" -> "user_id")
		tagName := strings.Split(tag, ",")[0]
		if tagName == fieldName {
			fv := rv.FieldByIndex(f.Index)
			return fv.Interface(), true
		}
	}

	return nil, false
}

// resolveMap handles map access by string key.
func resolveMap(rv reflect.Value, key string) (any, bool) {
	mapKey := reflect.ValueOf(key)
	v := rv.MapIndex(mapKey)
	if !v.IsValid() {
		return nil, false
	}
	return v.Interface(), true
}

// resolveSliceIndex handles slice/array access by numeric index.
func resolveSliceIndex(rv reflect.Value, indexStr string) (any, bool) {
	idx, err := strconv.Atoi(indexStr)
	if err != nil || idx < 0 || idx >= rv.Len() {
		return nil, false
	}
	return rv.Index(idx).Interface(), true
}

// CanDescend returns true if v is a type that can have fields/elements accessed.
func CanDescend(v any) bool {
	if v == nil {
		return false
	}

	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return false
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Struct, reflect.Map, reflect.Slice, reflect.Array:
		return true
	default:
		return false
	}
}
