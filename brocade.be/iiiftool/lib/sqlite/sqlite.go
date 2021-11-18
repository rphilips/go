package sqlite

import (
	"fmt"

	identifier "brocade.be/iiiftool/lib/identifier"
)

// Given a IIIF identifier and some files
// store the files in the appropriate SQLite archive
func Store(id identifier.Identifier, files []string) error {
	path := id.Location()
	fmt.Println(path, files)
	return nil
}
