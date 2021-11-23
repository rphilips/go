package iiif

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"brocade.be/base/mumps"
	identifier "brocade.be/iiiftool/lib/identifier"
)

type mResponse struct {
	Identifier string
	Images     []string
}

// Harvest IIIF metadata from MUMPS
func Meta(
	id identifier.Identifier,
	loiType string,
	urlty string,
	imgty string,
	access string,
	mime string) (mResponse, error) {

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

	var result mResponse
	oreader, _, err := mumps.Reader("d %Action^iiisori(.RApayload)", payload)
	if err != nil {
		return result, fmt.Errorf("mumps error:\n%s", err)
	}
	out, err := ioutil.ReadAll(oreader)
	if err != nil {
		return result, fmt.Errorf("mumps error:\n%s", err)
	}
	json.Unmarshal(out, &result)
	return result, nil
}
