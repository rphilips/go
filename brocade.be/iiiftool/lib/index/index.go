package index

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	qfs "brocade.be/base/fs"
	"brocade.be/base/mumps"
	"brocade.be/base/parallel"
	"brocade.be/base/registry"
	"brocade.be/iiiftool/lib/sqlite"
)

var iifBaseDir = registry.Registry["iiif-base-dir"]
var iiifIndexDb = registry.Registry["iiif-index-db"]
var iiifMaxPar, _ = strconv.Atoi(registry.Registry["iiif-max-parallel"])

const createIndexes = `
CREATE TABLE indexes (
	key INTEGER PRIMARY KEY AUTOINCREMENT,
	loi TEXT,
	digest TEXT,
	iiifsys TEXT,
	location TEXT,
	metatime TEXT,
	sqlartime TEXT
);`

type IndexData struct {
	LOIs      []string
	Digest    string
	Iiifsys   string
	Metatime  string
	Sqlartime string
	Location  string
	Sortcode  string
}

// Update IIIF index (1 archive, SQLite and MUMPS)
func Update(sqlitefile string) error {

	meta, err := sqlite.ReadMetaTable(sqlitefile)
	if err != nil {
		return fmt.Errorf("cannot read meta table: %v", err)
	}

	// create iiifIndexDb if necessary
	if !qfs.Exists(iiifIndexDb) {
		return Rebuild(false)
	}

	// open indexes

	db, err := sql.Open("sqlite", iiifIndexDb)
	if err != nil {
		return fmt.Errorf("cannot open index database: %v", err)
	}
	defer db.Close()

	mpipe, err := mumps.Open("")
	if err != nil {
		return fmt.Errorf("mumps open error:\n%s", err)
	}
	defer mpipe.Close()

	// data
	var indexdata IndexData
	indexdata.LOIs = strings.Split(meta.Indexes, "^")
	indexdata.Digest = meta.Digest
	indexdata.Iiifsys = meta.Iiifsys
	indexdata.Location = sqlitefile
	indexdata.Sortcode = meta.Sortcode

	metatime, err := sqlite.QueryTime(sqlitefile, "meta")
	if err != nil {
		return fmt.Errorf("error reading meta update time in file: %s: %v", sqlitefile, err)
	}
	indexdata.Metatime = metatime

	sqlartime, err := sqlite.QueryTime(sqlitefile, "sqlar")
	if err != nil {
		return fmt.Errorf("error reading sqlar update time in file: %s: %v", sqlitefile, err)
	}
	indexdata.Sqlartime = sqlartime

	err = SetIndex(indexdata, db, mpipe, "update")
	if err != nil {
		return fmt.Errorf("cannot write to index: %v", err)
	}

	return nil
}

// Rebuild IIIF index (all archives, SQLite and MUMPS)
func Rebuild(verbose bool) error {

	// Remove old indices

	os.Remove(iiifIndexDb)
	KillMIndex()

	// Create SQLite index
	// Caution: do not use "index" (= reserved keyword) as table name!

	index, err := sql.Open("sqlite", iiifIndexDb)
	if err != nil {
		return fmt.Errorf("cannot open index database: %v", err)
	}
	defer index.Close()

	_, err = index.Exec(createIndexes)
	if err != nil {
		return fmt.Errorf("cannot create index database: %v", err)
	}

	// Create mpipe

	mpipe, err := mumps.Open("")
	if err != nil {
		return fmt.Errorf("mumps open error:\n%s", err)
	}
	defer mpipe.Close()

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
	indexdatas := make([]IndexData, len(archives))
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

		indexdatas[n].Metatime = metatime
		indexdatas[n].Iiifsys = meta.Iiifsys
		indexdatas[n].Sqlartime = sqlartime
		indexdatas[n].Location = archives[n]
		indexdatas[n].Digest = meta.Digest
		indexdatas[n].Sortcode = meta.Sortcode
		indexdatas[n].LOIs = strings.Split(meta.Indexes, "^")

		return nil, nil
	}

	// Read (in parallel)
	_, errors := parallel.NMap(len(archives), iiifMaxPar, handleFile)
	for _, err := range errors {
		if err != nil {
			return fmt.Errorf("cannot read data in from databases: %v", errors)
		}
	}

	// Write (sequentially for SQLite!)
	for _, data := range indexdatas {
		err := SetIndex(data, index, mpipe, "rebuild")
		if err != nil {
			return fmt.Errorf("cannot write to index: %v", err)
		}
	}

	return nil
}
