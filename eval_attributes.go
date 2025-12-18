package vuego

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"

	"github.com/titpetric/vuego/internal/helpers"
)

func (v *Vue) evalAttributes(ctx VueContext, n *html.Node) (map[string]any, error) {
	if n.Type != html.ElementNode {
		return nil, nil
	}

	results := map[string]any{}

	var newAttrs []html.Attribute

	// First pass: collect static attributes and evaluate bound ones
	for _, a := range n.Attr {
		key := a.Key
		val := strings.TrimSpace(a.Val)

		boundName := key
		// literal bindings
		if strings.HasPrefix(key, "\\:") {
			boundName = boundName[1:]
			newAttrs = append(newAttrs, html.Attribute{
				Key: boundName,
				Val: val,
			})
			continue
		}
		if strings.HasPrefix(key, "[") && strings.HasSuffix(key, "]") {
			boundName = boundName[1 : len(boundName)-1]
			newAttrs = append(newAttrs, html.Attribute{
				Key: boundName,
				Val: val,
			})
			continue
		}
		if strings.HasPrefix(key, ":") {
			boundName = boundName[1:]
		}
		if strings.HasPrefix(key, "v-bind:") {
			boundName = boundName[7:]
		}

		switch {
		case boundName != key:
			boundValue, err := v.evalBoundAttribute(ctx, boundName, val)
			if err != nil {
				return nil, fmt.Errorf("error evaluating attr %s: %w", boundName, err)
			}
			if !helpers.IsTruthy(boundValue) {
				continue
			}
			results[boundName] = boundValue
		default:
			var err error
			boundValue := val
			if containsInterpolation(val) {
				boundValue, err = v.interpolate(ctx, val)
				if err != nil {
					return nil, fmt.Errorf("error evaluating attr %s: %w", boundName, err)
				}
			}
			newAttrs = append(newAttrs, html.Attribute{
				Key: boundName,
				Val: boundValue,
			})
		}
	}

	// Second pass: merge bound attributes with static ones
	for attrName, boundValue := range results {
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
				newAttrs[staticIdx].Val = fmt.Sprintf("%s %s", newAttrs[staticIdx].Val, boundValue)
				continue
			}
			if attrName == "style" {
				// Merge styles, with bound value taking precedence
				staticStyle := newAttrs[staticIdx].Val
				mergedStyle := v.mergeStyles(staticStyle, boundValue.(string))
				newAttrs[staticIdx].Val = mergedStyle
				continue
			}
			// For other attributes, bound value replaces static
			newAttrs[staticIdx].Val = fmt.Sprint(boundValue)
		} else {
			if helpers.IsTruthy(boundValue) {
				newAttrs = append(newAttrs, html.Attribute{
					Key: attrName,
					Val: fmt.Sprint(boundValue),
				})
				continue
			}
		}
	}

	n.Attr = newAttrs

	// Don't overwrite results with stringified attributes
	// The results map already contains the correctly typed bound values from line 39
	// Only add static attributes to results (they are naturally strings from interpolation)
	for _, v := range newAttrs {
		// Only add if not already in results (bound attributes take precedence and keep their type)
		if _, exists := results[v.Key]; !exists {
			results[v.Key] = v.Val
		}
	}

	return results, nil
}

// evalBoundAttribute evaluates a bound attribute, handling objects for class/style.
func (v *Vue) evalBoundAttribute(ctx VueContext, attrName, expr string) (any, error) {
	expr = strings.TrimSpace(expr)

	// Evaluate interpolation
	if containsInterpolation(expr) {
		interpolated, err := v.interpolate(ctx, expr)
		if err != nil {
			return "", err
		}
		return interpolated, nil
	}

	// Check if the expression is an object literal
	if strings.HasPrefix(expr, "{") && strings.HasSuffix(expr, "}") {
		return v.evalObjectBinding(ctx, attrName, expr), nil
	}

	// Regular variable binding
	valResolved, ok := ctx.stack.Resolve(expr)
	if ok {
		return valResolved, nil
	}
	return "", nil
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

	switch attrName {
	case "class":
		return v.buildClassString(pairs)
	case "style":
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
