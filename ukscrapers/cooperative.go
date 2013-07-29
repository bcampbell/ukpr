package ukscrapers

import (
	//	"errors"
	"github.com/bcampbell/ukpr/prscrape"
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
func (scraper *CooperativeScraper) FetchList() ([]*prscrape.PressRelease, error) {
	url := "http://www.co-operative.coop/corporate/Press/Press-releases/"
	sel := "#divNewsList h2 a"
	return prscrape.GenericFetchList(scraper.Name(), url, sel)
}

func (scraper *CooperativeScraper) Scrape(pr *prscrape.PressRelease, raw_html string) error {
	title := "#ctl00_ctl00_Content_contentDiv h1"
	pubDate := "#ctl00_ctl00_Content_contentDiv .publishDate"
	content := "#ctl00_ctl00_Content_contentDiv"
	// TODO: kill everything after: "Additional Information:"
	cruft := "script, noscript, .TwitterTweetFacebookLike, .CrumbTrail, .main-content, .NewsItemDate, .NewsItemFooter, .sendToAFriendBelowContent"
	return prscrape.GenericScrape(scraper.Name(), pr, raw_html, title, content, cruft, pubDate)
}
