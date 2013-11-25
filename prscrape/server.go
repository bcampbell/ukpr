package prscrape

import (
	"fmt"
	"github.com/donovanhide/eventsource"
	"github.com/golang/glog"
	//	"github.com/gorilla/mux"
	"errors"
	"flag"
	"io/ioutil"
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
func doit(scraper Scraper, store Store, sseSrv *eventsource.Server) {
	if glog.V(1) {
		glog.Infof("%s: Discover", scraper.Name())
	}
	pressReleases, err := scraper.Discover()
	if err != nil {
		glog.Errorf("%s: Discover failed: %s", scraper.Name(), err)
		return
	}

	// cull out the ones we've already got
	oldCount := len(pressReleases)
	pressReleases = store.WhichAreNew(pressReleases)
	if glog.V(1) {
		glog.Infof("%s: %d releases (%d new)", scraper.Name(), oldCount, len(pressReleases))
	}
	// for all the new ones:
	for _, pr := range pressReleases {
		if !pr.complete {
			err = scrape(scraper, pr)
			if err != nil {
				glog.Errorf("%s: %s %s\n", scraper.Name(), err, pr.Permalink)
				continue
			}
			pr.complete = true
		}

		// stash the new press release
		ev, err := store.Stash(pr)
		if err != nil {
			glog.Errorf("%s: failed to stash %s (%s)", scraper.Name(), pr.Permalink, err)
		} else {
			glog.Infof("%s: added %s", scraper.Name(), pr.Permalink)
		}

		if sseSrv != nil {
			// broadcast it to any connected clients
			sseSrv.Publish([]string{pr.Source}, ev)
		}
	}
}

// ServerMain is the entry point for running the server.
// handles commandline flags and all that stuff - the idea is that you can
// easily write a new server with a different bunch of scrapers. The real
// main() would just be a small stub which instantiates a bunch of scrapers,
// then passes control over to here. See ukpr/main.go for an example
func ServerMain(configfunc ConfigureFunc) {
	var port = flag.Int("port", 9998, "port to run server on")
	var interval = flag.Int("interval", 60*10, "interval at which to poll source sites for new releases (in seconds)")
	var testMode = flag.Bool("t", false, "Test mode - dumping to stdout. Doesn't run server or alter the database.")
	var briefFlag = flag.Bool("b", false, "Brief (testing mode output)")
	var listFlag = flag.Bool("l", false, "List scrapers and exit")
	var historicalFlag = flag.Bool("historical", false, "Run historical version of scrapers, where available")

	flag.Parse()

	// set up scrapers
	scraperList := configfunc(*historicalFlag)

	if *listFlag {
		// list scrapers and exit
		for _, scraper := range scraperList {
			fmt.Println(scraper.Name())
		}
		return
	}

	allScrapers := make(map[string]Scraper)
	for _, scraper := range scraperList {
		name := scraper.Name()
		allScrapers[name] = scraper
	}

	activeScrapers := make(map[string]Scraper)
	if len(flag.Args()) > 0 {
		// user asked for a subset of scrapers
		for _, name := range flag.Args() {
			scraper, ok := allScrapers[name]
			if !ok {
				panic(fmt.Sprintf("%s: unknown scraper", name))
			}
			activeScrapers[name] = scraper
		}
	} else {
		activeScrapers = allScrapers
	}

	// set up store and SSE server
	// using a common store for all scrapers
	// but no reason they couldn't all have their own store
	var store Store
	var sseSrv *eventsource.Server

	if *testMode {
		store = NewTestStore(*briefFlag)
		sseSrv = nil
	} else {
		store = NewDBStore("./prstore.db")
		sseSrv = eventsource.NewServer()
		for name, _ := range activeScrapers {
			sseSrv.Register(name, store)
			http.Handle("/"+name+"/", sseSrv.Handler(name))
		}
	}

	//
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		glog.Fatal(err)
	}
	defer l.Close()

	// cheesy task to periodically run the active scrapers
	go func() {
		for {
			for _, scraper := range activeScrapers {
				doit(scraper, store, sseSrv)
			}
			if glog.V(1) {
				glog.Info("sleeping")
			}
			time.Sleep(time.Duration(*interval) * time.Second)
		}
	}()

	glog.Infof("running on port %d", *port)
	http.Serve(l, nil)
}
