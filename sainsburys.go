package main

import (
//	"errors"
)

// scraper to grab Sainsburys press releases
type SainsburysScraper struct{}

func NewSainsburysScraper() *SainsburysScraper {
	var s SainsburysScraper
	return &s
}

func (scraper *SainsburysScraper) Name() string {
	return "sainsburys"
}

// fetches a list of latest press releases from Sainsburys
func (scraper *SainsburysScraper) FetchList() ([]*PressRelease, error) {
	url := "http://www.j-sainsbury.co.uk/media/latest-stories/"
	sel := "#content_container a.title"
	return GenericFetchList(scraper.Name(), url, sel)
}

func (scraper *SainsburysScraper) Scrape(pr *PressRelease, raw_html string) error {
	title := "#page_container h1"
	content := "#page_container .richTextFormat"
	// TODO: kill everything after: "Notes to Editors"
	cruft := ""
	pubDate := "#page_container .nm_right .list_plain, #page_container .blog_author"
	return GenericScrape(scraper.Name(), pr, raw_html, title, content, cruft, pubDate)
}
