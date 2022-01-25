package iiif

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"brocade.be/base/fs"
	"brocade.be/base/mumps"
	"brocade.be/base/registry"
	"brocade.be/iiiftool/lib/util"
)

var iifBaseDir = registry.Registry["iiif-base-dir"]

type MResponse struct {
	Digest     string              `json:"digest"`
	Identifier string              `json:"identifier"`
	Iiifsys    string              `json:"iiifsys"`
	Images     []map[string]string `json:"images"`
	Imgloi     string              `json:"imgloi"`
	Indexes    []string            `json:"index"`
	Manifest   interface{}         `json:"manifest"`
}

// Harvest IIIF metadata from MUMPS
func Meta(
	id string,
	loiType string,
	urlty string,
	imgty string,
	access string,
	mime string,
	iiifsys string) (MResponse, error) {

	payload := make(map[string]string)
	payload["loi"] = id
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
		return mResponse, fmt.Errorf("mumps reader error:\n%s", err)
	}
	out, err := ioutil.ReadAll(oreader)
	if err != nil {
		return mResponse, fmt.Errorf("cannot read MUMPS response:\n%s", err)
	}
	err = json.Unmarshal(out, &mResponse)
	if err != nil {
		return mResponse, fmt.Errorf("json error:\n%s", err)
	}
	return mResponse, nil
}

// Given a IIIF digest, formulate its appropriate location in the filesystem
// -- regardless of whether the location is an existing filepath or not.
func Digest2Location(digest string) string {
	digest = util.StrReverse(digest)
	folder := digest[0:2]
	subfolder := digest[2:4]
	location := filepath.Join(iifBaseDir, folder[0:2], subfolder, digest, "db.sqlite")
	return location
}

// Delete a IIIF archive
func DigestDelete(digest string) error {
	location := Digest2Location(digest)
	if !fs.Exists(location) {
		return fmt.Errorf("invalid location:\n%s", location)
	}

	directory := filepath.Dir(location)
	err := fs.Rmpath(directory)
	if err != nil {
		return fmt.Errorf("error deleting directory:\n%s", err)
	}

	return nil
}
