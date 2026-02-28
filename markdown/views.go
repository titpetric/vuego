package markdown

import (
	"embed"
	"io/fs"
)

//go:embed all:markdown
var views embed.FS

// Templates returns the embedded markdown template filesystem.
// Templates are nested under the "markdown/" directory, matching
// the convention used by layouts/, components/, etc. in OverlayFS.
func Templates() fs.FS {
	return views
}
