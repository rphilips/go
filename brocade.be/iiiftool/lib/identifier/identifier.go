package identifier

import (
	"crypto/sha1"
	"encoding/hex"
	"path/filepath"

	"brocade.be/base/registry"
)

type Identifier string

var iifBaseDir = registry.Registry["iiif-base-dir"]

func (id Identifier) String() string {
	return string(id)
}

// Given an identifier, formulate its appropriate location in the filesystem
// -- regardless of whether this location is an existing filepath or not.
func (id Identifier) Location(digest string) string {
	idEnc := ""
	if digest != "" {
		idEnc = digest
	} else {
		idEnc = id.Encode()
	}
	folder := idEnc[0:2]
	subfolder := idEnc[2:4]
	basename := idEnc[0:12] + ".sqlite"
	location := filepath.Join(iifBaseDir, folder[0:2], subfolder, basename)
	return location
}

// Encode using SHA1
func (id Identifier) Encode() string {
	hash := sha1.New()
	hash.Write([]byte(id.String()))
	encodedString := hex.EncodeToString(hash.Sum(nil))
	return encodedString
}
