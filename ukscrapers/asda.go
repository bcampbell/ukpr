package ukscrapers

import (
	"github.com/bcampbell/ukpr/prscrape"
)

// scraper to grab Asda press releases
func NewAsdaScraper() prscrape.Scraper {
	s := prscrape.DefaultScraper{
		ScraperName: "asda",
		IndexURL:    "http://your.asda.com/press-centre/",
		LinkSel:     "#main h2 a",
		TitleSel:    "#main .article-content .title h1",
		ContentSel:  "#main .article-content .body",
		PubdateSel:  "#main .article-content .posted-by",
	}
	return &s
}
