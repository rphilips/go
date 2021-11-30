package index

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"brocade.be/base/registry"
	"brocade.be/iiiftool/lib/sqlite"
)

var iifBaseDir = registry.Registry["iiif-base-dir"]

// Rebuild IIIF index
func Rebuild() error {

	indexDb := filepath.Join(iifBaseDir, "index.sqlite")

	os.Remove(indexDb)

	index, err := sql.Open("sqlite", indexDb)
	if err != nil {
		return fmt.Errorf("cannot open index database: %v", err)
	}
	defer index.Close()

	// note: do not use "index" or "indexes" as table name in SQLite!
	_, err = index.Exec(`
		CREATE TABLE indexes (
			id TEXT PRIMARY KEY,
			digest TEXT
		);`)
	if err != nil {
		return fmt.Errorf("cannot create index database: %v", err)
	}

	fn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking over file: %v", err)
		}
		ext := filepath.Ext(path)
		if ext != ".sqlite" {
			return nil
		}

		db, err := sql.Open("sqlite", path)
		if err != nil {
			return fmt.Errorf("cannot open archive: %v", err)
		}
		defer db.Close()

		row := db.QueryRow("SELECT * FROM meta")

		var meta sqlite.Meta
		// do not throw error, to let "sql: no rows in result set" pass
		_ = sqlite.ReadMetaRow(row, &meta)

		stmt1, err := index.Prepare("INSERT INTO indexes (id, digest) Values($1,$2)")
		if err != nil {
			return fmt.Errorf("cannot prepare insert1: %v", err)
		}
		defer stmt1.Close()

		indexes := strings.Split(meta.Indexes, "^")
		for _, index := range indexes {
			if index == "" {
				continue
			}
			index = strings.ToLower(index)
			unsafeRegexp := regexp.MustCompile(`[^a-z0-9]`)
			index = unsafeRegexp.ReplaceAllString(index, "_")
			_, err = stmt1.Exec(index, meta.Digest)
			if err != nil {
				// do not throw error
				fmt.Printf("Error executing stmt1: %v: %s\n", err, index)
			}
		}

		return nil
	}

	err = filepath.Walk(iifBaseDir, fn)
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}

	return nil
}
