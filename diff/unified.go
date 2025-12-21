package diff

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

// FormatToNormalizedHTML converts HTML content to a nicely formatted string for diff display.
func FormatToNormalizedHTML(content []byte) (string, error) {
	doc, err := html.Parse(bytes.NewReader(content))
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	renderNodeForDiff(doc, 0, &buf)
	return buf.String(), nil
}

// renderNodeForDiff renders an HTML node tree in a readable format for diffs.
func renderNodeForDiff(n *html.Node, depth int, buf *strings.Builder) {
	if n == nil {
		return
	}

	indent := strings.Repeat("  ", depth)

	switch n.Type {
	case html.DocumentNode:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			renderNodeForDiff(c, depth, buf)
		}

	case html.ElementNode:
		buf.WriteString(indent)
		buf.WriteString("<")
		buf.WriteString(n.Data)

		// Write attributes
		for _, attr := range n.Attr {
			buf.WriteString(" ")
			buf.WriteString(attr.Key)
			if attr.Val != "" {
				buf.WriteString("=\"")
				buf.WriteString(attr.Val)
				buf.WriteString("\"")
			}
		}
		buf.WriteString(">\n")

		// Write children
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			renderNodeForDiff(c, depth+1, buf)
		}

		buf.WriteString(indent)
		buf.WriteString("</")
		buf.WriteString(n.Data)
		buf.WriteString(">\n")

	case html.TextNode:
		text := strings.TrimSpace(n.Data)
		if text != "" {
			buf.WriteString(indent)
			buf.WriteString(text)
			buf.WriteString("\n")
		}

	case html.CommentNode:
		buf.WriteString(indent)
		buf.WriteString("<!-- ")
		buf.WriteString(strings.TrimSpace(n.Data))
		buf.WriteString(" -->\n")
	}
}

// GenerateUnifiedDiff creates a unified diff format output (compatible with colordiff).
func GenerateUnifiedDiff(file1, file2, content1, content2 string) string {
	lines1 := strings.Split(content1, "\n")
	lines2 := strings.Split(content2, "\n")

	var buf strings.Builder
	buf.WriteString("--- ")
	buf.WriteString(file1)
	buf.WriteString("\n")
	buf.WriteString("+++ ")
	buf.WriteString(file2)
	buf.WriteString("\n")

	// Simple line-by-line diff
	maxLines := len(lines1)
	if len(lines2) > maxLines {
		maxLines = len(lines2)
	}

	for i := 0; i < maxLines; i++ {
		var line1, line2 string
		if i < len(lines1) {
			line1 = lines1[i]
		}
		if i < len(lines2) {
			line2 = lines2[i]
		}

		if line1 != line2 {
			if line1 != "" {
				buf.WriteString("- ")
				buf.WriteString(line1)
				buf.WriteString("\n")
			}
			if line2 != "" {
				buf.WriteString("+ ")
				buf.WriteString(line2)
				buf.WriteString("\n")
			}
		} else if line1 != "" {
			buf.WriteString("  ")
			buf.WriteString(line1)
			buf.WriteString("\n")
		}
	}

	return buf.String()
}
