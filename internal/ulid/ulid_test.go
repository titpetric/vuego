package ulid_test

import (
	"testing"

	"github.com/titpetric/vuego/internal/ulid"
	"github.com/titpetric/vuego/tests/assert"
)

func TestULID(t *testing.T) {
	t.Run("returns a non-zero value", func(t *testing.T) {
		id := ulid.ULID()
		assert.NotEqual(t, "00000000000000000000000000", id.String())
	})

	t.Run("successive values differ", func(t *testing.T) {
		seen := make(map[string]bool, 100)
		for i := 0; i < 100; i++ {
			s := ulid.ULID().String()
			assert.False(t, seen[s], "duplicate ULID %s at iteration %d", s, i)
			seen[s] = true
		}
	})
}

func TestString(t *testing.T) {
	t.Run("is 26 characters of Crockford base32", func(t *testing.T) {
		s := ulid.String()
		assert.Len(t, s, 26)
		for _, r := range s {
			assert.True(t, r >= '0' && r <= '9' || r >= 'A' && r <= 'Z', "unexpected character %q in %s", r, s)
		}
	})

	t.Run("round-trips through Parse", func(t *testing.T) {
		s := ulid.String()
		id, err := ulid.Parse(s)
		assert.NoError(t, err)
		assert.Equal(t, s, id.String())
	})
}

func TestParse(t *testing.T) {
	t.Run("parses a value it produced", func(t *testing.T) {
		want := ulid.ULID()
		got, err := ulid.Parse(want.String())
		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("rejects a short string", func(t *testing.T) {
		_, err := ulid.Parse("01ARZ3NDEKTSV4RRFFQ69G5F")
		assert.Error(t, err)
	})

	t.Run("rejects an empty string", func(t *testing.T) {
		_, err := ulid.Parse("")
		assert.Error(t, err)
	})

	t.Run("rejects a timestamp that overflows 128 bits", func(t *testing.T) {
		_, err := ulid.Parse("8ZZZZZZZZZZZZZZZZZZZZZZZZZ")
		assert.Error(t, err)
	})

	t.Run("accepts characters outside the alphabet", func(t *testing.T) {
		// oklog/ulid.Parse decodes without the strict alphabet check, so
		// only the length and the 128-bit bound are enforced here.
		_, err := ulid.Parse("01ARZ3NDEKTSV4RRFFQ69G5FA!")
		assert.NoError(t, err)
	})
}

func TestValid(t *testing.T) {
	tests := []struct {
		name  string
		in    string
		valid bool
	}{
		{name: "generated value", in: ulid.String(), valid: true},
		{name: "known good value", in: "01ARZ3NDEKTSV4RRFFQ69G5FAV", valid: true},
		{name: "empty string", in: "", valid: false},
		{name: "too short", in: "01ARZ3NDEKTSV4RRFFQ69G5F", valid: false},
		{name: "too long", in: "01ARZ3NDEKTSV4RRFFQ69G5FAVX", valid: false},
		{name: "overflows 128 bits", in: "8ZZZZZZZZZZZZZZZZZZZZZZZZZ", valid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.valid, ulid.Valid(test.in))
		})
	}
}
