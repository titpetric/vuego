package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/titpetric/vuego"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s file.tpl file.json\n", os.Args[0])
		os.Exit(1)
	}

	tplFile := os.Args[1]
	jsonFile := os.Args[2]

	tplContent, err := os.ReadFile(tplFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading template file: %v\n", err)
		os.Exit(1)
	}

	jsonContent, err := os.ReadFile(jsonFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading JSON file: %v\n", err)
		os.Exit(1)
	}

	var data map[string]any
	if err := json.Unmarshal(jsonContent, &data); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Render template to stdout
	if err := vuego.Render(os.Stdout, string(tplContent), data); err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering template: %v\n", err)
		os.Exit(1)
	}
}
