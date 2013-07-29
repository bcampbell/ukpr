package main

import (
	"bytes"
	"code.google.com/p/cascadia"
	"code.google.com/p/go.net/html"
	"github.com/bcampbell/fuzzytime"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
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

// compressSpace reduces all whitespace sequences (space, tabs, newlines etc) in a string to a single space.
// Leading/trailing space is trimmed.
// Has the effect of converting multiline strings to one line.
func compressSpace(s string) string {
	multispacePat := regexp.MustCompile(`[\s]+`)
	s = strings.TrimSpace(multispacePat.ReplaceAllLiteralString(s, " "))
	return s
}

// GenericFetchList extracts links from a given page.
func GenericFetchList(scraperName, pageUrl, linkSelector string) ([]*PressRelease, error) {
	page, err := url.Parse(pageUrl)
	if err != nil {
		return nil, err // TODO: wrap up as ScrapeError?
	}

	linkSel := cascadia.MustCompile(linkSelector)
	resp, err := http.Get(pageUrl)
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
		link, err := page.Parse(getAttr(a, "href")) // extend to absolute url if needed
		if err != nil {
			// TODO: log a warning?
			continue
		}
		pr := PressRelease{Source: scraperName, Permalink: link.String()}
		docs = append(docs, &pr)
	}
	return docs, nil
}

// scrape a press release based on a bunch of css selector strings
func GenericScrape(source string, pr *PressRelease, raw_html, title, content, cruft, pubDate string) error {
	titleSel := cascadia.MustCompile(title)
	contentSel := cascadia.MustCompile(content)

	r := strings.NewReader(string(raw_html))
	root, err := html.Parse(r)
	if err != nil {
		return err // TODO: wrap up as ScrapeError?
	}

	pr.Source = source

	// title
	pr.Title = compressSpace(getTextContent(titleSel.MatchAll(root)[0]))

	// pubdate - only needs to contain a valid date string, doesn't matter
	// if there's other crap in there too.
	if pubDate != "" {
		pubDateSel := cascadia.MustCompile(pubDate)
		dateTxt := getTextContent(pubDateSel.MatchAll(root)[0])
		pr.PubDate, err = fuzzytime.Parse(dateTxt)
		if err != nil {
			return err
		}
	} else {
		// if time isn't already set, just fudge using current time
		if pr.PubDate.IsZero() {
			pr.PubDate = time.Now()
		}
	}

	// content
	contentEl := contentSel.MatchAll(root)[0]
	if cruft != "" {
		cruftSel := cascadia.MustCompile(cruft)
		for _, cruft := range cruftSel.MatchAll(contentEl) {
			cruft.Parent.RemoveChild(cruft)
		}
	}
	var out bytes.Buffer
	err = html.Render(&out, contentEl)
	if err != nil {
		return err
	}
	pr.Content = out.String()
	return nil
}
