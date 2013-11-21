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

// BuildGenericDiscover returns a DiscoverFunc which fetches a page and extracts matching links.
// TODO: pageUrl should be an array
func BuildGenericDiscover(scraperName, pageUrl, linkSelector string) (DiscoverFunc, error) {
	linkSel, err := cascadia.Compile(linkSelector)
	if err != nil {
		return nil, err
	}

	// parse the url to make sure it's a good 'un
	page, err := url.Parse(pageUrl)
	if err != nil {
		return nil, err
	}

	return func() ([]*PressRelease, error) {

		root, err := fetchPage(page)
		if err != nil {
			return nil, err
		}

		return getLinks(root, page, scraperName, linkSel)
	}, nil
}

// BuildPaginatedGenericDiscover returns a DiscoverFunc which fetches links
// and steps through multiple pages.
func BuildPaginatedGenericDiscover(scraperName, startUrl, nextPageSelector, linkSelector string) (DiscoverFunc, error) {
	linkSel, err := cascadia.Compile(linkSelector)
	if err != nil {
		return nil, err
	}
	nextPageSel, err := cascadia.Compile(nextPageSelector)
	if err != nil {
		return nil, err
	}

	return func() ([]*PressRelease, error) {
		docs := make([]*PressRelease, 0)
		// parse the url to make sure it's a good 'un
		page, err := url.Parse(startUrl)
		if err != nil {
			return nil, err
		}
		for {
			fmt.Printf("fetch %s\n", page.String())
			root, err := fetchPage(page)
			if err != nil {
				return nil, err
			}

			foo, err := getLinks(root, page, scraperName, linkSel)
			if err != nil {
				return docs, err
			}
			docs = append(docs, foo...)

			// more pages?
			nexts := nextPageSel.MatchAll(root)
			if len(nexts) == 0 {
				break
			}
			// extend to absolute url if needed
			nextLink, err := page.Parse(GetAttr(nexts[0], "href"))
			if err != nil {
				return docs, err
			}
			page = nextLink
		}
		return docs, nil
	}, nil
}

// fetch and parse a page
func fetchPage(page *url.URL) (*html.Node, error) {
	resp, err := http.Get(page.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = errors.New(fmt.Sprintf("HTTP code %d (%s)", resp.StatusCode, page.String()))
		return nil, err
	}

	root, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}
	return root, nil
}

// getLinks grabs all links matching linkSel
func getLinks(root *html.Node, baseURL *url.URL, scraperName string, linkSel cascadia.Selector) ([]*PressRelease, error) {
	docs := make([]*PressRelease, 0)
	for _, a := range linkSel.MatchAll(root) {
		link, err := baseURL.Parse(GetAttr(a, "href")) // extend to absolute url if needed
		if err != nil {
			// TODO: log a warning?
			continue
		}

		// stay on same site
		if link.Host != baseURL.Host {
			// TODO: log a warning?
			//fmt.Printf("SKIP link to different site %s\n", link.String())
			continue
		}

		pr := PressRelease{Source: scraperName, Permalink: link.String()}
		docs = append(docs, &pr)
	}
	return docs, nil
}

// BuildGenericScrape builds a function which scrapes a press release from raw_html based on a bunch of css selector strings
func BuildGenericScrape(source, title, content, cruft, pubDate string) (ScrapeFunc, error) {

	// precompile all the selectors, to catch config errors early
	titleSel, err := cascadia.Compile(title)
	if err != nil {
		return nil, err
	}
	contentSel, err := cascadia.Compile(content)
	if err != nil {
		return nil, err
	}

	var pubDateSel cascadia.Selector = nil
	if pubDate != "" {
		pubDateSel, err = cascadia.Compile(pubDate)
		if err != nil {
			return nil, err
		}
	}

	return func(pr *PressRelease, rawHTML string) (err error) {

		r := strings.NewReader(string(rawHTML))
		root, err := html.Parse(r)
		if err != nil {
			return err // TODO: wrap up as ScrapeError?
		}

		pr.Type = "press release"
		pr.Source = source

		// title
		pr.Title = CompressSpace(GetTextContent(titleSel.MatchAll(root)[0]))

		// pubdate - only needs to contain a valid date string, doesn't matter
		// if there's other crap in there too.
		if pubDateSel != nil {
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
	}, nil
}

// TODO: kill this once a proper config parser is in place
func MustBuildGenericDiscover(scraperName, pageUrl, linkSelector string) DiscoverFunc {
	fn, err := BuildGenericDiscover(scraperName, pageUrl, linkSelector)
	if err != nil {
		panic(err)
	}
	return fn
}

// TODO: kill this once a proper config parser is in place
func MustBuildGenericScrape(source, title, content, cruft, pubDate string) ScrapeFunc {
	fn, err := BuildGenericScrape(source, title, content, cruft, pubDate)
	if err != nil {
		panic(err)
	}
	return fn
}

func MustBuildPaginatedGenericDiscover(scraperName, startUrl, nextPageSelector, linkSelector string) DiscoverFunc {
	fn, err := BuildPaginatedGenericDiscover(scraperName, startUrl, nextPageSelector, linkSelector)
	if err != nil {
		panic(err)
	}
	return fn
}
