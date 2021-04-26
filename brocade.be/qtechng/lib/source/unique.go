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
	if found == ndigest {
		return true
	}
	return false
}

// StoreUnique stores a reference to the basename
func StoreUnique(version *qserver.Release, name string) {
	base := filepath.Base(name)
	digest := qutil.Digest([]byte(base))
	ndigest := qutil.Digest([]byte(name))
	fs := version.FS("/unique")
	fname := "/" + digest[:2] + "/" + digest[2:] + "/" + ndigest
	m := map[string]string{"path": name}
	fs.Store(fname, m, "qtech")
}

// UnlinkUnique stores a reference to the basename
func UnlinkUnique(version *qserver.Release, name string) {
	base := filepath.Base(name)
	digest := qutil.Digest([]byte(base))
	ndigest := qutil.Digest([]byte(name))
	fs := version.FS("/unique")
	fname := "/" + digest[:2] + "/" + digest[2:] + "/" + ndigest
	fs.Waste(fname)
}
