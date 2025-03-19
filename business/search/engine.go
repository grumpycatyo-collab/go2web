package search

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/grumpycatyo-collab/go2web/business/http"
	"golang.org/x/net/html"
)

type SearchResult struct {
	Title       string
	URL         string
	Description string
}

func PerformSearch(client *http.Client, term string) ([]SearchResult, error) {
	encodedTerm := url.QueryEscape(term)

	searchURL := fmt.Sprintf("https://lite.duckduckgo.com/lite?q=%s", encodedTerm)

	resp, err := client.Get(searchURL)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}

	results, err := parseSearchResults(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to extract search results: %w", err)
	}

	if len(results) > 10 {
		results = results[:10]
	}

	return results, nil
}

func parseSearchResults(htmlContent string) ([]SearchResult, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	var results []SearchResult

	var findResultLinks func(*html.Node)
	findResultLinks = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			isResultLink := false
			var href string

			for _, attr := range n.Attr {
				if attr.Key == "class" && attr.Val == "result-link" {
					isResultLink = true
				}
				if attr.Key == "href" {
					href = attr.Val
				}
			}

			if isResultLink && href != "" {

				title := extractTextContent(n)

				resultRow := findParentByTagName(n, "tr")
				if resultRow != nil {
					var description, originalURL string

					nextRow := findNextSibling(resultRow)
					if nextRow != nil {
						snippetCell := findElementByClass(nextRow, "result-snippet")
						if snippetCell != nil {
							description = extractTextContent(snippetCell)
						}
					}

					urlRow := findNextSibling(nextRow)
					if urlRow != nil {
						linkTextSpan := findElementByClass(urlRow, "link-text")
						if linkTextSpan != nil {
							originalURL = extractTextContent(linkTextSpan)
						}
					}

					actualURL := extractActualURL(href)
					if actualURL == "" {
						actualURL = originalURL
					}

					if title != "" && actualURL != "" {
						results = append(results, SearchResult{
							Title:       title,
							URL:         actualURL,
							Description: description,
						})
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findResultLinks(c)
		}
	}

	findResultLinks(doc)
	return results, nil
}

func extractTextContent(n *html.Node) string {
	if n == nil {
		return ""
	}

	var sb strings.Builder

	var extract func(*html.Node)
	extract = func(node *html.Node) {
		if node.Type == html.TextNode {
			sb.WriteString(node.Data)
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}

	extract(n)
	return strings.TrimSpace(sb.String())
}

func findParentByTagName(n *html.Node, tagName string) *html.Node {
	if n == nil {
		return nil
	}

	current := n.Parent
	for current != nil {
		if current.Type == html.ElementNode && current.Data == tagName {
			return current
		}
		current = current.Parent
	}

	return nil
}

func findNextSibling(n *html.Node) *html.Node {
	if n == nil {
		return nil
	}

	current := n.NextSibling
	for current != nil {
		if current.Type == html.ElementNode {
			return current
		}
		current = current.NextSibling
	}

	return nil
}

func findElementByClass(n *html.Node, className string) *html.Node {
	if n == nil {
		return nil
	}

	var find func(*html.Node) *html.Node
	find = func(node *html.Node) *html.Node {
		if node.Type == html.ElementNode {
			for _, attr := range node.Attr {
				if attr.Key == "class" && attr.Val == className {
					return node
				}
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if result := find(c); result != nil {
				return result
			}
		}

		return nil
	}

	return find(n)
}

func extractActualURL(ddgURL string) string {
	if strings.Contains(ddgURL, "duckduckgo.com/l/?uddg=") {
		parts := strings.Split(ddgURL, "uddg=")
		if len(parts) > 1 {
			encodedURL := parts[1]
			if idx := strings.Index(encodedURL, "&"); idx > 0 {
				encodedURL = encodedURL[:idx]
			}

			if decodedURL, err := url.QueryUnescape(encodedURL); err == nil {
				return decodedURL
			}
		}
	}

	if !strings.HasPrefix(ddgURL, "//") {
		return ddgURL
	}

	return "https:" + ddgURL
}
