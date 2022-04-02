package util

import (
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	qfnmatch "brocade.be/base/fnmatch"
	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
)

type Profile struct {
	Comment string `json:"comment"`
	Pattern string `json:"pattern"`
	Model   string `json:"model"`
	Hint    string `json:"hint"`
}

func FileCreate(fname string, hint string) error {
	profiledir := filepath.Join(qregistry.Registry["qtechng-support-dir"], "profiles")
	profilefile := filepath.Join(profiledir, "profiles.json")

	blob, err := qfs.Fetch(profilefile)
	if err != nil {
		blob = []byte("[]")
	}
	profiles := make([]Profile, 0)
	err = json.Unmarshal(blob, &profiles)

	if err != nil {
		return err
	}

	basename := filepath.Base(fname)
	model := ""

	for _, tripel := range profiles {
		pat := tripel.Pattern
		mod := tripel.Model
		h := tripel.Hint
		if hint != "" && h != hint {
			continue
		}
		if qfnmatch.Match(pat, basename) {
			model = mod
			break
		}
	}
	fmt.Println("model:", model)
	blob = []byte("")
	if model != "" {
		blob, _ = qfs.Fetch(filepath.Join(qregistry.Registry["qtechng-support-dir"], "profiles", model))
	}

	filler := make(map[string]string)
	filler["basename"] = basename
	filler["ext"] = strings.TrimPrefix(path.Ext(basename), ".")
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
	err = qfs.Store(fname, sblob, "qtech")
	return err
}
