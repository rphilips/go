package index

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"brocade.be/base/mumps"
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

// Rebuild IIIF index (SQLite and MUMPS)
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
			key INTEGER PRIMARY KEY AUTOINCREMENT,
			id TEXT,
			digest TEXT,
			location TEXT
		);`)
	if err != nil {
		return fmt.Errorf("cannot create index database: %v", err)
	}

	Mindices := make(map[string]string)

	fn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking over file: %v", err)
		}
		if filepath.Ext(path) != ".sqlite" {
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

		stmt1, err := index.Prepare("INSERT INTO indexes (key, id, digest, location) Values($1,$2,$3,$4)")
		if err != nil {
			return fmt.Errorf("cannot prepare insert1: %v", err)
		}
		defer stmt1.Close()

		indexes := strings.Split(meta.Indexes, "^")
		for _, index := range indexes {
			if index == "" {
				continue
			}
			Mindices[index] = meta.Digest
			index = safe(index)
			_, err = stmt1.Exec(nil, index, meta.Digest, path)
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

	if len(Mindices) > 0 {
		err = SetMIndex(Mindices)
		if err != nil {
			return fmt.Errorf("error: %v", err)
		}
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

// Remove a IIIF digest and its entries from the index database
func RemoveDigest(digest string) error {
	index, err := sql.Open("sqlite", iiifIndexDb)
	if err != nil {
		return fmt.Errorf("cannot open index database: %v", err)
	}
	defer index.Close()

	_, err = index.Exec("DELETE FROM indexes where digest=?", digest)
	if err != nil {
		return fmt.Errorf("cannot delete digest from index database: %v", err)
	}

	return nil
}

// Search the index database for a search string
func Search(search string) ([][]string, error) {

	result := make([][]string, 0)
	index, err := sql.Open("sqlite", iiifIndexDb)
	if err != nil {
		return result, fmt.Errorf("cannot open index database: %v", err)
	}
	defer index.Close()

	search = safe(search)

	query := "SELECT * FROM indexes where id='" + search + "' or digest='" + search + "'"
	rows, err := index.Query(query)
	if err != nil {
		return result, fmt.Errorf("cannot query index database: %v", err)
	}
	result, err = sqlite.ReadIndexRows(rows)
	if err != nil {
		return result, fmt.Errorf("cannot read result: %v", err)
	}

	return result, nil
}

// Log index info in MUMPS
func SetMIndex(indices map[string]string) error {
	mpipe, err := mumps.Open("")
	if err != nil {
		return fmt.Errorf("mumps error:\n%s", err)
	}
	defer mpipe.Close()

	for index, value := range indices {
		cmd := `s ^BIIIF("index",2,"` + index + `")="` + value + `"`
		err = mpipe.WriteExec(cmd)
		if err != nil {
			return fmt.Errorf("mumps error:\n%s", err)
		}
	}

	cmds := []string{
		`k ^BIIIF("index",1)`,
		`m ^BIIIF("index",1)=^BIIIF("index",2)`,
		`k ^BIIIF("index",2)`}
	for _, cmd := range cmds {
		err = mpipe.WriteExec(cmd)
		if err != nil {
			return fmt.Errorf("mumps error:\n%s", err)
		}
	}

	return nil
}
