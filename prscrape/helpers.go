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

// GenericDiscover returns a DiscoverFunc which fetches a page and extracts matching links.
// TODO: pageUrl should be an array
func GenericDiscover(scraperName, pageUrl, linkSelector string) DiscoverFunc {
	return func() ([]*PressRelease, error) {

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
}

// GenericScrape builds a function which scrapes a press release from raw_html based on a bunch of css selector strings
func GenericScrape(source, title, content, cruft, pubDate string) ScrapeFunc {
	return func(pr *PressRelease, rawHTML string) (err error) {

		r := strings.NewReader(string(rawHTML))
		root, err := html.Parse(r)
		if err != nil {
			return err // TODO: wrap up as ScrapeError?
		}

		pr.Type = "press release"
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
}
