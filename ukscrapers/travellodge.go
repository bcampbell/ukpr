package ukscrapers

import (
	//	"errors"
	"github.com/bcampbell/ukpr/prscrape"
)

// scraper to grab TravelLodge press releases
type TravelLodgeScraper struct{}

func NewTravelLodgeScraper() *TravelLodgeScraper {
	var s TravelLodgeScraper
	return &s
}

func (scraper *TravelLodgeScraper) Name() string {
	return "travellodge"
}

// fetches a list of latest press releases from TravelLodge
func (scraper *TravelLodgeScraper) FetchList() ([]*prscrape.PressRelease, error) {
	// just the most recent
	url := "http://www.travelodge.co.uk/news/category/press-releases/"
	sel := ".hentry h2 a"
	return prscrape.GenericFetchList(scraper.Name(), url, sel)
}

func (scraper *TravelLodgeScraper) Scrape(pr *prscrape.PressRelease, raw_html string) error {
	title := ".hentry h1"
	pubDate := ".hentry .date"
	content := ".hentry .post-box"
	cruft := "script, address"

	return prscrape.GenericScrape(scraper.Name(), pr, raw_html, title, content, cruft, pubDate)
}
