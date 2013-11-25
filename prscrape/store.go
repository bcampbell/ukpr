package prscrape

import (
	_ "code.google.com/p/go-sqlite/go1/sqlite3"
	"database/sql"
	"encoding/json"
	"github.com/donovanhide/eventsource"
	//_ "github.com/mattn/go-sqlite3"
	"fmt"
	"strconv"
)

type Store interface {
	WhichAreNew(incoming []*PressRelease) []*PressRelease
	Stash(pr *PressRelease) (*pressReleaseEvent, error)
	Replay(channel, lastEventId string) chan eventsource.Event
}

type TestStore struct {
	briefMode bool
}

func NewTestStore(brief bool) *TestStore {
	store := &TestStore{brief}
	return store
}

func (store *TestStore) WhichAreNew(incoming []*PressRelease) []*PressRelease {
	return incoming
}

func (store *TestStore) Stash(pr *PressRelease) (*pressReleaseEvent, error) {
	if store.briefMode {
		fmt.Printf("%s %s\n", pr.Title, pr.Permalink)
	} else {
		fmt.Printf("%s\n %s\n %s\n", pr.Title, pr.PubDate, pr.Permalink)
		fmt.Println("")
		fmt.Println(pr.Content)
		fmt.Println("------------------------------")
	}

	id := 0
	return &pressReleaseEvent{pr, int(id)}, nil
}

func (store *TestStore) Replay(channel, lastEventId string) chan eventsource.Event {
	panic("unsupported")
	return nil
}

// NOTE: github.com/mattn/go-sqlite3 and code.google.com/p/go-sqlite/go1/sqlite3
// seem to use different (and incompatable) formats for storing time.Time values.

// DBStore manages an archive of recent press releases in an sqlite db.
// It also implements eventsource.Repository to allow the press releases to be
// streamed out as server side events.
// Can stash away press releases for multiple sources.
type DBStore struct {
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

func NewDBStore(dbfile string) *DBStore {
	store := new(DBStore)
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
func (store *DBStore) WhichAreNew(incoming []*PressRelease) []*PressRelease {
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
func (store *DBStore) Stash(pr *PressRelease) (*pressReleaseEvent, error) {

	res, err := store.db.Exec("INSERT INTO press_release (title,source,permalink,pubdate,content) VALUES ($1,$2,$3,$4,$5)", pr.Title, pr.Source, pr.Permalink, pr.PubDate, pr.Content)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &pressReleaseEvent{pr, int(id)}, nil
}

// Replay to handle last-event-id catchups
// note: channel contains the source (eg 'tesco'...)
func (store *DBStore) Replay(channel, lastEventId string) (out chan eventsource.Event) {
	var err error
	var rows *sql.Rows

	fields := "id,title,source,permalink,pubdate,content"
	if lastEventId == "" {
		// no last eventid, just replay everything
		if rows, err = store.db.Query("SELECT "+fields+" FROM press_release WHERE source=$1", channel); err != nil {
			panic(err)
		}
	} else {
		// only fetch events after lastEventId
		id, err := strconv.Atoi(lastEventId)
		if err != nil {
			panic(err)
		}
		if rows, err = store.db.Query("SELECT "+fields+" FROM press_release WHERE id>$1 AND source=$2", id, channel); err != nil {
			panic(err)
		}
	}

	out = make(chan eventsource.Event)
	go func() {
		defer close(out)
		defer rows.Close()
		for rows.Next() {
			var id int
			var pr PressRelease
			pr.Type = "press release"
			if err := rows.Scan(&id, &pr.Title, &pr.Source, &pr.Permalink, &pr.PubDate, &pr.Content); err != nil {
				panic(err)
			}

			out <- &pressReleaseEvent{&pr, id}
		}
	}()
	return
}
