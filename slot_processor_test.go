package vuego_test

import (
	"testing"

	"github.com/titpetric/vuego"
	"github.com/titpetric/vuego/tests/assert"
)

func TestGetSlotName(t *testing.T) {
	tests := []struct {
		name string
		attr string
		want string
	}{
		{name: "empty attribute falls back to default", attr: "", want: "default"},
		{name: "a name is returned as given", attr: "header", want: "header"},
		{name: "the literal default", attr: "default", want: "default"},
		{name: "a kebab-case name", attr: "card-footer", want: "card-footer"},
		{name: "whitespace is not trimmed", attr: " header ", want: " header "},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, vuego.GetSlotName(test.attr))
		})
	}
}
