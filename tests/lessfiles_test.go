package tests_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/titpetric/lessgo/dst"
	"github.com/titpetric/lessgo/renderer"
)

// TestAllLessFilesRender verifies all LESS files in the repo can be rendered successfully.
func TestAllLessFilesRender(t *testing.T) {
	lessFiles := findLessFiles(t, "../")

	for _, lessFile := range lessFiles {
		t.Run(lessFile, func(t *testing.T) {
			content, err := os.ReadFile(lessFile)
			require.NoError(t, err, "failed to read file: %s", lessFile)

			// Parse LESS
			parser := dst.NewParser(bytes.NewReader(content))
			file, err := parser.Parse()
			require.NoError(t, err, "failed to parse LESS file: %s", lessFile)

			// Render to CSS
			r := renderer.NewRenderer()
			css, err := r.Render(file)
			require.NoError(t, err, "failed to render LESS file: %s", lessFile)

			// Basic validation: CSS should not be empty
			require.NotEmpty(t, css, "rendered CSS is empty for: %s", lessFile)

			// Check for common syntax errors
			validateCSS(t, css, lessFile)
		})
	}
}

// findLessFiles recursively finds all .less files in a directory.
func findLessFiles(t *testing.T, root string) []string {
	var lessFiles []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".less") {
			lessFiles = append(lessFiles, path)
		}
		return nil
	})
	require.NoError(t, err, "failed to walk directory")
	return lessFiles
}

// validateCSS checks for common CSS syntax issues.
func validateCSS(t *testing.T, css string, filename string) {
	// Count opening and closing braces
	openBraces := strings.Count(css, "{")
	closeBraces := strings.Count(css, "}")

	require.Equal(t, openBraces, closeBraces,
		"mismatched braces in %s: %d opening, %d closing\nCSS:\n%s",
		filename, openBraces, closeBraces, css)

	// Check for common issues
	lines := strings.Split(css, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for unclosed rules
		if strings.Contains(trimmed, "{") && !strings.Contains(trimmed, "}") {
			// Multi-line rule, check later lines
			found := false
			for j := i + 1; j < len(lines) && j < i+50; j++ {
				if strings.Contains(lines[j], "}") {
					found = true
					break
				}
			}
			require.True(t, found, "potential unclosed rule in %s at line %d: %s",
				filename, i+1, trimmed)
		}
	}

	// Verify specific class exists with proper closing braces (for forms.less check)
	if strings.Contains(filename, "forms.less") {
		checkBtnPrimaryClass(t, css, filename)
	}
}

// checkBtnPrimaryClass verifies .btn-primary is properly formed.
func checkBtnPrimaryClass(t *testing.T, css string, filename string) {
	// Find .btn-primary class
	idx := strings.Index(css, ".btn-primary")
	require.Greater(t, idx, -1, "expected to find .btn-primary in %s", filename)

	// Extract the rule (from .btn-primary to the closing brace)
	startIdx := strings.LastIndex(css[:idx], "\n") + 1
	endIdx := strings.Index(css[idx:], "}")
	require.Greater(t, endIdx, -1, "expected closing brace for .btn-primary in %s", filename)

	rule := css[startIdx : idx+endIdx+1]
	t.Logf("btn-primary rule:\n%s", rule)

	// Verify the rule has proper braces
	openCount := strings.Count(rule, "{")
	closeCount := strings.Count(rule, "}")
	require.Equal(t, openCount, closeCount,
		"mismatched braces in .btn-primary rule in %s: %d opening, %d closing",
		filename, openCount, closeCount)

	// Extract CSS properties
	propStart := strings.Index(rule, "{") + 1
	propEnd := strings.LastIndex(rule, "}")
	props := rule[propStart:propEnd]

	t.Logf("btn-primary properties: %s", props)

	// Check background color
	require.Contains(t, props, "background-color:", "expected 'background-color:' property in .btn-primary")

	t.Logf("✓ .btn-primary is properly formed with closing brace and correct properties")
}
