package prscrape

import (
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
	// if this is a fully-filled out press release, complete is set
	complete bool
}

// Scraper is the interface to implement to add a new scraper to the system
// (although in practice, they're all likely to be ComposedScraper :-)
// TODO: dump Scraper interface entirely and just use ComposedScraper instead?
type Scraper interface {
	Name() string

	// Fetch a list of 'current' press releases.
	// (via RSS feed, or by scraping an index page or whatever)
	// The results are passed back as PressRelease structs. At the very least,
	// the Permalink field must be set to the URL of the press release,
	// But there's no reason Discover() can't fill out all the fields if the
	// data is available (eg some rss feeds have everything required).
	// For incomplete PressReleases, the framework will fetch the HTML from
	// the Permalink URL, and invoke Scrape() to complete the data.
	Discover() ([]*PressRelease, error)

	// scrape a single press release from raw html passed in as a string
	Scrape(*PressRelease, string) error
}

type ConfigureFunc func() []Scraper

type DiscoverFunc func() ([]*PressRelease, error)
type ScrapeFunc func(pr *PressRelease, rawHTML string) error

// ComposedScraper lets you pick-and-mix various discover and scrape functions
type ComposedScraper struct {
	ScraperName string
	DoDiscover  DiscoverFunc
	DoScrape    ScrapeFunc
}

func (scraper *ComposedScraper) Name() string { return scraper.ScraperName }
func (scraper *ComposedScraper) Discover() (found []*PressRelease, err error) {
	return scraper.DoDiscover()
}
func (scraper *ComposedScraper) Scrape(pr *PressRelease, rawHTML string) (err error) {
	return scraper.DoScrape(pr, rawHTML)
}
