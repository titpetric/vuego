package vuego_test

import (
	"testing"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/tests/assert"
)

// Nested fixtures for the path resolution table below.
type (
	testConfig struct {
		Name        string
		Version     string
		Port        int
		Database    *testDatabase
		Cache       *testCache
		Services    []*testService
		SecretKey   string `json:"secret_key"`
		Environment string `json:"env"`
		Debug       bool
		Admins      []string
		Tags        map[string]string
	}

	testDatabase struct {
		Host     string
		Port     int
		Name     string
		Replicas []*testReplica
		Options  *testDatabaseOptions
		MaxConns int `json:"max_connections"`
	}

	testReplica struct {
		Host   string
		Port   int
		Region string
		Status string
	}

	testDatabaseOptions struct {
		SSL           bool
		Timeout       int
		PoolSize      int `json:"pool_size"`
		RetryAttempts int `json:"retry_attempts"`
	}

	testCache struct {
		Type    string
		Host    string
		Port    int
		Servers []*testCacheServer
		TTL     int
	}

	testCacheServer struct {
		Address string
		Weight  int
		Active  bool
	}

	testService struct {
		Name    string
		URL     string
		Timeout int
		Enabled bool
		Config  *testServiceConfig
	}

	testServiceConfig struct {
		APIKey    string `json:"api_key"`
		RateLimit int    `json:"rate_limit"`
		Endpoints []string
	}
)

// fullConfig is the fixture the resolution table walks. Every field it sets is
// reachable from at least one expression below.
func fullConfig() *testConfig {
	return &testConfig{
		Name:        "ProductionApp",
		Port:        8080,
		SecretKey:   "supersecret",
		Environment: "production",
		Admins:      []string{"alice@example.com", "bob@example.com", "charlie@example.com"},
		Database: &testDatabase{
			Host: "db.prod.local",
			Port: 5432,
			Name: "maindb",
			Options: &testDatabaseOptions{
				SSL:           true,
				PoolSize:      100,
				Timeout:       30,
				RetryAttempts: 5,
			},
			Replicas: []*testReplica{
				{Host: "replica1.local", Port: 5433, Region: "us-east-1", Status: "active"},
				{Host: "replica2.local", Port: 5433, Region: "us-west-1", Status: "active"},
				{Host: "replica3.local", Port: 5433, Region: "eu-west-1", Status: "standby"},
			},
		},
		Cache: &testCache{
			Type: "redis",
			Host: "cache.local",
			Port: 6379,
			Servers: []*testCacheServer{
				{Address: "cache1:6379", Weight: 100, Active: true},
				{Address: "cache2:6379", Weight: 50, Active: true},
				{Address: "cache3:6379", Weight: 25, Active: false},
			},
		},
		Services: []*testService{
			{
				Name: "AuthService", URL: "http://auth:3000", Timeout: 5, Enabled: true,
				Config: &testServiceConfig{APIKey: "secret-key-123", RateLimit: 1000, Endpoints: []string{"/login", "/logout", "/verify"}},
			},
			{
				Name: "PaymentService", URL: "http://payment:3001", Timeout: 10, Enabled: true,
				Config: &testServiceConfig{APIKey: "payment-key-456", RateLimit: 500, Endpoints: []string{"/charge", "/refund", "/status"}},
			},
		},
	}
}

func TestStack_Resolve(t *testing.T) {
	config := fullConfig()

	s := vuego.NewStack(map[string]any{
		"config":  config,
		"user":    &testUser{FirstName: "Alice", LastName: "Smith"},
		"name":    "Alice",
		"nested":  map[string]any{"name": "Bob"},
		"strings": []string{"a", "b"},
		"ints":    []int{10, 20},
		"floats":  []float64{1.5, 2.5},
		"maps":    []map[string]any{{"key": "val"}},
		"items": []any{
			map[string]any{"title": "First"},
			map[string]any{"title": "Second"},
		},
		"hostmap": map[string]string{"host": "localhost"},
		"nothing": nil,
		"empty":   map[string]any{},
	})

	tests := []struct {
		name string
		expr string
		want any
		ok   bool
	}{
		// Plain identifiers and maps.
		{name: "simple key", expr: "name", want: "Alice", ok: true},
		{name: "nested map", expr: "nested.name", want: "Bob", ok: true},
		{name: "map[string]string", expr: "hostmap.host", want: "localhost", ok: true},

		// Indexing.
		{name: "index into []any", expr: "items[1].title", want: "Second", ok: true},
		{name: "index into []string", expr: "strings[0]", want: "a", ok: true},
		{name: "index into []int", expr: "ints[1]", want: 20, ok: true},
		{name: "index into []float64", expr: "floats[0]", want: 1.5, ok: true},
		{name: "index into []map", expr: "maps[0].key", want: "val", ok: true},

		// Struct fields by Go name.
		{name: "struct field", expr: "config.Name", want: "ProductionApp", ok: true},
		{name: "struct int field", expr: "config.Port", want: 8080, ok: true},
		{name: "two levels down", expr: "config.Database.Host", want: "db.prod.local", ok: true},
		{name: "three levels down", expr: "config.Database.Options.PoolSize", want: 100, ok: true},
		{name: "bool field", expr: "config.Database.Options.SSL", want: true, ok: true},

		// Struct fields by json tag.
		{name: "json tag", expr: "config.secret_key", want: "supersecret", ok: true},
		{name: "renaming json tag", expr: "config.env", want: "production", ok: true},
		{name: "nested json tag", expr: "config.Database.Options.pool_size", want: 100, ok: true},
		{name: "json tag on a zero value", expr: "config.Database.max_connections", want: 0, ok: true},
		{name: "json tag on a top-level struct", expr: "user.first_name", want: "Alice", ok: true},
		{name: "field name still works when a json tag exists", expr: "user.FirstName", want: "Alice", ok: true},

		// Indexing through struct fields.
		{name: "index a string slice field", expr: "config.Admins[0]", want: "alice@example.com", ok: true},
		{name: "index a struct slice field", expr: "config.Database.Replicas[1].Region", want: "us-west-1", ok: true},
		{name: "index the last element", expr: "config.Database.Replicas[2].Status", want: "standby", ok: true},
		{name: "index two struct slices deep", expr: "config.Cache.Servers[1].Address", want: "cache2:6379", ok: true},
		{name: "json tag under an index", expr: "config.Services[1].Config.api_key", want: "payment-key-456", ok: true},

		// Failures.
		{name: "index out of range", expr: "items[5]", ok: false},
		{name: "index a struct slice out of range", expr: "config.Admins[10]", ok: false},
		{name: "nil intermediate", expr: "nothing.name", ok: false},
		{name: "missing key in a map", expr: "empty.missing", ok: false},
		{name: "unknown root", expr: "unknown.field", ok: false},
		{name: "unknown identifier", expr: "unknown", ok: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := s.Resolve(test.expr)
			assert.Equal(t, test.ok, ok)
			if test.ok {
				assert.Equal(t, test.want, got)
			}
		})
	}

	t.Run("a slice field resolves to the typed slice", func(t *testing.T) {
		got, ok := s.Resolve("config.Database.Replicas")
		assert.True(t, ok)
		assert.Len(t, got.([]*testReplica), 3)
	})

	t.Run("a slice under an index resolves to the typed slice", func(t *testing.T) {
		got, ok := s.Resolve("config.Services[0].Config.Endpoints")
		assert.True(t, ok)
		assert.Equal(t, []string{"/login", "/logout", "/verify"}, got)
	})
}

// The path cache is exercised through Resolve, which is the only caller of it.
func TestResolvePathParsing(t *testing.T) {
	s := vuego.NewStack(map[string]any{
		"items":    []any{"first", "second"},
		"nested":   map[string]any{"a": map[string]any{"b": "deep"}},
		"hyphen":   map[string]any{"a-b": "dashed"},
		"dotted":   map[string]any{"a.b": "literal"},
		"matrix":   []any{[]any{"r0c0", "r0c1"}, []any{"r1c0", "r1c1"}},
		"unusual[": "literal bracket",
	})

	tests := []struct {
		name string
		expr string
		want any
		ok   bool
	}{
		{name: "bracket index", expr: "items[0]", want: "first", ok: true},
		{name: "dotted index", expr: "items.1", want: "second", ok: true},
		{name: "two dots", expr: "nested.a.b", want: "deep", ok: true},
		{name: "bracket after bracket", expr: "matrix[1][0]", want: "r1c0", ok: true},
		{name: "bracket then dot", expr: "matrix[0].1", want: "r0c1", ok: true},
		{name: "single-quoted bracket key", expr: "hyphen['a-b']", want: "dashed", ok: true},
		{name: "double-quoted bracket key", expr: `hyphen["a-b"]`, want: "dashed", ok: true},
		{name: "surrounding whitespace is trimmed", expr: "  nested.a.b  ", want: "deep", ok: true},
		{name: "whitespace inside brackets is trimmed", expr: "items[ 0 ]", want: "first", ok: true},
		{name: "repeated dots collapse", expr: "nested..a..b", want: "deep", ok: true},
		{name: "a trailing dot is dropped", expr: "nested.a.b.", want: "deep", ok: true},
		{name: "empty brackets are dropped", expr: "nested[].a.b", want: "deep", ok: true},
		// A dot inside brackets still splits, so a key holding a dot is not
		// reachable this way.
		{name: "a dotted key is not reachable", expr: "dotted['a.b']", ok: false},
		// An unterminated bracket is kept as a literal character, so the whole
		// expression becomes one key.
		{name: "unterminated bracket", expr: "items[", ok: false},
		{name: "unterminated bracket names a key", expr: "unusual[", want: "literal bracket", ok: true},
		{name: "empty expression", expr: "", ok: false},
		{name: "only dots", expr: "...", ok: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := s.Resolve(test.expr)
			assert.Equal(t, test.ok, ok)
			if test.ok {
				assert.Equal(t, test.want, got)
			}
		})
	}

	t.Run("a repeated expression resolves the same way", func(t *testing.T) {
		// The second call is served from the cache.
		first, ok := s.Resolve("nested.a.b")
		assert.True(t, ok)

		second, ok := s.Resolve("nested.a.b")
		assert.True(t, ok)
		assert.Equal(t, first, second)
	})

	t.Run("a cached path is reused across stacks", func(t *testing.T) {
		other := vuego.NewStack(map[string]any{
			"nested": map[string]any{"a": map[string]any{"b": "other"}},
		})

		got, ok := other.Resolve("nested.a.b")
		assert.True(t, ok)
		assert.Equal(t, "other", got)
	})
}

func TestStack_GetString(t *testing.T) {
	s := vuego.NewStack(map[string]any{
		"name":    "Alice",
		"int":     42,
		"int8":    int8(8),
		"int64":   int64(100),
		"uint":    uint(7),
		"float":   3.14,
		"boolean": true,
		"slice":   []int{1, 2},
		"nothing": nil,
	})

	tests := []struct {
		expr string
		want string
		ok   bool
	}{
		{expr: "name", want: "Alice", ok: true},
		{expr: "int", want: "42", ok: true},
		{expr: "int8", want: "8", ok: true},
		{expr: "int64", want: "100", ok: true},
		{expr: "uint", want: "7", ok: true},
		{expr: "float", want: "3.14", ok: true},
		{expr: "boolean", want: "true", ok: true},
		{expr: "slice", want: "[1 2]", ok: true},
		{expr: "nothing", ok: false},
		{expr: "missing", ok: false},
	}

	for _, test := range tests {
		t.Run(test.expr, func(t *testing.T) {
			got, ok := s.GetString(test.expr)
			assert.Equal(t, test.ok, ok)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestStack_GetInt(t *testing.T) {
	s := vuego.NewStack(map[string]any{
		"int":     42,
		"int8":    int8(8),
		"int16":   int16(16),
		"int32":   int32(32),
		"int64":   int64(64),
		"uint":    uint(100),
		"float32": float32(2.5),
		"float64": 3.7,
		"numeric": "123",
		"words":   "not-a-number",
		"boolean": true,
		"nothing": nil,
	})

	tests := []struct {
		expr string
		want int
		ok   bool
	}{
		{expr: "int", want: 42, ok: true},
		{expr: "int8", want: 8, ok: true},
		{expr: "int16", want: 16, ok: true},
		{expr: "int32", want: 32, ok: true},
		{expr: "int64", want: 64, ok: true},
		{expr: "uint", want: 100, ok: true},
		{expr: "float32", want: 2, ok: true},
		{expr: "float64", want: 3, ok: true},
		{expr: "numeric", want: 123, ok: true},
		{expr: "words", ok: false},
		{expr: "boolean", ok: false},
		{expr: "nothing", ok: false},
		{expr: "missing", ok: false},
	}

	for _, test := range tests {
		t.Run(test.expr, func(t *testing.T) {
			got, ok := s.GetInt(test.expr)
			assert.Equal(t, test.ok, ok)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestStack_GetSlice(t *testing.T) {
	s := vuego.NewStack(map[string]any{
		"any":     []any{1, "two", 3.0},
		"strings": []string{"a", "b"},
		"ints":    []int{10, 20},
		"floats":  []float64{1.5, 2.5},
		"maps":    []map[string]any{{"key": "val"}},
		"array":   [2]string{"x", "y"},
		"scalar":  "value",
		"nothing": nil,
	})

	tests := []struct {
		expr string
		want []any
		ok   bool
	}{
		{expr: "any", want: []any{1, "two", 3.0}, ok: true},
		{expr: "strings", want: []any{"a", "b"}, ok: true},
		{expr: "ints", want: []any{10, 20}, ok: true},
		{expr: "floats", want: []any{1.5, 2.5}, ok: true},
		{expr: "maps", want: []any{map[string]any{"key": "val"}}, ok: true},
		{expr: "array", want: []any{"x", "y"}, ok: true},
		{expr: "scalar", ok: false},
		{expr: "nothing", ok: false},
		{expr: "missing", ok: false},
	}

	for _, test := range tests {
		t.Run(test.expr, func(t *testing.T) {
			got, ok := s.GetSlice(test.expr)
			assert.Equal(t, test.ok, ok)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestStack_GetMap(t *testing.T) {
	s := vuego.NewStack(map[string]any{
		"anyMap":    map[string]any{"key": "value"},
		"stringMap": map[string]string{"key": "value"},
		"intMap":    map[string]int{"key": 1},
		"scalar":    "value",
		"nothing":   nil,
	})

	tests := []struct {
		expr string
		want map[string]any
		ok   bool
	}{
		{expr: "anyMap", want: map[string]any{"key": "value"}, ok: true},
		{expr: "stringMap", want: map[string]any{"key": "value"}, ok: true},
		// Only map[string]any and map[string]string are converted.
		{expr: "intMap", ok: false},
		{expr: "scalar", ok: false},
		{expr: "nothing", ok: false},
		{expr: "missing", ok: false},
	}

	for _, test := range tests {
		t.Run(test.expr, func(t *testing.T) {
			got, ok := s.GetMap(test.expr)
			assert.Equal(t, test.ok, ok)
			assert.Equal(t, test.want, got)
		})
	}
}
