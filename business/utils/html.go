package utils

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

func StripHTMLTags(htmlContent string) (string, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	extractText(doc, &buf)

	result := buf.String()
	result = strings.TrimSpace(result)
	return result, nil
}

func extractText(n *html.Node, buf *bytes.Buffer) {
	if n.Type == html.TextNode {
		buf.WriteString(strings.TrimSpace(n.Data))
		buf.WriteString(" ")
	}

	if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style") {
		return
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractText(c, buf)
	}

	if n.Type == html.ElementNode {
		switch n.Data {
		case "div", "p":
			buf.WriteString("\n")
		case "h1", "h2", "h3", "h4", "h5", "h6":
			buf.WriteString("\n")
		case "br", "li":
			buf.WriteString("\n")
		}
	}
}
