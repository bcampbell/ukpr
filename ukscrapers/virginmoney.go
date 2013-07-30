package ukscrapers

// TODO: VirginMoney cms is really messed up. Scraper needs some more work.

import (
	//	"errors"
	"github.com/bcampbell/ukpr/prscrape"
)

// scraper to grab VirginMoney press releases
type VirginMoneyScraper struct{}

func NewVirginMoneyScraper() *VirginMoneyScraper {
	var s VirginMoneyScraper
	return &s
}

func (scraper *VirginMoneyScraper) Name() string {
	return "virginmoney"
}

// fetches a list of latest press releases from VirginMoney
func (scraper *VirginMoneyScraper) FetchList() ([]*prscrape.PressRelease, error) {
	// just the most recent
	url := "http://uk.virginmoney.com/virgin/news-centre/"
	sel := ".section>.padding .content:nth-child(2) p>a"
	/*
		// full archive
		url := "http://uk.virginmoney.com/virgin/news-centre/press-releases.jsp"
		sel := ".section>.padding>.content a"
	*/
	return prscrape.GenericFetchList(scraper.Name(), url, sel)
}

func (scraper *VirginMoneyScraper) Scrape(pr *prscrape.PressRelease, raw_html string) error {
	title := ".section>.padding>.content h2"
	pubDate := ".section>.padding>.content .prdate"
	content := ".section>.padding>.content"
	cruft := title + ", " + pubDate

	// TODO: stop content at "ENDS" or " - ENDS - "
	return prscrape.GenericScrape(scraper.Name(), pr, raw_html, title, content, cruft, pubDate)
}
