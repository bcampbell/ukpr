# ukpr (working title)

This program runs a server which
1) periodically scrapes a bunch of press release sources
2) serves up those press releases as server side event endpoints

The scraped press releases are persistant, in a sqlite db. The idea is
that eventually it'll be set up to keep just a week or so archive, to let
consumers have a chance to catch up if they go down for a day or two.
(but for now, it just keeps adding to the db)

Clients connect to:

    http://<host>:<port>/<source>/

where source is (currently) one of:
    tesco
    72point

Clients can send a last-event-id header to access archived press releases.
eg:
    $ curl http://localhost:9998/72point/ -H "Last-Event-ID: 0"
Will serve up _all_ the stored 72point press releases.

Without last-event-id, the client will be served only new press
releases as they come in.


## TODOs

 - proper logging and error handling (kill all the panics!)
 - split up into separate packages (in particular, make it easy to build
   a new app with a diffferent bunch of scrapers)
 - we've already got a http server running, so should implement a simple
   browsing interface for visual sanity-checking of press releases.
