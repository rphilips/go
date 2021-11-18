package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	fs "brocade.be/base/fs"
	registry "brocade.be/base/registry"
	identifier "brocade.be/iiiftool/lib/identifier"

	_ "modernc.org/sqlite"
)

var osSep = registry.Registry["os-sep"]
var user = registry.Registry["user-default"]

// Given a IIIF identifier and some files
// store the files in the appropriate SQLite archive
func Store(id identifier.Identifier, files []string) error {
	sqlitefile := id.Location()

	append := fs.Exists(sqlitefile)

	for _, file := range files {
		if !fs.IsFile(file) {
			return errors.New("file is not valid:\n" + file)
		}
	}

	path := strings.Split(sqlitefile, osSep)
	dirname := strings.Join(path[0:(len(path)-1)], osSep)
	err := fs.Mkdir(dirname, "process")
	if err != nil {
		fmt.Println(err)
		return errors.New("cannot make dir")
	}

	db, err := sql.Open("sqlite", sqlitefile)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("cannot open file: %v", err)
	}
	defer db.Close()

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
			return err
		}

		if _, err = db.Exec(`
		CREATE TABLE info (
			value TEXT PRIMARY KEY,
			label TEXT,
			user TEXT
		);`); err != nil {
			return err
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

	sqlar := func(name string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		data, err := fs.Fetch(name)

		if err != nil {
			return fmt.Errorf("cannot get content of `%s`: %v", name, err)
		}
		mt, err := fs.GetMTime(name)
		if err != nil {
			return fmt.Errorf("cannot get mtime of `%s`: %v", name, err)
		}
		utime := time.Now().Format(time.RFC3339)
		sz, err := fs.GetSize(name)
		if err != nil {
			return fmt.Errorf("cannot get size of `%s`: %v", name, err)
		}
		mode, err := fs.GetPerm(name)
		if err != nil {
			return fmt.Errorf("cannot get access permissions of `%s`: %v", name, err)
		}
		mtime := mt.Unix()
		_, err = stmt1.Exec(name, uint32(mode), mtime, utime, sz, data)
		if err != nil {
			fmt.Println(err)
			return fmt.Errorf("cannot exec: %v", err)
		}

		return nil
	}

	for _, file := range files {
		info, err := os.Stat(file)
		sqlar(file, info, err)
	}

	return nil
}
