package source

import (
	"path/filepath"

	qserver "brocade.be/qtechng/lib/server"
	qutil "brocade.be/qtechng/lib/util"
)

// IsUnique returns true if only one or zero such a basname exists
func IsUnique(version *qserver.Release, name string) bool {
	base := filepath.Base(name)
	digest := qutil.Digest([]byte(base))
	fs := version.FS("/unique")
	dir := "/" + digest[:2] + "/" + digest[2:]
	exists, _ := fs.Exists(dir)
	if !exists {
		return true
	}
	fi, err := fs.Stat(dir)
	if err != nil {
		return true
	}
	if !fi.IsDir() {
		return true
	}
	d, err := fs.Open(dir)
	if err != nil {
		return true
	}
	defer d.Close()
	names, _ := d.Readdirnames(-1)
	if len(names) == 0 {
		return true
	}
	if len(names) > 1 {
		return false
	}
	found := filepath.Base(names[0])
	ndigest := qutil.Digest([]byte(name))
	return found == ndigest
}

// StoreUnique stores a reference to the basename
func UniqueStore(version *qserver.Release, qpath string) {
	version.UniqueStore(qpath)
}

// UnlinkUnique stores a reference to the basename
func UniqueUnlink(version *qserver.Release, qpath string) {
	version.UniqueUnlink(qpath)
}
