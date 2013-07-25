package main

// TODOs
// - proper logging and error handling (kill all the panics!)
// - split up into separate packages
//

import (
	"fmt"
	"github.com/donovanhide/eventsource"
	//	"github.com/gorilla/mux"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

// TODO: support multiple urls
type PressRelease struct {
	Title     string
	Source    string
	Permalink string
	PubDate   time.Time
	Content   string
	// if this is a fully-filled out press release, complete is set
	complete bool
}

// Scraper is the interface to implement to add a new scraper to the system
type Scraper interface {
	// Fetch a list of 'current' press releases.
	// (via RSS feed, or by scraping an index page or whatever)
	// The results are passed back as PressRelease structs. At the very least,
	// the Permalink field must be set to the URL of the press release,
	// But there's no reason FetchList() can't fill out all the fields if the
	// data is available (eg some rss feeds have everything required).
	// For incomplete PressReleases, the framework will fetch the HTML from
	// the Permalink URL, and invoke Scrape() to complete the data.
	FetchList() ([]*PressRelease, error)

	// scrape a single press release from raw html passed in as a string
	Scrape(*PressRelease, string) error
}

// helper to fetch and scrape an individual press release
func scrape(scraper Scraper, pr *PressRelease) error {
	resp, err := http.Get(pr.Permalink)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// TODO: collect redirects

	err = scraper.Scrape(pr, string(html))
	if err != nil {
		return err
	}
	return nil
}

func doit(scraper Scraper, store *Store, sseSrv *eventsource.Server, channel string) {
	pressReleases, err := scraper.FetchList()
	if err != nil {
		panic(err)
	}

	// cull out the ones we've already got
	pressReleases = store.WhichAreNew(pressReleases)
	fmt.Printf("%d new releases\n", len(pressReleases))
	// for all the new ones:
	for _, pr := range pressReleases {
		if !pr.complete {
			err = scrape(scraper, pr)
			if err != nil {
				panic(err)
			}
		}

		fmt.Printf("stashing '%s'\n", pr.Title)
		// stash the new press release
		ev := store.Stash(pr)
		sseSrv.Publish([]string{channel}, ev)
		fmt.Printf("stashed %s: %s\n%s\n%s\n", ev.Id(), pr.Title, pr.PubDate, pr.Permalink)
	}
}

var port = flag.Int("port", 9998, "port to run server on")

func testHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "Hello\n")
}

func main() {
	flag.Parse()

	tescoScraper := NewTescoScraper()
	tescoStore := NewStore("./tesco.db")
	sseSrv := eventsource.NewServer()
	sseSrv.Register("tesco", tescoStore)

	http.HandleFunc("/test", testHandler)
	http.Handle("/tesco/", sseSrv.Handler("tesco"))

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		panic(err) //glog.Fatal(err)
	}
	defer l.Close()
	fmt.Printf("running.\n")
	log.Printf("running on port %d", *port)

	go func() {
		for {
			time.Sleep(5 * time.Second)
			doit(tescoScraper, tescoStore, sseSrv, "tesco") // every hour or so :-)
		}
	}()

	http.Serve(l, nil)
}
