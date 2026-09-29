package vuego

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/titpetric/vuego/internal/helpers"
)

// LoadOption is a functional option for configuring Load().
type LoadOption func(*Vue)

// WithFS returns a LoadOption that sets the filesystem for template loading.
func WithFS(templateFS fs.FS) LoadOption {
	return func(vue *Vue) {
		vue.templateFS = templateFS
		vue.loader = NewLoader(templateFS)
	}
}

// WithFuncs returns a LoadOption that merges custom template functions into the existing funcmap.
func WithFuncs(funcMap FuncMap) LoadOption {
	return func(vue *Vue) {
		vue.Funcs(funcMap)
	}
}

// WithLessProcessor returns a LoadOption that registers a LESS processor.
func WithLessProcessor() LoadOption {
	return func(vue *Vue) {
		vue.RegisterNodeProcessor(NewLessProcessor(vue.templateFS))
	}
}

// WithProcessor returns a LoadOption that registers a custom node processor.
func WithProcessor(processor NodeProcessor) LoadOption {
	return func(vue *Vue) {
		vue.RegisterNodeProcessor(processor)
	}
}

// WithComponents returns a LoadOption that registers every component shorthand.
// It walks the components folder and its subfolders for .vuego files, turning
// each path into a kebab-case tag name from the directory and the filename.
func WithComponents() LoadOption {
	return func(vue *Vue) {
		// Nothing to walk without a filesystem; New() builds a Vue with none.
		if vue.templateFS == nil {
			return
		}

		// Walk the components directory recursively
		err := fs.WalkDir(vue.templateFS, "components", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				// Skip directories that don't exist or can't be accessed
				return nil
			}

			// Skip directories, only process files
			if d.IsDir() {
				return nil
			}

			// Only process .vuego files
			if !strings.HasSuffix(d.Name(), ".vuego") {
				return nil
			}

			// Extract the relative path from components folder
			relPath := strings.TrimPrefix(path, "components/")
			relPath = strings.TrimSuffix(relPath, ".vuego")

			// Convert to kebab-case for the tag name
			// Split by path separators and convert each part
			parts := strings.Split(relPath, string(filepath.Separator))
			for i, part := range parts {
				parts[i] = helpers.CamelToKebab(part)
			}
			tagName := strings.Join(parts, "-")

			// Store the mapping
			vue.RegisterComponent(tagName, path)

			return nil
		})

		// Silently ignore if components directory doesn't exist
		_ = err
	}
}
