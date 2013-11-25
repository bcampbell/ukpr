// prscrape is a framework for scraping press releases.
//
// It's designed to be run as a server, and will:
// 1) periodically scrape a bunch of press release sources
// 2) hold an archive of press releases sufficient to cover a few days
// 3) serves up the scraped press releases via HTTP, as server side event
//    endpoints
//
// The scraped press releases are persistent, in a sqlite db. The idea is
// that eventually it'll be set up to keep just a week or so archive, to let
// consumers of the stream have a chance to catch up if they go down for a
// day or two.
// (but for now, it just keeps adding to the db)
//
// Clients connect to:
//
//   http://<host>:<port>/<source>/
//
// where source is the name of one of the installed scrapers.
//
// Clients can send a last-event-id header to access archived press releases.
// eg:
//  $ curl http://localhost:9998/72point/ -H "Last-Event-ID: 0"
// Will serve up _all_ the stored 72point press releases.
//
// Without last-event-id, the client will be served only new press
// releases as they come in.
//

package prscrape
