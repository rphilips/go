package index

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"brocade.be/base/mumps"
	"brocade.be/base/parallel"
	"brocade.be/base/registry"
	qtime "brocade.be/base/time"
	"brocade.be/iiiftool/lib/sqlite"
)

var iifBaseDir = registry.Registry["iiif-base-dir"]
var iiifIndexDb = registry.Registry["iiif-index-db"]
var iiifMaxPar, _ = strconv.Atoi(registry.Registry["iiif-max-parallel"])

const createIndexes = `
CREATE TABLE indexes (
	key INTEGER PRIMARY KEY AUTOINCREMENT,
	identifier TEXT,
	iiifsys TEXT,
	digest TEXT,
	location TEXT,
	metatime TEXT,
	sqlartime TEXT
);`

type IndexData struct {
	Index     *sql.DB
	Meta      sqlite.Meta
	Metatime  string
	Sqlartime string
	Path      string
}

// Write IIIF index information to SQLite index database
// Return Mindices to store in MUMPS index db
func WriteIndexes(index IndexData) (map[string][]string, error) {

	Mindices := make(map[string][]string)

	db := index.Index
	insert, err := db.Prepare("INSERT INTO indexes (key, identifier, digest, location, metatime, sqlartime) Values($1,$2,$3,$4,$5,$6)")
	if err != nil {
		return Mindices, fmt.Errorf("cannot prepare insert: %v", err)
	}
	defer insert.Close()

	meta := index.Meta

	_, err = insert.Exec(nil, meta.Identifier, meta.Digest, index.Path, index.Metatime, index.Sqlartime)
	if err != nil {
		// do not throw error, but allow to continue
		fmt.Printf("error executing insert: %v", err)
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

	metatime, err := sqlite.QueryTime(sqlitefile, "meta")
	if err != nil {
		return fmt.Errorf("error reading meta update time in file: %s: %v", sqlitefile, err)
	}

	sqlartime, err := sqlite.QueryTime(sqlitefile, "sqlar")
	if err != nil {
		return fmt.Errorf("error reading sqlar update time in file: %s: %v", sqlitefile, err)
	}

	var indexInfo IndexData

	indexInfo.Index = index
	indexInfo.Meta = meta
	indexInfo.Metatime = metatime
	indexInfo.Sqlartime = sqlartime
	indexInfo.Path = sqlitefile

	Mindices, err := WriteIndexes(indexInfo)
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
func Rebuild(verbose bool) error {

	os.Remove(iiifIndexDb)

	index, err := sql.Open("sqlite", iiifIndexDb)
	if err != nil {
		return fmt.Errorf("cannot open index database: %v", err)
	}
	defer index.Close()

	// Caution: do not use "index" (= reserved keyword) as table name!
	_, err = index.Exec(createIndexes)
	if err != nil {
		return fmt.Errorf("cannot create index database: %v", err)
	}

	// Collect archives
	var archives []string
	err = filepath.Walk(iifBaseDir, func(path string, info os.FileInfo, err error) error {
		if path == iiifIndexDb {
			return nil
		}
		if err != nil {
			return fmt.Errorf("error walking over file %s: %v", path, err)
		}
		if filepath.Ext(path) != ".sqlite" {
			return nil
		}
		archives = append(archives, path)
		return nil
	})
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}

	// Collect index information
	indexInfos := make([]IndexData, len(archives))
	handleFile := func(n int) (interface{}, error) {

		if verbose {
			fmt.Println(archives[n])
		}

		meta, err := sqlite.ReadMetaTable(archives[n])
		if err != nil {
			return nil, fmt.Errorf("error reading meta in file: %s: %v", archives[n], err)
		}

		metatime, err := sqlite.QueryTime(archives[n], "meta")
		if err != nil {
			return nil, fmt.Errorf("error reading meta update time in file: %s: %v", archives[n], err)
		}

		sqlartime, err := sqlite.QueryTime(archives[n], "sqlar")
		if err != nil {
			return nil, fmt.Errorf("error reading sqlar update time in file: %s: %v", archives[n], err)
		}

		indexInfos[n].Index = index
		indexInfos[n].Meta = meta
		indexInfos[n].Metatime = metatime
		indexInfos[n].Sqlartime = sqlartime
		indexInfos[n].Path = archives[n]

		return nil, nil
	}

	// Read in parallel
	_, errors := parallel.NMap(len(archives), iiifMaxPar, handleFile)
	for _, err := range errors {
		if err != nil {
			return fmt.Errorf("cannot read data in from databases: %v", errors)
		}
	}

	// Write sequentially
	Mindices := make(map[string][]string)
	for _, info := range indexInfos {
		Mindex, err := WriteIndexes(info)
		if err != nil {
			return fmt.Errorf("cannot write index data in index database: %v", err)
		}
		for key, values := range Mindex {
			Mindices[key] = values
		}
	}

	// Set M index
	if len(Mindices) > 0 {
		err = SetMIndex(Mindices, true)
		if err != nil {
			return fmt.Errorf("cannot write index data to M: %v", err)
		}
	}

	return nil
}

// Given a IIIF identifier, lookup its digest
// in the index database
func LookupId(identifier string) (string, error) {
	index, err := sql.Open("sqlite", iiifIndexDb)
	if err != nil {
		return "", fmt.Errorf("cannot open index database: %v", err)
	}
	defer index.Close()

	row := index.QueryRow("SELECT digest FROM indexes where identifier=?", identifier)
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

	for identifier := range result {
		identifiers = append(identifiers, identifier)
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

	query := "SELECT * FROM indexes where identifier='" + search + "' or digest='" + search + "'"
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

	for _, identifier := range identifiers {
		loi := strings.Split(identifier, ",")[0]
		loi = strings.TrimRight(loi, ",")
		cmds = append(cmds, `k ^BIIIF("index",1,"id2digest","`+loi+`","`+loi+`","`+digest+`")`)
		cmds = append(cmds, `k ^BIIIF("index",1,"id2digest","`+loi+`","`+identifier+`","`+digest+`")`)
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
func SetMIndex(indices map[string][]string, kill bool) error {
	return nil
	mpipe, err := mumps.Open("")
	if err != nil {
		return fmt.Errorf("mumps open error:\n%s", err)
	}
	defer mpipe.Close()

	for identifier, values := range indices {
		loi := strings.Split(identifier, ",")[0]
		loi = strings.TrimRight(loi, ",")
		digest := values[0]
		metaTimeObject, _ := time.Parse(time.RFC3339, values[1])
		sqlarTimeObject, _ := time.Parse(time.RFC3339, values[2])
		metatime := qtime.H(metaTimeObject)
		sqlartime := qtime.H(sqlarTimeObject)

		cmds := []string{
			`s ^BIIIF("index",2,"id2digest","` + loi + `","` + loi + `","` + digest + `")="` + metatime + `^` + sqlartime + `"`,
			`s ^BIIIF("index",2,"id2digest","` + loi + `","` + identifier + `","` + digest + `")="` + metatime + `^` + sqlartime + `"`,
			`s ^BIIIF("index",2,"digest2id","` + digest + `","` + loi + `","` + identifier + `")="` + metatime + `^` + sqlartime + `"`}

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
