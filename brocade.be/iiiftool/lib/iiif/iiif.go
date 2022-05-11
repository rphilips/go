package iiif

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"

	"brocade.be/base/fs"
	"brocade.be/base/mumps"
	"brocade.be/base/registry"
	"brocade.be/iiiftool/lib/util"
)

var iifBaseDir = registry.Registry["iiif-base-dir"]

const validator = "https://presentation-validator.iiif.io/validate?"

type validateResponse struct {
	Url      string      `json:"url"`
	Okay     int         `json:"okay"`
	Error    string      `json:"error"`
	Warnings interface{} `json:"warnings"`
}
type IIIFmeta struct {
	Images   []map[string]string `json:"images"`
	Imgloi   string              `json:"imgloi"`
	Indexes  []string            `json:"index"`
	Info     map[string]string   `json:"info"`
	Iiifsys  string              `json:"iiifsys"`
	Manifest interface{}         `json:"manifest"`
}

// Harvest IIIF metadata from MUMPS
func Meta(
	loi string,
	iiifsys string) (IIIFmeta, error) {

	payload := make(map[string]string)
	payload["loi"] = loi
	payload["iiifsys"] = iiifsys

	var iiifMeta IIIFmeta
	oreader, _, err := mumps.Reader("d %Action^iiisori(.RApayload)", payload)
	if err != nil {
		return iiifMeta, fmt.Errorf("mumps reader error:\n%s", err)
	}
	out, err := ioutil.ReadAll(oreader)
	if err != nil {
		return iiifMeta, fmt.Errorf("cannot read MUMPS response:\n%s", err)
	}
	err = json.Unmarshal(out, &iiifMeta)
	if err != nil {
		return iiifMeta, fmt.Errorf("json error:\n%s", err)
	}
	return iiifMeta, nil
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

// Function that validates a IIIF manifest
func Validate(manifestUrl string, version string) (validateResponse, error) {

	var result validateResponse

	URL := validator + "version=" + version + "&url=" + manifestUrl

	response, err := http.Get(URL)
	if err != nil {
		return result, fmt.Errorf("error validating:%s", err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return result, fmt.Errorf("error reading response:%s", err)
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return result, fmt.Errorf("json error:%s", err)
	}

	return result, nil

}
