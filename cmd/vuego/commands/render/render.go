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
func Run(args []string) error {
	fs := flag.NewFlagSet("render", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: vuego render <file.vuego> <data.json/yml>\n")
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	positional := fs.Args()
	if len(positional) != 2 {
		fs.Usage()
		return fmt.Errorf("render: requires exactly 2 arguments")
	}

	tplFile := positional[0]
	dataFile := positional[1]

	dataContent, err := os.ReadFile(dataFile)
	if err != nil {
		return fmt.Errorf("reading data file: %w", err)
	}

	var data map[string]any
	if err := yaml.Unmarshal(dataContent, &data); err != nil {
		return fmt.Errorf("parsing data file: %w", err)
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

// Usage returns the usage string for the render command.
func Usage() string {
	return `vuego render <file.vuego> <data.json/yml>

Render a vuego template with data.

Examples:
  vuego render index.vuego data.json
  vuego render index.vuego data.yml`
}
