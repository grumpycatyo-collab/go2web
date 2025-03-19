package utils

import (
	"bytes"
	"strings"

	"golang.org/x/net/html"
)

func ParseHTML(htmlContent string) (*html.Node, error) {
	return html.Parse(strings.NewReader(htmlContent))
}

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

func FindElementsByClass(n *html.Node, className string) []*html.Node {
	var elements []*html.Node

	var finder func(*html.Node)
	finder = func(node *html.Node) {
		if node.Type == html.ElementNode {
			for _, attr := range node.Attr {
				if attr.Key == "class" && containsClass(attr.Val, className) {
					elements = append(elements, node)
					break
				}
			}
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			finder(c)
		}
	}

	finder(n)
	return elements
}

func FindElements(n *html.Node, tagName string) []*html.Node {
	var elements []*html.Node

	var finder func(*html.Node)
	finder = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == tagName {
			elements = append(elements, node)
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			finder(c)
		}
	}

	finder(n)
	return elements
}

func FindElementWithin(n *html.Node, tagName string) *html.Node {
	if n == nil {
		return nil
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == tagName {
			return c
		}

		if result := FindElementWithin(c, tagName); result != nil {
			return result
		}
	}

	return nil
}

func FindElementsWithin(n *html.Node, tagName string) []*html.Node {
	var elements []*html.Node

	if n == nil {
		return elements
	}

	var finder func(*html.Node)
	finder = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == tagName {
			elements = append(elements, node)
		}

		for c := node.FirstChild; c != nil; c = c.NextSibling {
			finder(c)
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		finder(c)
	}

	return elements
}

func FindElementByClassWithin(n *html.Node, className string) *html.Node {
	if n == nil {
		return nil
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode {
			for _, attr := range c.Attr {
				if attr.Key == "class" && containsClass(attr.Val, className) {
					return c
				}
			}
		}

		if result := FindElementByClassWithin(c, className); result != nil {
			return result
		}
	}

	return nil
}

func GetAttribute(n *html.Node, attrName string) string {
	if n == nil {
		return ""
	}

	for _, attr := range n.Attr {
		if attr.Key == attrName {
			return attr.Val
		}
	}

	return ""
}

func GetTextContent(n *html.Node) string {
	if n == nil {
		return ""
	}

	var buf bytes.Buffer
	extractText(n, &buf)
	return strings.TrimSpace(buf.String())
}

func GetParent(n *html.Node) *html.Node {
	if n == nil {
		return nil
	}
	return n.Parent
}

func containsClass(classes, className string) bool {
	for _, class := range strings.Fields(classes) {
		if class == className {
			return true
		}
	}
	return false
}

func RenderNode(n *html.Node) string {
	var buf bytes.Buffer
	err := html.Render(&buf, n)
	if err != nil {
		return ""
	}
	return buf.String()
}
