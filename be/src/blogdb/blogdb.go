package blogdb

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	log "main/logger"
)

type DB struct {
	path string
	db *sql.DB
}

type BlogEntry struct {
	Id int 			`json:"uniqueId"`
	Title string 	`json:"title"`
	Body string 	`json:"body"`
	Posted string 	`json:"posted"`
	Tags []string 	`json:"tags"`
}

func New(path string) *DB {
	var err error

	d := DB{ path: path }
	d.db, err = sql.Open("sqlite3", d.path)
	if err != nil {
		log.Fatal("failed to open %s: %s", path, err)
	}

	return &d
}

func (d *DB) Close() {
	log.Info("d.Close")
	d.db.Close()
}

func (d *DB) Entries() []BlogEntry {
	// log.Trace("blogdb.Entries")

	rows := d.getEntryRows()
	if rows == nil {
		log.Error("Entries failed to get rows")
		return nil
	}
	defer rows.Close()

	entries := d.getList(rows)
	if entries == nil {
		log.Error("Entries failed to get list")
		return nil
	}
	if len(entries) == 0 {
		// no records is ok
		return entries
	}

	if !d.getTags(entries) {
		// a db problem?
		log.Error("failed to get tags")
		return nil
	}

	return entries
}

// TODO: fix this
func (d *DB) getTags(entries []BlogEntry) bool {
	// log.Trace("blogdb.getTags")

	uids := []string{}
	for idx := range entries {
		uids = append(uids, fmt.Sprintf("%d", entries[idx].Id))
	}
	if len(uids) == 0 {
		// no entries is ok
		return true
	}

	rows := d.getTagRows(uids)
	if rows == nil {
		// no tags is ok
		return true
	}

	var (
		uid int
		tag string
	)
	mm := make(map[int][]string)
	for rows.Next() {
		err := rows.Scan(&uid, &tag)
		if err != nil {
			log.Error("failed to scan: %s", err)
			return false
		}

		mm[uid] = append(mm[uid], tag)
	}

	for idx := range entries {
		entries[idx].Tags = mm[entries[idx].Id]
	}

	return true
}

func (d *DB) getEntryRows() *sql.Rows {
	rows := d.getRows("SELECT id, title, body, posted FROM Entries WHERE visible ORDER BY posted DESC")
	if rows == nil {
		log.Error("getEntryRows failed")
	}
	return rows
}

func (d *DB) getTagRows(uids []string) *sql.Rows {
	// log.Trace("blogdb.getTagRows")

	sql := strings.Join(uids, "', '")
	sql = "SELECT entryUid, tag FROM Tags WHERE entryUid IN ('" + sql + "')"
	return d.getRows(sql)
}

func (d *DB) getRows(sql string) *sql.Rows {
	rows, err := d.db.Query(sql)
	if err != nil {
		log.Error("failed to query %s: %s", d.path, err)
		return nil
	}

	return rows
}

func (d *DB) getList(rows *sql.Rows) []BlogEntry {
	// log.Trace("blogdb.getList")

	var (
		id int
		title, body, posted string
	)

	entries := []BlogEntry{}
	for rows.Next() {
		err := rows.Scan(&id, &title, &body, &posted)
		if err != nil {
			log.Error("failed to scan: %s", err)
			return nil
		}

		entries = append(entries, BlogEntry{Id: id, Title: title, Body: body, Posted: posted, Tags: []string{}})
	}

	return entries
}
