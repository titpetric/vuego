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
// If no layout is specified on the first template, defaults to layouts/base.vuego if available.
// Requires that Load() has been called first.
func (t *template) Layout(ctx context.Context, w io.Writer) error {
	if !t.filenameLoaded {
		return fmt.Errorf("no template loaded; call Load() first")
	}

	data := t.stack.EnvMap()
	filename := t.filename
	isFirstTemplate := true

	// Build layout chain and render intermediate templates
	for {
		var buf bytes.Buffer
		tpl := t.Load(filename).Fill(data)
		if err := tpl.Render(ctx, &buf); err != nil {
			return err
		}

		data["content"] = buf.String()
		layout := tpl.Get("layout")

		if layout == "" {
			// No layout specified
			if isFirstTemplate {
				// Default to base layout for first template
				filename = "layouts/base.vuego"
				isFirstTemplate = false
				delete(data, "layout")
				continue
			}
			// No more layouts, render final template
			_, err := io.Copy(w, &buf)
			return err
		}

		// Continue with next layout in chain
		isFirstTemplate = false
		delete(data, "layout")
		filename = "layouts/" + layout + ".vuego"
	}
}
