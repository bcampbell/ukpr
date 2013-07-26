package main

import (
	"bytes"
	"code.google.com/p/cascadia"
	"code.google.com/p/go.net/html"
	"net/http"

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

// fetches a list of latest press releases from tesco plc
func (scraper *SeventyTwoPointScraper) FetchList() ([]*PressRelease, error) {
	// (could also access archives, about 160 pages
	// eg    http://www.72point.com/coverage/page/2/)
	resp, err := http.Get("http://www.72point.com/coverage/")
	linkSel := cascadia.MustCompile(".items .item .content .links a")

	// the rest of this should be snipped out into a helper function.
	// Most scrapers can probably just use this as-is.
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
		pr := PressRelease{Source: scraper.Name(), Permalink: link}
		docs = append(docs, &pr)
	}
	return docs, nil
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
