package identifier

import (
	"encoding/base32"
	"log"
	"path/filepath"
	"strings"

	util "brocade.be/iiiftool/lib/util"

	registry "brocade.be/base/registry"
)

type Identifier string

// var iifBaseDir = registry.Registry("iiif-base-dir")
var iifBaseDir = registry.Registry["scratch-dir"]

var osSep = registry.Registry["os-sep"]

func (id Identifier) String() string {
	return string(id)
}

// Given an identifier, formulate its appropriate location in the filesystem
// -- regardless of whether this location is an existing filepath or not.
func (id Identifier) Location() string {
	idEnc := id.Encode()
	folder := strings.ReplaceAll(idEnc, "=", "")
	folder = strings.ToLower(folder)
	folder = util.StrReverse(folder)
	basename := strings.ToLower(idEnc) + ".sqlite"
	basename = strings.ReplaceAll(basename, "=", "8")
	location := filepath.Join(iifBaseDir, folder[0:2], folder, basename)
	return location
}

// Given an location, formulate its corresponding identifier
// -- regardless of whether this location is an existing filepath or not.
func ReverseLocation(location string) string {
	parts := strings.Split(location, osSep)
	id := parts[(len(parts) - 1)]
	id = strings.ReplaceAll(id, ".sqlite", "")
	id = strings.ToUpper(id)
	id = strings.ReplaceAll(id, "8", "=")
	return Identifier(id).Decode()
}

// Encode in base32
func (id Identifier) Encode() string {
	return base32.StdEncoding.EncodeToString([]byte(id))
}

// Decode from base32
func (id Identifier) Decode() string {
	dec, err := base32.StdEncoding.DecodeString(id.String())
	if err != nil {
		log.Fatal("error decoding string:\n", err)
	}
	return string(dec)
}
