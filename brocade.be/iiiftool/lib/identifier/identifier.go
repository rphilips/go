package identifier

import (
	"encoding/base64"
	"log"
	"path/filepath"
	"strings"

	util "brocade.be/iiiftool/lib/util"

	registry "brocade.be/base/registry"
)

type Identifier string

func (id Identifier) String() string {
	return string(id)
}

// Given an identifier, formulate its appropriate location in the filesystem
// -- regardless of whether this location is an existing filepath or not.
func (id Identifier) Location() string {
	baseDir := registry.Registry["scratch-dir"]
	// baseDir := registry.Registry("iiif-base-dir")
	idEnc := id.Encode()
	idEnc = strings.ReplaceAll(idEnc, "=", "")
	idEnc = strings.ToLower(idEnc)
	idEnc = util.StrReverse(idEnc)
	location := filepath.Join(baseDir, idEnc[0:2], idEnc, idEnc+".sqlite")
	return location
}

// Encode in base64url
func (id Identifier) Encode() string {
	return base64.URLEncoding.EncodeToString([]byte(id))
}

// Decode from base64url
func (id Identifier) Decode() string {
	dec, err := base64.URLEncoding.DecodeString(id.String())
	if err != nil {
		log.Fatal("error decoding string:\n", err)
	}
	return string(dec)
}

// basedir ansible, groupid db, backup -> Luc
// basedir/2char of hex van SH1 enkel lowercase/omgevormde identifier/omgevormde identifier.sqlite3

// metadata steken in identifier
// omvorming naar jpk2 ook in iiiftool
