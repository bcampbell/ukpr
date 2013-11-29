package ukscrapers

import (
	"code.google.com/p/cascadia"
	"code.google.com/p/go.net/html"
	//"fmt"
	"github.com/bcampbell/fuzzytime"
	"github.com/bcampbell/ukpr/prscrape"
	"regexp"
)

// currently got a custom scrape func for tesco, to try and snip off
// trailing PR contacts cruft, but really this should be generalised.

func NewTescoScraper() *prscrape.Scraper {
	name := "tesco"
	feeds := []string{"http://www.tescoplc.com/tescoplcnews.xml"}

	return &prscrape.Scraper{
		name,
		prscrape.MustBuildRSSDiscover(name, feeds),
		scrape,
	}
}

func scrape(pr *prscrape.PressRelease, root *html.Node) error {
	containerSel := cascadia.MustCompile(".pagecontent")
	titleSel := cascadia.MustCompile(".newstitle")
	dateSel := cascadia.MustCompile(".greydate")
	cruftSel := cascadia.MustCompile(".sharebuttons, .greydate, .newstitle, .boilerplate")

	endPat := regexp.MustCompile(`\bENDS\b`)

	contentSel := cascadia.MustCompile("p, ul")
	// :contains() not in css3 spec, but cascadia supports it
	//	noteSepSel := cascadia.MustCompile(`strong:contains("Notes to editors:"), strong:contains("ENDS")`)

	div := containerSel.MatchAll(root)[0]

	// get the title
	pr.Title = prscrape.GetTextContent(titleSel.MatchAll(div)[0])

	//
	dateTxt := prscrape.GetTextContent(dateSel.MatchAll(div)[0])
	var err error
	pr.PubDate, err = fuzzytime.Parse(dateTxt)
	if err != nil {
		return err
	}

	// cull out the rubbish
	for _, cruft := range cruftSel.MatchAll(div) {
		cruft.Parent.RemoveChild(cruft)
	}

	// content
	pr.Content = ""
	for _, n := range contentSel.MatchAll(div) {
		foo := prscrape.RenderText(n)
		// break content when we hit "ENDS"
		if endPat.MatchString(foo) {
			break
		}
		pr.Content = pr.Content + foo
	}
	return nil
}
