package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v3"

	"github.com/titpetric/vuego"
)

func main() {
	if err := start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func start() error {
	if len(os.Args) < 2 {
		return printUsage()
	}

	command := os.Args[1]

	switch command {
	case "fmt":
		return formatCommand(os.Args[2:])
	case "render":
		return renderCommand(os.Args[2:])
	case "--help", "-h", "help":
		return printHelp()
	default:
		// Backward compatibility: treat first arg as template file
		if len(os.Args) != 3 {
			return printUsage()
		}
		return renderCommand(os.Args[1:])
	}
}

func printUsage() error {
	return fmt.Errorf("Usage: vuego [fmt|render] <args...>\n       vuego <template.tpl> <data.json>")
}

func printHelp() error {
	fmt.Fprintf(os.Stderr, `vuego - Template formatter and renderer

Usage:
  vuego fmt <file.vuego>           Format a vuego template file
  vuego render <file.tpl> <data>   Render a template with data

Examples:
  vuego fmt layout.vuego
  vuego render index.vuego data.json

Commands:
  fmt      Format vuego template files
  render   Render templates with data
  help     Show this help message
`)
	return nil
}

func formatCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("fmt: missing file argument\nUsage: vuego fmt <file.vuego> [file2.vuego ...]")
	}

	formatter := vuego.NewFormatter()
	var lastErr error

	for _, file := range args {
		content, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", file, err)
			lastErr = err
			continue
		}

		formatted, err := formatter.Format(string(content))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting %s: %v\n", file, err)
			lastErr = err
			continue
		}

		// Only write and report if content changed
		if string(content) != formatted {
			if err := os.WriteFile(file, []byte(formatted), 0o644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", file, err)
				lastErr = err
				continue
			}
			fmt.Println(file)
		}
	}

	return lastErr
}

func renderCommand(args []string) error {
	fs := flag.NewFlagSet("render", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: vuego render <file.tpl> <data>\n")
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
	jsonFile := positional[1]

	dataContent, err := os.ReadFile(jsonFile)
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
