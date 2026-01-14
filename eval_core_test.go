package vuego_test

import (
	"bytes"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/diff"
)

func TestVue_EvaluateTextNode(t *testing.T) {
	template := []byte("<p>{{ message }}</p>")
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{Data: template},
	}
	vue := vuego.NewVue(fs)

	var buf bytes.Buffer
	err := vue.RenderFragment(&buf, "test.vuego", map[string]any{"message": "hello"})
	require.NoError(t, err)
	require.True(t, diff.EqualHTML(t, []byte("<p>hello</p>"), buf.Bytes(), nil, nil))
}

func TestVue_EvaluateElementNode(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{Data: []byte("<div class=\"container\">content</div>")},
	}
	vue := vuego.NewVue(fs)

	var buf bytes.Buffer
	err := vue.RenderFragment(&buf, "test.vuego", nil)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "container")
	require.Contains(t, buf.String(), "content")
}

func TestVue_EvaluateChildren(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{Data: []byte("<div><p>{{ a }}</p><p>{{ b }}</p></div>")},
	}
	vue := vuego.NewVue(fs)

	data := map[string]any{"a": "first", "b": "second"}
	var buf bytes.Buffer
	err := vue.RenderFragment(&buf, "test.vuego", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "first")
	require.Contains(t, buf.String(), "second")
}

func TestVue_EvaluateTemplateTag(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{Data: []byte("<div><template><span>hidden</span></template></div>")},
	}
	vue := vuego.NewVue(fs)

	var buf bytes.Buffer
	err := vue.RenderFragment(&buf, "test.vuego", nil)
	require.NoError(t, err)
	// template tag should be omitted, but content should render
	require.Contains(t, buf.String(), "hidden")
	require.NotContains(t, buf.String(), "template")
}

func TestVue_EvaluateWithVIf(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{Data: []byte("<div v-if=\"show\">visible</div><div v-if=\"hide\">hidden</div>")},
	}
	vue := vuego.NewVue(fs)

	data := map[string]any{"show": true, "hide": false}
	var buf bytes.Buffer
	err := vue.RenderFragment(&buf, "test.vuego", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "visible")
	require.NotContains(t, buf.String(), "hidden")
}

func TestVue_EvaluateWithVFor(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{Data: []byte("<span v-for=\"item in items\">{{ item }}</span>")},
	}
	vue := vuego.NewVue(fs)

	data := map[string]any{"items": []string{"a", "b", "c"}}
	var buf bytes.Buffer
	err := vue.RenderFragment(&buf, "test.vuego", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "a")
	require.Contains(t, buf.String(), "b")
	require.Contains(t, buf.String(), "c")
}

func TestVue_EvaluateWithVHtml(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{Data: []byte("<div v-html=\"html\"></div>")},
	}
	vue := vuego.NewVue(fs)

	data := map[string]any{"html": "<strong>bold</strong>"}
	var buf bytes.Buffer
	err := vue.RenderFragment(&buf, "test.vuego", data)
	require.NoError(t, err)
	require.Contains(t, buf.String(), "<strong>bold</strong>")
}

func TestVue_EvaluateComplexNesting(t *testing.T) {
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{Data: []byte(`
<div class="outer">
  <div v-for="item in items" class="item">
    <span v-if="item.show">{{ item.name }}</span>
  </div>
</div>
`)},
	}
	vue := vuego.NewVue(fs)

	data := map[string]any{
		"items": []map[string]any{
			{"name": "One", "show": true},
			{"name": "Two", "show": false},
			{"name": "Three", "show": true},
		},
	}

	var buf bytes.Buffer
	err := vue.RenderFragment(&buf, "test.vuego", data)
	require.NoError(t, err)
	output := buf.String()
	require.Contains(t, output, "One")
	require.NotContains(t, output, "Two")
	require.Contains(t, output, "Three")
}

func TestVue_EvaluateCloneNode(t *testing.T) {
	// Verify that nodes are cloned and not mutated
	fs := fstest.MapFS{
		"test.vuego": &fstest.MapFile{Data: []byte("<div>{{ value }}</div>")},
	}
	vue := vuego.NewVue(fs)

	var buf1 bytes.Buffer
	err := vue.RenderFragment(&buf1, "test.vuego", map[string]any{"value": "first"})
	require.NoError(t, err)

	var buf2 bytes.Buffer
	err = vue.RenderFragment(&buf2, "test.vuego", map[string]any{"value": "second"})
	require.NoError(t, err)

	require.Contains(t, buf1.String(), "first")
	require.Contains(t, buf2.String(), "second")
}
