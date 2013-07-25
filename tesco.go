package main

import (
	"bytes"
	"code.google.com/p/cascadia"
	"code.google.com/p/go.net/html"
	"fmt"
	rss "github.com/jteeuwen/go-pkg-rss"
	//	"github.com/SlyMarbo/rss"
	"github.com/bcampbell/fuzzytime"
	"net/url"
	"regexp"
	"strings"
)

// scraper to grab tesco press releases
type TescoScraper struct{}

func NewTescoScraper() *TescoScraper {
	var s TescoScraper
	return &s
}

func (scraper *TescoScraper) FetchList() ([]*PressRelease, error) {
	feed := rss.New(0, false, nil, nil)
	// TODO: ensure this DOES NOT go through an http proxy!
	// (use FetchClient)
	err := feed.Fetch("http://www.tescoplc.com/tescoplcnews.xml", nil)
	if err != nil {
		return nil, err
	}

	docs := make([]*PressRelease, 0)
	for _, channel := range feed.Channels {
		for _, item := range channel.Items {
			//	fmt.Printf("%s '%s' %v\n", item.Link, item.Title, item.Date)
			itemURL := item.Links[0].Href // TODO: scrub

			u, err := url.Parse(itemURL)
			if err != nil {
				return nil, err
			}
			if u.Host != "www.tescoplc.com" && u.Host != "tescoplc.com" {
				fmt.Printf("SKIP %s\n", itemURL)
				continue
			}

			pubDate, err := parseTime(item.PubDate)
			if err != nil {
				panic(err)
			}
			pr := PressRelease{Title: item.Title, Source: "tesco", Permalink: itemURL, PubDate: pubDate}
			docs = append(docs, &pr)
		}
	}
	return docs, nil
}

func (scraper *TescoScraper) Scrape(pr *PressRelease, raw_html string) error {
	r := strings.NewReader(string(raw_html))
	root, err := html.Parse(r)
	if err != nil {
		return err // TODO: wrap up as ScrapeError?
	}

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
	pr.Title = getTextContent(titleSel.MatchAll(div)[0])

	//
	dateTxt := getTextContent(dateSel.MatchAll(div)[0])
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
		var out bytes.Buffer
		err = html.Render(&out, n)
		if err != nil {
			return err
		}
		foo := out.String()
		// break content when we hit "ENDS"
		if endPat.MatchString(foo) {
			break
		}
		pr.Content = pr.Content + foo + "\n"
	}
	return nil
}
