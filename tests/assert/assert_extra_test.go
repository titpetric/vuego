package assert_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/titpetric/vuego/tests/assert"
)

// assertExtra invokes the assertions in assert_extra.go on the path where
// they hold. TestAssert runs it.
func assertExtra(t *testing.T) {
	var typed *int

	assert.Nil(t, nil)
	assert.Nil(t, typed)
	assert.NotNil(t, &struct{}{})

	assert.Len(t, "abc", 3)
	assert.Len(t, []int{1, 2}, 2)
	assert.Len(t, map[string]int{"a": 1}, 1)

	assert.Empty(t, nil)
	assert.Empty(t, "")
	assert.Empty(t, []int{})
	assert.Empty(t, 0)
	assert.NotEmpty(t, "x")

	assert.Greater(t, 2, 1)
	assert.Greater(t, uint(2), uint(1))
	assert.Greater(t, 2.5, 1.5)
	assert.Greater(t, "b", "a")

	assert.NotContains(t, "haystack", "needle")
	assert.NotContains(t, []string{"a"}, "b")
	assert.NotContains(t, map[string]int{"a": 1}, "b")

	assert.ErrorIs(t, fmt.Errorf("open: %w", os.ErrNotExist), os.ErrNotExist)

	assert.IsIncreasing(t, []int{1, 2, 3})
	assert.IsIncreasing(t, []string{"a", "b"})
}
