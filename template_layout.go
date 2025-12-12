package vuego

import (
	"bytes"
	"context"
	"fmt"
	"io"
)

// Layout loads a template, and if the template contains "layout" in the metadata, it will
// load another template from layouts/%s.vuego; Layouts can be chained so one layout can
// again trigger another layout, like `blog.vuego -> layouts/post.vuego -> layouts/base.vuego`.
// Requires that Load() has been called first.
func (t *template) Layout(ctx context.Context, w io.Writer) error {
	if !t.filenameLoaded {
		return fmt.Errorf("no template loaded; call Load() first")
	}

	data := t.stack.EnvMap()
	filename := t.filename

	for {
		var buf bytes.Buffer
		tpl := t.Load(filename).Fill(data)
		if err := tpl.Render(ctx, &buf); err != nil {
			return err
		}

		content := buf.String()
		data["content"] = content

		layout := tpl.Get("layout")
		if layout == "" {
			filename = "layouts/base.vuego"
			break
		}

		delete(data, "layout")
		filename = "layouts/" + layout + ".vuego"
		continue
	}

	var buf bytes.Buffer
	tpl := t.Load(filename).Fill(data)
	if err := tpl.Render(ctx, &buf); err != nil {
		return err
	}
	_, err := io.Copy(w, &buf)
	return err
}
