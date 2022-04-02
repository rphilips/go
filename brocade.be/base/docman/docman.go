package docman

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
)

type DocmanID string

// DB retrieves the Docman database associated with a DocmanID
func (id DocmanID) DB() string {
	ids := string(id)
	if strings.HasPrefix(ids, "/docman/") {
		ids = ids[7:]
	}
	dlm := "."
	sub := 0
	if strings.HasPrefix(ids, "/") {
		dlm = "/"
		sub = 1
	}
	parts := strings.SplitN(ids, dlm, -1)
	if len(parts) < sub {
		return ""
	}
	return parts[sub]
}

// Location retrieves the absolute filepath to the file refered to by the DocmanID.
// If the file does not exists, the empty string is returned
// This function honours the shadow property
func (id DocmanID) Location() string {
	ids := string(id)
	if strings.HasPrefix(ids, "/docman/") {
		ids = ids[7:]
	}
	dlm := "."
	old := true
	basename := ids
	db := ""
	if strings.HasPrefix(ids, "/") {
		dlm = "/"
		old = false
	}
	parts := strings.SplitN(ids, dlm, -1)

	dir1 := ""
	dir2 := ""

	if !old {
		if len(parts) < 2 {
			return ""
		}
		db = parts[1]
		basename = parts[len(parts)-1]
		if len(parts) == 3 {
			dirhex := fmt.Sprintf("%x", md5.Sum([]byte(basename)))[:3]
			dir1 = "y" + dirhex[0:1]
			dir2 = dirhex[1:3]
		} else {
			dirhex := parts[2]
			basename = dirhex[3:] + basename
			dir1 = "x" + dirhex[0:1]
			dir2 = dirhex[1:3]
		}

	} else {
		db = parts[0]
		dir1 = parts[1]
		dir2 = parts[2]
	}
	path1 := filepath.Join(qregistry.Registry["docman-db"], db, dir1, dir2, basename)
	shadow := filepath.Join(qregistry.Registry["docman-db"], db, "__shadow__")
	if !qfs.IsFile(shadow) {
		return getFile(path1)
	}
	bshadowdb, err := qfs.Fetch(shadow)
	if err != nil || len(bshadowdb) == 0 {
		return getFile(path1)
	}
	shadowdb := string(bshadowdb)
	path2 := filepath.Join(qregistry.Registry["docman-db"], shadowdb, dir1, dir2, basename)
	place2 := getFile(path2)
	if place2 != "" {
		return place2
	}
	return getFile(path1)
}

func getFile(file string) (place string) {

	if qfs.IsFile(file) {
		return file
	}
	pattern := file + ".*"
	matches, _ := filepath.Glob(pattern)
	if len(matches) == 0 {
		return ""
	}
	return matches[0]
}

func (id DocmanID) Reader() (io.ReadCloser, error) {
	ids := string(id)
	if strings.HasPrefix(ids, "/docman/") {
		ids = ids[7:]
	}
	location := id.Location()
	if location != "" {
		if qfs.IsFile(location) {
			file, err := os.Open(location)
			if err == nil {
				return file, err
			}
			if qregistry.Registry["docman-secondary-url"] == "" {
				return nil, err
			}
		}
	}
	baseurl := qregistry.Registry["docman-secondary-url"]

	if baseurl == "" {
		return nil, fmt.Errorf("cannot find docman `%s`", ids)
	}
	baseurl = strings.TrimRight(baseurl, "/")
	if !strings.Contains(baseurl, "://") {
		baseurl = "https://" + baseurl
	}
	url := baseurl + "/docman" + ids

	response, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	return response.Body, nil
}
