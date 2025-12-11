package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/titpetric/vuego"
)

func main() {
	if err := start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func start() error {
	if len(os.Args) != 3 {
		return fmt.Errorf("Usage: %s file.tpl file.json", os.Args[0])
	}

	tplFile := os.Args[1]
	jsonFile := os.Args[2]

	jsonContent, err := os.ReadFile(jsonFile)
	if err != nil {
		return fmt.Errorf("reading JSON file: %w", err)
	}

	var data map[string]any
	if err := json.Unmarshal(jsonContent, &data); err != nil {
		return fmt.Errorf("parsing JSON: %w", err)
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
