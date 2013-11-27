# ukpr (working title)

Ben Campbell (ben@scumways.com), at the [Media Standards Trust](http://mediastandardstrust.org)

## Overview

This program:

  1. periodically scrapes a bunch of press release sources
  2. stores them in a database
  3. serves up the press releases to any interested clients via HTTP (as
     [server-sent events](http://dev.w3.org/html5/eventsource/)).

The idea is eventually it'll be set up to keep an archive of a week or so
to let clients have a chance to catch up if they go down.

When ukpr is running, clients can connect to:

    http://<host>:<port>/<source>/

where source is one of the scrapers. You can get a list using:

    $ ukpr -l

Connected clients receive a stream of press releases, as they are scraped.
Clients can send a last-event-id header to access archived press releases,
or to resume after a disconnection

You can connect and view the raw stream like using any http client, eg:

    $ curl http://localhost:9998/72point/ -H "Last-Event-ID: 0"

Will serve up _all_ the stored 72point press releases.

Without last-event-id, the client will be served only new press
releases as they come in.


## Usage


    ukpr <flags> [scraper1 scraper2 ...]

Specific scrapers can be listed after the flags - only those scrapers will
be used. By default all scrapers will be used.

flags:

    -l
    list available scrapers and exit

    -historical
    use the history-collecting version of all scrapers which
    have one (only 72point at the moment)

    -t
    test mode. Scrape, but output to stdout and don't touch
    the database. Also turns off the SSE serving.

    -b
    brief output (for test mode only) - just dump out title of press
    releases to stdout rather than the whole thing.

It uses [glog](https://github.com/golang/glog) for logging, so also
supports all the standard glog flags.


## TODOs

 - we've already got a http server running, so should implement a simple
   browsing interface for visual sanity-checking of press releases.
 - implement a proper config file system
 - run the scrapers in parallel with proper interval timing

## Motivation & Goals

The main aim for this is to provide press releases for use by
 http://churnalism.com, hence the UK bias.

A major goal is to make it simple enough to customise for coverage of any
set of press release sources you like.

It's not designed to be a full historical archive of press releases, merely
a conduit to stream them out to interested clients, with a bit of buffering
to make it things more fault-tolerant.




