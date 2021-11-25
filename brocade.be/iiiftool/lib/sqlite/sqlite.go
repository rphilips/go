package sqlite

import (
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"brocade.be/base/fs"
	"brocade.be/base/registry"
	"brocade.be/iiiftool/lib/identifier"
	"brocade.be/iiiftool/lib/util"

	_ "modernc.org/sqlite"
)

var osSep = registry.Registry["os-sep"]
var user = registry.Registry["user-default"]

// Given a IIIF identifier and some io.Readers
// store the contents in the appropriate SQLite archive
func Store(id identifier.Identifier, sqlitefile string, filestream map[string]io.Reader, cwd string) error {

	if cwd == "" {
		path := strings.Split(sqlitefile, osSep)
		dirname := strings.Join(path[0:(len(path)-1)], osSep)
		err := fs.Mkdir(dirname, "process")
		if err != nil {
			return fmt.Errorf("cannot make dir")
		}
	} else {
		if !fs.IsDir(cwd) {
			return fmt.Errorf("cwd is not valied")
		}
		sqlitefile = filepath.Join(cwd, filepath.Base(sqlitefile))
	}

	db, err := sql.Open("sqlite", sqlitefile)
	if err != nil {
		return fmt.Errorf("cannot open file: %v", err)
	}
	defer db.Close()

	append := fs.Exists(sqlitefile)
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
		CREATE TABLE info (
			value TEXT PRIMARY KEY,
			label TEXT,
			user TEXT
		);`); err != nil {
			return fmt.Errorf("cannot create table info: %v", err)
		}

		if _, err = db.Exec(`
		CREATE TABLE files (
			key INTEGER PRIMARY KEY AUTOINCREMENT,
			original_name TEXT,
			name TEXT
		);`); err != nil {
			return fmt.Errorf("cannot create table file: %v", err)
		}
	}

	stmt1, err := db.Prepare("INSERT INTO sqlar (name, mode, mtime, sz, data) Values($1,$2,$3,$4,$5)")
	if err != nil {
		return fmt.Errorf("cannot prepare insert1: %v", err)
	}
	defer stmt1.Close()

	stmt2, err := db.Prepare("INSERT INTO info (value, label, user) Values($1,$2,$3)")
	if err != nil {
		return fmt.Errorf("cannot prepare insert2: %v", err)
	}
	defer stmt2.Close()

	if !append {
		h := time.Now()
		stmt2.Exec(h.Format(time.RFC3339), "created", user)
		stmt2.Exec(id.String(), "identifier", user)
	}
	h := time.Now()
	stmt2.Exec(h.Format(time.RFC3339), "modified", user)

	stmt3, err := db.Prepare("INSERT INTO files (key, original_name, name) Values($1,$2,$3)")
	if err != nil {
		return fmt.Errorf("cannot prepare insert2: %v", err)
	}
	defer stmt3.Close()

	sqlar := func(originalName string, name string, stream io.Reader) error {

		row := db.QueryRow("SELECT name FROM sqlar WHERE name =?", name)
		if err != nil {
			return fmt.Errorf("cannot check whether file already exists in archive: %v", err)
		}

		update := !(util.ReadRow(row) == "")
		if update {
			_, err = db.Exec("DELETE FROM sqlar WHERE name=?", name)
			if err != nil {
				return fmt.Errorf("cannot delete file from archive: %v", err)
			}
		}

		data, _ := ioutil.ReadAll(stream)
		mtime := time.Now().Unix()
		mode := 0777
		_, err := stmt1.Exec(name, mode, mtime, len(data), data)
		if err != nil {
			return fmt.Errorf("cannot exec stmt1: %v", err)
		}
		if !update {
			_, err = stmt3.Exec(nil, originalName, name)
			if err != nil {
				return fmt.Errorf("cannot exec stmt3: %v", err)
			}
		}
		return nil
	}

	imageIndex := 0
	for name, stream := range filestream {
		originalName := name
		ext := filepath.Ext(name)
		if ext != ".json" {
			imageIndex++
			name = util.ImageName(name, imageIndex)
		}
		err = sqlar(originalName, name, stream)
		if err != nil {
			return err
		}
	}

	return nil
}

// Given a SQLite archive and a table name show the contents of that table
func Inspect(sqlitefile string, table string) ([]interface{}, error) {

	result := make([]interface{}, 0)
	db, err := sql.Open("sqlite", sqlitefile)
	if err != nil {
		return result, fmt.Errorf("cannot open file: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM " + table)
	if table == "sqlar" {
		rows, err = db.Query("SELECT name, mode, mtime, sz FROM sqlar")
	}
	defer rows.Close()
	if err != nil {
		return result, fmt.Errorf("cannot open file: %v", err)
	}

	return util.ReadRows(rows)
}
