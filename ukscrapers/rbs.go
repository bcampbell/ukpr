package ukscrapers

// TODO: RBS cms is really messed up. Scraper needs some more work.

import (
	//	"errors"
	"github.com/bcampbell/ukpr/prscrape"
)

// scraper to grab RBS press releases
type RBSScraper struct{}

func NewRBSScraper() *RBSScraper {
	var s RBSScraper
	return &s
}

func (scraper *RBSScraper) Name() string {
	return "rbs"
}

// fetches a list of latest press releases from RBS
func (scraper *RBSScraper) FetchList() ([]*prscrape.PressRelease, error) {

	//	foo := prscrape.PressRelease{Permalink: "http://www.rbs.com/news/2013/07/postcode-lending.html"}
	//	return []*prscrape.PressRelease{&foo}, nil
	url := "http://www.rbs.com/news.html"
	sel := ".news-row h3 a"
	return prscrape.GenericFetchList(scraper.Name(), url, sel)
}

func (scraper *RBSScraper) Scrape(pr *prscrape.PressRelease, raw_html string) error {
	// NOTE: RBS seem to serve different a couple of differing HTML formats...
	// not sure if they are browser sniffing or what...
	// The muppets.

	title := "article .title h2, #main h1"
	content := "article .rbs-rich-text, #main .mainpara"
	pubDate := "article .rbs-rich-text:first-child h2, #main .mainpara"
	cruft := pubDate
	return prscrape.GenericScrape(scraper.Name(), pr, raw_html, title, content, cruft, pubDate)
}
