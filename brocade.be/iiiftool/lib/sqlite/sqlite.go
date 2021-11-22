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

	_ "modernc.org/sqlite"
)

var osSep = registry.Registry["os-sep"]
var user = registry.Registry["user-default"]

func readRow(row *sql.Row) string {
	data := ""
	err := row.Scan(&data)
	if err != nil {
		return data
	}
	return data
}

// Given a IIIF identifier and some io.Readers
// store the contents in the appropriate SQLite archive
func Store(id identifier.Identifier, files map[string]io.Reader, cwd string) error {
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
			utime TEXT,
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

	stmt1, err := db.Prepare("INSERT INTO sqlar (name, mode, mtime, utime, sz, data) Values($1,$2,$3,$4,$5,$6)")
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

	sqlar := func(name string, filestream io.Reader) error {
		row := db.QueryRow("SELECT name FROM sqlar WHERE name =?", name)
		if err != nil {
			return fmt.Errorf("cannot check whether file already exists in archive: %v", err)
		}

		update := readRow(row)
		if update != "" {
			_, err = db.Exec("DELETE FROM sqlar WHERE name=?", name)
			if err != nil {
				return fmt.Errorf("cannot delete file from archive: %v", err)
			}
		}

		data, _ := ioutil.ReadAll(filestream)

		// mt, err := fs.GetMTime(file)
		// if err != nil {
		// 	return fmt.Errorf("cannot get mtime of `%s`: %v", file, err)
		// }
		// utime := time.Now().Format(time.RFC3339)
		// sz, err := fs.GetSize(file)
		// if err != nil {
		// 	return fmt.Errorf("cannot get size of `%s`: %v", file, err)
		// }
		// mode, err := fs.GetPerm(file)
		// if err != nil {
		// 	return fmt.Errorf("cannot get access permissions of `%s`: %v", file, err)
		// }
		// mtime := mt.Unix()
		_, err = stmt1.Exec(name, 1, 2, "test", 1, data)
		if err != nil {
			return fmt.Errorf("cannot exec stmt1: %v", err)
		}

		return nil
	}

	for name, filestream := range files {
		sqlar(name, filestream)
	}

	return nil
}
