package iiif

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"brocade.be/base/mumps"
	"brocade.be/base/registry"
	"brocade.be/iiiftool/lib/identifier"
	"brocade.be/iiiftool/lib/util"
)

var iifBaseDir = registry.Registry["iiif-base-dir"]

type MResponse struct {
	Digest     string
	Identifier string
	Iiifsys    string
	Images     []string
	Imgloi     string
}

// Harvest IIIF metadata from MUMPS
func Meta(
	id identifier.Identifier,
	loiType string,
	urlty string,
	imgty string,
	access string,
	mime string) (MResponse, error) {

	payload := make(map[string]string)
	payload["loi"] = id.String()
	switch loiType {
	case "c", "o":
		payload["urlty"] = urlty
	case "tg":
		payload["imgty"] = imgty
	}
	if access != "" {
		payload["access"] = access
	}
	if mime != "" {
		payload["mime"] = mime
	}

	var mResponse MResponse
	oreader, _, err := mumps.Reader("d %Action^iiisori(.RApayload)", payload)
	if err != nil {
		return mResponse, fmt.Errorf("mumps error:\n%s", err)
	}
	out, err := ioutil.ReadAll(oreader)
	if err != nil {
		return mResponse, fmt.Errorf("mumps error:\n%s", err)
	}
	json.Unmarshal(out, &mResponse)
	return mResponse, nil
}

// Given an IIIF, formulate its appropriate location in the filesystem
// -- regardless of whether this location is an existing filepath or not.
func Digest2Location(digest string) string {
	digest = util.StrReverse(digest)
	folder := digest[0:2]
	subfolder := digest[2:4]
	basename := digest[0:12] + ".sqlite"
	location := filepath.Join(iifBaseDir, folder[0:2], subfolder, basename)
	return location
}

// Encode using SHA1
// func (id Identifier) Encode() string {
// 	hash := sha1.New()
// 	hash.Write([]byte(id.String()))
// 	encodedString := hex.EncodeToString(hash.Sum(nil))
// 	return encodedString
// }
