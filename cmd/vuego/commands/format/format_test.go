package format_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	fmtcmd "github.com/titpetric/vuego/cmd/vuego/commands/fmt"
)

func TestRun_NoFiles(t *testing.T) {
	err := fmtcmd.Run([]string{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing file argument")
}

func TestRun_FormatsFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.vuego")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("<div><span>hello</span></div>")
	require.NoError(t, err)
	tmpFile.Close()

	err = fmtcmd.Run([]string{tmpFile.Name()})
	require.NoError(t, err)
}

func TestUsage(t *testing.T) {
	usage := fmtcmd.Usage()
	require.NotEmpty(t, usage)
	require.Contains(t, usage, "vuego fmt")
}
