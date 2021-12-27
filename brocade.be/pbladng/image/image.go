package image

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// ImageMap creates a map with the identifier (lowercase) mapped to the relpath to dir
func ImageMap(dir string) map[string]string {
	m := make(map[string]string)
	fn := func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		name := info.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".jpg" && ext != ".jpeg" {
			return nil
		}
		k := strings.IndexAny(name, "-_.")
		index := name[:k]
		_, ok := m[index]
		if ok {
			return nil
		}
		m[index] = path
		return nil
	}
	filepath.WalkDir(dir, fn)
	return m
}
