package vuego

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/net/html"
)

// VueGo drives template evaluation using a scope (no reflection in main flow).
type VueGo struct {
	scope *Scope
}

// New creates a new VueGo.
func New() *VueGo {
	return &VueGo{}
}

// Evaluate initializes the scope stack and renders the provided node (typically the parsed document root).
func (d *VueGo) Evaluate(w io.Writer, node *html.Node, data map[string]any) error {
	d.scope = NewScope(data)
	return d.renderNode(w, node)
}

// renderNode handles dispatch for document/element/text/comment nodes.
func (d *VueGo) renderNode(w io.Writer, node *html.Node) error {
	switch node.Type {
	case html.DocumentNode:
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if err := d.renderNode(w, c); err != nil {
				return err
			}
		}
		return nil
	case html.ElementNode:
		// v-for takes priority: expand first, then render each instance (v-if evaluated per-instance)
		vFor := ""
		for _, a := range node.Attr {
			if a.Key == "v-for" {
				vFor = a.Val
				break
			}
		}
		if vFor != "" {
			vars, collectionExpr, err := parseVFor(vFor)
			if err != nil {
				return err
			}
			// try map[string]any directly first
			colRaw, ok := d.scope.Resolve(collectionExpr)
			if !ok {
				// nothing to iterate -> render nothing
				return nil
			}
			if m, ok := colRaw.(map[string]any); ok {
				for k, val := range m {
					iterScope := map[string]any{}
					if len(vars) == 2 {
						iterScope[vars[0]] = k
						iterScope[vars[1]] = val
					} else {
						iterScope[vars[0]] = val
					}
					d.scope.Push(iterScope)
					if err := d.renderElementSkippingVFor(w, node); err != nil {
						d.scope.Pop()
						return err
					}
					d.scope.Pop()
				}
				return nil
			}
			// otherwise use scope.ForEach which supports many slice/map types
			err = d.scope.ForEach(collectionExpr, func(i int, val any) error {
				iterScope := map[string]any{}
				if len(vars) == 2 {
					iterScope[vars[0]] = i
					iterScope[vars[1]] = val
				} else {
					iterScope[vars[0]] = val
				}
				d.scope.Push(iterScope)
				if err := d.renderElementSkippingVFor(w, node); err != nil {
					d.scope.Pop()
					return err
				}
				d.scope.Pop()
				return nil
			})
			return err
		}
		// no v-for -> render once (v-if checked inside renderElementSkippingVFor)
		return d.renderElementSkippingVFor(w, node)
	case html.TextNode:
		// interpolate {expr} inside text nodes, then escape and write
		out := d.interpolateText(node.Data)
		_, err := w.Write([]byte(html.EscapeString(out)))
		return err
	case html.CommentNode:
		// ignore comments
		return nil
	default:
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if err := d.renderNode(w, c); err != nil {
				return err
			}
		}
		return nil
	}
}

// renderElementSkippingVFor renders an element node but ignores any v-for attribute on it.
// This allows callers to perform repeated rendering (for v-for) without re-processing v-for.
func (d *VueGo) renderElementSkippingVFor(w io.Writer, node *html.Node) error {
	// First: respect v-if (if present). Evaluate using current scope; falsy -> skip element.
	for _, a := range node.Attr {
		if a.Key == "v-if" {
			val, ok := d.scope.Resolve(a.Val)
			if !ok || !isTruthy(val) {
				// condition false -> do not render element at all
				return nil
			}
			// condition true -> proceed (do not write v-if attr)
			break
		}
	}

	// prepare attributes
	var attrs []html.Attribute
	// v-html, if present, will override children and write raw HTML
	vHtmlExpr := ""
	hasVHtml := false

	for _, a := range node.Attr {
		// skip v-for and v-if attributes entirely in output
		if a.Key == "v-for" || a.Key == "v-if" {
			continue
		}
		// v-html: evaluate expression to string, write raw HTML as inner content
		if a.Key == "v-html" {
			vHtmlExpr = a.Val
			hasVHtml = true
			continue
		}
		// :attr or v-bind:attr -> evaluate entire expression as attribute value
		if strings.HasPrefix(a.Key, ":") {
			attrName := strings.TrimPrefix(a.Key, ":")
			if s, ok := d.scope.GetString(a.Val); ok {
				attrs = append(attrs, html.Attribute{Key: attrName, Val: s})
			} else {
				attrs = append(attrs, html.Attribute{Key: attrName, Val: ""})
			}
			continue
		}
		if strings.HasPrefix(a.Key, "v-bind:") {
			attrName := strings.TrimPrefix(a.Key, "v-bind:")
			if s, ok := d.scope.GetString(a.Val); ok {
				attrs = append(attrs, html.Attribute{Key: attrName, Val: s})
			} else {
				attrs = append(attrs, html.Attribute{Key: attrName, Val: ""})
			}
			continue
		}
		// literal attribute: allow {expr} interpolation inside the attribute value
		interpolated := d.interpolateText(a.Val)
		attrs = append(attrs, html.Attribute{Key: a.Key, Val: interpolated})
	}

	// write start tag and attributes
	if _, err := w.Write([]byte("<" + node.Data)); err != nil {
		return err
	}
	for _, a := range attrs {
		escaped := html.EscapeString(a.Val)
		if _, err := w.Write([]byte(fmt.Sprintf(` %s="%s"`, a.Key, escaped))); err != nil {
			return err
		}
	}

	// treat void tags simply (no children)
	voidTags := map[string]bool{
		"area": true, "base": true, "br": true, "col": true, "embed": true,
		"hr": true, "img": true, "input": true, "link": true, "meta": true,
		"param": true, "source": true, "track": true, "wbr": true,
	}
	if voidTags[strings.ToLower(node.Data)] {
		_, err := w.Write([]byte(" />"))
		return err
	}

	if _, err := w.Write([]byte(">")); err != nil {
		return err
	}

	// children or v-html content
	if hasVHtml {
		if s, ok := d.scope.GetString(vHtmlExpr); ok {
			// write unescaped HTML (v-html semantics). Caller must ensure safety.
			if _, err := w.Write([]byte(s)); err != nil {
				return err
			}
		} else {
			// unresolved -> write nothing
		}
	} else {
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if err := d.renderNode(w, c); err != nil {
				return err
			}
		}
	}

	// end tag
	if _, err := w.Write([]byte("</" + node.Data + ">")); err != nil {
		return err
	}
	return nil
}

// parseVFor parses v-for strings like:
//
//	"item in items"
//	"(i, v) in items"
func parseVFor(s string) ([]string, string, error) {
	s = strings.TrimSpace(s)
	parts := strings.SplitN(s, " in ", 2)
	if len(parts) != 2 {
		return nil, "", fmt.Errorf("invalid v-for expression: %q", s)
	}
	left := strings.TrimSpace(parts[0])
	right := strings.TrimSpace(parts[1])

	var vars []string
	if strings.HasPrefix(left, "(") && strings.HasSuffix(left, ")") {
		inside := strings.TrimSpace(left[1 : len(left)-1])
		for _, p := range strings.Split(inside, ",") {
			vars = append(vars, strings.TrimSpace(p))
		}
	} else {
		vars = []string{left}
	}
	if len(vars) == 0 || vars[0] == "" {
		return nil, "", fmt.Errorf("no iteration variables in v-for: %q", s)
	}
	return vars, right, nil
}

// isTruthy implements a simple truthiness evaluator for v-if:
// - nil -> false
// - bool -> its value
// - string -> non-empty true
// - int/float -> non-zero true (several common numeric types supported)
// - slices/maps -> len>0 true (supported concrete types)
// - otherwise non-nil -> true
func isTruthy(v any) bool {
	if v == nil {
		return false
	}
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return t != ""
	case int:
		return t != 0
	case int8:
		return t != 0
	case int16:
		return t != 0
	case int32:
		return t != 0
	case int64:
		return t != 0
	case uint:
		return t != 0
	case uint8:
		return t != 0
	case uint16:
		return t != 0
	case uint32:
		return t != 0
	case uint64:
		return t != 0
	case float32:
		return t != 0
	case float64:
		return t != 0
	case []any:
		return len(t) > 0
	case []map[string]any:
		return len(t) > 0
	case []string:
		return len(t) > 0
	case []int:
		return len(t) > 0
	case []float64:
		return len(t) > 0
	case map[string]any:
		return len(t) > 0
	case map[string]string:
		return len(t) > 0
	default:
		// unknown concrete type but non-nil -> treat as true
		return true
	}
}
