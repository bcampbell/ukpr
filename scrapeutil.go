package main

import (
	"code.google.com/p/cascadia"
	"code.google.com/p/go.net/html"
	"net/http"
)

// getAttr retrieved the value of an attribute on a node.
// Returns empty string if attribute doesn't exist.
func getAttr(n *html.Node, attr string) string {
	for _, a := range n.Attr {
		if a.Key == attr {
			return a.Val
		}
	}
	return ""
}

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

// GenericFetchList extracts links from a given page.
func GenericFetchList(scraperName, url, linkSelector string) ([]*PressRelease, error) {
	linkSel := cascadia.MustCompile(linkSelector)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	root, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err // TODO: wrap up as ScrapeError?
	}
	docs := make([]*PressRelease, 0)
	for _, a := range linkSel.MatchAll(root) {
		link := getAttr(a, "href")
		pr := PressRelease{Source: scraperName, Permalink: link}
		docs = append(docs, &pr)
	}
	return docs, nil
}
