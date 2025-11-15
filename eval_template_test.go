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
	vue := vuego.NewVue(os.DirFS("testdata"))

	data := map[string]any{
		// Both pageTitle and pageAuthor are missing to trigger validation errors
		// The component requires both "title" and "author"
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "required-error-test/page.vuego", data)

	require.Equal(t, "error in required-error-test/component.vuego (included from required-error-test/page.vuego): required attribute 'title' not provided", err.Error())
}

// TestRootLevelTemplateRequired tests that a root-level template with :required attribute
// validates the required attribute at the document root level.
func TestRootLevelTemplateRequired(t *testing.T) {
	vue := vuego.NewVue(os.DirFS("testdata"))

	data := map[string]any{
		// Missing 'title' which is required by the template
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "root-template-required/page.vuego", data)

	require.Equal(t, "required attribute 'title' not provided", err.Error())
}

// TestRootLevelTemplateRequiredSuccess tests that a root-level template with :required attribute
// renders successfully when the required attribute is provided.
func TestRootLevelTemplateRequiredSuccess(t *testing.T) {
	vue := vuego.NewVue(os.DirFS("testdata"))

	data := map[string]any{
		"title": "My Page Title",
	}

	var buf bytes.Buffer
	err := vue.Render(&buf, "root-template-required/page.vuego", data)

	require.NoError(t, err)
	require.Contains(t, buf.String(), "<h1>My Page Title</h1>")
}
