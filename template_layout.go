package vuego

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"

	"github.com/titpetric/vuego/internal/parser"
)

// extractSlotsFromDOM extracts named slot definitions from a DOM tree before rendering.
// Looks for <template #slotname> or <template v-slot:slotname> elements and returns
// a SlotScope with their content ready for use.
// This is called during template parsing to preserve slot structure before rendering.
func extractSlotsFromDOM(nodes []*html.Node) *SlotScope {
	slotScope := NewSlotScope()

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "template" {
			// Check for named slot markers
			slotName := ""
			for _, attr := range n.Attr {
				// Handle #slotname shorthand
				if len(attr.Key) > 0 && attr.Key[0] == '#' {
					slotName = attr.Key[1:]
					break
				}
				// Handle v-slot:slotname
				if strings.HasPrefix(attr.Key, "v-slot:") {
					slotName = strings.TrimPrefix(attr.Key, "v-slot:")
					break
				}
			}

			if slotName != "" {
				// Collect the children of the template node as slot content
				var childNodes []*html.Node
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					childNodes = append(childNodes, c)
				}

				slotScope.SetSlot(slotName, &SlotContent{
					Nodes:        childNodes,
					TemplateNode: n,
				})
			}
		}

		// Traverse children
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	for _, node := range nodes {
		walk(node)
	}

	return slotScope
}

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
	var inheritedSlotScope *SlotScope // Slots defined in child templates (as DOM nodes)

	// Build layout chain and render intermediate templates
	for {
		if depth >= maxDepth {
			return fmt.Errorf("layout chain depth exceeded maximum of %d, possible circular dependency", maxDepth)
		}
		depth++

		// Create a fresh buffer for each iteration
		buf := new(bytes.Buffer)
		tplInterface := t.Load(filename).Fill(data)
		tpl := tplInterface.(*template)

		// Extract slot definitions from the template DOM before rendering (only for first template)
		if isFirstTemplate {
			// Parse the template bytes to get DOM nodes
			templateNodes, err := parser.ParseTemplateBytes(tpl.templateBytes)
			if err == nil {
				inheritedSlotScope = extractSlotsFromDOM(templateNodes)
			}
		}

		if err := tpl.renderWithoutLayout(ctx, buf); err != nil {
			return err
		}

		contentHTML := buf.String()

		data["content"] = contentHTML
		// Pass inherited slots to the layout via the SlotScope so they can be used by <slot> elements
		if inheritedSlotScope != nil && len(inheritedSlotScope.Slots) > 0 {
			data["__slotScope__"] = inheritedSlotScope
		}

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
			_, err := io.Copy(w, buf)
			return err
		}

		// Continue with next layout in chain
		isFirstTemplate = false
		delete(data, "layout")
		filename = t.resolveLayoutPath(layout, filename)
	}
}
