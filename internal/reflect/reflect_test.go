package reflect_test

import (
	"testing"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/internal/reflect"
	"github.com/titpetric/vuego/tests/assert"
)

// Test data structures with multiple levels of nesting.
type (
	Config struct {
		AppName  string
		Database *Database
		Tags     []string
	}

	Database struct {
		Name    string
		Options *Options
	}

	Options struct {
		ServerAddr string
		PoolSize   int
	}

	User struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email_address,omitempty"`
		Profile   *Profile
		password  string //nolint:unused // unexported fields must stay out of StructToMap
	}

	Profile struct {
		Bio      string `json:"bio"`
		Location string `json:"location"`
	}
)

func testConfig() *Config {
	return &Config{
		AppName: "MyApp",
		Database: &Database{
			Name: "mydb",
			Options: &Options{
				ServerAddr: "localhost:5432",
				PoolSize:   100,
			},
		},
		Tags: []string{"api", "production", "critical"},
	}
}

func testUser() *User {
	return &User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Profile: &Profile{
			Bio:      "Software engineer",
			Location: "San Francisco",
		},
		password: "secret",
	}
}

func TestResolveValue(t *testing.T) {
	config := testConfig()
	user := testUser()

	tests := []struct {
		name  string
		value any
		field string
		want  any
		ok    bool
	}{
		// Struct fields by Go field name.
		{name: "top-level field", value: config, field: "AppName", want: "MyApp", ok: true},
		{name: "first-level nested field", value: config.Database, field: "Name", want: "mydb", ok: true},
		{name: "second-level nested field", value: config.Database.Options, field: "ServerAddr", want: "localhost:5432", ok: true},
		{name: "nonexistent field", value: config, field: "NotExists", ok: false},

		// Struct fields by JSON tag.
		{name: "json tag on top level", value: user, field: "first_name", want: "John", ok: true},
		{name: "json tag on nested struct", value: user.Profile, field: "bio", want: "Software engineer", ok: true},
		{name: "json tag with omitempty option", value: user, field: "email_address", want: "john@example.com", ok: true},
		{name: "field name wins when a json tag also exists", value: user, field: "FirstName", want: "John", ok: true},
		{name: "nonexistent json tag", value: user, field: "phone_number", ok: false},

		// Slices and arrays by numeric index.
		{name: "first element", value: config.Tags, field: "0", want: "api", ok: true},
		{name: "middle element", value: config.Tags, field: "1", want: "production", ok: true},
		{name: "last element", value: config.Tags, field: "2", want: "critical", ok: true},
		{name: "array element", value: [3]string{"x", "y", "z"}, field: "1", want: "y", ok: true},
		{name: "index out of bounds", value: config.Tags, field: "5", ok: false},
		{name: "negative index", value: config.Tags, field: "-1", ok: false},
		{name: "index is not a number", value: config.Tags, field: "first", ok: false},

		// Maps by string key.
		{name: "existing map key", value: map[string]string{"host": "localhost"}, field: "host", want: "localhost", ok: true},
		{name: "map value of any type", value: map[string]any{"port": 5432}, field: "port", want: 5432, ok: true},
		{name: "nonexistent map key", value: map[string]string{"host": "localhost"}, field: "timeout", ok: false},

		// Nil and empty inputs.
		{name: "nil value", value: nil, field: "Field", ok: false},
		{name: "empty field name", value: &Config{}, field: "", ok: false},
		{name: "nil pointer", value: (*Config)(nil), field: "Database", ok: false},
		{name: "nil nested pointer", value: &Config{}, field: "Database", want: (*Database)(nil), ok: true},
		{name: "scalar cannot be descended", value: 42, field: "Field", ok: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := reflect.ResolveValue(test.value, test.field)
			assert.Equal(t, test.ok, ok)
			if ok {
				assert.Equal(t, test.want, got)
			}
		})
	}

	t.Run("dereferences pointers while traversing", func(t *testing.T) {
		db, ok := reflect.ResolveValue(config, "Database")
		assert.True(t, ok)

		opts, ok := reflect.ResolveValue(db, "Options")
		assert.True(t, ok)
		assert.Equal(t, "localhost:5432", opts.(*Options).ServerAddr)
	})

	t.Run("resolves struct fields held in a Stack", func(t *testing.T) {
		stack := vuego.NewStack(map[string]any{"config": config})

		configVal, ok := stack.Lookup("config")
		assert.True(t, ok)

		appName, ok := reflect.ResolveValue(configVal, "AppName")
		assert.True(t, ok)
		assert.Equal(t, "MyApp", appName)

		db, ok := reflect.ResolveValue(configVal, "Database")
		assert.True(t, ok)

		opts, ok := reflect.ResolveValue(db.(*Database), "Options")
		assert.True(t, ok)
		assert.Equal(t, 100, opts.(*Options).PoolSize)
	})
}

func TestCanDescend(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{name: "struct pointer", value: &Config{}, want: true},
		{name: "struct", value: Config{}, want: true},
		{name: "map", value: map[string]string{}, want: true},
		{name: "slice", value: []string{}, want: true},
		{name: "array", value: [3]string{}, want: true},
		{name: "string", value: "hello", want: false},
		{name: "int", value: 42, want: false},
		{name: "nil", value: nil, want: false},
		{name: "nil pointer", value: (*Config)(nil), want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, reflect.CanDescend(test.value))
		})
	}
}

func TestStructToMap(t *testing.T) {
	t.Run("keys by json tag, falling back to the field name", func(t *testing.T) {
		m := reflect.StructToMap(testUser())
		assert.Equal(t, "John", m["first_name"])
		assert.Equal(t, "Doe", m["last_name"])
		assert.Equal(t, "john@example.com", m["email_address"])
		assert.NotNil(t, m["Profile"])
	})

	t.Run("skips unexported fields", func(t *testing.T) {
		m := reflect.StructToMap(testUser())
		assert.Nil(t, m["password"])
		assert.Len(t, m, 4)
	})

	t.Run("converts nested structs to maps", func(t *testing.T) {
		profile, ok := reflect.StructToMap(testUser())["Profile"].(map[string]any)
		assert.True(t, ok, "Profile is a map")
		assert.Equal(t, "Software engineer", profile["bio"])
		assert.Equal(t, "San Francisco", profile["location"])
	})

	t.Run("converts a struct value as well as a pointer", func(t *testing.T) {
		assert.Equal(t,
			reflect.StructToMap(Options{ServerAddr: "a", PoolSize: 1}),
			reflect.StructToMap(&Options{ServerAddr: "a", PoolSize: 1}),
		)
	})

	t.Run("a nil nested pointer becomes an empty map", func(t *testing.T) {
		nested, ok := reflect.StructToMap(&Database{Name: "mydb"})["Options"].(map[string]any)
		assert.True(t, ok, "Options is a map")
		assert.Len(t, nested, 0)
	})

	t.Run("leaves non-struct fields alone", func(t *testing.T) {
		m := reflect.StructToMap(testConfig())
		assert.Equal(t, "MyApp", m["AppName"])
		assert.Equal(t, []string{"api", "production", "critical"}, m["Tags"])
	})

	tests := []struct {
		name  string
		input any
	}{
		{name: "nil", input: nil},
		{name: "nil pointer", input: (*Config)(nil)},
		{name: "map", input: map[string]any{"a": 1}},
		{name: "slice", input: []string{"a"}},
		{name: "scalar", input: 42},
	}

	for _, test := range tests {
		t.Run(test.name+" yields an empty map", func(t *testing.T) {
			m := reflect.StructToMap(test.input)
			assert.NotNil(t, m)
			assert.Len(t, m, 0)
		})
	}
}

func TestIsSlice(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  bool
	}{
		{name: "[]string", value: []string{"a", "b"}, want: true},
		{name: "[]int", value: []int{1, 2}, want: true},
		{name: "[]any", value: []any{}, want: true},
		{name: "array", value: [3]string{}, want: true},
		{name: "string", value: "not a slice", want: false},
		{name: "int", value: 42, want: false},
		{name: "map", value: map[string]int{}, want: false},
		{name: "nil", value: nil, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, reflect.IsSlice(test.value))
		})
	}
}

func TestSliceToAny(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  []any
	}{
		{name: "[]string", input: []string{"a", "b", "c"}, want: []any{"a", "b", "c"}},
		{name: "[]int", input: []int{1, 2, 3}, want: []any{1, 2, 3}},
		{name: "[]float64", input: []float64{1.5, 2.5, 3.5}, want: []any{1.5, 2.5, 3.5}},
		{name: "[]map[string]any", input: []map[string]any{{"key": "value"}}, want: []any{map[string]any{"key": "value"}}},
		{name: "array", input: [3]string{"x", "y", "z"}, want: []any{"x", "y", "z"}},
		{name: "empty slice", input: []string{}, want: []any{}},
		{name: "nil slice", input: nil, want: nil},
		{name: "non-slice", input: "not a slice", want: nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, reflect.SliceToAny(test.input))
		})
	}
}
