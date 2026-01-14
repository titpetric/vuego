package render

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v3"

	"github.com/titpetric/vuego"
)

// Run executes the render command with the given arguments.
// Supports the following invocations:
//
//   - vuego render file.vuego (loads file.yaml or file.yml if it exists)
//   - vuego render file.vuego data.yaml (uses explicit data file)
//   - vuego render file.vuego data.json (uses explicit JSON data file).
func Run(args []string) error {
	fs := flag.NewFlagSet("render", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: vuego render <file.vuego> [data.yaml|data.json]\n")
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	positional := fs.Args()
	if len(positional) < 1 || len(positional) > 2 {
		fs.Usage()
		return fmt.Errorf("render: requires 1 or 2 arguments")
	}

	tplFile := positional[0]
	var dataFile string

	// Determine data file
	if len(positional) == 2 {
		// Explicitly provided data file
		dataFile = positional[1]
	} else {
		// Auto-discover data file from template name
		dataFile = findDataFile(tplFile)
	}

	// Load data
	var data map[string]any
	if dataFile != "" {
		dataContent, err := os.ReadFile(dataFile)
		if err != nil {
			return fmt.Errorf("reading data file: %w", err)
		}

		if err := yaml.Unmarshal(dataContent, &data); err != nil {
			return fmt.Errorf("parsing data file: %w", err)
		}
	}

	// Initialize with empty map if no data
	if data == nil {
		data = make(map[string]any)
	}

	// Open template file for efficient streaming
	tplReader, err := os.Open(tplFile)
	if err != nil {
		return fmt.Errorf("opening template file: %w", err)
	}
	defer tplReader.Close()

	// Load template with data and render directly from file reader
	templateFS := os.DirFS(filepath.Dir(tplFile))
	tmpl := vuego.NewFS(templateFS).Fill(data)

	if err := tmpl.RenderReader(context.Background(), os.Stdout, tplReader); err != nil {
		return fmt.Errorf("rendering template: %w", err)
	}

	return nil
}

// findDataFile looks for a data file with the same base name as the template.
// Returns the path if found (checking .yaml first, then .yml, then .json), or empty string if not found.
func findDataFile(tplFile string) string {
	base := tplFile[:len(tplFile)-len(filepath.Ext(tplFile))]
	dir := filepath.Dir(tplFile)

	// Check for .yaml
	candidates := []string{
		filepath.Join(dir, filepath.Base(base)+".yaml"),
		filepath.Join(dir, filepath.Base(base)+".yml"),
		filepath.Join(dir, filepath.Base(base)+".json"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}

// Usage returns the usage string for the render command.
func Usage() string {
	return `vuego render <file.vuego> [data.yaml|data.json]

Render a vuego template with optional data.

If no data file is specified, the command looks for a file with the same
base name as the template (e.g., index.vuego → index.yaml, index.yml, or index.json).
If no data file is found, the template is rendered with empty data.

Examples:
  vuego render index.vuego                  (auto-loads index.yaml if it exists)
  vuego render index.vuego data.json        (explicit data file)
  vuego render index.vuego data.yml         (explicit data file)`
}
