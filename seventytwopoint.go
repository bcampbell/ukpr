package main

import (
	"bytes"
	"code.google.com/p/cascadia"
	"code.google.com/p/go.net/html"
	"github.com/bcampbell/fuzzytime"
	"strings"
)

// scraper to grab 72point press releases
type SeventyTwoPointScraper struct{}

func NewSeventyTwoPointScraper() *SeventyTwoPointScraper {
	var s SeventyTwoPointScraper
	return &s
}

func (scraper *SeventyTwoPointScraper) Name() string {
	return "72point"
}

// fetches a list of latest press releases from 72point
func (scraper *SeventyTwoPointScraper) FetchList() ([]*PressRelease, error) {
	// (could also access archives, about 160 pages
	// eg    http://www.72point.com/coverage/page/2/)

	url := "http://www.72point.com/coverage/"
	sel := ".items .item .content .links a"
	return GenericFetchList(scraper.Name(), url, sel)
}

func (scraper *SeventyTwoPointScraper) Scrape(pr *PressRelease, raw_html string) error {
	titleSel := cascadia.MustCompile("#content h3.title")
	contentSel := cascadia.MustCompile("#content .item .content")
	cruftSel := cascadia.MustCompile(".addthis_toolbox")
	pubDateSel := cascadia.MustCompile("#content .item .meta")

	pr.Source = "72point"

	// the rest of this should be snipped out into a helper function.
	// Most scrapers can probably just use this as-is.
	r := strings.NewReader(string(raw_html))
	root, err := html.Parse(r)
	if err != nil {
		return err // TODO: wrap up as ScrapeError?
	}
	// content
	contentEl := contentSel.MatchAll(root)[0]
	for _, cruft := range cruftSel.MatchAll(contentEl) {
		cruft.Parent.RemoveChild(cruft)
	}

	var out bytes.Buffer
	err = html.Render(&out, contentEl)
	if err != nil {
		return err
	}
	pr.Content = out.String()

	// title
	pr.Title = getTextContent(titleSel.MatchAll(root)[0])

	// pubdate
	dateTxt := getTextContent(pubDateSel.MatchAll(root)[0])
	pr.PubDate, err = fuzzytime.Parse(dateTxt)
	if err != nil {
		return err
	}

	return nil
}
