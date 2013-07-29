package main

import (
	"database/sql"
	"encoding/json"
	"github.com/donovanhide/eventsource"
	_ "code.google.com/p/go-sqlite/go1/sqlite3"
	"strconv"
)

// Store manages an archive of recent press releases.
// It also implements eventsource.Repository to allow the press releases to be
// streamed out as server side events.
// Can stash away press releases for multiple sources.
type Store struct {
	db *sql.DB
}

// pressReleaseEvent wraps up a PressRelease for use as a server-sent event.
type pressReleaseEvent struct {
	payload *PressRelease
	id      int
}

func (ev *pressReleaseEvent) Id() string {
	return strconv.Itoa(ev.id)
}

func (ev *pressReleaseEvent) Event() string {
	return "press_release"
}

func (ev *pressReleaseEvent) Data() string {
	out, _ := json.Marshal(*ev.payload)
	return string(out)
}

func NewStore(dbfile string) *Store {
	store := new(Store)
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		panic(err)
	}
	store.db = db

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS press_release (
         id INTEGER PRIMARY KEY,
         title TEXT NOT NULL,
         source TEXT NOT NULL,
         permalink TEXT NOT NULL,
         pubdate DATETIME NOT NULL,
         content TEXT NOT NULL )`)
	if err != nil {
		panic(err)
	}

	return store
}

// returns a list of press releases with the ones already in the store culled out
func (store *Store) WhichAreNew(incoming []*PressRelease) []*PressRelease {
	var unseen []*PressRelease
	// should really just use a single sql query ("WHERE permalink IN (...)" but hey.
	for _, pr := range incoming {
		var id int64
		err := store.db.QueryRow("SELECT id FROM press_release WHERE permalink=$1 AND source=$2", pr.Permalink, pr.Source).Scan(&id)
		switch err {
		case nil:
			// it's already in db
		case sql.ErrNoRows:
			// it's a new one
			unseen = append(unseen, pr)
		default:
			panic(err)
		}
	}
	return unseen
}

// Stash adds a press release into the store
func (store *Store) Stash(pr *PressRelease) *pressReleaseEvent {

	res, err := store.db.Exec("INSERT INTO press_release (title,source,permalink,pubdate,content) VALUES ($1,$2,$3,$4,$5)", pr.Title, pr.Source, pr.Permalink, pr.PubDate, pr.Content)
	if err != nil {
		panic(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		panic(err)
	}
	return &pressReleaseEvent{pr, int(id)}
}

// Get to help handle last-event-id catchups
// note: channel contains the source (eg 'tesco'...)
func (store *Store) Get(channel, eventId string) eventsource.Event {
	id, err := strconv.Atoi(eventId)
	if err != nil {
		panic(err)
	}
	row := store.db.QueryRow(`SELECT title,source,permalink,pubdate,content FROM press_release WHERE id=$1 AND source=$2`, id, channel)
	var pr PressRelease
	if err := row.Scan(&pr.Title, &pr.Source, &pr.Permalink, &pr.PubDate, &pr.Content); err != nil {
		panic(err) // TODO: log and return nil
	}
	pr.complete = true

	return &pressReleaseEvent{&pr, id}
}

// Replay to handle last-event-id catchups
// note: channel contains the source (eg 'tesco'...)
func (store *Store) Replay(channel, lastEventId string) (ids chan string) {
	var err error
	var rows *sql.Rows
	if lastEventId == "" {
		// no last eventid, just replay everything
		if rows, err = store.db.Query("SELECT id FROM press_release WHERE source=$1", channel); err != nil {
			panic(err)
		}
	} else {
		// only fetch events after lastEventId
		id, err := strconv.Atoi(lastEventId)
		if err != nil {
			panic(err)
		}
		if rows, err = store.db.Query("SELECT id FROM press_release WHERE id>$1 AND source=$2", id, channel); err != nil {
			panic(err)
		}
	}

	ids = make(chan string)
	go func() {
		defer rows.Close()
		for rows.Next() {
			var id int
			err := rows.Scan(&id)
			if err != nil {
				panic(err)
			}
			ids <- strconv.Itoa(id)
		}
		close(ids)
	}()
	return ids
}
