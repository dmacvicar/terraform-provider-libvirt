package docindex

import (
	"fmt"
	stdhtml "html"
	"io"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

var (
	// Match XML element names in angle brackets
	elementPattern = regexp.MustCompile(`<([a-zA-Z][a-zA-Z0-9_-]*)>`)
	// Match XML elements in code/literal spans
	codeElementPattern = regexp.MustCompile(`<code[^>]*>([a-zA-Z][a-zA-Z0-9_-]*)</code>`)
	// Match literal span content that looks like an XML name
	literalNamePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_.:-]*$`)
)

// ParseHTML extracts documentation sections from an HTML file
func ParseHTML(r io.Reader, baseURL string) (FileIndex, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return FileIndex{}, fmt.Errorf("parsing HTML: %w", err)
	}

	var sections []Section

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		// Check for <section> elements with IDs
		if n.Type == html.ElementNode && n.Data == "section" {
			id := getAttr(n, "id")
			if id != "" {
				// Extract section info
				title := extractSectionTitle(n)
				section := Section{
					ID:    id,
					Title: title,
					URL:   baseURL + "#" + id,
				}
				finalizeSection(&section, n)
				sections = append(sections, section)
			}
		}

		// Recurse
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(doc)

	return FileIndex{Sections: sections}, nil
}

// extractSectionTitle finds the first heading in a section
func extractSectionTitle(n *html.Node) string {
	var find func(*html.Node) string
	find = func(node *html.Node) string {
		if node.Type == html.ElementNode {
			switch node.Data {
			case "h1", "h2", "h3", "h4", "h5", "h6":
				return extractText(node)
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if title := find(c); title != "" {
				return title
			}
		}
		return ""
	}
	return find(n)
}

// extractSectionContent gets all text content from p, dd, li elements
func extractSectionContent(n *html.Node) string {
	var content strings.Builder
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			switch node.Data {
			case "p", "dd", "li":
				text := extractText(node)
				if text != "" {
					content.WriteString(text)
					content.WriteString(" ")
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return content.String()
}

// finalizeSection extracts element names and creates preview
func finalizeSection(section *Section, node *html.Node) {
	content := extractSectionContent(node)

	// Extract element names from content
	elements := make(map[string]bool)

	// Find <element> patterns
	for _, match := range elementPattern.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			elements[match[1]] = true
		}
	}

	// Find elements in code spans
	for _, match := range codeElementPattern.FindAllStringSubmatch(content, -1) {
		if len(match) > 1 {
			elements[match[1]] = true
		}
	}

	// Scan literal blocks and spans for escaped XML element names
	collectLiteralElements(node, elements)

	// Convert to slice
	for elem := range elements {
		section.Keywords = append(section.Keywords, elem)
	}

	// Create preview (first 200 chars)
	preview := strings.TrimSpace(content)
	if len(preview) > 200 {
		preview = preview[:200] + "..."
	}
	section.Preview = preview
}

// extractText recursively extracts all text from a node
func extractText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}

	if shouldSkipNode(n) {
		return ""
	}

	var text strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text.WriteString(extractText(c))
	}
	return text.String()
}

func collectLiteralElements(n *html.Node, elements map[string]bool) {
	if n == nil {
		return
	}

	if n.Type == html.ElementNode {
		if n.Data == "pre" && hasClass(n, "literal-block") {
			text := extractLiteralBlockText(n)
			addElementsFromText(text, elements)
		}

		if n.Data == "span" && (hasClass(n, "docutils") || hasClass(n, "literal") || hasClass(n, "pre")) {
			text := strings.TrimSpace(extractText(n))
			if literalNamePattern.MatchString(text) {
				// unescape to ensure consistent casing
				text = stdhtml.UnescapeString(text)
				elements[text] = true
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		collectLiteralElements(c, elements)
	}
}

func extractLiteralBlockText(n *html.Node) string {
	var text strings.Builder
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.TextNode {
			text.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return stdhtml.UnescapeString(text.String())
}

func addElementsFromText(text string, elements map[string]bool) {
	for _, match := range elementPattern.FindAllStringSubmatch(text, -1) {
		if len(match) > 1 {
			elements[match[1]] = true
		}
	}
}

func shouldSkipNode(n *html.Node) bool {
	if n.Type != html.ElementNode {
		return false
	}

	if n.Data == "a" && hasClass(n, "headerlink") {
		return true
	}

	return false
}

func hasClass(n *html.Node, class string) bool {
	for _, attr := range n.Attr {
		if attr.Key != "class" {
			continue
		}
		for _, value := range strings.Fields(attr.Val) {
			if value == class {
				return true
			}
		}
	}
	return false
}

// getAttr gets an attribute value from a node
func getAttr(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}
