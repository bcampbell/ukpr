package main

import (
//	"errors"
)

// scraper to grab Cooperative press releases
type CooperativeScraper struct{}

func NewCooperativeScraper() *CooperativeScraper {
	var s CooperativeScraper
	return &s
}

func (scraper *CooperativeScraper) Name() string {
	return "cooperative"
}

// fetches a list of latest press releases from Cooperative
func (scraper *CooperativeScraper) FetchList() ([]*PressRelease, error) {
	url := "http://www.co-operative.coop/corporate/Press/Press-releases/"
	sel := "#divNewsList h2 a"
	return GenericFetchList(scraper.Name(), url, sel)
}

func (scraper *CooperativeScraper) Scrape(pr *PressRelease, raw_html string) error {
	title := "#ctl00_ctl00_Content_contentDiv h1"
	pubDate := "#ctl00_ctl00_Content_contentDiv .publishDate"
	content := "#ctl00_ctl00_Content_contentDiv"
	// TODO: kill everything after: "Additional Information:"
	cruft := "script, noscript, .TwitterTweetFacebookLike, .CrumbTrail, .main-content, .NewsItemDate, .NewsItemFooter, .sendToAFriendBelowContent"
	return GenericScrape(scraper.Name(), pr, raw_html, title, content, cruft, pubDate)
}
