package prscrape

// utility functions to help out with scraping html

import (
	"code.google.com/p/go.net/html"
	//"fmt"
	"regexp"
	"strings"
)

// GetAttr retrieved the value of an attribute on a node.
// Returns empty string if attribute doesn't exist.
func GetAttr(n *html.Node, attr string) string {
	for _, a := range n.Attr {
		if a.Key == attr {
			return a.Val
		}
	}
	return ""
}

// GetTextContent recursively fetches the text for a node
func GetTextContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	txt := ""
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		txt += GetTextContent(child)
	}

	return txt
}

// Contains returns true if n is a descendant of container
func Contains(container *html.Node, n *html.Node) bool {
	n = n.Parent
	for ; n != nil; n = n.Parent {
		if n == container {
			return true
		}
	}
	return false
}

// CompressSpace reduces all whitespace sequences (space, tabs, newlines etc) in a string to a single space.
// Leading/trailing space is trimmed.
// Has the effect of converting multiline strings to one line.
func CompressSpace(s string) string {
	multispacePat := regexp.MustCompile(`[\s]+`)
	s = strings.TrimSpace(multispacePat.ReplaceAllLiteralString(s, " "))
	return s
}

func StripComments(n *html.Node) {
	if n.Type == html.CommentNode {
		n.Parent.RemoveChild(n)
		return
	}

	child := n.FirstChild
	for child != nil {
		next := child.NextSibling
		StripComments(child)
		child = next
	}
}

var inlineTags = map[string]bool{"a": true, "abbr": true, "acronym": true, "b": true, "basefont": true, "bdo": true, "big": true,
	"br":   true,
	"cite": true, "code": true, "dfn": true, "em": true, "font": true, "i": true, "img": true, "input": true,
	"kbd": true, "label": true, "q": true, "s": true, "samp": true, "select": true, "small": true, "span": true,
	"strike": true, "strong": true, "sub": true, "sup": true, "textarea": true, "tt": true, "u": true, "var": true,
	"applet": true, "button": true, "del": true, "iframe": true, "ins": true, "map": true, "object": true,
	"script": true}

// RenderText returns the text, using whitespace and line breaks
func innerRenderText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}

	if n.Type != html.ElementNode {
		return ""
	}

	txt := ""
	tag := strings.ToLower(n.DataAtom.String())
	_, inline := inlineTags[tag]
	if !inline {
		txt += "\n"
	}
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		txt += innerRenderText(child)
	}

	if tag == "br" || !inline {
		txt += "\n"
	}
	// TODO: strip out excessive whitespace
	return txt
}

func RenderText(n *html.Node) string {
	txt := innerRenderText(n)
	txt = regexp.MustCompile(`[\r\n]\s+[\r\n]`).ReplaceAllLiteralString(txt, "\n\n")
	txt = regexp.MustCompile(`[\r\n]{2,}`).ReplaceAllLiteralString(txt, "\n\n")
	return txt
}
