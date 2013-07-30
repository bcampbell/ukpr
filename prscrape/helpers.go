package prscrape

// generic scraping functions for PR scrapers to use

import (
	"bytes"
	"code.google.com/p/cascadia"
	"code.google.com/p/go.net/html"
	"github.com/bcampbell/fuzzytime"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GenericFetchList fetches a page, and extracts matching links.
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
		link, err := page.Parse(GetAttr(a, "href")) // extend to absolute url if needed
		if err != nil {
			// TODO: log a warning?
			continue
		}
		pr := PressRelease{Source: scraperName, Permalink: link.String()}
		docs = append(docs, &pr)
	}
	return docs, nil
}

// GenericScrape scrapes a press release from raw_html based on a bunch of css selector strings
func GenericScrape(source string, pr *PressRelease, raw_html, title, content, cruft, pubDate string) error {

	r := strings.NewReader(string(raw_html))
	root, err := html.Parse(r)
	if err != nil {
		return err // TODO: wrap up as ScrapeError?
	}

	pr.Source = source

	// title
	titleSel := cascadia.MustCompile(title)
	pr.Title = CompressSpace(GetTextContent(titleSel.MatchAll(root)[0]))

	// pubdate - only needs to contain a valid date string, doesn't matter
	// if there's other crap in there too.
	if pubDate != "" {
		pubDateSel := cascadia.MustCompile(pubDate)
		dateTxt := GetTextContent(pubDateSel.MatchAll(root)[0])
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
	contentSel := cascadia.MustCompile(content)
	contentElements := contentSel.MatchAll(root)
	if cruft != "" {
		cruftSel := cascadia.MustCompile(cruft)
		for _, el := range contentElements {
			for _, cruft := range cruftSel.MatchAll(el) {
				cruft.Parent.RemoveChild(cruft)
			}
		}
	}
	var out bytes.Buffer
	for _, el := range contentElements {
		err = html.Render(&out, el)
		if err != nil {
			return err
		}
	}

	pr.Content = out.String()
	return nil
}
