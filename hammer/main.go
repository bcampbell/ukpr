package main

import (
	"fmt"
	"github.com/bcampbell/ukpr/prscrape"
	"math/rand"
	"strings"
	"time"
)

// numbe of dummy scrapers to create
const NSCRAPERS int = 20

// number of dummy press releases emited by each scraper per run
const NRELEASES int = 100

func main() {
	rand.Seed(time.Now().UnixNano())

	prscrape.ServerMain("hammer.db", configure)
}

func configure(historical bool) []*prscrape.Scraper {
	out := make([]*prscrape.Scraper, 0)
	for i := 0; i < NSCRAPERS; i++ {
		out = append(out, NewHammerScraper(i))
	}

	return out
}

// scraper to hammer the db
func NewHammerScraper(sourcenum int) *prscrape.Scraper {
	name := fmt.Sprintf("hammer_%d", sourcenum)
	hammer := func() ([]*prscrape.PressRelease, error) {
		out := make([]*prscrape.PressRelease, 0, NRELEASES)
		for i := 0; i < NRELEASES; i++ {
			r := rand.Int()
			pr := prscrape.PressRelease{
				Type:      "press release",
				Title:     fmt.Sprintf("Test Press Release %d", r),
				Source:    name,
				Permalink: fmt.Sprintf("http://example.com/testpr/%d", r),
				Content:   strings.Repeat("blah blah blah blah blah blah blah blah blah blah blah blah blah blah blah\n", 20),
				PubDate:   time.Now(),
			}
			out = append(out, &pr)
		}
		return out, nil
	}
	s := prscrape.Scraper{
		name,
		hammer,
		nil}
	return &s
}
