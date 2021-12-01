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
	"brocade.be/iiiftool/lib/util"
)

var iifBaseDir = registry.Registry["iiif-base-dir"]
var iiifIndexDb = registry.Registry["iiif-index-db"]

// Make identifier safe for index
func safe(id string) string {
	id = strings.ToLower(id)
	unsafeRegexp := regexp.MustCompile(`[^a-z0-9]`)
	id = unsafeRegexp.ReplaceAllString(id, "_")
	return id
}

// Rebuild IIIF index
func Rebuild() error {

	os.Remove(iiifIndexDb)

	index, err := sql.Open("sqlite", iiifIndexDb)
	if err != nil {
		return fmt.Errorf("cannot open index database: %v", err)
	}
	defer index.Close()

	// note: do not use "index" or "indexes" as table name in SQLite!
	_, err = index.Exec(`
		CREATE TABLE indexes (
			id TEXT PRIMARY KEY,
			digest TEXT,
			location TEXT
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

		stmt1, err := index.Prepare("INSERT INTO indexes (id, digest, location) Values($1,$2,$3)")
		if err != nil {
			return fmt.Errorf("cannot prepare insert1: %v", err)
		}
		defer stmt1.Close()

		indexes := strings.Split(meta.Indexes, "^")
		for _, index := range indexes {
			if index == "" {
				continue
			}
			index = safe(index)
			_, err = stmt1.Exec(index, meta.Digest, path)
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

// Given a IIIF identifier, lookup its digest
// in the index database
func LookupId(id string) (string, error) {
	index, err := sql.Open("sqlite", iiifIndexDb)
	if err != nil {
		return "", fmt.Errorf("cannot open index database: %v", err)
	}
	defer index.Close()

	id = safe(id)
	row := index.QueryRow("SELECT digest FROM indexes where id=?", id)
	digest := util.ReadStringRow(row)

	return digest, nil
}
