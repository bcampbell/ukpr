package main

import (
//	"errors"
)

// scraper to grab MarksAndSpencer press releases
type MarksAndSpencerScraper struct{}

func NewMarksAndSpencerScraper() *MarksAndSpencerScraper {
	var s MarksAndSpencerScraper
	return &s
}

func (scraper *MarksAndSpencerScraper) Name() string {
	return "marksandspencer"
}

// fetches a list of latest press releases from MarksAndSpencer
func (scraper *MarksAndSpencerScraper) FetchList() ([]*PressRelease, error) {
	url := "http://corporate.marksandspencer.com/media/press_releases"
	sel := "#press-releases .item h2 a"
	return GenericFetchList(scraper.Name(), url, sel)
}

func (scraper *MarksAndSpencerScraper) Scrape(pr *PressRelease, raw_html string) error {
	title := "#main h2"
	content := "#pr_article"
	// TODO: kill everything after: "-ENDS-"
	cruft := "p.back-top, p.reference"
	pubDate := "#main" // TODO: a more specific selector would be nice!
	return GenericScrape(scraper.Name(), pr, raw_html, title, content, cruft, pubDate)
}
