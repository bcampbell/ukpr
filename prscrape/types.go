package prscrape

import (
	"code.google.com/p/go.net/html"
	"time"
)

// PressRelease is the data we're scraping and storing.
// TODO: support multiple urls
type PressRelease struct {
	Title     string    `json:"title"`
	Source    string    `json:"source"`
	Permalink string    `json:"permalink"`
	PubDate   time.Time `json:"published"`
	Content   string    `json:"text"`
	Type      string    `json:"type"`
}

type ConfigureFunc func(historical bool) []*Scraper

// DiscoverFunc is for fetching a list of 'current' press releases.
// (via RSS feed, or by scraping an index page or whatever)
// The results are passed back as PressRelease structs. At the very least,
// the Permalink field must be set to the URL of the press release,
// But there's no reason Discover() can't fill out all the fields if the
// data is available (eg some rss feeds have everything required).
// For incomplete PressReleases, the framework will fetch the HTML from
// the Permalink URL, and invoke Scrape() to complete the data.
type DiscoverFunc func() ([]*PressRelease, error)

// ScrapeFunc is for scraping a single press release from html
type ScrapeFunc func(pr *PressRelease, doc *html.Node) error

// ComposedScraper lets you pick-and-mix various discover and scrape functions
type Scraper struct {
	Name     string
	Discover DiscoverFunc
	Scrape   ScrapeFunc
}
