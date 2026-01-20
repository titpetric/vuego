package vuego

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// resolveLayoutPath resolves a layout name to a file path.
// It first checks if the layout exists relative to the current template's directory,
// then falls back to the layouts/ directory.
// If the layout name already includes a .vuego extension, it's used as-is for relative resolution.
func (t *template) resolveLayoutPath(layout, currentFile string) string {
	currentDir := filepath.Dir(currentFile)

	// Check if layout has .vuego extension (explicit relative path)
	if strings.HasSuffix(layout, ".vuego") {
		relativePath := filepath.Join(currentDir, layout)
		if t.vue.loader.Stat(relativePath) == nil {
			return relativePath
		}
	}

	// Try relative path without extension
	relativePath := filepath.Join(currentDir, layout+".vuego")
	if t.vue.loader.Stat(relativePath) == nil {
		return relativePath
	}

	// Fall back to layouts/ directory
	return "layouts/" + layout + ".vuego"
}

// layout loads a template, and if the template contains "layout" in the metadata, it will
// load another template from layouts/%s.vuego; Layouts can be chained so one layout can
// again trigger another layout, like `blog.vuego -> layouts/post.vuego -> layouts/base.vuego`.
// If no layout is specified on the first template, defaults to layouts/base.vuego if available.
// Layout paths are resolved relative to the current template first, then fall back to layouts/.
// Requires that Load() has been called first.
func (t *template) layout(ctx context.Context, w io.Writer) error {
	if !t.filenameLoaded {
		return fmt.Errorf("no template loaded; call Load() first")
	}

	data := t.stack.EnvMap()
	filename := t.filename
	isFirstTemplate := true
	maxDepth := 100
	depth := 0

	// Build layout chain and render intermediate templates
	for {
		if depth >= maxDepth {
			return fmt.Errorf("layout chain depth exceeded maximum of %d, possible circular dependency", maxDepth)
		}
		depth++

		var buf bytes.Buffer
		tplInterface := t.Load(filename).Fill(data)
		tpl := tplInterface.(*template)
		if err := tpl.renderWithoutLayout(ctx, &buf); err != nil {
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
		filename = t.resolveLayoutPath(layout, filename)
	}
}
