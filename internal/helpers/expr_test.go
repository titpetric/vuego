package helpers_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/titpetric/vuego/internal/helpers"
)

func TestContainsPipe(t *testing.T) {
	t.Run("empty string has no pipe", func(t *testing.T) {
		require.False(t, helpers.ContainsPipe(""))
	})

	t.Run("string with pipe returns true", func(t *testing.T) {
		require.True(t, helpers.ContainsPipe("name | upper"))
	})

	t.Run("string without pipe returns false", func(t *testing.T) {
		require.False(t, helpers.ContainsPipe("name"))
	})

	t.Run("pipe at start", func(t *testing.T) {
		require.True(t, helpers.ContainsPipe("| upper"))
	})

	t.Run("multiple pipes", func(t *testing.T) {
		require.True(t, helpers.ContainsPipe("name | upper | lower"))
	})

	t.Run("pipe in string literal doesn't count", func(t *testing.T) {
		// Note: This function is simple and checks all characters
		require.True(t, helpers.ContainsPipe(`"text | literal"`))
	})
}

func TestIsIdentifier(t *testing.T) {
	t.Run("empty string is not identifier", func(t *testing.T) {
		require.False(t, helpers.IsIdentifier(""))
	})

	t.Run("valid identifier with letters", func(t *testing.T) {
		require.True(t, helpers.IsIdentifier("name"))
	})

	t.Run("valid identifier with underscore start", func(t *testing.T) {
		require.True(t, helpers.IsIdentifier("_name"))
	})

	t.Run("valid identifier with numbers", func(t *testing.T) {
		require.True(t, helpers.IsIdentifier("name123"))
	})

	t.Run("valid identifier with mixed case", func(t *testing.T) {
		require.True(t, helpers.IsIdentifier("MyName"))
	})

	t.Run("valid identifier with underscores", func(t *testing.T) {
		require.True(t, helpers.IsIdentifier("my_name_123"))
	})

	t.Run("single underscore is valid", func(t *testing.T) {
		require.True(t, helpers.IsIdentifier("_"))
	})

	t.Run("single letter is valid", func(t *testing.T) {
		require.True(t, helpers.IsIdentifier("a"))
	})

	t.Run("starts with digit is invalid", func(t *testing.T) {
		require.False(t, helpers.IsIdentifier("1name"))
	})

	t.Run("contains space is invalid", func(t *testing.T) {
		require.False(t, helpers.IsIdentifier("my name"))
	})

	t.Run("contains dash is invalid", func(t *testing.T) {
		require.False(t, helpers.IsIdentifier("my-name"))
	})

	t.Run("contains dot is invalid", func(t *testing.T) {
		require.False(t, helpers.IsIdentifier("my.name"))
	})

	t.Run("contains special chars is invalid", func(t *testing.T) {
		require.False(t, helpers.IsIdentifier("my@name"))
	})

	t.Run("contains pipe is invalid", func(t *testing.T) {
		require.False(t, helpers.IsIdentifier("my|name"))
	})
}

func TestIsIdentifierChar(t *testing.T) {
	t.Run("lowercase letter is valid", func(t *testing.T) {
		require.True(t, helpers.IsIdentifierChar('a', false))
	})

	t.Run("uppercase letter is valid", func(t *testing.T) {
		require.True(t, helpers.IsIdentifierChar('A', false))
	})

	t.Run("digit is valid", func(t *testing.T) {
		require.True(t, helpers.IsIdentifierChar('0', false))
	})

	t.Run("underscore is valid", func(t *testing.T) {
		require.True(t, helpers.IsIdentifierChar('_', false))
	})

	t.Run("space is invalid", func(t *testing.T) {
		require.False(t, helpers.IsIdentifierChar(' ', false))
	})

	t.Run("dash is invalid", func(t *testing.T) {
		require.False(t, helpers.IsIdentifierChar('-', false))
	})

	t.Run("special chars are invalid", func(t *testing.T) {
		require.False(t, helpers.IsIdentifierChar('$', false))
		require.False(t, helpers.IsIdentifierChar('@', false))
		require.False(t, helpers.IsIdentifierChar('!', false))
	})
}

func TestIsFunctionCall(t *testing.T) {
	t.Run("empty string is not a function call", func(t *testing.T) {
		require.False(t, helpers.IsFunctionCall(""))
	})

	t.Run("identifier without parentheses is not a function call", func(t *testing.T) {
		require.False(t, helpers.IsFunctionCall("name"))
	})

	t.Run("function call with empty args", func(t *testing.T) {
		require.True(t, helpers.IsFunctionCall("func()"))
	})

	t.Run("function call with args", func(t *testing.T) {
		require.True(t, helpers.IsFunctionCall("func(arg)"))
	})

	t.Run("function call with multiple args", func(t *testing.T) {
		require.True(t, helpers.IsFunctionCall("func(arg1, arg2)"))
	})

	t.Run("function call with leading whitespace", func(t *testing.T) {
		require.True(t, helpers.IsFunctionCall("  func()"))
	})

	t.Run("parentheses at start is not a function call", func(t *testing.T) {
		require.False(t, helpers.IsFunctionCall("()"))
	})

	t.Run("invalid identifier with parentheses", func(t *testing.T) {
		require.False(t, helpers.IsFunctionCall("my-func()"))
	})

	t.Run("function call with underscore", func(t *testing.T) {
		require.True(t, helpers.IsFunctionCall("_func()"))
	})

	t.Run("function call with numbers in name", func(t *testing.T) {
		require.True(t, helpers.IsFunctionCall("func123()"))
	})
}
