package index

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

// Read IIIF meta table from archive
func ReadMeta(path string) (sqlite.Meta, error) {
	var meta sqlite.Meta

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return meta, fmt.Errorf("cannot open archive: %v", err)
	}
	defer db.Close()

	row := db.QueryRow("SELECT * FROM meta")

	// do not throw error, to let "sql: no rows in result set" pass
	_ = sqlite.ReadMetaRow(row, &meta)

	return meta, nil
}

// Write IIIF index information to SQLite index database
// Return Mindices to store in MUMPS index db
func WriteIndexes(index *sql.DB, meta sqlite.Meta, path string) (map[string]string, error) {

	Mindices := make(map[string]string)

	insert, err := index.Prepare("INSERT INTO indexes (key, id, digest, location) Values($1,$2,$3,$4)")
	if err != nil {
		return Mindices, fmt.Errorf("cannot prepare insert1: %v", err)
	}
	defer insert.Close()

	indexes := strings.Split(meta.Indexes, "^")
	for _, entry := range indexes {
		if entry == "" {
			continue
		}
		Mindices[entry] = meta.Digest
		entry = safe(entry)
		_, err = insert.Exec(nil, entry, meta.Digest, path)
		if err != nil {
			// do not throw error, but allow to continue
			fmt.Printf("Error executing insert: %v: %s\n", err, entry)
		}
	}

	return Mindices, nil

}

// Update IIIF index (1 archive, SQLite and MUMPS)
func Update(sqlitefile string) error {

	meta, err := ReadMeta(sqlitefile)
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

	// Update L (delete, and write new)

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

		meta, err := ReadMeta(path)
		if err != nil {
			return fmt.Errorf("cannot read meta: %v", err)
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

	err = filepath.Walk(iifBaseDir, fn)
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

	id = safe(id)
	row := index.QueryRow("SELECT digest FROM indexes where id=?", id)
	digest := util.ReadStringRow(row)

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
		return fmt.Errorf("mumps error:\n%s", err)
	}
	out, err := ioutil.ReadAll(oreader)

	if err != nil {
		return fmt.Errorf("mumps error:\n%s", err)
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

	err = KillfromMIndex(digest, identifiers)
	if err != nil {
		return fmt.Errorf("cannot delete digest from MUMPS index database: %v", err)
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

// Remove index info in MUMPS
func KillfromMIndex(digest string, identifiers []string) error {
	mpipe, err := mumps.Open("")
	if err != nil {
		return fmt.Errorf("mumps error:\n%s", err)
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
			return fmt.Errorf("mumps error:\n%s", err)
		}
	}

	return nil
}

// Log index info in MUMPS
// kill=true rebuild the index from scratch
func SetMIndex(indices map[string]string, kill bool) error {
	mpipe, err := mumps.Open("")
	if err != nil {
		return fmt.Errorf("mumps error:\n%s", err)
	}
	defer mpipe.Close()

	for id, digest := range indices {
		cmds := []string{
			`s ^BIIIF("index",2,"id2digest","` + id + `","` + digest + `")=""`,
			`s ^BIIIF("index",2,"digest2id","` + digest + `","` + id + `")=""`}

		for _, cmd := range cmds {
			err = mpipe.WriteExec(cmd)
			if err != nil {
				return fmt.Errorf("mumps error:\n%s", err)
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
