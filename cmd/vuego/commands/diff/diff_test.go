package diff_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego/cmd/vuego/commands/diff"
)

func TestRun_WrongNumberOfArguments(t *testing.T) {
	err := diff.Run([]string{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "requires exactly 2 file arguments")

	err = diff.Run([]string{"file1.html"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "requires exactly 2 file arguments")

	err = diff.Run([]string{"a", "b", "c"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "requires exactly 2 file arguments")
}

func TestUsage(t *testing.T) {
	usage := diff.Usage()
	require.NotEmpty(t, usage)
	require.Contains(t, usage, "vuego diff")
}
