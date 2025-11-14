package vuego_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/titpetric/vuego"
)

func TestNewStack(t *testing.T) {
	t.Run("with nil root", func(t *testing.T) {
		s := vuego.NewStack(nil)
		require.NotNil(t, s)
		_, ok := s.Lookup("nonexistent")
		require.False(t, ok)
	})

	t.Run("with initial data", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"key": "value"})
		val, ok := s.Lookup("key")
		require.True(t, ok)
		require.Equal(t, "value", val)
	})
}

func TestStack_Push(t *testing.T) {
	s := vuego.NewStack(map[string]any{"base": "value"})

	s.Push(map[string]any{"top": "data"})
	val, ok := s.Lookup("top")
	require.True(t, ok)
	require.Equal(t, "data", val)

	val, ok = s.Lookup("base")
	require.True(t, ok)
	require.Equal(t, "value", val)
}

func TestStack_Pop(t *testing.T) {
	t.Run("pop to root", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"base": "value"})
		s.Push(map[string]any{"top": "data"})
		s.Pop()

		_, ok := s.Lookup("top")
		require.False(t, ok)

		val, ok := s.Lookup("base")
		require.True(t, ok)
		require.Equal(t, "value", val)
	})

	t.Run("pop on empty stack creates empty root", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"key": "value"})
		s.Pop()
		s.Pop()

		_, ok := s.Lookup("key")
		require.False(t, ok)

		s.Set("new", "value")
		val, ok := s.Lookup("new")
		require.True(t, ok)
		require.Equal(t, "value", val)
	})
}

func TestStack_Set(t *testing.T) {
	s := vuego.NewStack(map[string]any{"existing": "old"})

	s.Set("new", "value")
	val, ok := s.Lookup("new")
	require.True(t, ok)
	require.Equal(t, "value", val)

	s.Set("existing", "updated")
	val, ok = s.Lookup("existing")
	require.True(t, ok)
	require.Equal(t, "updated", val)
}

func TestStack_Lookup(t *testing.T) {
	t.Run("finds in top scope", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"base": "bottom"})
		s.Push(map[string]any{"key": "top"})

		val, ok := s.Lookup("key")
		require.True(t, ok)
		require.Equal(t, "top", val)
	})

	t.Run("finds in parent scope", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"parent": "value"})
		s.Push(map[string]any{"child": "data"})

		val, ok := s.Lookup("parent")
		require.True(t, ok)
		require.Equal(t, "value", val)
	})

	t.Run("top scope shadows parent", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"key": "parent"})
		s.Push(map[string]any{"key": "child"})

		val, ok := s.Lookup("key")
		require.True(t, ok)
		require.Equal(t, "child", val)
	})

	t.Run("not found", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"key": "value"})
		_, ok := s.Lookup("nonexistent")
		require.False(t, ok)
	})
}

func TestStack_Resolve(t *testing.T) {
	t.Run("simple key", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"name": "Alice"})
		val, ok := s.Resolve("name")
		require.True(t, ok)
		require.Equal(t, "Alice", val)
	})

	t.Run("nested map dot notation", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{
			"user": map[string]any{"name": "Bob"},
		})
		val, ok := s.Resolve("user.name")
		require.True(t, ok)
		require.Equal(t, "Bob", val)
	})

	t.Run("array index", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{
			"items": []any{"first", "second", "third"},
		})
		val, ok := s.Resolve("items[1]")
		require.True(t, ok)
		require.Equal(t, "second", val)
	})

	t.Run("nested array and map", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{
			"items": []any{
				map[string]any{"title": "First"},
				map[string]any{"title": "Second"},
			},
		})
		val, ok := s.Resolve("items[0].title")
		require.True(t, ok)
		require.Equal(t, "First", val)
	})

	t.Run("map[string]string support", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{
			"config": map[string]string{"host": "localhost"},
		})
		val, ok := s.Resolve("config.host")
		require.True(t, ok)
		require.Equal(t, "localhost", val)
	})

	t.Run("typed slices", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{
			"strings": []string{"a", "b"},
			"ints":    []int{10, 20},
			"floats":  []float64{1.5, 2.5},
			"maps":    []map[string]any{{"key": "val"}},
		})

		val, ok := s.Resolve("strings[0]")
		require.True(t, ok)
		require.Equal(t, "a", val)

		val, ok = s.Resolve("ints[1]")
		require.True(t, ok)
		require.Equal(t, 20, val)

		val, ok = s.Resolve("floats[0]")
		require.True(t, ok)
		require.Equal(t, 1.5, val)

		val, ok = s.Resolve("maps[0].key")
		require.True(t, ok)
		require.Equal(t, "val", val)
	})

	t.Run("invalid index", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"items": []any{"one"}})
		_, ok := s.Resolve("items[5]")
		require.False(t, ok)
	})

	t.Run("nil intermediate value", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"user": nil})
		_, ok := s.Resolve("user.name")
		require.False(t, ok)
	})

	t.Run("nonexistent key", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"user": map[string]any{}})
		_, ok := s.Resolve("user.missing")
		require.False(t, ok)
	})
}

func TestStack_GetString(t *testing.T) {
	t.Run("string value", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"name": "Alice"})
		val, ok := s.GetString("name")
		require.True(t, ok)
		require.Equal(t, "Alice", val)
	})

	t.Run("int to string", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{
			"a": 42,
			"b": int8(8),
			"c": int64(100),
		})
		val, ok := s.GetString("a")
		require.True(t, ok)
		require.Equal(t, "42", val)
	})

	t.Run("float to string", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"pi": 3.14})
		val, ok := s.GetString("pi")
		require.True(t, ok)
		require.Contains(t, val, "3.14")
	})

	t.Run("bool to string", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"flag": true})
		val, ok := s.GetString("flag")
		require.True(t, ok)
		require.Equal(t, "true", val)
	})

	t.Run("nil value", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"key": nil})
		_, ok := s.GetString("key")
		require.False(t, ok)
	})

	t.Run("missing key", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{})
		_, ok := s.GetString("missing")
		require.False(t, ok)
	})
}

func TestStack_GetInt(t *testing.T) {
	t.Run("int value", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"count": 42})
		val, ok := s.GetInt("count")
		require.True(t, ok)
		require.Equal(t, 42, val)
	})

	t.Run("various int types", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{
			"i8":  int8(8),
			"i16": int16(16),
			"i32": int32(32),
			"i64": int64(64),
			"u":   uint(100),
		})

		val, ok := s.GetInt("i8")
		require.True(t, ok)
		require.Equal(t, 8, val)

		val, ok = s.GetInt("i64")
		require.True(t, ok)
		require.Equal(t, 64, val)
	})

	t.Run("float to int", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"f": 3.7})
		val, ok := s.GetInt("f")
		require.True(t, ok)
		require.Equal(t, 3, val)
	})

	t.Run("string to int", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"str": "123"})
		val, ok := s.GetInt("str")
		require.True(t, ok)
		require.Equal(t, 123, val)
	})

	t.Run("invalid string", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"str": "not-a-number"})
		_, ok := s.GetInt("str")
		require.False(t, ok)
	})

	t.Run("nil value", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"key": nil})
		_, ok := s.GetInt("key")
		require.False(t, ok)
	})
}

func TestStack_GetSlice(t *testing.T) {
	t.Run("[]any", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"items": []any{1, "two", 3.0}})
		val, ok := s.GetSlice("items")
		require.True(t, ok)
		require.Len(t, val, 3)
		require.Equal(t, 1, val[0])
	})

	t.Run("[]string", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"items": []string{"a", "b"}})
		val, ok := s.GetSlice("items")
		require.True(t, ok)
		require.Len(t, val, 2)
		require.Equal(t, "a", val[0])
	})

	t.Run("[]int", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"items": []int{10, 20}})
		val, ok := s.GetSlice("items")
		require.True(t, ok)
		require.Len(t, val, 2)
		require.Equal(t, 10, val[0])
	})

	t.Run("[]float64", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"items": []float64{1.5, 2.5}})
		val, ok := s.GetSlice("items")
		require.True(t, ok)
		require.Len(t, val, 2)
		require.Equal(t, 1.5, val[0])
	})

	t.Run("[]map[string]any", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{
			"items": []map[string]any{{"key": "val"}},
		})
		val, ok := s.GetSlice("items")
		require.True(t, ok)
		require.Len(t, val, 1)
	})

	t.Run("not a slice", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"key": "value"})
		_, ok := s.GetSlice("key")
		require.False(t, ok)
	})

	t.Run("nil value", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"key": nil})
		_, ok := s.GetSlice("key")
		require.False(t, ok)
	})
}

func TestStack_GetMap(t *testing.T) {
	t.Run("map[string]any", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{
			"obj": map[string]any{"key": "value"},
		})
		val, ok := s.GetMap("obj")
		require.True(t, ok)
		require.Equal(t, "value", val["key"])
	})

	t.Run("map[string]string", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{
			"obj": map[string]string{"key": "value"},
		})
		val, ok := s.GetMap("obj")
		require.True(t, ok)
		require.Equal(t, "value", val["key"])
	})

	t.Run("not a map", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"key": "value"})
		_, ok := s.GetMap("key")
		require.False(t, ok)
	})

	t.Run("nil value", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"key": nil})
		_, ok := s.GetMap("key")
		require.False(t, ok)
	})
}

func TestStack_ForEach(t *testing.T) {
	t.Run("[]any", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"items": []any{"a", "b", "c"}})
		var results []string
		err := s.ForEach("items", func(i int, v any) error {
			results = append(results, v.(string))
			return nil
		})
		require.NoError(t, err)
		require.Equal(t, []string{"a", "b", "c"}, results)
	})

	t.Run("[]string", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"items": []string{"x", "y"}})
		var results []string
		err := s.ForEach("items", func(i int, v any) error {
			results = append(results, v.(string))
			return nil
		})
		require.NoError(t, err)
		require.Equal(t, []string{"x", "y"}, results)
	})

	t.Run("[]int", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"nums": []int{1, 2, 3}})
		sum := 0
		err := s.ForEach("nums", func(i int, v any) error {
			sum += v.(int)
			return nil
		})
		require.NoError(t, err)
		require.Equal(t, 6, sum)
	})

	t.Run("[]map[string]any", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{
			"items": []map[string]any{{"id": 1}, {"id": 2}},
		})
		count := 0
		err := s.ForEach("items", func(i int, v any) error {
			count++
			return nil
		})
		require.NoError(t, err)
		require.Equal(t, 2, count)
	})

	t.Run("map[string]any", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{
			"obj": map[string]any{"a": 1, "b": 2},
		})
		count := 0
		err := s.ForEach("obj", func(i int, v any) error {
			count++
			return nil
		})
		require.NoError(t, err)
		require.Equal(t, 2, count)
	})

	t.Run("map[string]string", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{
			"config": map[string]string{"host": "localhost", "port": "8080"},
		})
		count := 0
		err := s.ForEach("config", func(i int, v any) error {
			count++
			return nil
		})
		require.NoError(t, err)
		require.Equal(t, 2, count)
	})

	t.Run("tracks correct indices for []any", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"items": []any{10, 20, 30}})
		var indices []int
		err := s.ForEach("items", func(i int, v any) error {
			indices = append(indices, i)
			return nil
		})
		require.NoError(t, err)
		require.Equal(t, []int{0, 1, 2}, indices)
	})

	t.Run("tracks correct indices for map[string]any", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{
			"obj": map[string]any{"a": 1, "b": 2, "c": 3},
		})
		count := 0
		var indices []int
		err := s.ForEach("obj", func(i int, v any) error {
			indices = append(indices, i)
			count++
			return nil
		})
		require.NoError(t, err)
		require.Equal(t, 3, count)
		require.Equal(t, []int{0, 1, 2}, indices)
	})

	t.Run("error stops iteration", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"items": []any{1, 2, 3}})
		count := 0
		testErr := errors.New("stop")
		err := s.ForEach("items", func(i int, v any) error {
			count++
			if count == 2 {
				return testErr
			}
			return nil
		})
		require.ErrorIs(t, err, testErr)
		require.Equal(t, 2, count)
	})

	t.Run("error stops iteration on []string", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"items": []string{"a", "b", "c"}})
		count := 0
		testErr := errors.New("abort")
		err := s.ForEach("items", func(i int, v any) error {
			count++
			if count == 2 {
				return testErr
			}
			return nil
		})
		require.ErrorIs(t, err, testErr)
		require.Equal(t, 2, count)
	})

	t.Run("error stops iteration on []int", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"nums": []int{10, 20, 30}})
		sum := 0
		testErr := errors.New("calc error")
		err := s.ForEach("nums", func(i int, v any) error {
			sum += v.(int)
			if sum > 25 {
				return testErr
			}
			return nil
		})
		require.ErrorIs(t, err, testErr)
		require.Equal(t, 30, sum)
	})

	t.Run("error stops iteration on map", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{
			"obj": map[string]any{"a": 1, "b": 2, "c": 3},
		})
		count := 0
		testErr := errors.New("map error")
		err := s.ForEach("obj", func(i int, v any) error {
			count++
			if count == 2 {
				return testErr
			}
			return nil
		})
		require.ErrorIs(t, err, testErr)
		require.Equal(t, 2, count)
	})

	t.Run("missing key is no-op", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{})
		err := s.ForEach("missing", func(i int, v any) error {
			t.Fatal("should not be called")
			return nil
		})
		require.NoError(t, err)
	})

	t.Run("unsupported type is no-op", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"key": "string"})
		err := s.ForEach("key", func(i int, v any) error {
			t.Fatal("should not be called")
			return nil
		})
		require.NoError(t, err)
	})

	t.Run("boolean type is no-op", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"flag": true})
		err := s.ForEach("flag", func(i int, v any) error {
			t.Fatal("should not be called")
			return nil
		})
		require.NoError(t, err)
	})

	t.Run("numeric type is no-op", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"count": 42})
		err := s.ForEach("count", func(i int, v any) error {
			t.Fatal("should not be called")
			return nil
		})
		require.NoError(t, err)
	})

	t.Run("empty []any", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"items": []any{}})
		count := 0
		err := s.ForEach("items", func(i int, v any) error {
			count++
			return nil
		})
		require.NoError(t, err)
		require.Equal(t, 0, count)
	})

	t.Run("empty map[string]any", func(t *testing.T) {
		s := vuego.NewStack(map[string]any{"obj": map[string]any{}})
		count := 0
		err := s.ForEach("obj", func(i int, v any) error {
			count++
			return nil
		})
		require.NoError(t, err)
		require.Equal(t, 0, count)
	})
}
