package main

import (
	"fmt"
	"os"

	"github.com/titpetric/vuego/cmd/vuego/commands/diff"
	fmtcmd "github.com/titpetric/vuego/cmd/vuego/commands/fmt"
	"github.com/titpetric/vuego/cmd/vuego/commands/render"
	"github.com/titpetric/vuego/cmd/vuego/commands/tour"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return printUsage()
	}

	command := os.Args[1]

	switch command {
	case "fmt":
		return fmtcmd.Run(os.Args[2:])
	case "render":
		return render.Run(os.Args[2:])
	case "diff":
		return diff.Run(os.Args[2:])
	case "tour":
		return tour.Run(os.Args[2:])
	case "--help", "-h", "help":
		return printHelp()
	default:
		// Backward compatibility: treat first arg as template file
		// Supports: vuego file.vuego [data.yaml]
		if len(os.Args) >= 2 && len(os.Args) <= 3 {
			return render.Run(os.Args[1:])
		}
		return printUsage()
	}
}

func printUsage() error {
	return fmt.Errorf("Usage: vuego <command> [args...]\n       vuego help")
}

func printHelp() error {
	fmt.Fprintf(os.Stderr, `vuego - Template formatter, renderer, and development server

Usage:
  vuego <command> [args...]

Commands:
  fmt      Format vuego template files
  render   Render templates with data (supports JSON and YAML)
  diff     Compare two HTML/vuego files using DOM comparison
  tour     Start the vuego tour server
  help     Show this help message

Examples:
  vuego fmt layout.vuego
  vuego render index.vuego                  (auto-loads index.yaml if it exists)
  vuego render index.vuego data.json
  vuego diff before.html after.html
  vuego tour ./tour

Run 'vuego <command> --help' for more information on a command.
`)
	return nil
}
