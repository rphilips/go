package fs

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	qpattern "brocade.be/base/pattern"
	qregistry "brocade.be/base/registry"

	fatomic "github.com/natefinch/atomic"
)

var (
	// ErrNoHome error "No HOME found"
	ErrNoHome = errors.New("No HOME found")
	// ErrNotPathMode indicates not a valid pathmode
	ErrNotPathMode = errors.New("Not a filemode")
)

// AbsPath creates absolute path for a filename
func AbsPath(pth string) (abspath string, err error) {

	if !strings.HasPrefix(pth, "~") {
		abspath, err = pth, nil
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			home = "~"
		}
		abspath = home + pth[1:]
	}
	abspath = os.ExpandEnv(abspath)
	abspath = filepath.FromSlash(abspath)
	abspath, err = filepath.Abs(abspath)
	if err != nil {
		abspath, _ = filepath.EvalSymlinks(abspath)
	}
	return abspath, err
}

// PathURI creates a file URI from a filename
func PathURI(pth string) (uri string, err error) {
	abspath, err := AbsPath(pth)
	if err != nil {
		return
	}
	uri = "file://"
	volume := filepath.VolumeName(abspath)
	if strings.HasSuffix(volume, ":") {
		volume = strings.ToUpper(volume)
		abspath = abspath[len(volume):]
	} else {
		if volume != "" {
			abspath = abspath[2:]
			volume = ""
		}
	}
	abspath = filepath.ToSlash(abspath)
	parts := strings.SplitN(abspath, "/", -1)
	iparts := []string{}
	for _, part := range parts {
		iparts = append(iparts, url.PathEscape(part))
	}
	uri += strings.Join(iparts, "/")

	return
}

type Property struct {
	UID  *user.User
	GID  *user.Group
	PERM os.FileMode
}

// Properties returns:
//    - uid (*user.User)
//    - gid / access for a pathmode according to the registry
func Properties(pathmode string) (prop Property, err error) {
	mode := pathmode
	switch {
	case strings.HasSuffix(pathmode, "dir"):
		mode = pathmode[:len(pathmode)-3]
	case strings.HasSuffix(pathmode, "file"):
		mode = pathmode[:len(pathmode)-4]
	}
	key := "fs-owner-" + mode
	suid := ""
	sgid := ""
	var perm os.FileMode
	value, ok := qregistry.Registry[key]
	if ok && strings.ContainsRune(value, ':') {
		parts := strings.Split(value, ":")
		suid = parts[0]
		sgid = parts[1]
	} else {
		err = ErrNotPathMode
		return
	}
	switch pathmode {
	case "webdir":
		perm = 0755
	case "webfile":
		perm = 0644
	case "webdavdir":
		perm = 0755
	case "webdavfile":
		perm = 0644
	case "scriptdir":
		perm = 0755
	case "scriptfile":
		perm = 0755
	case "processdir":
		perm = 0770
	case "daemonfile":
		perm = 0755
	case "daemondir":
		perm = 0755
	case "processfile":
		perm = 0770
	case "tempdir":
		perm = 0755
	case "tempfile":
		perm = 0664
	case "qtechdir":
		perm = 0770
	case "qtechfile":
		perm = 0660
	case "nakeddir":
		perm = 0777
	case "nakedfile":
		perm = 0776
	default:
		err = ErrNotPathMode
		return
	}
	uid, err := user.Lookup(suid)
	gid, err := user.LookupGroup(sgid)

	return Property{uid, gid, perm}, err
}

// SetPathMode assigns the ownership and access modes to a path
func SetPathmode(pth string, pathmode string) (err error) {
	if pathmode == "" {
		return ErrNotPathMode
	}
	p, err := Properties(pathmode)
	if err == nil && runtime.GOOS != "windows" {
		gi, _ := strconv.Atoi(p.GID.Gid)
		ui, _ := strconv.Atoi(p.UID.Uid)
		_ = os.Chown(pth, ui, gi)
		_ = os.Chmod(pth, p.PERM.Perm())
	}
	return err
}

// Store writes a file contents atomically
func Store(fname string, data interface{}, pathmode string) (err error) {

	var r io.Reader
	switch v := data.(type) {
	case []byte:
		r = bytes.NewReader(v)
	case string:
		r = strings.NewReader(v)
	case *string:
		r = strings.NewReader(*v)
	case bytes.Buffer:
		r = &v
	case *bytes.Buffer:
		r = v
	case io.Reader:
		r = v
	default:
		b, err := json.Marshal(data)
		if err != nil {
			return err
		}
		r = bytes.NewReader(b)
	}
	err = fatomic.WriteFile(fname, r)
	if err != nil {
		return
	}
	if pathmode == "" {
		return nil
	}
	if !strings.HasSuffix(pathmode, "file") {
		pathmode = pathmode + "file"
	}
	return SetPathmode(fname, pathmode)
}

// Fetch returns the contents of a file as a slice of bytes
var Fetch = os.ReadFile

// Mkdir makes a directory and sets the access
func Mkdir(dirname string, pathmode string) (err error) {
	if !strings.HasSuffix(pathmode, "dir") {
		pathmode = pathmode + "dir"
	}
	perm, err := Properties(pathmode)
	if err != nil {
		return
	}
	prm := perm.PERM
	err = os.Mkdir(dirname, prm)
	if err == nil {
		return SetPathmode(dirname, pathmode)
	}
	if os.IsExist(err) {
		return nil
	}
	parent := path.Dir(dirname)
	if parent == "" || dirname == parent {
		return err
	}
	Mkdir(parent, pathmode)
	os.Mkdir(dirname, prm)
	return SetPathmode(dirname, pathmode)
}

// Rmpath removes a file or a dirctory tree except root
func Rmpath(dirname string) (err error) {
	dirname, err = AbsPath(dirname)
	if err == nil {
		parent := path.Dir(dirname)
		if parent == dirname {
			err = errors.New("Cannot delete a root")
		} else {
			err = os.RemoveAll(dirname)
		}
	}
	return err
}

// EmptyDir Empties a directory
func EmptyDir(dir string) (err error) {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return
}

// Exists checks if a file exists
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return !os.IsNotExist(err)
}

// IsDir checks if a path is an existing directory
func IsDir(path string) bool {
	fi, err := os.Stat(path)
	if err == nil {
		return fi.Mode().IsDir()
	}
	return false
}

// IsFile checks if a path is an existing regular file
func IsFile(path string) bool {
	fi, err := os.Stat(path)
	if err == nil {
		return fi.Mode().IsRegular()
	}
	return false
}

// GetSize return the size of a file
func GetSize(path string) (size int64, err error) {
	fi, err := os.Stat(path)
	if err != nil {
		return
	}
	size = fi.Size()
	return
}

// GetMTime returns the last modification time of a path
func GetMTime(path string) (mtime time.Time, err error) {
	fi, err := os.Stat(path)
	if err != nil {
		return
	}
	mtime = fi.ModTime()
	return
}

// SetMTimes sets the last modification time
func SetMTime(path string, mtime time.Time) (err error) {
	ztime := time.Time{}
	if mtime == ztime {
		mtime = time.Now()
	}
	err = os.Chtimes(path, mtime, mtime)
	return
}

// GetSHA1 return SHA-1 (hex) of file contents (or "" if error)
func GetSHA1(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	//Open a new SHA1 hash interface to write to
	h := sha1.New()

	if _, err := io.Copy(h, file); err != nil {
		return ""
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// IsDirEmpty checks if a directory is empty
func IsDirEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()
	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

// TempDir creates a temporary subdirectory with prefix in dir or scratch-dir (if dir is empty string)
func TempDir(dir string, prefix string) (name string, err error) {
	if dir == "" {
		dir = qregistry.Registry["scratch-dir"]
	}
	name, err = os.MkdirTemp(dir, prefix)
	if err != nil {
		err = SetPathmode(name, "tempdir")
	}
	return
}

// TempFile creates a temporary file  in dir or scratch-dir
func TempFile(dir, prefix string) (name string, err error) {
	if dir == "" {
		dir = qregistry.Registry["scratch-dir"]
	}
	f, err := os.CreateTemp(dir, prefix)

	if err != nil {
		return
	}
	defer f.Close()
	name = f.Name()
	if err != nil {
		err = SetPathmode(name, "tempfile")
	}
	return
}

// CopyFile copies a file to another file or to a directory
// - src: sourcefile
// - dst: destination (if a directory, the basename of src is appended)
// - pathmode:
//   - "": nothing extra will be done
//   - "=": the same values of src will be applied
//   - otherwise a rgeistere dpathmode will be applied
// - keepmtime: ctime/mtime of src will be applied to dst
func CopyFile(src, dst, pathmode string, keepmtime bool) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)
	if IsDir(dst) {
		base := filepath.Base(src)
		dst = filepath.Join(dst, base)
	}

	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	if pathmode != "" && pathmode != "=" {
		err = SetPathmode(dst, pathmode)
		if err != nil {
			return
		}
	}
	if !keepmtime && pathmode != "=" {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	if pathmode == "=" {
		err = os.Chmod(dst, si.Mode())
		if err != nil {
			return
		}
		if runtime.GOOS != "windows" {
			uid, gid, ok := uidgid(si)
			if ok {
				err = os.Chown(dst, uid, gid)
				if err != nil {
					return
				}
			}
		}
	}
	if keepmtime {
		mtime := si.ModTime()
		atime := time.Now()
		err = os.Chtimes(dst, atime, mtime)
	}

	return
}

// CopyMeta owner, group, permissions from src to dst
func CopyMeta(src string, dst string, keepmtime bool) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	di, err := os.Stat(dst)
	if err != nil {
		return err
	}

	iss := IsDir(src)
	isd := IsDir(dst)

	if iss && !isd {
		return fmt.Errorf("`%s` is a directory, `%s` is not", src, dst)
	}

	if !iss && isd {
		return fmt.Errorf("`%s` is a directory, `%s` is not", dst, src)
	}

	var suid int
	var sgid int
	var duid int
	var dgid int

	if runtime.GOOS != "windows" {
		suid, sgid, _ = uidgid(si)
		duid, dgid, _ = uidgid(di)
	}

	if suid != duid || sgid != dgid {
		err = os.Chown(dst, suid, sgid)
		if err != nil {
			return
		}
	}
	fms := si.Mode().Perm()
	fmd := si.Mode().Perm()

	if fms == fmd {
		return nil
	}

	err = os.Chmod(dst, fms)

	if keepmtime {
		mtime := si.ModTime()
		atime := time.Now()
		err = os.Chtimes(dst, atime, mtime)
	}

	return

}

func CopyDir(src string, dst string, pathmode string, keepmtime bool) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return
	}
	if err == nil {
		return fmt.Errorf("destination already exists")
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return
	}
	if pathmode != "" && pathmode != "=" {
		err = SetPathmode(dst, pathmode)
		if err != nil {
			return
		}
	}
	if pathmode == "=" {
		err = CopyMeta(src, dst, false)
		if err != nil {
			return
		}
	}
	if keepmtime {
		mtime := si.ModTime()
		atime := time.Now()
		err = os.Chtimes(dst, atime, mtime)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return
	}

	for _, ery := range entries {
		entry, e := ery.Info()
		if e != nil {
			continue
		}
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = CopyDir(srcPath, dstPath, pathmode, keepmtime)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.

			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = CopyFile(srcPath, dstPath, pathmode, keepmtime)
			if err != nil {
				return
			}
		}
	}
	return
}

// cleanGlobPath prepares path for glob matching.
func cleanGlobPath(path string) string {
	switch path {
	case "":
		return "."
	default:
		return path[0 : len(path)-1] // chop off trailing separator
	}
}

// glob searches for files matching pattern in the directory dir
// and appends them to matches, returning the updated slice.
// If the directory cannot be opened, glob returns the existing matches.
// New matches are added in lexicographical order.
func glob(fsys fs.FS, root string, dir string, patterns []string, matches []string, recurse bool, files bool, dirs bool) (m []string, e error) {
	m = matches
	infos, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return // ignore I/O error
	}
	for _, info := range infos {
		n := info.Name()
		fmt.Println(n)
		isdir := info.IsDir()
		matched := len(patterns) == 0
		for _, pattern := range patterns {
			matched, err = path.Match(pattern, n)
			if err != nil {
				return m, err
			}
			if matched {
				break
			}
		}
		if matched {
			if (files && !isdir) || (dirs && isdir) {
				m = append(m, path.Join(root, dir, n))
			}
		}
		if isdir && recurse {
			newdir := dir + "/" + n
			if dir == "" || dir == "." {
				newdir = n
			}
			m, err = glob(fsys, root, newdir, patterns, m, true, files, dirs)
			if err != nil {
				return m, err
			}
		}
	}
	return
}

// hasMeta reports whether path contains any of the magic characters
// recognized by path.Match.
func hasMeta(path string) bool {
	for i := 0; i < len(path); i++ {
		c := path[i]
		if c == '*' || c == '?' || c == '[' || runtime.GOOS == "windows" && c == '\\' {
			return true
		}
	}
	return false
}

// Find lists regular files matching one of a list of patterns on the basename
//      if there are no patterns, all files are listed
//      results start with root
func Find(root string, patterns []string, recurse bool, files bool, dirs bool) (matches []string, err error) {
	fsys := os.DirFS(root)
	for _, pattern := range patterns {
		if _, err := path.Match(pattern, ""); err != nil {
			return nil, err
		}
	}
	return glob(fsys, root, ".", patterns, nil, recurse, files, dirs)
}

// AsyncWork works on all keys in a slice in parallel and returns a result (map indexed on key)
func AsyncWork(keys []string, fn func(key string) interface{}) (results map[string]interface{}) {
	if len(keys) == 0 {
		return
	}
	results = make(map[string]interface{})

	if len(keys) == 1 {
		p := keys[0]
		results[p] = fn(p)
		return
	}

	type res struct {
		key    string
		result interface{}
	}
	out := make(chan res)
	var wg sync.WaitGroup

	maxopen := len(keys)
	borrow, release, finish := qpattern.Number(maxopen)

	for _, key := range keys {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			borrow()
			defer release()
			r := fn(p)
			out <- res{
				key:    p,
				result: r,
			}
		}(key)
	}
	go func() {
		defer finish()
		wg.Wait()
		close(out)
	}()
	for re := range out {
		results[re.key] = re.result
	}
	return

}

// FilesDirs gets the regular files and dirs in directory
func FilesDirs(dir string) (files []os.FileInfo, dirs []os.FileInfo, err error) {
	fis, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	files = make([]os.FileInfo, 0)
	dirs = make([]os.FileInfo, 0)
	for _, f := range fis {
		fi, e := f.Info()
		if e != nil {
			continue
		}
		if fi.Mode().IsDir() {
			dirs = append(dirs, fi)
			continue
		}
		if fi.Mode().IsRegular() {
			files = append(files, fi)
		}
	}
	return
}
