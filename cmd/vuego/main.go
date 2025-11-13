package main

import (
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

	templateFS := os.DirFS(filepath.Dir(tplFile))
	vue := vuego.NewVue(templateFS)

	if err := vue.Render(os.Stdout, filepath.Base(tplFile), data); err != nil {
		return fmt.Errorf("rendering template: %w", err)
	}

	return nil
}
