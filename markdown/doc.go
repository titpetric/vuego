// Package markdown parses GFM-flavoured markdown and renders each element
// through vuego templates. Default templates are embedded and can be overridden
// via [vuego.OverlayFS]. Post-processors allow per-block-type HTML transformation.
//
// # Usage
//
//	contentFS := os.DirFS("content")
//	md := markdown.New(contentFS)
//
//	doc, err := md.Load("post.md")
//	if err != nil {
//		log.Fatal(err)
//	}
//	doc.Render(os.Stdout)
//
// # Post-processing
//
// Post-processors transform rendered HTML for a specific block type.
// The processor receives the rendered HTML and returns the transformed result:
//
//	md.PostProcess("paragraph", func(src string) string {
//		return strings.ReplaceAll(src, "foo", "bar")
//	})
//
// # Template customization
//
// Default templates are embedded via [Templates] and can be overridden by
// placing custom .vuego files in the content filesystem. The content filesystem
// is overlaid on top of the embedded defaults using [vuego.OverlayFS], so any
// matching filename in the content filesystem takes precedence:
//
//	// contentFS/paragraph.vuego overrides the default paragraph template
//	md := markdown.New(os.DirFS("my-theme"))
package markdown
