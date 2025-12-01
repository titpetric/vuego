package vuego

import (
	"bytes"
	"context"
	"os"
	"testing"
)

func TestLoadWithLessProcessor(t *testing.T) {
	root := os.DirFS("testdata")

	tpl, err := Load(root, "pages/components/Button.vuego", WithLessProcessor())
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if tpl == nil {
		t.Fatal("expected template to be loaded")
	}
}

func TestLoadWithMultipleProcessors(t *testing.T) {
	root := os.DirFS("testdata")

	tpl, err := Load(root, "pages/components/Button.vuego", WithLessProcessor(), WithLessProcessor())
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if tpl == nil {
		t.Fatal("expected template to be loaded")
	}
}

func TestLoadWithoutProcessors(t *testing.T) {
	root := os.DirFS("testdata")

	tpl, err := Load(root, "pages/components/Button.vuego")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if tpl == nil {
		t.Fatal("expected template to be loaded")
	}
}

func TestTemplateRegisterProcessor(t *testing.T) {
	root := os.DirFS("testdata")

	tpl, err := Load(root, "pages/components/Button.vuego")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	result := tpl.RegisterProcessor(NewLessProcessor(root))

	// Verify chaining returns the template
	if result != tpl {
		t.Error("RegisterProcessor() should return the template for chaining")
	}
}

func TestTemplateFuncs(t *testing.T) {
	root := os.DirFS("testdata")

	tpl, err := Load(root, "pages/components/Button.vuego")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

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

	tpl, err := Load(root, "pages/components/Button.vuego", WithLessProcessor())
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify chaining works
	result := tpl.
		Assign("key", "value").
		RegisterProcessor(NewLessProcessor(root)).
		Funcs(FuncMap{"test": func() string { return "test" }})

	if result != tpl {
		t.Error("method chaining failed")
	}

	if tpl.GetString("key") != "value" {
		t.Error("assignment failed")
	}
}

func TestLoadTemplateRender(t *testing.T) {
	root := os.DirFS("testdata")

	tpl, err := Load(root, "pages/components/Button.vuego", WithLessProcessor())
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Assign required attributes - Button component requires name, variant, and title
	tpl.Assign("name", "TestButton").
		Assign("variant", "primary").
		Assign("title", "Click me")

	var buf bytes.Buffer
	err = tpl.Render(context.Background(), &buf)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("rendered output is empty")
	}
}
