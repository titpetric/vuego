package reflect_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/internal/reflect"
)

// TestResolveValue_NestedStructByFieldName resolves nested struct fields using Go field names.
func TestResolveValue_NestedStructByFieldName(t *testing.T) {
	config := &Config{
		Database: &Database{
			Options: &Options{
				ServerAddr: "localhost:5432",
				PoolSize:   100,
			},
			Name: "mydb",
		},
		AppName: "MyApp",
	}

	tests := []struct {
		name      string
		value     any
		fieldName string
		want      any
		ok        bool
	}{
		{
			name:      "top-level field",
			value:     config,
			fieldName: "AppName",
			want:      "MyApp",
			ok:        true,
		},
		{
			name:      "first-level nested field",
			value:     config.Database,
			fieldName: "Name",
			want:      "mydb",
			ok:        true,
		},
		{
			name:      "second-level nested field",
			value:     config.Database.Options,
			fieldName: "ServerAddr",
			want:      "localhost:5432",
			ok:        true,
		},
		{
			name:      "nonexistent field",
			value:     config,
			fieldName: "NotExists",
			want:      nil,
			ok:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := reflect.ResolveValue(tt.value, tt.fieldName)
			require.Equal(t, tt.ok, ok)
			if ok {
				require.Equal(t, tt.want, got)
			}
		})
	}
}

// TestResolveValue_NestedStructByJSONTag resolves nested struct fields using JSON tags.
func TestResolveValue_NestedStructByJSONTag(t *testing.T) {
	user := &User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Profile: &Profile{
			Bio:      "Software engineer",
			Location: "San Francisco",
		},
	}

	tests := []struct {
		name      string
		value     any
		fieldName string
		want      any
		ok        bool
	}{
		{
			name:      "json tag on top level",
			value:     user,
			fieldName: "first_name",
			want:      "John",
			ok:        true,
		},
		{
			name:      "json tag on nested struct",
			value:     user.Profile,
			fieldName: "bio",
			want:      "Software engineer",
			ok:        true,
		},
		{
			name:      "json tag with omitempty option",
			value:     user,
			fieldName: "email_address",
			want:      "john@example.com",
			ok:        true,
		},
		{
			name:      "nonexistent json tag",
			value:     user,
			fieldName: "phone_number",
			want:      nil,
			ok:        false,
		},
		{
			name:      "field name when json tag exists",
			value:     user,
			fieldName: "FirstName",
			want:      "John",
			ok:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := reflect.ResolveValue(tt.value, tt.fieldName)
			require.Equal(t, tt.ok, ok)
			if ok {
				require.Equal(t, tt.want, got)
			}
		})
	}
}

// TestResolveValue_PointerDereference dereferences pointers automatically during traversal.
func TestResolveValue_PointerDereference(t *testing.T) {
	config := &Config{
		Database: &Database{
			Options: &Options{
				ServerAddr: "localhost:5432",
			},
			Name: "mydb",
		},
	}

	// config is *Config, should still work
	got, ok := reflect.ResolveValue(config, "AppName")
	require.True(t, ok)
	require.Equal(t, config.AppName, got)

	// Nested pointer access
	db, _ := reflect.ResolveValue(config, "Database")
	got, ok = reflect.ResolveValue(db, "Options")
	require.True(t, ok)
	opts := got.(*Options)
	require.Equal(t, "localhost:5432", opts.ServerAddr)
}

// TestResolveValue_SliceIndexing resolves slice elements by numeric index.
func TestResolveValue_SliceIndexing(t *testing.T) {
	config := &Config{
		Tags: []string{"api", "production", "critical"},
	}

	tests := []struct {
		name      string
		value     any
		fieldName string
		want      any
		ok        bool
	}{
		{
			name:      "first element",
			value:     config.Tags,
			fieldName: "0",
			want:      "api",
			ok:        true,
		},
		{
			name:      "middle element",
			value:     config.Tags,
			fieldName: "1",
			want:      "production",
			ok:        true,
		},
		{
			name:      "last element",
			value:     config.Tags,
			fieldName: "2",
			want:      "critical",
			ok:        true,
		},
		{
			name:      "out of bounds",
			value:     config.Tags,
			fieldName: "5",
			want:      nil,
			ok:        false,
		},
		{
			name:      "negative index",
			value:     config.Tags,
			fieldName: "-1",
			want:      nil,
			ok:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := reflect.ResolveValue(tt.value, tt.fieldName)
			require.Equal(t, tt.ok, ok)
			if ok {
				require.Equal(t, tt.want, got)
			}
		})
	}
}

// TestResolveValue_MapAccess resolves map values by string key.
func TestResolveValue_MapAccess(t *testing.T) {
	m := map[string]string{
		"host":     "localhost",
		"port":     "5432",
		"database": "mydb",
	}

	tests := []struct {
		name      string
		value     any
		fieldName string
		want      any
		ok        bool
	}{
		{
			name:      "existing key",
			value:     m,
			fieldName: "host",
			want:      "localhost",
			ok:        true,
		},
		{
			name:      "another existing key",
			value:     m,
			fieldName: "port",
			want:      "5432",
			ok:        true,
		},
		{
			name:      "nonexistent key",
			value:     m,
			fieldName: "timeout",
			want:      nil,
			ok:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := reflect.ResolveValue(tt.value, tt.fieldName)
			require.Equal(t, tt.ok, ok)
			if ok {
				require.Equal(t, tt.want, got)
			}
		})
	}
}

// TestResolveValue_NilAndEmpty handles nil values and empty inputs gracefully.
func TestResolveValue_NilAndEmpty(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		fieldName string
		ok        bool
	}{
		{
			name:      "nil value",
			value:     nil,
			fieldName: "Field",
			ok:        false,
		},
		{
			name:      "empty field name",
			value:     &Config{},
			fieldName: "",
			ok:        false,
		},
		{
			name:      "nil pointer",
			value:     (*Config)(nil),
			fieldName: "Database",
			ok:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := reflect.ResolveValue(tt.value, tt.fieldName)
			require.Equal(t, tt.ok, ok)
		})
	}
}

// TestCanDescend checks if a value can be descended into.
func TestCanDescend(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{
			name:  "struct pointer",
			value: &Config{},
			want:  true,
		},
		{
			name:  "struct",
			value: Config{},
			want:  true,
		},
		{
			name:  "map",
			value: map[string]string{},
			want:  true,
		},
		{
			name:  "slice",
			value: []string{},
			want:  true,
		},
		{
			name:  "array",
			value: [3]string{},
			want:  true,
		},
		{
			name:  "string",
			value: "hello",
			want:  false,
		},
		{
			name:  "int",
			value: 42,
			want:  false,
		},
		{
			name:  "nil",
			value: nil,
			want:  false,
		},
		{
			name:  "nil pointer",
			value: (*Config)(nil),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := reflect.CanDescend(tt.value)
			require.Equal(t, tt.want, got)
		})
	}
}

// TestResolveValue_WithVueStack demonstrates integration with the Vue stack.
// This test shows how struct field resolution would work within a rendering context.
func TestResolveValue_WithVueStack(t *testing.T) {
	config := &Config{
		AppName: "MyApp",
		Database: &Database{
			Name: "mydb",
			Options: &Options{
				ServerAddr: "localhost:5432",
				PoolSize:   100,
			},
		},
	}

	// Create a Vue stack with the config as data
	stack := vuego.NewStack(map[string]any{
		"config": config,
	})

	// Resolve the top-level config object from the stack
	configVal, ok := stack.Lookup("config")
	require.True(t, ok)
	require.NotNil(t, configVal)

	// Now use ResolveValue to access nested fields
	appName, ok := reflect.ResolveValue(configVal, "AppName")
	require.True(t, ok)
	require.Equal(t, "MyApp", appName)

	// Access nested struct
	db, ok := reflect.ResolveValue(configVal, "Database")
	require.True(t, ok)
	dbVal := db.(*Database)
	require.Equal(t, "mydb", dbVal.Name)

	// Access deeply nested field
	opts, ok := reflect.ResolveValue(dbVal, "Options")
	require.True(t, ok)
	optsVal := opts.(*Options)
	require.Equal(t, "localhost:5432", optsVal.ServerAddr)
	require.Equal(t, 100, optsVal.PoolSize)
}

// Test data structures with multiple levels of nesting

type Config struct {
	AppName  string
	Database *Database
	Tags     []string
}

type Database struct {
	Name    string
	Options *Options
}

type Options struct {
	ServerAddr string
	PoolSize   int
}

type User struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email_address,omitempty"`
	Profile   *Profile
}

type Profile struct {
	Bio      string `json:"bio"`
	Location string `json:"location"`
}

// TestIsSlice checks if a value is a slice or array.
func TestIsSlice(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{
			name:  "[]string",
			value: []string{"a", "b"},
			want:  true,
		},
		{
			name:  "[]int",
			value: []int{1, 2},
			want:  true,
		},
		{
			name:  "[]any",
			value: []any{},
			want:  true,
		},
		{
			name:  "array",
			value: [3]string{},
			want:  true,
		},
		{
			name:  "string",
			value: "not a slice",
			want:  false,
		},
		{
			name:  "int",
			value: 42,
			want:  false,
		},
		{
			name:  "map",
			value: map[string]int{},
			want:  false,
		},
		{
			name:  "nil",
			value: nil,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := reflect.IsSlice(tt.value)
			require.Equal(t, tt.want, got)
		})
	}
}

// TestSliceToAny converts typed slices to []any.
func TestSliceToAny(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  []any
	}{
		{
			name:  "[]string",
			input: []string{"a", "b", "c"},
			want:  []any{"a", "b", "c"},
		},
		{
			name:  "[]int",
			input: []int{1, 2, 3},
			want:  []any{1, 2, 3},
		},
		{
			name:  "[]float64",
			input: []float64{1.5, 2.5, 3.5},
			want:  []any{1.5, 2.5, 3.5},
		},
		{
			name:  "[]map[string]any",
			input: []map[string]any{{"key": "value"}},
			want:  []any{map[string]any{"key": "value"}},
		},
		{
			name:  "empty slice",
			input: []string{},
			want:  []any{},
		},
		{
			name:  "nil slice",
			input: nil,
			want:  nil,
		},
		{
			name:  "non-slice",
			input: "not a slice",
			want:  nil,
		},
		{
			name:  "array",
			input: [3]string{"x", "y", "z"},
			want:  []any{"x", "y", "z"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := reflect.SliceToAny(tt.input)
			require.Equal(t, tt.want, got)
		})
	}
}
