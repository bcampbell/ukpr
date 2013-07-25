package main

import (
	"code.google.com/p/go.net/html"
)

// getTextContent recursively fetches the text for a node
func getTextContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	txt := ""
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		txt += getTextContent(child)
	}

	return txt
}

// contains returns true if is a descendant of container
func contains(container *html.Node, n *html.Node) bool {
	n = n.Parent
	for ; n != nil; n = n.Parent {
		if n == container {
			return true
		}
	}
	return false
}
