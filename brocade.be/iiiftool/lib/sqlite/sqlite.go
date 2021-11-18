package sqlite

import (
	"errors"
	"fmt"

	fs "brocade.be/base/fs"
	identifier "brocade.be/iiiftool/lib/identifier"
)

// Given a IIIF identifier and some files
// store the files in the appropriate SQLite archive
func Store(id identifier.Identifier, files []string) error {
	path := id.Location()

	if fs.Exists(path) {
		return errors.New("location already has data")
		// to do: provide append mode?
	}
	for _, file := range files {
		if !fs.IsFile(file) {
			return errors.New("file is not valid:\n" + file)
		}
	}

	fmt.Println("okay")

	// store files at location

	return nil
}
