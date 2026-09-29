# Node Post-Processing with NodeProcessor

The `NodeProcessor` interface enables custom post-processing of rendered HTML nodes. This allows you to extend Vue with custom tags and attributes that transform the final DOM tree.

## Overview

Once template evaluation ends, Vuego calls each registered `NodeProcessor` with the rendered DOM. By then it has handled every Vue directive and interpolation, which lets a processor:

- Add custom tags such as `<script type="text/css+less">`
- Transform HTML nodes based on custom attributes
- Compile embedded content (CSS, JavaScript, etc.)
- Change node structure or attributes in place

## The NodeProcessor Interface

```go
type NodeProcessor interface {
	// Process receives the rendered nodes and may modify them in place.
	// It should return an error if processing fails.
	Process(nodes []*html.Node) error
}
```

## Registering a Processor

```go
v := vuego.NewVue(templateFS)
v.RegisterNodeProcessor(myProcessor)
```

Each processor runs in registration order once template evaluation ends.

## Example: LESS CSS Compiler

This example implements a processor that compiles LESS CSS embedded in `<script type="text/css+less">` tags:

```go
import (
	"bytes"
	"github.com/titpetric/lessgo/dst"
	"github.com/titpetric/lessgo/renderer"
	"golang.org/x/net/html"
)

type LessProcessor struct {
	fs fs.FS // Optional filesystem for @import statements
}

func (lp *LessProcessor) Process(nodes []*html.Node) error {
	for _, node := range nodes {
		if err := lp.processNode(node); err != nil {
			return err
		}
	}
	return nil
}

func (lp *LessProcessor) processNode(node *html.Node) error {
	if node == nil {
		return nil
	}

	// Check for script tag with type="text/css+less"
	if node.Type == html.ElementNode && node.Data == "script" {
		if lp.isLessScriptTag(node) {
			if err := lp.compileLessTag(node); err != nil {
				return err
			}
		}
	}

	// Recursively process children
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if err := lp.processNode(c); err != nil {
			return err
		}
	}

	return nil
}

func (lp *LessProcessor) isLessScriptTag(node *html.Node) bool {
	for _, attr := range node.Attr {
		if attr.Key == "type" && attr.Val == "text/css+less" {
			return true
		}
	}
	return false
}

func (lp *LessProcessor) compileLessTag(scriptNode *html.Node) error {
	// Extract LESS content
	lessContent := ""
	for c := scriptNode.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			lessContent += c.Data
		}
	}

	if lessContent == "" {
		return nil
	}

	// Parse and compile
	parser := dst.NewParser(bytes.NewReader([]byte(lessContent)))
	file, err := parser.Parse()
	if err != nil {
		return fmt.Errorf("failed to parse LESS: %w", err)
	}

	r := renderer.NewRenderer()
	css, err := r.Render(file)
	if err != nil {
		return fmt.Errorf("failed to render LESS to CSS: %w", err)
	}

	// Replace script tag with style tag
	scriptNode.Data = "style"
	scriptNode.Attr = filterAttrs(scriptNode.Attr, "type")
	scriptNode.FirstChild = &html.Node{
		Type: html.TextNode,
		Data: css,
	}
	scriptNode.LastChild = scriptNode.FirstChild

	return nil
}

func filterAttrs(attrs []html.Attribute, excludeKey string) []html.Attribute {
	var result []html.Attribute
	for _, attr := range attrs {
		if attr.Key != excludeKey {
			result = append(result, attr)
		}
	}
	return result
}
```

### Usage

```go
import (
    "os"
    "github.com/titpetric/vuego"
)

templateFS := os.DirFS("templates")
v := vuego.NewVue(templateFS)
v.RegisterNodeProcessor(&LessProcessor{})

// Now templates can use:
// <script type="text/css+less">
//   @primary-color: #ff0000;
//   .button { color: @primary-color; }
// </script>
```

## Processing Order

- Template evaluation (v-if, v-for, interpolation, etc.)
- Custom node processors (in registration order)
- HTML rendering

A processor may assume Vuego already evaluated every Vue directive.

## Best Practices

1. **Recursive Processing**: Always recursively process child nodes
2. **Error Handling**: Return meaningful errors that include context
3. **In-place editing**: change nodes where they sit rather than rebuilding them
4. **Document Node Types**: Check `node.Type == html.ElementNode` before accessing attributes
5. **Keep the structure**: where you can, leave the DOM shape alone and change the content

## Testing NodeProcessors

Create a test template in `testdata/` and render it with your processor:

```go
func TestMyProcessor(t *testing.T) {
	templateFS := os.DirFS("testdata/myprocessor")
	v := vuego.NewVue(templateFS)
	v.RegisterNodeProcessor(NewMyProcessor())

	var buf bytes.Buffer
	err := v.RenderTemplate(&buf, "template.html", map[string]any{})
	require.NoError(t, err)

	output := buf.String()
	// Assert on output...
}
```
