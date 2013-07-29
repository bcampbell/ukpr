package ukscrapers

import (
	//	"errors"
	"github.com/bcampbell/ukpr/prscrape"
)

// scraper to grab Asda press releases
type AsdaScraper struct{}

func NewAsdaScraper() *AsdaScraper {
	var s AsdaScraper
	return &s
}

func (scraper *AsdaScraper) Name() string {
	return "asda"
}

// fetches a list of latest press releases from Asda
func (scraper *AsdaScraper) FetchList() ([]*prscrape.PressRelease, error) {
	url := "http://your.asda.com/press-centre/"
	sel := "#main h2 a"
	return prscrape.GenericFetchList(scraper.Name(), url, sel)
}

func (scraper *AsdaScraper) Scrape(pr *prscrape.PressRelease, raw_html string) error {
	title := "#main .article-content .title h1"
	content := "#main .article-content .body"
	cruft := ""
	pubDate := "#main .article-content .posted-by"
	return prscrape.GenericScrape(scraper.Name(), pr, raw_html, title, content, cruft, pubDate)
}
