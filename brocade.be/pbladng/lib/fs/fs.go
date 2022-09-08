package fs

import (
	"os"
	"path/filepath"

	bfs "brocade.be/base/fs"
	pregistry "brocade.be/pbladng/lib/registry"
)

var Base = pregistry.Registry["base-dir"].(string)

func Store(fname string, data interface{}) error {
	name := FName(fname)
	dir := filepath.Dir(name)
	bfs.MkdirAll(dir, "process")
	return bfs.Store(name, data, "process")
}

func Fetch(fname string) ([]byte, error) {
	name := FName(fname)
	return os.ReadFile(name)
}

func CopyFileToDir(fname, dir string) error {
	cname := FName(fname)
	cdir := FName(dir)
	bfs.MkdirAll(cdir, "process")
	return bfs.CopyFile(cname, cdir, "process", false)
}

func FName(fname string) string {
	fname = filepath.FromSlash(fname)
	return filepath.Join(Base, fname)
}

func Exists(fname string) bool {
	name := FName(fname)
	return bfs.Exists(name)
}
