package blogdb

import (
	"database/sql"
	"flag"
	"fmt"
	"io"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

const dbDriver = "sqlite3"
var testDbName, testDbTyoe string

func TestMain(m *testing.M) {
	flag.Parse()

	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		fmt.Printf("failed to make temp dir\n")
		os.Exit(1)
	}

	db := createTestDb(tmpDir)
	if db == nil {
		fmt.Printf("failed to create test db\n")
		os.Exit(1)
	}

	status := m.Run()

	cleanupTemp(tmpDir)

	os.Exit(status)
}

func createTestDb(tmpDir string) (*sql.DB) {
	const dbName = "BlogTest.db"

	setupSql := []string {
		"00_create.sql",	// TODO: should be $ROOT/db/00_create.sql
		"01_load.sql",		// TODO: should be $ROOT/db/01_load.sql
	}

	testDbName = tmpDir + "/" + dbName
	testDbTyoe = dbDriver
	db, err := sql.Open(testDbTyoe, testDbName)
	if err != nil {
		fmt.Printf("failed to open %s: %s\n", testDbName, err)
		return nil
	}

	for _, fn := range setupSql {
		err := execFile(db, fn)
		if err != nil {
			fmt.Printf("failed to exec %s: %s\n", fn, err)
			return nil
		}
	}

	db.Close()
	return db
}

func execFile(db *sql.DB, fn string) error {
	var err error

	fh, err := os.Open(fn)
	if err != nil {
		fmt.Printf("failed to open %s: %s\n", fn, err)
		return err
	}
	defer fh.Close()

	src, err := io.ReadAll(fh)
	if err != nil {
		fmt.Printf("failed to read %s: %s\n", fn, err)
		return err
	}

	_, err = db.Exec(string(src))
	if err != nil {
		fmt.Printf("failed to exec: %s\n", err)
		return err
	}

	return nil
}

func cleanupTemp(tmpDir string) {
	os.RemoveAll(tmpDir)
}


func TestEntries(t *testing.T) {
	db := New(testDbName)
	if db == nil {
		t.Fatalf("failed to open %s", testDbName)
	}

	entries := db.Entries()
	if entries == nil {
		t.Fatalf("failed to get entries")
	}
	if len(entries) == 0 {
		t.Fatalf("empty entries")
	}
	if len(entries) != 4 {
		t.Fatalf("wrong entries count")
	}

	mm := make(map[string]int)
	for idx, entry := range entries {
		mm[entry.Title] = idx
	}

	entry := entries[mm["Praesent iaculis nisi"]]
	if entry.Body != "7efd22f6afd98fe6f289ca5c3bf57098" || entry.Posted != "2023-01-02 00:00:00" {
		t.Fatalf("bad record: %s", entry.Title)
	}

	cnt := 0
	for _, tag := range entry.Tags {
		switch tag {
		case "baker", "foxtrot", "george", "jig", "love", "mike":
			cnt++
		}
	}
	if cnt != 6 {
		t.Fatalf("unexpected tags")
	}
}
