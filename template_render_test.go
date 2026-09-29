package vuego_test

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/tests/assert"
)

func TestTemplate_Render(t *testing.T) {
	fixtures := os.DirFS("testdata/fixtures")

	t.Run("renders a loaded template", func(t *testing.T) {
		tpl := vuego.NewFS(fixtures)
		tpl.Fill(map[string]any{"title": "Hello", "user": map[string]any{"name": "Alice"}})

		var buf bytes.Buffer
		assert.NoError(t, tpl.Load("interpolation-basic.vuego").Render(t.Context(), &buf))
		assert.Contains(t, buf.String(), "Hello")
		assert.Contains(t, buf.String(), "Hi Alice!")
	})

	t.Run("front matter values reach the output", func(t *testing.T) {
		tpl := vuego.NewFS(fixtures)

		var buf bytes.Buffer
		assert.NoError(t, tpl.Load("frontmatter-basic.vuego").Render(t.Context(), &buf))
		assert.Contains(t, buf.String(), "From Front-Matter")
	})

	t.Run("rendering without Load is an error", func(t *testing.T) {
		var buf bytes.Buffer
		err := vuego.NewFS(fixtures).Render(t.Context(), &buf)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no template loaded")
	})

	t.Run("a load error surfaces on render", func(t *testing.T) {
		var buf bytes.Buffer
		err := vuego.NewFS(fixtures).Load("missing.vuego").Render(t.Context(), &buf)
		assert.Error(t, err)
	})

	t.Run("a cancelled context is reported before any work", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		tpl := vuego.NewFS(fixtures)
		tpl.Assign("title", "Hello")

		var buf bytes.Buffer
		assert.Equal(t, context.Canceled, tpl.Load("interpolation-basic.vuego").Render(ctx, &buf))
		assert.Equal(t, 0, buf.Len())
	})

	t.Run("a front matter layout wraps the template", func(t *testing.T) {
		tpl := vuego.NewFS(fstest.MapFS{
			"layouts/base.vuego": {Data: []byte(`<main><template v-html="content"></template></main>`)},
			"page.vuego":         {Data: []byte("---\nlayout: layouts/base.vuego\n---\n<p>body</p>")},
		})

		var buf bytes.Buffer
		assert.NoError(t, tpl.Load("page.vuego").Render(t.Context(), &buf))
		assert.Contains(t, buf.String(), "<main>")
		assert.Contains(t, buf.String(), "<p>body</p>")
	})
}

func TestTemplate_RenderFile(t *testing.T) {
	fixtures := os.DirFS("testdata/fixtures")

	t.Run("loads and renders in one call", func(t *testing.T) {
		tpl := vuego.NewFS(fixtures)
		tpl.Assign("title", "Hello")

		var buf bytes.Buffer
		assert.NoError(t, tpl.RenderFile(t.Context(), &buf, "interpolation-basic.vuego"))
		assert.Contains(t, buf.String(), "Hello")
	})

	t.Run("reports a missing file", func(t *testing.T) {
		var buf bytes.Buffer
		assert.Error(t, vuego.NewFS(fixtures).RenderFile(t.Context(), &buf, "missing.vuego"))
	})

	t.Run("leaves the receiver loadable again", func(t *testing.T) {
		tpl := vuego.NewFS(fixtures)
		tpl.Assign("title", "Hello")

		var first, second bytes.Buffer
		assert.NoError(t, tpl.RenderFile(t.Context(), &first, "interpolation-basic.vuego"))
		assert.NoError(t, tpl.RenderFile(t.Context(), &second, "interpolation-basic.vuego"))
		assert.Equal(t, first.String(), second.String())
	})
}

func TestTemplate_RenderString(t *testing.T) {
	t.Run("renders source given inline", func(t *testing.T) {
		tpl := vuego.NewFS(os.DirFS("testdata/fixtures"))
		tpl.Assign("message", "Hello from String")

		var buf bytes.Buffer
		assert.NoError(t, tpl.RenderString(t.Context(), &buf, "<p>{{ message }}</p>"))
		assert.Contains(t, buf.String(), "Hello from String")
	})

	t.Run("works without a filesystem", func(t *testing.T) {
		tpl := vuego.New()
		tpl.Assign("message", "Hello from String")

		var buf bytes.Buffer
		assert.NoError(t, tpl.RenderString(t.Context(), &buf, "<p>{{ message }}</p>"))
		assert.Contains(t, buf.String(), "Hello from String")
	})

	t.Run("empty source writes nothing", func(t *testing.T) {
		var buf bytes.Buffer
		assert.NoError(t, vuego.New().RenderString(t.Context(), &buf, ""))
		assert.Equal(t, "", buf.String())
	})

	t.Run("a cancelled context is reported", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		var buf bytes.Buffer
		assert.Equal(t, context.Canceled, vuego.New().RenderString(ctx, &buf, "<p>x</p>"))
	})
}

func TestTemplate_RenderByte(t *testing.T) {
	t.Run("renders source given as bytes", func(t *testing.T) {
		tpl := vuego.New()
		tpl.Assign("message", "Hello from Bytes")

		var buf bytes.Buffer
		assert.NoError(t, tpl.RenderByte(t.Context(), &buf, []byte("<p>{{ message }}</p>")))
		assert.Contains(t, buf.String(), "Hello from Bytes")
	})

	t.Run("nil source writes nothing", func(t *testing.T) {
		var buf bytes.Buffer
		assert.NoError(t, vuego.New().RenderByte(t.Context(), &buf, nil))
		assert.Equal(t, "", buf.String())
	})

	t.Run("matches RenderString for the same source", func(t *testing.T) {
		const source = `<p>{{ message }}</p>`

		var fromBytes, fromString bytes.Buffer
		assert.NoError(t, vuego.New().Assign("message", "x").RenderByte(t.Context(), &fromBytes, []byte(source)))
		assert.NoError(t, vuego.New().Assign("message", "x").RenderString(t.Context(), &fromString, source))
		assert.Equal(t, fromString.String(), fromBytes.String())
	})
}

func TestTemplate_RenderReader(t *testing.T) {
	t.Run("renders source from a reader", func(t *testing.T) {
		tpl := vuego.New()
		tpl.Assign("message", "Hello from Reader")

		var buf bytes.Buffer
		assert.NoError(t, tpl.RenderReader(t.Context(), &buf, strings.NewReader("<p>{{ message }}</p>")))
		assert.Contains(t, buf.String(), "Hello from Reader")
	})

	t.Run("renders source from a file", func(t *testing.T) {
		path := writeTempTemplate(t, "<p>{{ message }}</p>")

		file, err := os.Open(path)
		assert.NoError(t, err)
		defer file.Close()

		tpl := vuego.New()
		tpl.Assign("message", "Hello from File Reader")

		var buf bytes.Buffer
		assert.NoError(t, tpl.RenderReader(t.Context(), &buf, file))
		assert.Contains(t, buf.String(), "Hello from File Reader")
	})

	t.Run("a cancelled context leaves the writer untouched", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		var buf bytes.Buffer
		assert.Equal(t, context.Canceled, vuego.New().RenderReader(ctx, &buf, strings.NewReader("<p>x</p>")))
		assert.Equal(t, 0, buf.Len())
	})

	t.Run("a render error leaves the writer untouched", func(t *testing.T) {
		tpl := vuego.New(vuego.WithProcessor(&errorProcessor{failOn: "post", err: errWriteFailed}))

		var buf bytes.Buffer
		assert.ErrorIs(t, tpl.RenderReader(t.Context(), &buf, strings.NewReader("<p>x</p>")), errWriteFailed)
		assert.Equal(t, 0, buf.Len())
	})

	t.Run("a writer error is reported", func(t *testing.T) {
		assert.ErrorIs(t,
			vuego.New().RenderReader(t.Context(), failingWriter{}, strings.NewReader("<p>x</p>")),
			errWriteFailed,
		)
	})
}

func TestNew_WithoutFilesystem(t *testing.T) {
	t.Run("Load reports the missing filesystem", func(t *testing.T) {
		tpl := vuego.New()
		tpl.Assign("message", "Hello")

		var buf bytes.Buffer
		err := tpl.Load("interpolation-basic.vuego").Render(t.Context(), &buf)
		assert.Equal(t, "error reading interpolation-basic.vuego: no filesystem configured", err.Error())
	})
}

// failingWriter fails every write, so a render error surfaces.
type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errWriteFailed
}

// writeTempTemplate writes content to a temporary file and returns its path.
func writeTempTemplate(t *testing.T, content string) string {
	t.Helper()

	file, err := os.CreateTemp(t.TempDir(), "template_*.vuego")
	assert.NoError(t, err)

	_, err = file.WriteString(content)
	assert.NoError(t, err)
	assert.NoError(t, file.Close())

	return file.Name()
}
