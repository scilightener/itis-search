package task

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/net/html"

	"search/internal/pkg"
)

func ParseHandler(_ context.Context, t *Task) *Task {
	const op = "task.parse.ParseHandler"

	if t.Finished {
		return t
	}

	err := parseDocument(t)
	if err != nil {
		t = t.Fail(fmt.Sprintf("%s: %s", op, err.Error()))
	}

	return t
}

func parseDocument(t *Task) error {
	baseURL, err := url.Parse(t.Document.URI)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	reader := bytes.NewReader(t.Document.Text)
	doc, err := html.Parse(reader)
	if err != nil {
		return fmt.Errorf("failed to parse HTML: %w", err)
	}

	textContent, links, err := extractContent(doc, baseURL)
	if err != nil {
		return fmt.Errorf("failed to extract content: %w", err)
	}

	t.Document.Text = []byte(textContent)
	t.Document.Links = links

	return nil
}

// extractContent extracts the text content and links from an HTML document.
func extractContent(doc *html.Node, baseURL *url.URL) (string, []string, error) {
	var textContent strings.Builder
	links := make(map[string]bool)

	var extract func(*html.Node)
	extract = func(n *html.Node) {
		if n.Type == html.ElementNode && isNonTextElement(n.Data) {
			return
		}

		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if text != "" {
				text = replaceHtmlEntities(text)
				textContent.WriteString(text)
				textContent.WriteString(" ")
			}
		}

		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					link := attr.Val
					if !isImageLink(link) {
						absoluteLink, err := resolveURL(baseURL, link)
						if err != nil {
							break
						}
						absoluteLink = pkg.PrepareLink(absoluteLink)
						links[absoluteLink] = true
					}
					break
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}

	extract(doc)

	var uniqueLinks []string
	for link := range links {
		uniqueLinks = append(uniqueLinks, link)
	}

	return textContent.String(), uniqueLinks, nil
}

// isNonTextElement checks if an HTML element is non-text.
func isNonTextElement(element string) bool {
	nonTextElements := []string{"script", "style", "iframe", "img", "noscript"}
	for _, e := range nonTextElements {
		if e == element {
			return true
		}
	}
	return false
}

// replaceHtmlEntities replaces HTML entities in a string.
func replaceHtmlEntities(text string) string {
	return strings.ReplaceAll(text, "\u00a0", " ")
}

var imageExtensions = []string{".png", ".jpg", ".jpeg", ".gif", ".bmp", ".tiff", ".svg", ".webp"}

// isImageLink checks if a link is an image link.
func isImageLink(link string) bool {
	for _, ext := range imageExtensions {
		if strings.HasSuffix(link, ext) {
			return true
		}
	}
	return false
}

// resolveURL resolves a URL relative to a base URL.
func resolveURL(baseURL *url.URL, link string) (string, error) {
	parsedLink, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	return baseURL.ResolveReference(parsedLink).String(), nil
}
