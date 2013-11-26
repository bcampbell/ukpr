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

// NOTE: currently in the process of migrating (back to!) a config-driven
// system, so most of this can be data driven, with a the odd custom function
// to handle particularly annoying sites.

func main() {
	prscrape.ServerMain(configure)
}

func configure(historical bool) []prscrape.Scraper {
	out := []prscrape.Scraper{
		// supermarkets
		uk.NewTescoScraper(), // our only custom scraper
		NewAsdaScraper(),
		NewWaitroseScraper(),
		NewMarksAndSpencerScraper(),
		NewSainsburysScraper(),
		NewMorrisonsScraper(),
		NewCooperativeScraper(),
		// banks
		NewBarclaysScraper(),
		//		uk.NewRBSScraper(), // needs more work!
		NewVirginMoneyScraper(),
		// Hotels
		NewTravelLodgeScraper(),
		// Culture
		NewTateScraper(),
		// Government
		NewGovUKAnnounceScraper(),
		// Science
		NewEurekalertScraper(),
		// pr general
		NewPRWebUKScraper(),
		NewPRNewsWireUKScraper(),
		// thinktanks
		NewPolicyExchangeScraper(),
		NewMigrationWatchScraper(),
		NewTaxpayersAllianceScraper(),
		// charities
		NewGreenpeaceUKScraper(),
		NewShelterScraper(),
	}

	if historical {
		out = append(out, NewHistoricalSeventyTwoPointScraper())
	} else {
		out = append(out, NewSeventyTwoPointScraper())
	}
	return out
}

// scraper to grab Asda press releases
func NewAsdaScraper() prscrape.Scraper {
	name := "asda"
	s := prscrape.ComposedScraper{
		name,
		prscrape.MustBuildGenericDiscover(name, "http://your.asda.com/press-centre/", "#main h2 a", false),
		prscrape.MustBuildGenericScrape(name,
			"#main .article-content .title h1",
			"#main .article-content .body",
			"",
			"#main .article-content .posted-by"),
	}
	return &s
}

// scraper to grab press releases from Barclays
func NewBarclaysScraper() prscrape.Scraper {
	name := "barclays"
	url := "http://www.newsroom.barclays.com/content/default.aspx?NewsAreaID=2"
	linkSel := ".individualResultListing h2 a"
	title := ".mainContent h1"
	content := ".mainContent .leadin, .mainContent .bodyCopy"
	cruft := ""
	pubDate := ".mainContent .titleDate"

	return &prscrape.ComposedScraper{
		name,
		prscrape.MustBuildGenericDiscover(name, url, linkSel, false),
		prscrape.MustBuildGenericScrape(name, title, content, cruft, pubDate),
	}
}

// scraper to grab press releases from Virgin Money
func NewVirginMoneyScraper() prscrape.Scraper {
	name := "virginmoney"
	url := "http://uk.virginmoney.com/virgin/news-centre/"
	linkSel := ".section>.padding .content:nth-child(2) p>a"

	title := ".section>.padding>.content h2"
	pubDate := ".section>.padding>.content .prdate"
	content := ".section>.padding>.content"
	cruft := title + ", " + pubDate
	// TODO: stop content at "ENDS" or " - ENDS - "

	return &prscrape.ComposedScraper{
		name,
		prscrape.MustBuildGenericDiscover(name, url, linkSel, false),
		prscrape.MustBuildGenericScrape(name, title, content, cruft, pubDate),
	}
}

// scraper to grab press releases from RBS
func NewRBSScraper() prscrape.Scraper {
	name := "rbs"
	url := "http://www.rbs.com/news.html"
	linkSel := ".news-row h3 a"

	// NOTE: RBS seem to serve different a couple of differing HTML formats...
	// not sure if they are browser sniffing or what...
	// The muppets.

	title := "article .title h2, #main h1"
	content := "article .rbs-rich-text, #main .mainpara"
	pubDate := "article .rbs-rich-text:first-child h2, #main .mainpara"
	cruft := pubDate

	return &prscrape.ComposedScraper{
		name,
		prscrape.MustBuildGenericDiscover(name, url, linkSel, false),
		prscrape.MustBuildGenericScrape(name, title, content, cruft, pubDate),
	}
}

// scraper to grab press releases from 72point
func NewSeventyTwoPointScraper() prscrape.Scraper {
	name := "72point"
	url := "http://www.72point.com/coverage/"
	linkSel := ".items .item .content .links a"

	title := "#content h3.title"
	content := "#content .item .content"
	cruft := ".addthis_toolbox"
	pubDate := "#content .item .meta"

	return &prscrape.ComposedScraper{
		name,
		prscrape.MustBuildGenericDiscover(name, url, linkSel, false),
		prscrape.MustBuildGenericScrape(name, title, content, cruft, pubDate),
	}
}

// scraper to grab press releases from 72point, including the whole historical archive
func NewHistoricalSeventyTwoPointScraper() prscrape.Scraper {
	// standard scraper, but replace the discover function
	name := "72point"
	url := "http://www.72point.com/coverage/"
	linkSel := ".items .item .content .links a"
	nextPageSel := "#system .pagination a.next"

	title := "#content h3.title"
	content := "#content .item .content"
	cruft := ".addthis_toolbox"
	pubDate := "#content .item .meta"

	return &prscrape.ComposedScraper{
		name,
		prscrape.MustBuildPaginatedGenericDiscover(name, url, nextPageSel, linkSel),
		prscrape.MustBuildGenericScrape(name, title, content, cruft, pubDate),
	}
}

// scraper to grab press releases from the Tate
func NewTateScraper() prscrape.Scraper {
	name := "tate"
	url := "http://www.tate.org.uk/about/press-office/releases"
	linkSel := ".tate-facet-search-result .result-title h3 a"
	title := "#page-title"
	pubDate := "#block-tate-blocks-created-date"
	content := "#region-content article .field-name-body"
	cruft := ""

	return &prscrape.ComposedScraper{
		name,
		prscrape.MustBuildGenericDiscover(name, url, linkSel, false),
		prscrape.MustBuildGenericScrape(name, title, content, cruft, pubDate),
	}
}

// scraper to grab press releases from Morrisons
func NewMorrisonsScraper() prscrape.Scraper {
	// TODO: morrisons press releases don't have dates on the individual pages.
	// should extract dates during discovery
	name := "morrisons"
	url := "http://www.morrisons-corporate.com/Media-centre/News-archive/"
	linkSel := ".news_summary_noimage h4 a"

	title := ".morrisons-header h2"
	content := ".morrisons-content .inside_left_block"
	cruft := "script, .button_divider, .featured_funnels, .block2Inner"
	pubDate := ""

	return &prscrape.ComposedScraper{
		name,
		prscrape.MustBuildGenericDiscover(name, url, linkSel, false),
		prscrape.MustBuildGenericScrape(name, title, content, cruft, pubDate),
	}
}

// scraper to grab press releases from MarksAndSpencer
func NewMarksAndSpencerScraper() prscrape.Scraper {
	name := "marksandspencer"
	url := "http://corporate.marksandspencer.com/media/press_releases"
	linkSel := "#press-releases .item h2 a"
	title := "#main h2"
	content := "#pr_article"
	// TODO: kill everything after: "-ENDS-"
	cruft := "p.back-top, p.reference"
	pubDate := "#main" // TODO: a more specific selector would be nice!

	return &prscrape.ComposedScraper{
		name,
		prscrape.MustBuildGenericDiscover(name, url, linkSel, false),
		prscrape.MustBuildGenericScrape(name, title, content, cruft, pubDate),
	}
}

// scraper to grab press releases from TravelLodge
func NewTravelLodgeScraper() prscrape.Scraper {
	name := "travellodge"
	url := "http://www.travelodge.co.uk/news/category/press-releases/"
	linkSel := ".hentry h2 a"

	title := ".hentry h1"
	pubDate := ".hentry .date"
	content := ".hentry .post-box"
	cruft := "script, address"

	return &prscrape.ComposedScraper{
		name,
		prscrape.MustBuildGenericDiscover(name, url, linkSel, false),
		prscrape.MustBuildGenericScrape(name, title, content, cruft, pubDate),
	}
}

// scraper to grab press releases from Cooperative
func NewCooperativeScraper() prscrape.Scraper {
	name := "cooperative"
	url := "http://www.co-operative.coop/corporate/Press/Press-releases/"
	linkSel := "#divNewsList h2 a"

	title := "#ctl00_ctl00_Content_contentDiv h1"
	pubDate := "#ctl00_ctl00_Content_contentDiv .publishDate"
	content := "#ctl00_ctl00_Content_contentDiv"
	// TODO: kill everything after: "Additional Information:"
	cruft := "script, noscript, .TwitterTweetFacebookLike, .CrumbTrail, .main-content, .NewsItemDate, .NewsItemFooter, .sendToAFriendBelowContent"

	return &prscrape.ComposedScraper{
		name,
		prscrape.MustBuildGenericDiscover(name, url, linkSel, false),
		prscrape.MustBuildGenericScrape(name, title, content, cruft, pubDate),
	}
}

// scraper to grab press releases from Waitrose
func NewWaitroseScraper() prscrape.Scraper {
	name := "waitrose"
	url := "http://www.waitrose.presscentre.com/content/default.aspx?NewsAreaID=2"
	linkSel := "#content .main .item h3 a"
	title := "#content h1"
	content := "#content .main .bodyCopy"
	// TODO: kill everything after: "-ENDS-"
	cruft := ""
	pubDate := "#content .date_release"

	return &prscrape.ComposedScraper{
		name,
		prscrape.MustBuildGenericDiscover(name, url, linkSel, false),
		prscrape.MustBuildGenericScrape(name, title, content, cruft, pubDate),
	}
}

// scraper to grab press releases from Sainsburys
func NewSainsburysScraper() prscrape.Scraper {
	name := "sainsburys"
	url := "http://www.j-sainsbury.co.uk/media/latest-stories/"
	linkSel := "#content_container a.title"
	title := "#page_container h1"
	content := "#page_container .richTextFormat"
	// TODO: kill everything after: "Notes to Editors"
	cruft := ""
	pubDate := "#page_container .nm_right .list_plain, #page_container .blog_author"

	return &prscrape.ComposedScraper{
		name,
		prscrape.MustBuildGenericDiscover(name, url, linkSel, false),
		prscrape.MustBuildGenericScrape(name, title, content, cruft, pubDate),
	}
}

// announcments from gov.uk
func NewGovUKAnnounceScraper() prscrape.Scraper {
	name := "gov.uk-announce"
	url := "https://www.gov.uk/government/announcements"
	linkSel := "#announcements-container h3 a"
	title := "#page article header h1"
	content := "#page article .govspeak"
	cruft := ""
	pubDate := "#page article .primary-metadata .date"

	return &prscrape.ComposedScraper{
		name,
		prscrape.MustBuildGenericDiscover(name, url, linkSel, false),
		prscrape.MustBuildGenericScrape(name, title, content, cruft, pubDate),
	}
}

func NewEurekalertScraper() prscrape.Scraper {
	name := "eurekalert.com"
	feeds := []string{"http://www.eurekalert.org/rss.xml"}

	title := "h1"
	content := "p"
	cruft := "table, .FA_Footer, .disclaimer"
	pubDate := "" // use date from rss
	return &prscrape.ComposedScraper{
		name,
		prscrape.MustBuildRSSDiscover(name, feeds),
		prscrape.MustBuildGenericScrape(name, title, content, cruft, pubDate),
	}

}

func NewPRWebUKScraper() prscrape.Scraper {
	name := "uk.prweb.com"
	// there is an rss feed, but it only holds 10 items (too few for such a high-volume source)
	url := "http://uk.prweb.com/recentnews/"
	linkSel := "#releases .release a"

	title := ".container .release .content h1.title"
	content := ".container .release .content p"
	cruft := ".footershare, .mediabox, .releaseDateline"
	pubDate := ".releaseDateline"
	return &prscrape.ComposedScraper{
		name,
		prscrape.MustBuildGenericDiscover(name, url, linkSel, false),
		prscrape.MustBuildGenericScrape(name, title, content, cruft, pubDate),
	}

}

func NewPRNewsWireUKScraper() prscrape.Scraper {
	name := "prnewswire.co.uk"

	feeds := []string{"http://www.prnewswire.co.uk/rss/english-releases-news.rss"}

	title := "#newsdetailnew h1"
	content := "#newsdetailnew .news-col p"
	cruft := ""
	pubDate := "#newsdetailnew .xn-chron"
	return &prscrape.ComposedScraper{
		name,
		prscrape.MustBuildRSSDiscover(name, feeds),
		prscrape.MustBuildGenericScrape(name, title, content, cruft, pubDate),
	}

}

func NewPolicyExchangeScraper() prscrape.Scraper {
	name := "policyexchange.org.uk"

	feeds := []string{"http://www.policyexchange.org.uk/media-centre/press-releases/category/feed/rss/press-releases?format=feed"}

	title := "#main .item h2"
	content := "#main .item .element"
	cruft := ""
	pubDate := "#main .item .event-date"
	return &prscrape.ComposedScraper{
		name,
		prscrape.MustBuildRSSDiscover(name, feeds),
		prscrape.MustBuildGenericScrape(name, title, content, cruft, pubDate),
	}

}

func NewMigrationWatchScraper() prscrape.Scraper {
	name := "migrationwatchuk.org"
	url := "http://www.migrationwatchuk.org/press-releases"
	linkSel := ".middleColumn a[href*=\"/press-release/\"]"

	title := ".mainColumn h1"
	content := ".mainColumn .article"
	pubDate := ".mainColumn .article em"
	cruft := ""
	return &prscrape.ComposedScraper{
		name,
		prscrape.MustBuildGenericDiscover(name, url, linkSel, false),
		prscrape.MustBuildGenericScrape(name, title, content, cruft, pubDate),
	}

}

func NewTaxpayersAllianceScraper() prscrape.Scraper {
	name := "taxpayersalliance.com"
	feeds := []string{"http://www.taxpayersalliance.com/rss"}

	title := ".entry h1"
	content := ".entry .entry_content"
	cruft := ".sharedaddy, .yarpp-related, .author-info"
	// sigh... rss has no date, and page uses a stupid format (eg "Nov 2013 23")
	// for now, just use current date
	// TODO: fix this!
	pubDate := ""
	return &prscrape.ComposedScraper{
		name,
		prscrape.MustBuildRSSDiscover(name, feeds),
		prscrape.MustBuildGenericScrape(name, title, content, cruft, pubDate),
	}
}

func NewGreenpeaceUKScraper() prscrape.Scraper {
	name := "greenpeace.org.uk"
	url := "http://www.greenpeace.org.uk/media/press-releases"
	linkSel := ".view-press-releases .views-row a"

	title := "#main h1"
	content := "#main .node .content .field-body"
	pubDate := "#main .node .content .field-date-published"
	cruft := ""
	return &prscrape.ComposedScraper{
		name,
		prscrape.MustBuildGenericDiscover(name, url, linkSel, false),
		prscrape.MustBuildGenericScrape(name, title, content, cruft, pubDate),
	}

}
func NewShelterScraper() prscrape.Scraper {
	name := "shelter.org.uk"
	url := "http://media.shelter.org.uk/home/press_releases"
	linkSel := "#mediaListingWrapper h3 a"

	title := ".news_story_body h1"
	content := ".news_story_body"
	// TODO: get pubdate from meta tag
	pubDate := ""
	cruft := ""
	return &prscrape.ComposedScraper{
		name,
		prscrape.MustBuildGenericDiscover(name, url, linkSel, false),
		prscrape.MustBuildGenericScrape(name, title, content, cruft, pubDate),
	}

}
