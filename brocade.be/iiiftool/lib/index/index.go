package index

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"brocade.be/base/mumps"
	"brocade.be/base/registry"
	"brocade.be/iiiftool/lib/sqlite"
	"brocade.be/iiiftool/lib/util"
)

var iifBaseDir = registry.Registry["iiif-base-dir"]
var iiifIndexDb = registry.Registry["iiif-index-db"]

const createIndexes = `
CREATE TABLE indexes (
	key INTEGER PRIMARY KEY AUTOINCREMENT,
	id TEXT,
	digest TEXT,
	location TEXT
);`

// Write IIIF index information to SQLite index database
// Return Mindices to store in MUMPS index db
func WriteIndexes(index *sql.DB, meta sqlite.Meta, path string) (map[string]string, error) {

	Mindices := make(map[string]string)

	insert, err := index.Prepare("INSERT INTO indexes (key, id, digest, location) Values($1,$2,$3,$4)")
	if err != nil {
		return Mindices, fmt.Errorf("cannot prepare insert: %v", err)
	}
	defer insert.Close()

	indexes := strings.Split(meta.Indexes, "^")
	for _, entry := range indexes {
		if entry == "" {
			continue
		}

		Mindices[entry] = meta.Digest

		// log both original and URL-safe version (for PHP endpoint)
		versions := []string{entry, util.URLSafe(entry)}
		for _, version := range versions {
			_, err = insert.Exec(nil, version, meta.Digest, path)
			if err != nil {
				// do not throw error, but allow to continue
				fmt.Printf("error executing insert: %v: %s\n", err, entry)
			}
		}
	}

	return Mindices, nil

}

// Update IIIF index (1 archive, SQLite and MUMPS)
func Update(sqlitefile string) error {

	meta, err := sqlite.ReadMetaTable(sqlitefile)
	if err != nil {
		return fmt.Errorf("cannot read meta table: %v", err)
	}

	index, err := sql.Open("sqlite", iiifIndexDb)
	if err != nil {
		return fmt.Errorf("cannot open index database: %v", err)
	}
	defer index.Close()

	_, err = index.Exec("DELETE FROM indexes WHERE digest=?", meta.Digest)
	if err != nil {
		return fmt.Errorf("cannot delete rows in index database: %v", err)
	}

	Mindices, err := WriteIndexes(index, meta, sqlitefile)
	if err != nil {
		return fmt.Errorf("cannot write index data in index database: %v", err)
	}

	err = SetMIndex(Mindices, false)
	if err != nil {
		return fmt.Errorf("cannot write MUMPS index database: %v", err)
	}

	return nil
}

// Rebuild IIIF index (all archives, SQLite and MUMPS)
func Rebuild() error {

	os.Remove(iiifIndexDb)

	index, err := sql.Open("sqlite", iiifIndexDb)
	if err != nil {
		return fmt.Errorf("cannot open index database: %v", err)
	}
	defer index.Close()

	// Caution: do not use "index" (= reserverd keyword) as table name!
	_, err = index.Exec(createIndexes)
	if err != nil {
		return fmt.Errorf("cannot create index database: %v", err)
	}

	Mindices := make(map[string]string)

	handleFile := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking over file %s: %v", path, err)
		}
		if filepath.Ext(path) != ".sqlite" {
			return nil
		}

		meta, err := sqlite.ReadMetaTable(path)
		if err != nil {
			return fmt.Errorf("error reading meta in file: %s: %v", path, err)
		}

		Mindex, err := WriteIndexes(index, meta, path)
		if err != nil {
			return fmt.Errorf("cannot write index data in index database: %v", err)
		}
		for key, value := range Mindex {
			Mindices[key] = value
		}

		return nil
	}

	err = filepath.Walk(iifBaseDir, handleFile)
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}

	if len(Mindices) > 0 {
		err = SetMIndex(Mindices, true)
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

	row := index.QueryRow("SELECT digest FROM indexes where id=?", id)
	digest, err := sqlite.ReadStringRow(row)
	if err != nil {
		return "", fmt.Errorf("error selecting digest: %v", err)
	}

	return digest, nil
}

// Remove a IIIF digest and its associated identifiers from the index database
func RemoveDigest(digest string) error {
	index, err := sql.Open("sqlite", iiifIndexDb)
	if err != nil {
		return fmt.Errorf("cannot open index database: %v", err)
	}
	defer index.Close()

	_, err = index.Exec("DELETE FROM indexes where digest=?", digest)
	if err != nil {
		return fmt.Errorf("cannot delete digest from SQLite index database: %v", err)
	}

	payload := make(map[string]string)
	rou := `d %GetIds^gbiiif(.RApayload,"` + digest + `",1)`
	oreader, _, err := mumps.Reader(rou, payload)
	if err != nil {
		return fmt.Errorf("mumps reader error:\n%s", err)
	}

	out, err := ioutil.ReadAll(oreader)
	if err != nil {
		return fmt.Errorf("error reading MUMPS response:\n%s", err)
	}

	var result map[string]string
	var identifiers []string

	err = json.Unmarshal(out, &result)
	if err != nil {
		return fmt.Errorf("json error:\n%s", err)
	}

	for id := range result {
		identifiers = append(identifiers, id)
	}

	err = KillinMIndex(digest, identifiers)
	if err != nil {
		return fmt.Errorf("error deleting digest from MUMPS index database: %v", err)
	}

	return nil
}

// Search the index database for a search string
func Search(search string) ([][]string, error) {
	result := make([][]string, 0)

	index, err := sql.Open("sqlite", iiifIndexDb)
	if err != nil {
		return result, fmt.Errorf("error opening index database: %v", err)
	}
	defer index.Close()

	query := "SELECT * FROM indexes where id='" + search + "' or digest='" + search + "'"
	rows, err := index.Query(query)
	if err != nil {
		return result, fmt.Errorf("error querying index database: %v", err)
	}
	result, err = sqlite.ReadIndexRows(rows)
	if err != nil {
		return result, fmt.Errorf("error reading result: %v", err)
	}

	return result, nil
}

// Remove index info in MUMPS
func KillinMIndex(digest string, identifiers []string) error {
	mpipe, err := mumps.Open("")
	if err != nil {
		return fmt.Errorf("mumps open error:\n%s", err)
	}
	defer mpipe.Close()

	cmds := []string{
		`k ^BIIIF("index",1,"digest2id","` + digest + `")`}

	for _, id := range identifiers {
		cmds = append(cmds, `k ^BIIIF("index",1,"id2digest","`+id+`")`)
	}

	for _, cmd := range cmds {
		err = mpipe.WriteExec(cmd)
		if err != nil {
			return fmt.Errorf("mumps exec error:\n%s", err)
		}
	}

	return nil
}

// Log index info in MUMPS
// kill=true rebuild the index from scratch
func SetMIndex(indices map[string]string, kill bool) error {
	mpipe, err := mumps.Open("")
	if err != nil {
		return fmt.Errorf("mumps open error:\n%s", err)
	}
	defer mpipe.Close()

	for id, digest := range indices {
		cmds := []string{
			`s ^BIIIF("index",2,"id2digest","` + id + `","` + digest + `")=""`,
			`s ^BIIIF("index",2,"digest2id","` + digest + `","` + id + `")=""`}

		for _, cmd := range cmds {
			err = mpipe.WriteExec(cmd)
			if err != nil {
				return fmt.Errorf("mumps exec error:\n%s", err)
			}
		}
	}

	cmds := []string{
		`m ^BIIIF("index",1)=^BIIIF("index",2)`,
		`k ^BIIIF("index",2)`}

	if kill {
		cmds = append([]string{`k ^BIIIF("index",1)`}, cmds...)
	}

	for _, cmd := range cmds {
		err = mpipe.WriteExec(cmd)
		if err != nil {
			return fmt.Errorf("mumps error:\n%s", err)
		}
	}

	return nil
}
