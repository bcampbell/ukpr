package prscrape

import (
	//	"bytes"
	"code.google.com/p/cascadia"
	"code.google.com/p/go-charset/charset"
	_ "code.google.com/p/go-charset/data"
	"code.google.com/p/go.net/html"
	"errors"
	"fmt"
	"github.com/bcampbell/fuzzytime"
	rss "github.com/jteeuwen/go-pkg-rss"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// BuildGenericDiscover returns a DiscoverFunc which fetches a page and extracts matching links.
// TODO: pageUrl should be an array
func BuildGenericDiscover(scraperName, pageUrl, linkSelector string, allowHostChange bool) (DiscoverFunc, error) {
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

		return getLinks(root, page, scraperName, linkSel, allowHostChange)
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
			root, err := fetchPage(page)
			if err != nil {
				return nil, err
			}

			foo, err := getLinks(root, page, scraperName, linkSel, true)
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

// fetches an HTML page, converts it to utf-8 and parses it
func fetchPage(page *url.URL) (*html.Node, error) {
	resp, err := http.Get(page.String())
	if err != nil {
		return nil, err
	}
	// TODO: collect redirects
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = errors.New(fmt.Sprintf("HTTP code %d (%s)", resp.StatusCode, page.String()))
		return nil, err
	}

	// read the page and devine the character encoding.
	// if it's not utf-8, convert it.
	rawHTML, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	enc := findCharset("", rawHTML)
	var r io.Reader
	r = strings.NewReader(string(rawHTML))
	if enc != "utf-8" {
		// we'll be translating to utf-8
		var err error
		r, err = charset.NewReader(enc, r)
		if err != nil {
			return nil, err
		}
	}

	root, err := html.Parse(r)
	if err != nil {
		return nil, err
	}
	return root, nil
}

// getLinks grabs all links matching linkSel
func getLinks(root *html.Node, baseURL *url.URL, scraperName string, linkSel cascadia.Selector, allowHostChange bool) ([]*PressRelease, error) {
	docs := make([]*PressRelease, 0)
	for _, a := range linkSel.MatchAll(root) {
		link, err := baseURL.Parse(GetAttr(a, "href")) // extend to absolute url if needed
		if err != nil {
			// TODO: log a warning?
			continue
		}

		// stay on same site?
		if !allowHostChange && (link.Host != baseURL.Host) {
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

	return func(pr *PressRelease, root *html.Node) (err error) {
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
				fmt.Printf("ZZZZ: '%s'\n", dateTxt)
				return err
			}
		} else {
			// if time isn't already set, just fudge using current time
			if pr.PubDate.IsZero() {
				pr.PubDate = time.Now()
			}
		}

		// content
		if cruft != "" {
			cruftSel, err := cascadia.Compile(cruft)
			if err != nil {
				return err
			}
			for _, cruftNode := range cruftSel.MatchAll(root) {
				cruftNode.Parent.RemoveChild(cruftNode)
			}
		}
		contentElements := contentSel.MatchAll(root)

		//var out bytes.Buffer
		pr.Content = ""
		for _, el := range contentElements {
			StripComments(el)

			txt := RenderText(el)
			txt = regexp.MustCompile(`^[\n]{2,}`).ReplaceAllLiteralString(txt, "")
			txt = regexp.MustCompile(`[\n]{2,}$`).ReplaceAllLiteralString(txt, "\n")
			pr.Content += txt
		}
		return nil
	}, nil
}

// BuildRSSDiscover returns a discover function which grabs links from rss feeds
func BuildRSSDiscover(scraperName string, feeds []string) (DiscoverFunc, error) {
	return func() ([]*PressRelease, error) {
		docs := make([]*PressRelease, 0)
		for _, feed := range feeds {
			foo, err := rssDiscover(scraperName, feed)
			if err != nil {
				return docs, err
			}
			docs = append(docs, foo...)
		}
		return docs, nil
	}, nil
}

func htmlToText(rawHTML string) string {
	r := strings.NewReader(rawHTML)
	doc, err := html.Parse(r)
	if err != nil {
		return ""
	}
	return RenderText(doc)
}

func rssDiscover(scraperName string, feedURL string) ([]*PressRelease, error) {
	feed := rss.New(0, false, nil, nil)
	// TODO: ensure this DOES NOT go through an http proxy!
	// (use FetchClient)
	// TODO: this is a bit brittle with badly-formed XML (eg badly-encoded characters)
	err := feed.Fetch(feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("rss: %s", err)
	}

	docs := make([]*PressRelease, 0)
	for _, channel := range feed.Channels {
		for _, item := range channel.Items {
			//fmt.Printf("%v\n", item)
			itemURL := item.Links[0].Href // TODO: scrub

			txt := htmlToText(item.Description)
			/*
				u, err := url.Parse(itemURL)
				if err != nil {
					return nil, err
				}
				if u.Host != "www.tescoplc.com" && u.Host != "tescoplc.com" {
					//fmt.Printf("SKIP %s\n", itemURL)
					continue
				}
			*/
			pubDate, err := ParseTime(item.PubDate)
			if err != nil {
				pubDate = time.Time{}
			}
			pr := PressRelease{Title: item.Title, Source: scraperName, Permalink: itemURL, PubDate: pubDate, Content: txt}
			docs = append(docs, &pr)
		}
	}
	return docs, nil
}

// TODO: kill this once a proper config parser is in place
func MustBuildGenericDiscover(scraperName, pageUrl, linkSelector string, allowHostChange bool) DiscoverFunc {
	fn, err := BuildGenericDiscover(scraperName, pageUrl, linkSelector, allowHostChange)
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

// TODO: kill this once a proper config parser is in place
func MustBuildPaginatedGenericDiscover(scraperName, startUrl, nextPageSelector, linkSelector string) DiscoverFunc {
	fn, err := BuildPaginatedGenericDiscover(scraperName, startUrl, nextPageSelector, linkSelector)
	if err != nil {
		panic(err)
	}
	return fn
}

// TODO: kill this once a proper config parser is in place
func MustBuildRSSDiscover(scraperName string, feeds []string) DiscoverFunc {
	fn, err := BuildRSSDiscover(scraperName, feeds)
	if err != nil {
		panic(err)
	}
	return fn
}
