package render_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego/cmd/vuego/commands/render"
)

func TestRun_WrongNumberOfArguments(t *testing.T) {
	err := render.Run([]string{})
	require.Error(t, err)
	require.Equal(t, "render: requires exactly 2 arguments", err.Error())

	err = render.Run([]string{"file.vuego"})
	require.Error(t, err)
	require.Equal(t, "render: requires exactly 2 arguments", err.Error())

	err = render.Run([]string{"a", "b", "c"})
	require.Error(t, err)
	require.Equal(t, "render: requires exactly 2 arguments", err.Error())
}

func TestUsage(t *testing.T) {
	usage := render.Usage()
	require.NotEmpty(t, usage)
	require.Contains(t, usage, "vuego render")
}
