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
	Digest     string              `json:"digest"`
	Identifier string              `json:"identifier"`
	Iiifsys    string              `json:"iiifsys"`
	Images     []map[string]string `json:"images"`
	Imgloi     string              `json:"imgloi"`
	Index      []string            `json:"index"`
	Manifest   interface{}
}

// Harvest IIIF metadata from MUMPS
func Meta(
	id identifier.Identifier,
	loiType string,
	urlty string,
	imgty string,
	access string,
	mime string,
	iiifsys string) (MResponse, error) {

	payload := make(map[string]string)
	payload["loi"] = id.String()
	payload["iiifsys"] = iiifsys
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
