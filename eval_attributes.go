package vuego

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"

	"github.com/titpetric/vuego/internal/helpers"
)

func (v *Vue) evalAttributes(ctx VueContext, n *html.Node) error {
	if n.Type != html.ElementNode {
		return nil
	}

	var newAttrs []html.Attribute
	boundAttrs := make(map[string]string) // Track bound attributes to merge with static ones

	// First pass: collect static attributes and evaluate bound ones
	for _, a := range n.Attr {
		key := a.Key
		val := strings.TrimSpace(a.Val)

		switch {
		case strings.HasPrefix(key, ":"):
			boundName := strings.TrimPrefix(key, ":")
			boundValue := v.evalBoundAttribute(ctx, boundName, val)
			if boundValue != "" {
				boundAttrs[boundName] = boundValue
			}

		case strings.HasPrefix(key, "v-bind:"):
			boundName := strings.TrimPrefix(key, "v-bind:")
			boundValue := v.evalBoundAttribute(ctx, boundName, val)
			if boundValue != "" {
				boundAttrs[boundName] = boundValue
			}

		default:
			// run interpolation on normal attributes
			interpolated, err := v.interpolate(ctx, val)
			if err != nil {
				// For now, use original value on error in attributes
				interpolated = val
			}
			newAttrs = append(newAttrs, html.Attribute{
				Key: key,
				Val: interpolated,
			})
		}
	}

	// Second pass: merge bound attributes with static ones
	for attrName, boundValue := range boundAttrs {
		// Check if there's a static attribute with the same name
		staticIdx := -1
		for i, a := range newAttrs {
			if a.Key == attrName {
				staticIdx = i
				break
			}
		}

		if staticIdx >= 0 {
			// Merge with static attribute (special handling for class and style)
			if attrName == "class" {
				newAttrs[staticIdx].Val = newAttrs[staticIdx].Val + " " + boundValue
			} else if attrName == "style" {
				// Merge styles, with bound value taking precedence
				staticStyle := newAttrs[staticIdx].Val
				mergedStyle := v.mergeStyles(staticStyle, boundValue)
				newAttrs[staticIdx].Val = mergedStyle
			} else {
				// For other attributes, bound value replaces static
				newAttrs[staticIdx].Val = boundValue
			}
		} else {
			// No static attribute, just add the bound one
			newAttrs = append(newAttrs, html.Attribute{
				Key: attrName,
				Val: boundValue,
			})
		}
	}

	n.Attr = newAttrs
	return nil
}

// evalBoundAttribute evaluates a bound attribute, handling objects for class/style.
func (v *Vue) evalBoundAttribute(ctx VueContext, attrName, expr string) string {
	// Check if the expression is an object literal
	if strings.HasPrefix(strings.TrimSpace(expr), "{") && strings.HasSuffix(strings.TrimSpace(expr), "}") {
		return v.evalObjectBinding(ctx, attrName, expr)
	}

	// Regular variable binding
	valResolved, _ := ctx.stack.Resolve(expr)
	if helpers.IsTruthy(valResolved) {
		return fmt.Sprint(valResolved)
	}
	return ""
}

// evalObjectBinding evaluates object literals like {display: "none"} or {active: true, error: false}
// For :class, treats values as booleans and includes keys where value is truthy.
// For :style, treats values as strings and builds CSS property:value pairs.
func (v *Vue) evalObjectBinding(ctx VueContext, attrName, expr string) string {
	expr = strings.TrimSpace(expr)
	if !strings.HasPrefix(expr, "{") || !strings.HasSuffix(expr, "}") {
		return ""
	}

	content := expr[1 : len(expr)-1] // Remove { }
	pairs := v.parseObjectPairs(ctx, content)

	if attrName == "class" {
		return v.buildClassString(pairs)
	} else if attrName == "style" {
		return v.buildStyleString(pairs)
	}

	// For other attributes, just concatenate all values
	var values []string
	for _, v := range pairs {
		if v != "" {
			values = append(values, v)
		}
	}
	return strings.Join(values, " ")
}

// parseObjectPairs parses key:value pairs from an object literal.
// Returns a slice of resolved values in order.
func (v *Vue) parseObjectPairs(ctx VueContext, content string) []string {
	var pairs []string

	// Split by comma, but respect quoted strings
	items := v.splitObjectItems(content)

	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		// Split by colon
		colonIdx := strings.Index(item, ":")
		if colonIdx == -1 {
			continue
		}

		key := strings.TrimSpace(item[:colonIdx])
		valueExpr := strings.TrimSpace(item[colonIdx+1:])

		// Try to resolve as expression first (handles literals and expressions)
		val, err := v.exprEval.Eval(valueExpr, ctx.stack.EnvMap())
		if err != nil {
			// Fall back to stack resolution for variable references
			var ok bool
			val, ok = ctx.stack.Resolve(valueExpr)
			if !ok {
				pairs = append(pairs, "")
				continue
			}
		}

		// Store both key and resolved value
		pairs = append(pairs, fmt.Sprintf("%s:%v", key, val))
	}

	return pairs
}

// splitObjectItems splits comma-separated items in an object, respecting quoted strings.
func (v *Vue) splitObjectItems(content string) []string {
	var items []string
	var current strings.Builder
	inQuotes := false
	quoteChar := rune(0)
	inBrackets := 0

	for _, ch := range content {
		switch {
		case (ch == '"' || ch == '\'') && !inQuotes:
			inQuotes = true
			quoteChar = ch
			current.WriteRune(ch)
		case ch == quoteChar && inQuotes:
			inQuotes = false
			current.WriteRune(ch)
		case ch == '{' && !inQuotes:
			inBrackets++
			current.WriteRune(ch)
		case ch == '}' && !inQuotes:
			inBrackets--
			current.WriteRune(ch)
		case ch == ',' && !inQuotes && inBrackets == 0:
			items = append(items, current.String())
			current.Reset()
		default:
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		items = append(items, current.String())
	}

	return items
}

// buildClassString builds a space-separated class string from key:value pairs.
// Includes key only if the boolean value is truthy.
func (v *Vue) buildClassString(pairs []string) string {
	var classes []string

	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		colonIdx := strings.Index(pair, ":")
		if colonIdx == -1 {
			continue
		}

		key := strings.TrimSpace(pair[:colonIdx])
		valueStr := strings.TrimSpace(pair[colonIdx+1:])

		// Check if value is truthy using the actual type
		val := parseValue(valueStr)
		if helpers.IsTruthy(val) {
			classes = append(classes, key)
		}
	}

	return strings.Join(classes, " ")
}

// buildStyleString builds a CSS style string from key:value pairs.
// Each pair becomes a property:value; entry.
// camelCase keys are automatically converted to kebab-case (e.g., fontSize -> font-size).
func (v *Vue) buildStyleString(pairs []string) string {
	var styles []string

	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		colonIdx := strings.Index(pair, ":")
		if colonIdx == -1 {
			continue
		}

		key := strings.TrimSpace(pair[:colonIdx])
		value := strings.TrimSpace(pair[colonIdx+1:])

		// Remove quotes if present
		value = strings.Trim(value, "\"'")

		if value != "" {
			// Convert camelCase to kebab-case if the key doesn't contain hyphens
			if !strings.Contains(key, "-") {
				key = camelToKebab(key)
			}
			styles = append(styles, key+":"+value+";")
		}
	}

	return strings.Join(styles, "")
}

// camelToKebab converts camelCase strings to kebab-case.
// e.g., "fontSize" -> "font-size", "backgroundColor" -> "background-color"
func camelToKebab(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('-')
			result.WriteRune(r + 32) // Convert to lowercase
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// parseValue converts a string representation to a Go value for truthiness check.
func parseValue(s string) interface{} {
	s = strings.TrimSpace(s)

	// Handle boolean strings
	switch s {
	case "true":
		return true
	case "false":
		return false
	}

	// Handle quoted strings
	if (strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"")) ||
		(strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")) {
		return strings.Trim(s, "\"'")
	}

	// Handle numbers - convert to int for proper truthiness checking
	if isNumeric(s) {
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			return int(i)
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return f
		}
	}

	return s
}

// isNumeric checks if a string is a numeric value.
func isNumeric(s string) bool {
	s = strings.TrimSpace(s)
	matched, _ := regexp.MatchString(`^-?\d+(\.\d+)?$`, s)
	return matched
}

// mergeStyles merges static and bound CSS styles, with bound values taking precedence.
func (v *Vue) mergeStyles(staticStyle, boundStyle string) string {
	// Parse both styles into maps
	staticMap := parseStyleMap(staticStyle)
	boundMap := parseStyleMap(boundStyle)

	// Merge: bound values override static ones
	for k, v := range boundMap {
		staticMap[k] = v
	}

	// Rebuild style string
	var styles []string
	for k, v := range staticMap {
		styles = append(styles, k+":"+v+";")
	}
	return strings.Join(styles, "")
}

// parseStyleMap parses a CSS style string into a map of properties to values.
func parseStyleMap(style string) map[string]string {
	result := make(map[string]string)
	if style == "" {
		return result
	}

	// Split by semicolon to get individual properties
	parts := strings.Split(style, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split by colon to get key-value pair
		kv := strings.SplitN(part, ":", 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			val := strings.TrimSpace(kv[1])
			result[key] = val
		}
	}

	return result
}
