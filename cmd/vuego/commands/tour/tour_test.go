package tour_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego/cmd/vuego/commands/tour"
)

func TestUsage(t *testing.T) {
	usage := tour.Usage()
	require.NotEmpty(t, usage)
	require.Contains(t, usage, "vuego tour")
	require.Contains(t, usage, "-addr")
}

func TestRun_InvalidFlag(t *testing.T) {
	err := tour.Run([]string{"--invalid-flag"})
	require.Error(t, err)
}
