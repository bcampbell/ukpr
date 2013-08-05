package prscrape

import (
	"fmt"
	"github.com/donovanhide/eventsource"
	//	"github.com/gorilla/mux"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"time"
)

// helper to fetch and scrape an individual press release
func scrape(scraper Scraper, pr *PressRelease) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New(fmt.Sprintf("%v", e))
		}
	}()
	resp, err := http.Get(pr.Permalink)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = errors.New(fmt.Sprintf("HTTP code %d (%s)", resp.StatusCode, pr.Permalink))
		return
	}
	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	// TODO: collect redirects

	err = scraper.Scrape(pr, string(html))
	if err != nil {
		return
	}
	return
}

// run a scraper
func doit(scraper Scraper, store *Store, sseSrv *eventsource.Server) {

	pressReleases, err := scraper.FetchList()
	if err != nil {
		panic(err)
	}

	// cull out the ones we've already got
	oldCount := len(pressReleases)
	pressReleases = store.WhichAreNew(pressReleases)
	log.Printf("%s: %d releases (%d new)", scraper.Name(), oldCount, len(pressReleases))
	// for all the new ones:
	for _, pr := range pressReleases {
		if !pr.complete {
			err = scrape(scraper, pr)
			if err != nil {
				log.Printf("ERROR '%s' %s\n", err, pr.Permalink)
				continue
			}
			pr.complete = true
		}

		// stash the new press release
		ev, err := store.Stash(pr)
		if err != nil {
			log.Printf("%s: ERROR stashing %s (%s)", scraper.Name(), pr.Permalink, err)
		} else {
			log.Printf("%s: stashed %s", scraper.Name(), pr.Permalink)
		}

		// broadcast it to any connected clients
		sseSrv.Publish([]string{pr.Source}, ev)
	}
}

// ServerMain is the entry point for running the server.
// handles commandline flags and all that stuff - the idea is that you can
// easily write a new server with a different bunch of scrapers. The real
// main() would just be a small stub which instantiates a bunch of scrapers,
// then passes control over to here. See ukpr/main.go for an example
func ServerMain(scraperList []Scraper) {
	var port = flag.Int("port", 9998, "port to run server on")
	var interval = flag.Int("interval", 60*10, "interval at which to poll source sites for new releases (in seconds)")
	var testScraper = flag.String("t", "", "Test run an individual scraper, dumping to stdout. Doesn't run server or alter the database.")
	var briefFlag = flag.Bool("b", false, "Brief (testing mode output)")
	var listFlag = flag.Bool("l", false, "List scrapers and exit")

	flag.Parse()

	scrapers := make(map[string]Scraper)

	for _, scraper := range scraperList {
		name := scraper.Name()
		scrapers[name] = scraper
	}

	if *listFlag {
		for name, _ := range scrapers {
			fmt.Println(name)
		}
		return
	}

	if *testScraper != "" {
		// run a single scraper, without server or store
		// TODO: merge the test implementation with doit()
		scraper, ok := scrapers[*testScraper]
		if !ok {
			log.Fatal("Unknown scraper")
		}
		pressReleases, err := scraper.FetchList()
		if err != nil {
			panic(err)
		}
		for _, pr := range pressReleases {
			if !pr.complete {
				//log.Printf("%s: scrape %s", scraper.Name(), pr.Permalink)
				err = scrape(scraper, pr)
				if err != nil {
					log.Printf("%s: ERROR '%s' %s\n", scraper.Name(), err, pr.Permalink)
					continue
				}
				pr.complete = true
			}

			if !*briefFlag {
				fmt.Printf("%s\n %s\n %s\n", pr.Title, pr.PubDate, pr.Permalink)
				fmt.Println("")
				fmt.Println(pr.Content)
				fmt.Println("------------------------------")
			} else {
				fmt.Printf("%s %s\n", pr.Title, pr.Permalink)
			}
		}
		return
	}

	// set up as server
	// using a common store for all scrapers
	// but no reason they couldn't all have their own store
	store := NewStore("./prstore.db")
	sseSrv := eventsource.NewServer()
	for name, _ := range scrapers {
		sseSrv.Register(name, store)
		http.Handle("/"+name+"/", sseSrv.Handler(name))
	}

	//
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		panic(err) //glog.Fatal(err)
	}
	defer l.Close()

	// cheesy task to periodically run the scrapers
	go func() {
		for {
			for _, scraper := range scrapers {
				doit(scraper, store, sseSrv)
			}
			time.Sleep(time.Duration(*interval) * time.Second)
		}
	}()

	log.Printf("running on port %d", *port)
	http.Serve(l, nil)
}
