package vuego_test

import (
	"bytes"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"
	"github.com/titpetric/vuego"
)

func TestPipe_ComparisonExpression(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{
			Data: []byte(`<div>{{ num > 5 ? "high" : "low" }}</div>`),
		},
	}

	vue := vuego.NewVue(fs)

	tests := map[string]struct {
		num  interface{}
		want string
	}{
		"high": {10, "high"},
		"low":  {3, "low"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			err := vue.Render(&buf, "test.vuego", map[string]any{"num": tt.num})
			require.NoError(t, err)
			require.Contains(t, buf.String(), tt.want)
		})
	}
}

func TestPipe_TernaryExpression(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{
			Data: []byte(`<div>{{ status == "active" ? "Enabled" : "Disabled" }}</div>`),
		},
	}

	vue := vuego.NewVue(fs)

	tests := map[string]struct {
		status string
		want   string
	}{
		"active":   {"active", "Enabled"},
		"inactive": {"inactive", "Disabled"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			err := vue.Render(&buf, "test.vuego", map[string]any{"status": tt.status})
			require.NoError(t, err)
			require.Contains(t, buf.String(), tt.want)
		})
	}
}

func TestPipe_FilterThenFilter(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{
			Data: []byte(`<div>{{ name | upper | title }}</div>`),
		},
	}

	vue := vuego.NewVue(fs).Funcs(vuego.DefaultFuncMap())

	var buf bytes.Buffer
	err := vue.Render(&buf, "test.vuego", map[string]any{"name": "john doe"})
	require.NoError(t, err)
	require.Contains(t, buf.String(), "John Doe")
}
