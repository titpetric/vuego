package format

import (
	"fmt"
	"os"

	"github.com/titpetric/vuego/formatter"
)

// Run executes the fmt command with the given arguments.
func Run(args []string) error {
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

// Usage returns the usage string for the fmt command.
func Usage() string {
	return `vuego fmt <file.vuego> [file2.vuego ...]

Format vuego template files.

Examples:
  vuego fmt layout.vuego
  vuego fmt *.vuego`
}
