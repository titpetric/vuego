package vuego

import (
	"bytes"
	"testing"
)

func TestDoubleEscapingBug(t *testing.T) {
	const template = `<div data-tooltip="{{value}}"></div>`

	var out bytes.Buffer
	err := Render(&out, template, map[string]any{
		"value": "A & B",
	})

	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	result := out.String()
	expected := `<div data-tooltip="A &amp; B"></div>`
	if result != expected {
		t.Errorf("Double escaping detected!\nExpected: %s\nGot:      %s", expected, result)
	}
}
