package main

import (
	"errors"
	"strings"
)

// scraper to grab Morrisons press releases
type MorrisonsScraper struct{}

func NewMorrisonsScraper() *MorrisonsScraper {
	var s MorrisonsScraper
	return &s
}

func (scraper *MorrisonsScraper) Name() string {
	return "morrisons"
}

// TODO: morrisons press releases don't have dates on the individual pages.
// should extract dates during FetchList()

// fetches a list of latest press releases from Morrisons
func (scraper *MorrisonsScraper) FetchList() ([]*PressRelease, error) {
	url := "http://www.morrisons-corporate.com/Media-centre/News-archive/"
	sel := ".news_summary_noimage h4 a"
	return GenericFetchList(scraper.Name(), url, sel)
}

func (scraper *MorrisonsScraper) Scrape(pr *PressRelease, raw_html string) error {
	if strings.Contains(raw_html, "<title>Sorry, page not available") {
		return errors.New("Borked link")
	}

	title := ".morrisons-header h2"
	content := ".morrisons-content .inside_left_block"
	cruft := "script, .button_divider, .featured_funnels, .block2Inner"
	pubDate := ""
	return GenericScrape(scraper.Name(), pr, raw_html, title, content, cruft, pubDate)
}
