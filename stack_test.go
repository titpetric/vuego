package vuego_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/tests/assert"
)

// Test helper structs for struct resolution tests.
type (
	testUser struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
)

func TestNewStack(t *testing.T) {
	t.Run("a nil root is an empty scope", func(t *testing.T) {
		s := vuego.NewStack(nil)
		assert.NotNil(t, s)

		_, ok := s.Lookup("nonexistent")
		assert.False(t, ok)
	})

	t.Run("the root map is the bottom scope", func(t *testing.T) {
		val, ok := vuego.NewStack(map[string]any{"key": "value"}).Lookup("key")
		assert.True(t, ok)
		assert.Equal(t, "value", val)
	})

	t.Run("no original data means no struct fallback", func(t *testing.T) {
		_, ok := vuego.NewStack(map[string]any{}).Lookup("FirstName")
		assert.False(t, ok)
	})
}

func TestNewStackWithData(t *testing.T) {
	user := &testUser{FirstName: "Alice", LastName: "Smith"}

	t.Run("the map is the bottom scope", func(t *testing.T) {
		val, ok := vuego.NewStackWithData(map[string]any{"key": "value"}, user).Lookup("key")
		assert.True(t, ok)
		assert.Equal(t, "value", val)
	})

	t.Run("a struct field is the fallback for an unknown name", func(t *testing.T) {
		val, ok := vuego.NewStackWithData(map[string]any{}, user).Lookup("FirstName")
		assert.True(t, ok)
		assert.Equal(t, "Alice", val)
	})

	t.Run("a json tag also resolves", func(t *testing.T) {
		val, ok := vuego.NewStackWithData(map[string]any{}, user).Lookup("first_name")
		assert.True(t, ok)
		assert.Equal(t, "Alice", val)
	})

	t.Run("the map wins over the struct field", func(t *testing.T) {
		val, ok := vuego.NewStackWithData(map[string]any{"FirstName": "Override"}, user).Lookup("FirstName")
		assert.True(t, ok)
		assert.Equal(t, "Override", val)
	})

	t.Run("a nil root map is allowed", func(t *testing.T) {
		val, ok := vuego.NewStackWithData(nil, user).Lookup("LastName")
		assert.True(t, ok)
		assert.Equal(t, "Smith", val)
	})
}

func TestStack_Push(t *testing.T) {
	t.Run("the new scope is searched first", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"base": "value", "key": "parent"})
		s.Push(map[string]any{"top": "data", "key": "child"})

		for name, want := range map[string]string{"top": "data", "base": "value", "key": "child"} {
			val, ok := s.Lookup(name)
			assert.True(t, ok, name)
			assert.Equal(t, want, val)
		}
	})

	t.Run("a nil map pushes an empty scope", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"base": "value"})
		s.Push(nil)
		s.Set("added", 1)

		val, ok := s.Lookup("added")
		assert.True(t, ok)
		assert.Equal(t, 1, val)

		s.Pop()
		_, ok = s.Lookup("added")
		assert.False(t, ok)
	})
}

func TestStack_Pop(t *testing.T) {
	t.Run("the top scope is discarded", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"base": "value"})
		s.Push(map[string]any{"top": "data"})
		s.Pop()

		_, ok := s.Lookup("top")
		assert.False(t, ok)

		val, ok := s.Lookup("base")
		assert.True(t, ok)
		assert.Equal(t, "value", val)
	})

	t.Run("popping past the root leaves an empty scope", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"key": "value"})
		s.Pop()
		s.Pop()

		_, ok := s.Lookup("key")
		assert.False(t, ok)

		s.Set("new", "value")
		val, ok := s.Lookup("new")
		assert.True(t, ok)
		assert.Equal(t, "value", val)
	})
}

func TestStack_Set(t *testing.T) {
	t.Run("adds and replaces in the top scope", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"existing": "old"})

		s.Set("new", "value")
		val, ok := s.Lookup("new")
		assert.True(t, ok)
		assert.Equal(t, "value", val)

		s.Set("existing", "updated")
		val, ok = s.Lookup("existing")
		assert.True(t, ok)
		assert.Equal(t, "updated", val)
	})

	t.Run("a write to a pushed scope does not reach the root", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"key": "root"})
		s.Push(nil)
		s.Set("key", "pushed")
		s.Pop()

		val, ok := s.Lookup("key")
		assert.True(t, ok)
		assert.Equal(t, "root", val)
	})
}

func TestStack_Lookup(t *testing.T) {
	t.Run("searches from the top down", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"parent": "value", "key": "bottom"})
		s.Push(map[string]any{"child": "data", "key": "top"})

		for name, want := range map[string]string{"key": "top", "parent": "value", "child": "data"} {
			val, ok := s.Lookup(name)
			assert.True(t, ok, name)
			assert.Equal(t, want, val)
		}
	})

	t.Run("an unknown name is not found", func(t *testing.T) {
		_, ok := vuego.NewStack(map[string]any{"key": "value"}).Lookup("nonexistent")
		assert.False(t, ok)
	})
}

func TestStack_Copy(t *testing.T) {
	t.Run("carries the flattened values", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"root": "value"})
		s.Push(map[string]any{"pushed": "data"})

		copied := s.Copy()
		for name, want := range map[string]string{"root": "value", "pushed": "data"} {
			val, ok := copied.Lookup(name)
			assert.True(t, ok, name)
			assert.Equal(t, want, val)
		}
	})

	t.Run("writes to the copy do not reach the original", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"key": "original"})

		copied := s.Copy()
		copied.Set("key", "copied")
		copied.Set("added", 1)

		val, _ := s.Lookup("key")
		assert.Equal(t, "original", val)

		_, ok := s.Lookup("added")
		assert.False(t, ok)
	})

	t.Run("the struct fallback is retained", func(t *testing.T) {
		copied := vuego.NewStackWithData(nil, &testUser{FirstName: "Alice"}).Copy()

		val, ok := copied.Lookup("FirstName")
		assert.True(t, ok)
		assert.Equal(t, "Alice", val)
	})

	t.Run("concurrent copies of one stack do not race", func(t *testing.T) {
		// A base template's stack is shared by every goroutine calling Load,
		// which copies it. Copy must not write the env cache.
		s := vuego.NewStackWithData(map[string]any{"root": "value"}, &testUser{FirstName: "Alice"})

		var wg sync.WaitGroup
		for i := 0; i < 32; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				copied := s.Copy()
				if val, ok := copied.Lookup("root"); !ok || val != "value" {
					t.Errorf("copied stack lost root: %v %v", val, ok)
				}
			}()
		}
		wg.Wait()
	})

	t.Run("the copy starts with one scope", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"root": "value"})
		s.Push(map[string]any{"pushed": "data"})

		copied := s.Copy()
		copied.Pop()

		// Popping the only scope leaves an empty one, so both keys are gone.
		_, ok := copied.Lookup("root")
		assert.False(t, ok)
	})
}

func TestStack_EnvMap(t *testing.T) {
	t.Run("flattens the scopes with the top winning", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"a": 1, "shared": "root"})
		s.Push(map[string]any{"b": 2, "shared": "pushed"})

		env := s.EnvMap()
		assert.Equal(t, 1, env["a"])
		assert.Equal(t, 2, env["b"])
		assert.Equal(t, "pushed", env["shared"])
	})

	t.Run("struct fields are keyed by json tag", func(t *testing.T) {
		env := vuego.NewStackWithData(map[string]any{}, &testUser{FirstName: "Alice", LastName: "Smith"}).EnvMap()

		assert.Equal(t, "Alice", env["first_name"])
		assert.Equal(t, "Smith", env["last_name"])
	})

	t.Run("struct fields override the stack values", func(t *testing.T) {
		env := vuego.NewStackWithData(
			map[string]any{"first_name": "FromMap"},
			&testUser{FirstName: "FromStruct"},
		).EnvMap()

		assert.Equal(t, "FromStruct", env["first_name"])
	})

	t.Run("the result is cached until the stack is mutated", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"a": 1})

		first := s.EnvMap()
		assert.Equal(t, first, s.EnvMap())

		s.Set("b", 2)
		assert.Equal(t, 2, s.EnvMap()["b"])
	})

	t.Run("Push and Pop invalidate the cache", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"a": 1})
		_ = s.EnvMap()

		s.Push(map[string]any{"b": 2})
		assert.Equal(t, 2, s.EnvMap()["b"])

		s.Pop()
		_, ok := s.EnvMap()["b"]
		assert.False(t, ok)
	})

	t.Run("an empty stack yields an empty map", func(t *testing.T) {
		assert.Len(t, vuego.NewStack(nil).EnvMap(), 0)
	})
}

func TestStack_ForEach(t *testing.T) {
	s := vuego.NewStack(map[string]any{
		"any":       []any{"a", "b", "c"},
		"strings":   []string{"x", "y"},
		"ints":      []int{1, 2, 3},
		"mapSlice":  []map[string]any{{"id": 1}, {"id": 2}},
		"anyMap":    map[string]any{"a": 1, "b": 2},
		"stringMap": map[string]string{"host": "localhost", "port": "8080"},
		"emptyAny":  []any{},
		"emptyMap":  map[string]any{},
		"scalar":    "string",
		"boolean":   true,
		"number":    42,
	})

	counts := []struct {
		expr string
		want int
	}{
		{expr: "any", want: 3},
		{expr: "strings", want: 2},
		{expr: "ints", want: 3},
		{expr: "mapSlice", want: 2},
		{expr: "anyMap", want: 2},
		{expr: "stringMap", want: 2},
		{expr: "emptyAny", want: 0},
		{expr: "emptyMap", want: 0},
		// A non-collection is a no-op rather than an error.
		{expr: "scalar", want: 0},
		{expr: "boolean", want: 0},
		{expr: "number", want: 0},
		{expr: "missing", want: 0},
	}

	for _, test := range counts {
		t.Run(test.expr, func(t *testing.T) {
			count := 0
			assert.NoError(t, s.ForEach(test.expr, func(int, any) error {
				count++
				return nil
			}))
			assert.Equal(t, test.want, count)
		})
	}

	t.Run("values arrive in slice order", func(t *testing.T) {
		var got []string
		assert.NoError(t, s.ForEach("any", func(_ int, v any) error {
			got = append(got, v.(string))
			return nil
		}))
		assert.Equal(t, []string{"a", "b", "c"}, got)
	})

	t.Run("indices count up from zero", func(t *testing.T) {
		for _, expr := range []string{"ints", "anyMap"} {
			var indices []int
			assert.NoError(t, s.ForEach(expr, func(i int, _ any) error {
				indices = append(indices, i)
				return nil
			}))
			assert.Equal(t, []int{0, 1, 2}[:len(indices)], indices)
		}
	})

	// An error from fn stops the walk and is passed through.
	stops := []string{"any", "strings", "ints", "anyMap"}

	for _, expr := range stops {
		t.Run("an error stops the walk over "+expr, func(t *testing.T) {
			stop := errors.New("stop")

			count := 0
			err := s.ForEach(expr, func(int, any) error {
				count++
				if count == 2 {
					return stop
				}
				return nil
			})

			assert.ErrorIs(t, err, stop)
			assert.Equal(t, 2, count)
		})
	}
}
