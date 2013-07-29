package ukscrapers

import (
	"github.com/bcampbell/ukpr/prscrape"
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
func (scraper *SeventyTwoPointScraper) FetchList() ([]*prscrape.PressRelease, error) {
	// (could also access archives, about 160 pages
	// eg    http://www.72point.com/coverage/page/2/)

	url := "http://www.72point.com/coverage/"
	sel := ".items .item .content .links a"
	return prscrape.GenericFetchList(scraper.Name(), url, sel)
}

func (scraper *SeventyTwoPointScraper) Scrape(pr *prscrape.PressRelease, raw_html string) error {
	title := "#content h3.title"
	content := "#content .item .content"
	cruft := ".addthis_toolbox"
	pubDate := "#content .item .meta"

	return prscrape.GenericScrape(scraper.Name(), pr, raw_html, title, content, cruft, pubDate)
}
