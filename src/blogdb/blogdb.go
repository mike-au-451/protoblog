package blogdb

import (
	"bytes"
	"database/sql"
	"fmt"
	// "runtime"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"

	"main/cache"
)

type DB struct {
	path string
	db *sql.DB
	cc *cache.Cache
}

type BlogEntry struct {
	Id int 			`json:"id"`
	Title string 	`json:"title"`
	Body string 	`json:"body"`
	Posted string 	`json:"posted"`
}

func NewDB(path string, cc *cache.Cache) *DB {
	var err error

	db := DB{ path: path, cc: cc }
	db.db, err = sql.Open("sqlite3", db.path)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("NewDB: failed to open: %s", err))
		return nil
	}
	// runtime.SetFinalizer(&db, db.db.Close())

	return &db;
}

func (bdb *DB) Close() {
	bdb.db.Close()
}

func (bdb *DB) GetEntries() []BlogEntry {
	rows := bdb.getRows()
	if rows == nil {
		log.Error().Msg("failed to get rows")
		return nil
	}
	defer rows.Close()

	entries := blogEntryList(rows)
	if entries == nil {
		log.Error().Msg("failed to list")
		return nil
	}

	if !bdb.getBodies(entries) {
		log.Error().Msg("failed to get bodies")
		return nil
	}

	return entries
}

func (bdb *DB) getRows() (*sql.Rows) {
	sql := `
		SELECT id, title, body, posted 
		FROM
			Entries ee
			JOIN (
				SELECT entryid, MAX(version) version
				FROM Entries
				WHERE visible
				GROUP BY entryid
			) kk ON ee.entryid = kk.entryid AND ee.version = kk.version
		ORDER BY posted DESC
`
	rows, err := bdb.db.Query(sql)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("failed to select: %s", err))
		return nil
	}

	return rows
}

func blogEntryList(rows *sql.Rows) []BlogEntry {
	var (
		id int
		title, body, posted string
	)

	entries := []BlogEntry{}
	for rows.Next() {
		err := rows.Scan(&id, &title, &body, &posted)
		if err != nil {
			log.Error().Msg(fmt.Sprintf("failed to scan: %s", err))
			return nil
		}
		entries = append(entries, BlogEntry{ id, title, body, posted })
	}

	return entries
}

func (bdb *DB) getBodies(entries []BlogEntry) bool {

	var body bytes.Buffer

	gm := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			highlighting.Highlighting,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)	

	for ii, _ := range entries {
		buf, ok := bdb.cc.Get(entries[ii].Body)
		if !ok {
			log.Error().Msg("failed to get body")
			return false
		}
		body.Reset()
		err := gm.Convert(buf, &body)
		if err != nil {
			log.Error().Msg(fmt.Sprintf("failed to convert: %s", err))
			return false
		}

		entries[ii].Body = body.String()
	}

	return true
}
