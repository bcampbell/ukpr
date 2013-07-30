package ukscrapers

import (
	//	"errors"
	"github.com/bcampbell/ukpr/prscrape"
)

// scraper to grab Barclays press releases
type BarclaysScraper struct{}

func NewBarclaysScraper() *BarclaysScraper {
	var s BarclaysScraper
	return &s
}

func (scraper *BarclaysScraper) Name() string {
	return "barclays"
}

// fetches a list of latest press releases from Barclays
func (scraper *BarclaysScraper) FetchList() ([]*prscrape.PressRelease, error) {
	url := "http://www.newsroom.barclays.com/content/default.aspx?NewsAreaID=2"
	sel := ".individualResultListing h2 a"
	return prscrape.GenericFetchList(scraper.Name(), url, sel)
}

func (scraper *BarclaysScraper) Scrape(pr *prscrape.PressRelease, raw_html string) error {
	title := ".mainContent h1"
	content := ".mainContent .leadin, .mainContent .bodyCopy"
	cruft := ""
	pubDate := ".mainContent .titleDate"
	return prscrape.GenericScrape(scraper.Name(), pr, raw_html, title, content, cruft, pubDate)
}
