package fs

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
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
	ErrNoHome = errors.New("no HOME found")
	// ErrNotPathMode indicates not a valid pathmode
	ErrNotPathMode = errors.New("not a filemode")
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
	if !QSetPathMode() {
		return prop, nil
	}
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
		if len(mode) == 9 {
			perm = calcPerm(mode)
			return Property{nil, nil, perm}, nil
		}

	}
	switch pathmode {
	case "webdir", "webdavdir", "scriptdir", "daemondir":
		perm = calcPerm("rwxr-x---")
	case "webfile":
		perm = calcPerm("rwxr-----")
	case "webdavfile":
		perm = calcPerm("rw-rw----")
	case "scriptfile", "daemonfile":
		perm = calcPerm("rwxr-x---")
	case "processdir", "tempdir", "qtechdir":
		perm = calcPerm("rwxrwx---")
	case "qtechfile", "processfile", "tempfile":
		perm = calcPerm("rw-rw----")
	case "nakeddir":
		perm = calcPerm("rwxrwxrwx")
	case "nakedfile":
		perm = calcPerm("rw-rw-rw-")
	default:
		err = ErrNotPathMode
		return
	}
	if suid == "" || sgid == "" {
		return Property{nil, nil, perm}, nil
	}
	uid, err := user.Lookup(suid)
	if err != nil {
		return
	}
	gid, err := user.LookupGroup(sgid)
	if err != nil {
		return
	}
	return Property{uid, gid, perm}, err
}

func calcPerm(nine string) os.FileMode {
	var perm os.FileMode = 0
	var plus os.FileMode = 0
	for i, c := range nine {
		perm *= 2
		if c == 's' && i == 2 {
			plus += 04000
		}
		if c == 's' && i == 5 {
			plus += 02000
		}
		if c == 't' && i == 8 {
			plus += 01000
		}

		if c == '-' {
			continue
		}
		perm++
	}
	return perm + plus
}

// SetPathMode assigns the ownership and access modes to a path
func SetPathmode(pth string, pathmode string) (err error) {
	if !QSetPathMode() {
		return nil
	}
	if pathmode == "" {
		return ErrNotPathMode
	}
	p, err := Properties(pathmode)
	if err == nil {
		if p.GID != nil && p.UID != nil {
			gi, _ := strconv.Atoi(p.GID.Gid)
			ui, _ := strconv.Atoi(p.UID.Uid)
			_ = os.Chown(pth, ui, gi)
		}
		_ = os.Chmod(pth, p.PERM.Perm())
	}
	return err
}

func QSetPathMode() bool {
	if runtime.GOOS == "windows" {
		return false
	}
	// if qregistry.Registry["qtechng-type"] == "W" {
	// 	return false
	// }
	return true
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
	if !QSetPathMode() {
		return nil
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

// Append write bytes to a file. File has to exist
func Append(fname string, tail []byte) (err error) {
	f, err := os.OpenFile(fname, os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	if len(tail) != 0 {
		if _, err = f.Write(tail); err != nil {
			return
		}
	} else {
		if _, err = f.Write([]byte{10}); err != nil {
			return
		}
		fi, err := f.Stat()
		if err != nil {
			return err
		}
		return f.Truncate(fi.Size() - 1)
	}
	return
}

// Mkdir makes a directory and sets the access
func Mkdir(dirname string, pathmode string) (err error) {
	if !strings.HasSuffix(pathmode, "dir") {
		pathmode = pathmode + "dir"
	}
	prm := fs.FileMode(0770)
	if QSetPathMode() {
		perm, e := Properties(pathmode)
		if e != nil {
			return e
		}
		prm = perm.PERM
	}
	err = os.Mkdir(dirname, prm)
	if err == nil {
		if !QSetPathMode() {
			return nil
		}
		return SetPathmode(dirname, pathmode)
	}
	if os.IsExist(err) {
		return nil
	}
	parent := filepath.Dir(dirname)
	if parent == "" || dirname == parent {
		return err
	}
	Mkdir(parent, pathmode)
	os.Mkdir(dirname, prm)
	if !QSetPathMode() {
		return nil
	}
	return SetPathmode(dirname, pathmode)
}

// MkdirAll makes a directory and sets the access
func MkdirAll(dirname string, pathmode string) (err error) {
	dirs := make([]string, 0)
	for {
		if Exists(dirname) {
			break
		}
		dirs = append(dirs, dirname)
		parent := filepath.Dir(dirname)
		if parent == "" || parent == dirname {
			break
		}
		dirname = parent
	}
	if len(dirname) == 0 {
		return Mkdir(dirname, pathmode)
	}
	for i := len(dirs) - 1; i > -1; i-- {
		dirname = dirs[i]
		Mkdir(dirname, pathmode)
	}
	return nil
}

// Rmpath removes a file or a dirctory tree except root
func Rmpath(dirname string) (err error) {
	dirname, err = AbsPath(dirname)
	if err == nil {
		parent := filepath.Dir(dirname)
		if parent == dirname || parent == "" {
			err = errors.New("cannot delete a root")
		} else {
			err = os.RemoveAll(dirname)
		}
	}
	return err
}

// RmpathUntil removes a file or a dirctory tree except root
func RmpathUntil(dirname string, until string) (err error) {
	if until == "" {
		return Rmpath(dirname)
	}
	dirname, err = AbsPath(dirname)
	if err != nil {
		return err
	}
	until, err = AbsPath(until)
	if err != nil {
		return err
	}
	if SameFile(dirname, until) {
		return nil
	}
	rel, err := filepath.Rel(until, dirname)
	if err != nil {
		return err
	}
	if rel == "" {
		return nil
	}
	up := ".." + string(os.PathSeparator)
	if strings.HasPrefix(rel, up) {
		return nil
	}

	parent := ""
	if err == nil {
		parent = filepath.Dir(dirname)
		if parent == dirname || parent == "" {
			err = errors.New("cannot delete a root")
		} else {
			err = os.RemoveAll(dirname)
		}
	}
	if err != nil {
		return err
	}
	if SameFile(parent, until) {
		return nil
	}
	empty, _ := IsDirEmpty(parent)
	if !empty {
		return nil
	}
	return RmpathUntil(parent, until)
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
	if err == nil {
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
	if !QSetPathMode() {
		return name, nil
	}
	err = SetPathmode(name, "tempfile")
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
	fmd := di.Mode().Perm()

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
		os.Chtimes(dst, atime, mtime)
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
		isdir := info.IsDir()
		matched := len(patterns) == 0
		for _, pattern := range patterns {
			matched, err = filepath.Match(pattern, n)
			if err != nil {
				return m, err
			}
			if matched {
				break
			}
		}
		if matched {
			if (files && !isdir) || (dirs && isdir) {
				m = append(m, filepath.Join(root, dir, n))
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

// Find lists regular files matching one of a list of patterns on the basename
//      if there are no patterns, all files are listed
//      results start with root
func Find(root string, patterns []string, recurse bool, files bool, dirs bool) (matches []string, err error) {
	fsys := os.DirFS(root)
	for _, pattern := range patterns {
		if _, err := filepath.Match(pattern, ""); err != nil {
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

//SameFile file1, file2
func SameFile(file1, file2 string) bool {
	old, err := os.Stat(file1)
	if err != nil {
		return false
	}
	new, err := os.Stat(file2)
	if err != nil {
		return false
	}
	return os.SameFile(old, new)
}

// Refresh executable
func RefreshEXE(oldexe string, newexe string) error {
	si, err := os.Stat(oldexe)
	if err != nil {
		return err
	}
	os.Remove(oldexe + ".bak")
	ftmp, err := TempFile(filepath.Dir(oldexe), "exe-")
	if err != nil {
		return err
	}
	err = os.Rename(newexe, ftmp)
	if err != nil {
		err = CopyFile(newexe, ftmp, "", false)
	}
	if err != nil {
		return err
	}
	newexe = ftmp

	for i := 0; i < 2; i++ {
		os.Rename(oldexe, oldexe+".bak")
		os.Remove(oldexe)
		err = os.Rename(newexe, oldexe)

		if err == nil || i == 1 {
			break
		}
		if err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}
	err = os.Chmod(oldexe, si.Mode())
	if err != nil {
		return err
	}
	return err
}

func GetURL(url string, fname string, pathmode string) error {
	out, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer out.Close()
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func IsSubDir(parent string, ksubdir string) bool {
	if ksubdir == "" {
		return false
	}
	parent, _ = AbsPath(parent)
	ksubdir, _ = AbsPath(ksubdir)
	if parent == ksubdir {
		return true
	}
	if strings.HasPrefix(ksubdir, strings.TrimSuffix(parent, string(os.PathSeparator))+string(os.PathSeparator)) {
		return true
	}
	rel, e := filepath.Rel(parent, ksubdir)
	if e != nil {
		return false
	}
	if strings.HasPrefix(rel, "..") {
		return false
	}
	return true
}

type reader struct{}

func (reader) Read(b []byte) (int, error) {
	for i := range b {
		b[i] = 0
	}
	return len(b), nil
}

var NullReader io.Reader

func Read(b []byte) (n int, err error) {
	return NullReader.Read(b)
}

var NullWriter io.Writer

func Write(p []byte) (n int, err error) {
	return NullWriter.Write(p)
}

func init() {
	NullReader = new(reader)
	NullWriter = ioutil.Discard
}

func Log(v ...interface{}) {
	filename := qregistry.Registry["qtechng-log"]
	if filename == "" {
		return
	}
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	fmt.Fprintln(f, v...)
	fmt.Fprintln(f, "===")
}