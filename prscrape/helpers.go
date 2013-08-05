package prscrape

import (
	//	"bytes"
	"code.google.com/p/cascadia"
	"code.google.com/p/go.net/html"
	"errors"
	"fmt"
	"github.com/bcampbell/fuzzytime"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DefaultScraper is a Scraper implementation to handle most common cases.
// It uses CSS selectors to identify the relevant data to extract from a page.
type DefaultScraper struct {
	ScraperName string

	// IndexURL is page to scrape to get list of press release URLs. LinkSel identifies the links on that page.
	IndexURL, LinkSel string

	// Selectors to pull out data from an individual page
	TitleSel, ContentSel, PubdateSel, CruftSel string
}

func (scraper *DefaultScraper) Name() string { return scraper.ScraperName }

func (scraper *DefaultScraper) FetchList() (found []*PressRelease, err error) {
	defer func() {
		if e := recover(); e != nil {
			found = nil
			err = e.(error)
		}
	}()
	found, err = GenericFetchList(scraper.Name(), scraper.IndexURL, scraper.LinkSel)
	return
}

func (scraper *DefaultScraper) Scrape(pr *PressRelease, rawHTML string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	err = GenericScrape(scraper.Name(), pr, rawHTML, scraper.TitleSel, scraper.ContentSel, scraper.CruftSel, scraper.PubdateSel)
	return
}

// GenericFetchList fetches a page, and extracts matching links.
func GenericFetchList(scraperName, pageUrl, linkSelector string) ([]*PressRelease, error) {
	page, err := url.Parse(pageUrl)
	if err != nil {
		return nil, err
	}

	linkSel, err := cascadia.Compile(linkSelector)
	if err != nil {
		return nil, err
	}
	resp, err := http.Get(pageUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = errors.New(fmt.Sprintf("HTTP code %d (%s)", resp.StatusCode, pageUrl))
		return nil, err
	}

	root, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}
	docs := make([]*PressRelease, 0)
	for _, a := range linkSel.MatchAll(root) {
		link, err := page.Parse(GetAttr(a, "href")) // extend to absolute url if needed
		if err != nil {
			// TODO: log a warning?
			continue
		}

		// stay on same site
		if link.Host != page.Host {
			// TODO: log a warning
			//fmt.Printf("SKIP link to different site %s\n", link.String())
			continue
		}

		pr := PressRelease{Source: scraperName, Permalink: link.String()}
		docs = append(docs, &pr)
	}
	return docs, nil
}

// GenericScrape scrapes a press release from raw_html based on a bunch of css selector strings
func GenericScrape(source string, pr *PressRelease, raw_html, title, content, cruft, pubDate string) error {

	r := strings.NewReader(string(raw_html))
	root, err := html.Parse(r)
	if err != nil {
		return err // TODO: wrap up as ScrapeError?
	}

	pr.Source = source

	// title
	titleSel, err := cascadia.Compile(title)
	if err != nil {
		return err
	}
	pr.Title = CompressSpace(GetTextContent(titleSel.MatchAll(root)[0]))

	// pubdate - only needs to contain a valid date string, doesn't matter
	// if there's other crap in there too.
	if pubDate != "" {
		pubDateSel, err := cascadia.Compile(pubDate)
		if err != nil {
			return err
		}
		dateTxt := GetTextContent(pubDateSel.MatchAll(root)[0])
		pr.PubDate, err = fuzzytime.Parse(dateTxt)
		if err != nil {
			return err
		}
	} else {
		// if time isn't already set, just fudge using current time
		if pr.PubDate.IsZero() {
			pr.PubDate = time.Now()
		}
	}

	// content
	contentSel, err := cascadia.Compile(content)
	if err != nil {
		return err
	}
	contentElements := contentSel.MatchAll(root)
	if cruft != "" {
		cruftSel, err := cascadia.Compile(cruft)
		if err != nil {
			return err
		}
		for _, el := range contentElements {
			for _, cruft := range cruftSel.MatchAll(el) {
				cruft.Parent.RemoveChild(cruft)
			}
		}
	}

	//var out bytes.Buffer
	pr.Content = ""
	for _, el := range contentElements {
		StripComments(el)

		pr.Content += GetTextContent(el)
		/*err = html.Render(&out, el)
		if err != nil {
			return err
		}
		*/
	}
	//pr.Content = out.String()
	return nil
}
