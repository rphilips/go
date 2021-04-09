package util

import (
	"encoding/json"
	"path"
	"strings"
	"time"

	qfnmatch "brocade.be/base/fnmatch"
	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
)

type Profile struct {
	Pattern string `json:"pattern"`
	Model   string `json:"model"`
}

func FileCreate(fname string) error {
	profiledir := path.Join(qregistry.Registry["qtechng-support-dir"], "profiles")
	profilefile := path.Join(profiledir, "profiles.json")

	blob, err := qfs.Fetch(profilefile)
	if err != nil {
		return err
	}
	profiles := make([]Profile, 0)
	err = json.Unmarshal(blob, &profiles)

	if err != nil {
		return err
	}

	basename := path.Base(fname)
	model := ""

	for _, pair := range profiles {
		pat := pair.Pattern
		mod := pair.Model
		if qfnmatch.Match(pat, basename) {
			model = mod
			break
		}
	}

	if model == "" {
		err = qfs.Store(fname, "", "process")
		return err
	}

	blob, err = qfs.Fetch(path.Join(qregistry.Registry["qtechng-support-dir"], "profiles", model))

	if err != nil {
		return err
	}

	filler := make(map[string]string)
	filler["basename"] = basename
	filler["ext"] = path.Ext(basename)
	filler["root"] = strings.TrimSuffix(basename, path.Ext(basename))
	filler["root1"] = filler["root"][1:]
	filler["user"] = qregistry.Registry["qtechng-user"]
	filler["time"] = time.Now().Format(time.RFC3339)

	sblob := string(blob)

	for key := range filler {
		value := filler[key]
		sblob = strings.ReplaceAll(sblob, "{"+key+"}", value)
		key = strings.ToUpper(key)
		value = strings.ToUpper(value)
		sblob = strings.ReplaceAll(sblob, "{"+key+"}", value)
	}

	err = qfs.Store(fname, sblob, "process")
	return err

}
