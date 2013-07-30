// This program runs a server which:
//  1) scrapes UK press releases
//  2) serves them up as HTTP server-sent
//  3) stashes them in a database for persistence, keeping a few days
//     worth of history (at least)
//
// For more details, see prscrape (which provides all implementation).
package main

import (
	"github.com/bcampbell/ukpr/prscrape"
	uk "github.com/bcampbell/ukpr/ukscrapers"
)

func main() {
	scrapers := [...]prscrape.Scraper{
		// supermarkets
		uk.NewTescoScraper(),
		uk.NewSeventyTwoPointScraper(),
		uk.NewAsdaScraper(),
		uk.NewWaitroseScraper(),
		uk.NewMarksAndSpencerScraper(),
		uk.NewSainsburysScraper(),
		uk.NewMorrisonsScraper(),
		uk.NewCooperativeScraper(),
		// banks
		uk.NewBarclaysScraper(),
		//		uk.NewRBSScraper(), // needs more work!
		uk.NewVirginMoneyScraper(),
		// Hotels
		uk.NewTravelLodgeScraper(),
		// Culture
		uk.NewTateScraper(),
	}
	prscrape.ServerMain(scrapers[:])
}
