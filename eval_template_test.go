package vuego_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/titpetric/vuego"
)

// TestRequiredAttributeError tests that a missing :required attribute
// produces an error. Ideally, the error should include file/line information
// from the source .vuego template.
//
// CURRENT LIMITATION: golang.org/x/net/html.Node does not track source line information.
// See: https://github.com/golang/go/issues/34302
//
// To add line tracking, we would need to either:
// 1. Use a forked version of html parser that tracks offsets/lines
// 2. Maintain a separate mapping of nodes to source positions
// 3. Wait for upstream support in golang.org/x/net/html
func TestRequiredAttributeError(t *testing.T) {
	vue := vuego.NewVue(os.DirFS("testdata/pages"))

	data := map[string]any{
		// Both pageTitle and pageAuthor are missing to trigger validation errors
		// The component requires both "title" and "author"
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "required-error-test/page.vuego", data)

	// We expect an error because required attributes are not provided
	require.Error(t, err, "Expected error for missing required attributes")

	// Check that the error includes the component filename
	require.Contains(t, err.Error(), "component.vuego", "Error should mention component filename")

	// Check that the error includes the parent template (page.vuego)
	require.Contains(t, err.Error(), "page.vuego", "Error should mention parent page filename")

	// Check that the error message mentions a required attribute
	require.Contains(t, err.Error(), "required attribute", "Error should mention required attribute")

	// Check that the error shows template inclusion chain
	require.Contains(t, err.Error(), "included from", "Error should show template inclusion chain")

	t.Logf("Error with template chain: %v", err)
}
