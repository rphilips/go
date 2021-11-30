package index

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"brocade.be/base/registry"
)

var iifBaseDir = registry.Registry["iiif-base-dir"]

// Rebuild IIIF index
func Rebuild() error {

	fn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking over file: %v", err)
		}
		ext := filepath.Ext(path)
		if ext != ".sqlite" {
			return nil
		}

		db, err := sql.Open("sqlite", path)
		if err != nil {
			return fmt.Errorf("cannot open file: %v", err)
		}
		defer db.Close()

		return nil
	}

	filepath.Walk(iifBaseDir, fn)

	// select sqlite "identifier" and "index" value from table meta
	// write identifier and index in new table

	return nil
}
