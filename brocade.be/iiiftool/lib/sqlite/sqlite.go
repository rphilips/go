package sqlite

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	basefs "brocade.be/base/fs"
	"brocade.be/base/registry"
	"brocade.be/iiiftool/lib/iiif"
	"brocade.be/iiiftool/lib/util"
	_ "modernc.org/sqlite"
)

// CONSTANTS
var osSep = registry.Registry["os-sep"]
var user = registry.Registry["qtechng-user"]

type Sqlar struct {
	Name   string
	Mode   int64
	Mtime  time.Time
	Sz     int64
	Reader *bytes.Reader
}

type Meta struct {
	Key        string
	Digest     string
	Identifier string
	Indexes    string
	Imgloi     string
	Iiifsys    string
	Manifest   string
}

// Given a IIIF identifier and some io.Readers
// store the contents in the appropriate SQLite archive
func Store(sqlitefile string,
	filestream []io.Reader,
	cwd string,
	mResponse iiif.MResponse) error {

	if cwd == "" {
		path := strings.Split(sqlitefile, osSep)
		dirname := strings.Join(path[0:(len(path)-1)], osSep)
		err := basefs.Mkdir(dirname, "process")
		if err != nil {
			return fmt.Errorf("cannot make dir")
		}
	} else {
		if !basefs.IsDir(cwd) {
			return fmt.Errorf("cwd is not valied")
		}
		sqlitefile = filepath.Join(cwd, filepath.Base(sqlitefile))
	}

	db, err := sql.Open("sqlite", sqlitefile)
	if err != nil {
		return fmt.Errorf("cannot open file: %v", err)
	}
	defer db.Close()

	append := basefs.Exists(sqlitefile)
	if !append {

		if _, err = db.Exec(`
		CREATE TABLE sqlar (
			name TEXT PRIMARY KEY,
			mode INT,
			mtime INT,
  			sz INT,
  			data BLOB
		);`); err != nil {
			return fmt.Errorf("cannot create table sqlar: %v", err)
		}

		if _, err = db.Exec(`
		CREATE TABLE admin (
			key INTEGER PRIMARY KEY AUTOINCREMENT,
			time TEXT,
			action TEXT,
			user TEXT
		);`); err != nil {
			return fmt.Errorf("cannot create table admin: %v", err)
		}

		if _, err = db.Exec(`
		CREATE TABLE files (
			key INTEGER PRIMARY KEY AUTOINCREMENT,
			docman TEXT,
			name TEXT
		);`); err != nil {
			return fmt.Errorf("cannot create table files: %v", err)
		}

		if _, err = db.Exec(`
		CREATE TABLE meta (
			key INTEGER PRIMARY KEY AUTOINCREMENT,
			digest TEXT,
			identifier TEXT,
			indexes TEXT,
			imgloi TEXT,
			iiifsys TEXT,
			manifest TEXT
		);`); err != nil {
			return fmt.Errorf("cannot create table meta: %v", err)
		}
	}

	stmt1, err := db.Prepare("INSERT INTO sqlar (name, mode, mtime, sz, data) Values($1,$2,$3,$4,$5)")
	if err != nil {
		return fmt.Errorf("cannot prepare insert1: %v", err)
	}
	defer stmt1.Close()

	stmt2, err := db.Prepare("INSERT INTO admin (key, time, action, user) Values($1,$2,$3,$4)")
	if err != nil {
		return fmt.Errorf("cannot prepare insert2: %v", err)
	}
	defer stmt2.Close()

	if !append {
		h := time.Now()
		_, err = stmt2.Exec(nil, h.Format(time.RFC3339), "created", user)
	}
	h := time.Now()
	stmt2.Exec(nil, h.Format(time.RFC3339), "modified", user)

	stmt3, err := db.Prepare("INSERT INTO files (key, docman, name) Values($1,$2,$3)")
	if err != nil {
		return fmt.Errorf("cannot prepare insert3: %v", err)
	}
	defer stmt3.Close()

	stmt4, err := db.Prepare("INSERT INTO meta (key, digest, identifier, indexes, iiifsys, imgloi, manifest) Values($1,$2,$3,$4,$5,$6,$7)")
	if err != nil {
		return fmt.Errorf("cannot prepare insert4: %v", err)
	}
	defer stmt4.Close()

	manifest, err := json.Marshal(mResponse.Manifest)
	index := strings.Join(mResponse.Index, "^")
	_, err = stmt4.Exec(nil, mResponse.Digest, mResponse.Identifier, index, mResponse.Iiifsys, mResponse.Imgloi, string(manifest))
	if err != nil {
		return fmt.Errorf("cannot exec stmt4: %v", err)
	}

	sqlar := func(docman string, name string, stream io.Reader) error {

		row := db.QueryRow("SELECT name FROM sqlar WHERE name =?", name)

		update := !(util.ReadStringRow(row) == "")
		if update {
			_, err = db.Exec("DELETE FROM sqlar WHERE name=?", name)
			if err != nil {
				return fmt.Errorf("cannot delete file from archive: %v", err)
			}
		}

		data, _ := ioutil.ReadAll(stream)
		mtime := time.Now().Unix()
		// mode := basefs.CalcPerm("rw-rw-rw-")
		mode := int64(33204)
		sz := int64(len(data))
		_, err = stmt1.Exec(name, mode, mtime, sz, data)
		if err != nil {
			return fmt.Errorf("cannot exec stmt1: %v", err)
		}
		if !update {
			_, err = stmt3.Exec(nil, docman, name)
			if err != nil {
				return fmt.Errorf("cannot exec stmt3: %v", err)
			}
		}
		return nil
	}

	for i, stream := range filestream {
		docman := ""
		if len(mResponse.Images) != 0 {
			docman = mResponse.Images[i]["loc"]
		}
		err = sqlar(docman, mResponse.Images[i]["name"], stream)
		if err != nil {
			return err
		}
	}

	return nil
}

// Given a SQLite archive and a table name show the contents of that table
// version with sqlite3
func Inspect(sqlitefile string, table string) (interface{}, error) {

	var query string
	switch {
	case table == "sqlar":
		query = "SELECT name, mode, mtime, sz FROM sqlar"
	case table == "":
		query = `SELECT m.name as tables, group_concat(p.name,';') as columns FROM sqlite_master AS m
		JOIN pragma_table_info(m.name) AS p
		GROUP BY m.name
		ORDER BY m.name, p.cid`
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

// Function that reads a single sqlar row sql.Row
func readSqlarRow(row *sql.Row, sqlar *Sqlar) error {
	var data []byte
	var mtime int64
	err := row.Scan(
		&sqlar.Name,
		&sqlar.Mode,
		mtime,
		&sqlar.Sz,
		&data)
	sqlar.Reader = bytes.NewReader(data)
	sqlar.Mtime = time.Unix(0, mtime)
	if err != nil {
		return err
	}
	return nil
}

// Function that reads a single meta row sql.Row
func ReadMetaRow(row *sql.Row, meta *Meta) error {
	err := row.Scan(
		&meta.Key,
		&meta.Digest,
		&meta.Identifier,
		&meta.Indexes,
		&meta.Imgloi,
		&meta.Iiifsys,
		&meta.Manifest)
	if err != nil {
		return err
	}
	return nil
}

// Function that reads multiple Index sql.Rows
func ReadIndexRows(rows *sql.Rows) ([][]string, error) {

	result := make([][]string, 0)

	// key|id|digest|location
	columns, err := rows.Columns()
	if err != nil {
		return result, err
	}

	for rows.Next() {

		data := make([]string, len(columns))
		err := rows.Scan(&data[0], &data[1], &data[2], &data[3])
		if err != nil {
			return result, err
		}

		result = append(result, data)
	}
	if err := rows.Err(); err != nil {
		return result, err
	}
	return result, nil
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
	err = readSqlarRow(row, sqlar)
	if err != nil {
		return fmt.Errorf("cannot read file contents from archive: %v", err)
	}

	return nil
}
