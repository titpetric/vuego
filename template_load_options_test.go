package vuego

import (
	"bytes"
	"context"
	"os"
	"testing"
)

func TestLoadWithLessProcessor(t *testing.T) {
	root := os.DirFS("testdata")

	tpl := NewFS(root, WithLessProcessor())

	if tpl == nil {
		t.Fatal("expected template to be loaded")
	}
}

func TestLoadWithMultipleProcessors(t *testing.T) {
	root := os.DirFS("testdata")

	tpl := NewFS(root, WithLessProcessor(), WithLessProcessor())

	if tpl == nil {
		t.Fatal("expected template to be loaded")
	}
}

func TestLoadWithoutProcessors(t *testing.T) {
	root := os.DirFS("testdata")

	tpl := NewFS(root)

	if tpl == nil {
		t.Fatal("expected template to be loaded")
	}
}

func TestTemplateFuncs(t *testing.T) {
	root := os.DirFS("testdata")

	tpl := NewFS(root)

	funcMap := FuncMap{
		"customFunc": func(s string) string {
			return "custom: " + s
		},
	}

	result := tpl.Funcs(funcMap)

	// Verify chaining returns the template
	if result != tpl {
		t.Error("Funcs() should return the template for chaining")
	}
}

func TestTemplateChaining(t *testing.T) {
	root := os.DirFS("testdata")

	tpl := NewFS(root, WithLessProcessor())

	// Verify chaining works
	result := tpl.
		Assign("key", "value").
		Funcs(FuncMap{"test": func() string { return "test" }})

	if result != tpl {
		t.Error("method chaining failed")
	}

	if tpl.Get("key") != "value" {
		t.Error("assignment failed")
	}
}

func TestLoadTemplateRender(t *testing.T) {
	root := os.DirFS("testdata")

	tpl := NewFS(root, WithLessProcessor())

	// Assign required attributes - Button component requires name, variant, and title
	tpl.Assign("name", "TestButton").
		Assign("variant", "primary").
		Assign("title", "Click me")

	var buf bytes.Buffer
	err := tpl.Load("pages/components/Button.vuego").Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("rendered output is empty")
	}
}
