package ukscrapers

import (
	//	"errors"
	"github.com/bcampbell/ukpr/prscrape"
)

// scraper to grab Tate press releases
type TateScraper struct{}

func NewTateScraper() *TateScraper {
	var s TateScraper
	return &s
}

func (scraper *TateScraper) Name() string {
	return "tate"
}

// fetches a list of latest press releases from Tate
func (scraper *TateScraper) FetchList() ([]*prscrape.PressRelease, error) {
	url := "http://www.tate.org.uk/about/press-office/releases"
	sel := ".tate-facet-search-result .result-title h3 a"
	return prscrape.GenericFetchList(scraper.Name(), url, sel)
}

func (scraper *TateScraper) Scrape(pr *prscrape.PressRelease, raw_html string) error {
	title := "#page-title"
	pubDate := "#block-tate-blocks-created-date"
	content := "#region-content article .field-name-body"
	cruft := ""

	return prscrape.GenericScrape(scraper.Name(), pr, raw_html, title, content, cruft, pubDate)
}
