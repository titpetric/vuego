package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	yaml "gopkg.in/yaml.v3"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/diff"
	"github.com/titpetric/vuego/formatter"
	"github.com/titpetric/vuego/internal/helpers"
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
	case "diff":
		return diffCommand(os.Args[2:])
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
	return fmt.Errorf("Usage: vuego [fmt|render] <args...>\n       vuego <template.vuego> <data.json/yml>")
}

func printHelp() error {
	fmt.Fprintf(os.Stderr, `vuego - Template formatter, renderer, and DOM comparator

Usage:
  vuego fmt <file.vuego>                    Format a vuego template file
  vuego render <file.vuego> <data>          Render a template with data
  vuego diff [--output FORMAT] <file1> <file2>  Compare two HTML/vuego files (DOM-aware)

Examples:
  vuego fmt layout.vuego
  vuego render index.vuego data.json
  vuego render index.vuego data.yml
  vuego diff before.html after.html
  vuego diff --output yaml file1.html file2.html | dyff between - -
  vuego diff --output simple file1.html file2.html | colordiff

Commands:
  fmt      Format vuego template files
  render   Render templates with data (supports JSON and YAML)
  diff     Compare two HTML/vuego files using DOM comparison (ignores formatting)
           Formats: simple, unified (default), yaml
           Default 'unified' format can be piped to colordiff
           'yaml' format can be used with dyff: vuego diff --output yaml a b | dyff between - -
  help     Show this help message
`)
	return nil
}

func formatCommand(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("fmt: missing file argument\nUsage: vuego fmt <file.vuego> [file2.vuego ...]")
	}

	f := formatter.NewFormatter()
	var lastErr error

	for _, file := range args {
		content, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", file, err)
			lastErr = err
			continue
		}

		formatted, err := f.Format(string(content))
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

func diffCommand(args []string) error {
	fs := flag.NewFlagSet("diff", flag.ContinueOnError)
	outputFormat := fs.String("output", "unified", "output format: simple, unified, yaml")
	fs.String("format", "unified", "deprecated: use --output instead")

	if err := fs.Parse(args); err != nil {
		return err
	}

	positional := fs.Args()
	if len(positional) != 2 {
		return fmt.Errorf("diff: requires exactly 2 file arguments\nUsage: vuego diff [--output FORMAT] <file1> <file2>\n\nFormats: simple, unified, yaml")
	}

	file1 := positional[0]
	file2 := positional[1]

	// Read both files
	content1, err := os.ReadFile(file1)
	if err != nil {
		return fmt.Errorf("reading %s: %w", file1, err)
	}

	content2, err := os.ReadFile(file2)
	if err != nil {
		return fmt.Errorf("reading %s: %w", file2, err)
	}

	// Compare using DOM-aware comparison
	isEqual := helpers.CompareHTML(content1, content2)

	switch *outputFormat {
	case "simple":
		if isEqual {
			fmt.Printf("✓ %s and %s have equivalent DOM trees\n", file1, file2)
			return nil
		}
		fmt.Printf("✗ %s and %s have different DOM trees\n", file1, file2)
		fmt.Printf("\nExpected (%s):\n%s\n", file1, content1)
		fmt.Printf("\nActual (%s):\n%s\n", file2, content2)
		return fmt.Errorf("DOM trees do not match")

	case "unified":
		// Generate unified diff format (can be piped to colordiff)
		formatted1, err := diff.FormatToNormalizedHTML(content1)
		if err != nil {
			return fmt.Errorf("formatting %s: %w", file1, err)
		}
		formatted2, err := diff.FormatToNormalizedHTML(content2)
		if err != nil {
			return fmt.Errorf("formatting %s: %w", file2, err)
		}

		unifiedDiff := diff.GenerateUnifiedDiff(file1, file2, formatted1, formatted2)
		fmt.Print(unifiedDiff)

		if !isEqual {
			return fmt.Errorf("DOM trees do not match")
		}
		return nil

	case "yaml":
		// Convert to YAML and use dyff for comparison
		yaml1 := diff.DomToYAML(content1)
		yaml2 := diff.DomToYAML(content2)

		// Create temp files for dyff
		tmp1, err := os.CreateTemp("", "vuego-*.yaml")
		if err != nil {
			return fmt.Errorf("failed to create temp file: %w", err)
		}
		defer os.Remove(tmp1.Name())

		tmp2, err := os.CreateTemp("", "vuego-*.yaml")
		if err != nil {
			return fmt.Errorf("failed to create temp file: %w", err)
		}
		defer os.Remove(tmp2.Name())

		if _, err := tmp1.WriteString(yaml1); err != nil {
			return fmt.Errorf("failed to write temp file: %w", err)
		}
		tmp1.Close()

		if _, err := tmp2.WriteString(yaml2); err != nil {
			return fmt.Errorf("failed to write temp file: %w", err)
		}
		tmp2.Close()

		// Run dyff
		cmd := exec.Command("dyff", "between", tmp1.Name(), tmp2.Name())
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Run() // Ignore error as dyff returns 1 when files differ

		if !isEqual {
			return fmt.Errorf("DOM trees do not match")
		}
		return nil

	default:
		return fmt.Errorf("unknown output format: %s (use: simple, unified, yaml)", *outputFormat)
	}
}
