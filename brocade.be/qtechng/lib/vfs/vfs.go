package vfs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"

	qfnmatch "brocade.be/base/fnmatch"
	qutil "brocade.be/qtechng/lib/util"
)

// QFs is afero filesystem for use in Qtech: there is extra functionality
type QFs struct {
	afero.Afero
	ReadOnly bool
}

// RealPath for QFs
func (fs QFs) RealPath(fname string) (path string, err error) {
	bp := fs.Fs.(*afero.BasePathFs)
	path, err = bp.RealPath(fname)
	if err == nil {
		path, err = filepath.Abs(path)
	}
	return
}

// RemoveAll is a stricter RemoveAll
func (fs QFs) RemoveAll(fname string) error {
	if fs.ReadOnly {
		return errors.New("Filesystem is readonly")
	}
	rp, err := fs.RealPath(fname)
	dirname := filepath.Dir(rp)
	parent := filepath.Dir(dirname)
	if parent == dirname || parent == "/" || parent == "." {
		return errors.New("Cannot delete a root")
	}
	err = os.RemoveAll(rp)
	return err
}

// Waste is a specialised delete: if the file is deleted
// and there are no other files in the directory then
// the directory is removed as well.
func (fs QFs) Waste(fname string) (change bool, err error) {
	if fs.ReadOnly {
		return false, errors.New("Filesystem is readonly")
	}
	dir := filepath.Dir(fname)

	e := fs.Remove(fname)

	if e != nil {
		if ok, _ := fs.Exists(fname); ok {
			return false, errors.New("Cannot delete " + fname)
		}
	}
	if ok, _ := fs.Exists(dir); !ok {
		return true, nil
	}
	for {
		if (dir == "/") || (dir == ".") || (dir == "") {
			break
		}
		files, erro := fs.ReadDir(dir)
		if len(files) != 0 || erro != nil {
			break
		}
		parent := filepath.Dir(dir)
		e := fs.Remove(dir)
		if e != nil {
			break
		}
		if parent == dir {
			break
		}
		dir = parent
	}
	return true, nil
}

// JSONLoad loads a json file in an object
func (fs QFs) JSONLoad(jname string, result interface{}) (err error) {
	content, err := fs.ReadFile(jname)
	if err != nil {
		return err
	}
	err = json.Unmarshal(content, result.(*interface{}))
	return err
}

// Store stores data in a qtech file
//       creates underlying directories as needed
//       only writes if the file does not exists
//            or has different content
func (fs QFs) Store(fname string, data interface{}, digest string) (changed bool, before []byte, after []byte, err error) {
	if fs.ReadOnly {
		return false, nil, nil, errors.New("Filesystem is readonly")
	}

	before, e := fs.ReadFile(fname)

	if e == nil && digest != "" {
		dg := qutil.Digest(before)
		if dg != digest {
			return false, before, nil, errors.New("Digest does not match")
		}
	}

	// reduce data to []byte

	after, err = qutil.MakeBytes(data)

	if err != nil {
		return false, before, nil, err
	}

	if bytes.Compare(before, after) == 0 && !os.IsNotExist(e) {
		return false, before, after, nil
	}

	// atomic write

	dir, name := filepath.Split(fname)
	err = fs.MkdirAll(dir, 0o770)
	if err != nil {
		return false, before, after, fmt.Errorf("cannot create directory: %v", err)
	}
	f, err := afero.TempFile(fs, dir, name)
	if err != nil {
		return false, before, after, fmt.Errorf("cannot create temp file: %v", err)
	}
	name = f.Name()
	defer func() {
		f.Close()
		if err != nil {
			// Don't leave the temp file lying around on error.
			_ = fs.Remove(name) // yes, ignore the error, not much we can do about it.
		}
	}()
	// ensure we always close f. Note that this does not conflict with the
	// close below, as close is idempotent.
	if _, err := io.Copy(f, bytes.NewReader(after)); err != nil {
		return false, before, after, fmt.Errorf("cannot write data to tempfile %q: %v", name, err)
	}
	if err := f.Close(); err != nil {
		return false, before, after, fmt.Errorf("can't close tempfile %q: %v", name, err)
	}

	if err := fs.Chmod(name, 0o660); err != nil {
		return false, before, after, fmt.Errorf("can't set filemode on tempfile %q: %v", name, err)
	}
	if err := fs.Rename(name, fname); err != nil {
		return false, before, after, fmt.Errorf("cannot replace %q with tempfile %q: %v", fname, name, err)
	}
	return true, before, after, nil
}

// Dir returns a slice with files/dirs in the directory
func (fs QFs) Dir(dir string, onlyfiles bool, onlydirs bool) (names []string) {
	fi, err := fs.Stat(dir)
	if err != nil {
		return
	}
	if !fi.IsDir() {
		return
	}

	d, err := fs.Open(dir)
	if err != nil {
		return
	}
	defer d.Close()
	names = make([]string, 0)
	fnames, _ := d.Readdirnames(-1)

	for _, n := range fnames {
		k := filepath.Join(dir, n)
		if onlyfiles || onlydirs {
			fi, err := fs.Stat(k)
			if err != nil {
				return
			}
			isdir := fi.IsDir()
			if (isdir && onlyfiles) || (!isdir && onlydirs) {
				continue
			}
		}
		names = append(names, k)
	}
	return
}

// Glob looks for files matchin a given pattern
func (fs QFs) Glob(dir string, patterns []string, matchonlybasename bool) (matches []string) {
	if len(patterns) == 0 {
		patterns = []string{"*"}
	}
	if len(patterns) == 1 {
		pattern := patterns[0]
		if !strings.ContainsAny(pattern, "[]*?") && !matchonlybasename {
			x := filepath.Join(dir, pattern)
			exists, _ := fs.Exists(x)
			if exists {
				return []string{x}
			}
			return
		}
		matches = make([]string, 0)
		fi, err := fs.Stat(dir)
		if err != nil {
			return
		}
		if !fi.IsDir() {
			return
		}
		d, err := fs.Open(dir)
		if err != nil {
			return
		}
		defer d.Close()
		names, _ := d.Readdirnames(-1)

		for _, n := range names {
			k := filepath.Join(dir, n)
			fi, err := fs.Stat(k)
			if err != nil {
				return
			}
			if fi.IsDir() {
				matches = append(matches, fs.Glob(k, patterns, matchonlybasename)...)
				continue
			}
			s := k
			if matchonlybasename {
				s = n
			}
			if qfnmatch.Match(pattern, s) {
				matches = append(matches, k)
			}
		}
		return matches
	}
	pats := []string{"*"}
	all := fs.Glob(dir, pats, matchonlybasename)
	if len(patterns) == 0 {
		return all
	}
	matches = make([]string, 0)
	for _, fname := range all {
		name := fname
		if matchonlybasename {
			name = filepath.Base(fname)
		}
		for _, pattern := range patterns {
			if qfnmatch.Match(pattern, name) {
				matches = append(matches, fname)
				break
			}
		}
	}
	return matches
}
