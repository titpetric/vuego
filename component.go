package vuego

import (
	"context"
	"io"
)

// Component is a renderable .vuego template component.
type Component interface {
	// Render renders the component template to the given writer.
	Render(ctx context.Context, w io.Writer, filename string, data any) error
}
