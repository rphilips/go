package sqlite

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	basefs "brocade.be/base/fs"
	"brocade.be/base/registry"
	qdigest "brocade.be/iiiftool/lib/digest"
	"brocade.be/iiiftool/lib/iiif"
	_ "modernc.org/sqlite"
)

// CONSTANTS
var user = registry.Registry["qtechng-user"]

const createSqlar = `
CREATE TABLE sqlar (
	name TEXT PRIMARY KEY,
	mode INT,
	mtime INT,
	sz INT,
	data BLOB
);`

const createAdmin = `
CREATE TABLE admin (
	key INTEGER PRIMARY KEY AUTOINCREMENT,
	time TEXT,
	action TEXT,
	user TEXT,
	fileshash TEXT
);`

const createFiles = `
CREATE TABLE files (
	key INTEGER PRIMARY KEY AUTOINCREMENT,
	docman TEXT,
	name TEXT
);`

const createMeta = `
CREATE TABLE meta (
	key INTEGER PRIMARY KEY AUTOINCREMENT,
	digest TEXT,
	indexes TEXT,
	imgloi TEXT,
	iiifsys TEXT,
	sortcode TEXT,
	manifest TEXT
);`

const selectDbInfo = `
SELECT m.name as tables, group_concat(p.name,';') as columns FROM sqlite_master AS m
JOIN pragma_table_info(m.name) AS p
GROUP BY m.name
ORDER BY m.name, p.cid`

const selectSqlar = "SELECT name, mode, mtime, sz FROM sqlar"

const insertMeta = "INSERT INTO meta (key, digest, indexes, iiifsys, sortcode, imgloi, manifest) Values($1,$2,$3,$4,$5,$6,$7)"

// Structs

type Sqlar struct {
	Sz     int64
	Name   string
	Mode   int64
	Mtime  time.Time
	Reader *bytes.Reader
}

type Meta struct {
	Key      string
	Digest   string
	Imgloi   string
	Indexes  string
	Iiifsys  string
	Sortcode string
	Manifest string
}

// Given a IIIF digest and an io.Reader
// create the appropriate SQLite archive
// and store the contents.
func Create(sqlitefile string,
	filestream []io.Reader,
	cwd string,
	iiifMeta iiif.IIIFmeta) (string, error) {

	// Validate input

	if cwd == "" {
		directory := filepath.Dir(sqlitefile)
		err := basefs.Mkdir(directory, "process")
		if err != nil {
			return "", fmt.Errorf("cannot make dir")
		}
	} else {
		if !basefs.IsDir(cwd) {
			return "", fmt.Errorf("cwd is not valied")
		}
		sqlitefile = filepath.Join(cwd, filepath.Base(sqlitefile))
	}

	if basefs.Exists(sqlitefile) {
		err := basefs.Rmpath(sqlitefile)
		if err != nil {
			return "", fmt.Errorf("cannot remove file: %v", err)
		}
	}

	// Create tables

	db, err := sql.Open("sqlite", sqlitefile)
	if err != nil {
		return "", fmt.Errorf("cannot open file: %v", err)
	}
	defer db.Close()

	if _, err = db.Exec(createSqlar); err != nil {
		return "", fmt.Errorf("cannot create table sqlar: %v", err)
	}

	if _, err = db.Exec(createAdmin); err != nil {
		return "", fmt.Errorf("cannot create table admin: %v", err)
	}

	if _, err = db.Exec(createFiles); err != nil {
		return "", fmt.Errorf("cannot create table files: %v", err)
	}

	if _, err = db.Exec(createMeta); err != nil {
		return "", fmt.Errorf("cannot create table meta: %v", err)
	}

	stmt1, err := db.Prepare("INSERT INTO sqlar (name, mode, mtime, sz, data) Values($1,$2,$3,$4,$5)")
	if err != nil {
		return "", fmt.Errorf("cannot prepare insert1: %v", err)
	}
	defer stmt1.Close()

	stmt2, err := db.Prepare("INSERT INTO admin (key, time, action, user, fileshash) Values($1,$2,$3,$4,$5)")
	if err != nil {
		return "", fmt.Errorf("cannot prepare insert2: %v", err)
	}
	defer stmt2.Close()

	stmt3, err := db.Prepare("INSERT INTO files (key, docman, name) Values($1,$2,$3)")
	if err != nil {
		return "", fmt.Errorf("cannot prepare insert3: %v", err)
	}
	defer stmt3.Close()

	stmt4, err := db.Prepare(insertMeta)
	if err != nil {
		return "", fmt.Errorf("cannot prepare insert4: %v", err)
	}
	defer stmt4.Close()

	// Insert into "admin", "files"

	content, err := json.Marshal(iiifMeta.Manifest)
	manifest := string(content)
	if err != nil {
		return "", fmt.Errorf("json error on stmt4: %v", err)
	}

	sqlar := func(docman string, name string, stream io.Reader) error {

		data, err := ioutil.ReadAll(stream)
		if err != nil {
			return fmt.Errorf("cannot read stream: %v", err)
		}
		mtime := time.Now().Unix() // must be int in sqlar specification!
		props, _ := basefs.Properties("nakedfile")
		mode := int64(props.PERM)
		sz := int64(len(data))
		_, err = stmt1.Exec(name, mode, mtime, sz, data)
		if err != nil {
			return fmt.Errorf("cannot exec stmt1: %v", err)
		}
		_, err = stmt3.Exec(nil, docman, name)
		if err != nil {
			return fmt.Errorf("cannot exec stmt3: %v", err)
		}
		return nil
	}

	for i, stream := range filestream {
		docman := ""
		if len(iiifMeta.Images) != 0 {
			docman = iiifMeta.Images[i]["loc"]
		}
		err = sqlar(docman, iiifMeta.Images[i]["name"], stream)
		if err != nil {
			return "", err
		}
	}

	// Insert into "meta", "admin"

	filesHash, digest, err := qdigest.CalculateDigest(filestream, manifest, iiifMeta.Iiifsys)
	if err != nil {
		return "", fmt.Errorf("cannot calculate digest: %v", err)
	}

	indexes := strings.Join(iiifMeta.Indexes, "^")
	manifest = strings.ReplaceAll(manifest, `\u003c-digest-\u003e`, digest)
	_, err = stmt4.Exec(nil, digest, indexes, iiifMeta.Iiifsys, iiifMeta.Info["sortid"], iiifMeta.Imgloi, manifest)
	if err != nil {
		return "", fmt.Errorf("cannot exec stmt4: %v", err)
	}

	h := time.Now()
	_, err = stmt2.Exec(nil, h.Format(time.RFC3339), "created", user, filesHash)
	if err != nil {
		return "", fmt.Errorf("cannot execute smt2: %v", err)
	}

	return digest, nil
}

// Given a SQLite archive and a table name show the contents of that table
// version with sqlite3
func Inspect(sqlitefile string, table string) (interface{}, error) {

	if !basefs.Exists(sqlitefile) {
		return "", fmt.Errorf("file does not exist: %s", sqlitefile)
	}

	var query string
	switch {
	case table == "sqlar":
		query = selectSqlar
	case table == "":
		query = selectDbInfo
	default:
		query = "SELECT * FROM " + table
	}

	cmd := exec.Command("sqlite3", sqlitefile, query, "-header")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("cannot inspect file %v: %s", sqlitefile, err)
	}

	return string(out), nil
}

// Given a IIIF harvest code, i.e. digest with filepath,
// e.g. a42f98d253ea3dd019de07870862cbdc62d6077c00000001.jp2.
// Return that filename as a stream
func Harvest(harvestcode string, sqlar *Sqlar) error {

	digest := harvestcode[0:40]
	sqlitefile := iiif.Digest2Location(digest)
	file := harvestcode[40:]

	db, err := sql.Open("sqlite", sqlitefile)
	if err != nil {
		return fmt.Errorf("cannot open file: %v", err)
	}
	defer db.Close()

	row := db.QueryRow("SELECT * FROM sqlar WHERE name =?", file)
	err = ReadSqlarRow(row, sqlar)
	if err != nil {
		return fmt.Errorf("cannot read file contents from archive: %v", err)
	}

	return nil
}

// Query IIIF sqlitefile for update times
func QueryTime(sqlitefile string, mode string) (string, error) {
	db, err := sql.Open("sqlite", sqlitefile)
	if err != nil {
		return "", fmt.Errorf("cannot open file: %v", err)
	}
	defer db.Close()

	result := ""

	if mode == "meta" {
		row := db.QueryRow("SELECT time FROM admin WHERE action='update meta' ORDER BY time DESC LIMIT 1", sqlitefile)
		result, err = ReadStringRow(row)
		if err != nil {
			row = db.QueryRow("SELECT time FROM admin WHERE action='created'", sqlitefile)
			result, err = ReadStringRow(row)
			if err != nil {
				return "", fmt.Errorf("cannot query meta update time from archive: %v", err)
			}
		}
	} else if mode == "sqlar" {
		var sqlar Sqlar
		row := db.QueryRow("SELECT * FROM sqlar ORDER BY mtime DESC LIMIT 1", sqlitefile)
		err = ReadSqlarRow(row, &sqlar)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// in case of empty sqlar table with no images
				return "", nil
			}
			return "", fmt.Errorf("cannot query sqlar time from archive: %v", err)
		}
		result = (sqlar.Mtime).Format(time.RFC3339)
	}

	return result, nil
}

func ReplaceMeta(sqlitefile string, iiifMeta iiif.IIIFmeta) error {
	db, err := sql.Open("sqlite", sqlitefile)
	if err != nil {
		return fmt.Errorf("cannot open file: %v", err)
	}
	defer db.Close()

	// update meta
	_, err = db.Exec("DELETE FROM meta")
	if err != nil {
		return fmt.Errorf("cannot delete meta from archive: %v", err)
	}

	stmt, err := db.Prepare(insertMeta)
	if err != nil {
		return fmt.Errorf("cannot prepare replacemeta insert statement: %v", err)
	}
	defer stmt.Close()

	data, err := json.Marshal(iiifMeta.Manifest)
	manifest := string(data)
	if err != nil {
		return fmt.Errorf("json error on replacemeta: %v", err)
	}
	indexes := strings.Join(iiifMeta.Indexes, "^")
	_, err = stmt.Exec(nil, iiifMeta.Info["olddigest"], indexes, iiifMeta.Iiifsys, iiifMeta.Info["sortid"], iiifMeta.Imgloi, manifest)
	if err != nil {
		return fmt.Errorf("cannot execute replacemeta statement: %v", err)
	}

	// update admin
	query := "SELECT admin.key, admin.fileshash FROM admin ORDER BY CAST(admin.key AS INT) DESC LIMIT 1"
	row := db.QueryRow(query, sqlitefile)
	fileshash, err := ReadStringRow(row)
	if err != nil {
		return fmt.Errorf("cannot query meta update time from archive: %v", err)
	}
	stmt2, err := db.Prepare("INSERT INTO admin (key, time, action, user, fileshash) Values($1,$2,$3,$4,$5)")
	if err != nil {
		return fmt.Errorf("cannot prepare insert2: %v", err)
	}
	defer stmt2.Close()

	h := time.Now()
	_, err = stmt2.Exec(nil, h.Format(time.RFC3339), "update meta", user, fileshash)
	if err != nil {
		return fmt.Errorf("cannot execute insert2: %v", err)
	}

	return nil

}
