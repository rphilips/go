package sqlite

import (
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	fs "brocade.be/base/fs"
	registry "brocade.be/base/registry"
	identifier "brocade.be/iiiftool/lib/identifier"
	"brocade.be/iiiftool/lib/util"

	_ "modernc.org/sqlite"
)

var osSep = registry.Registry["os-sep"]
var user = registry.Registry["user-default"]

// Given a IIIF identifier and some io.Readers
// store the contents in the appropriate SQLite archive
func Store(id identifier.Identifier, filestream map[string]io.Reader, cwd string) error {
	sqlitefile := id.Location()

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

	sqlar := func(name string, stream io.Reader) error {
		row := db.QueryRow("SELECT name FROM sqlar WHERE name =?", name)
		if err != nil {
			return fmt.Errorf("cannot check whether file already exists in archive: %v", err)
		}

		update := util.ReadRow(row)
		if update != "" {
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
		return nil
	}

	for name, stream := range filestream {
		err = sqlar(name, stream)
		if err != nil {
			return err
		}
	}

	return nil
}
