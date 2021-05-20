package server

import (
	"archive/tar"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/spf13/afero"

	qfs "brocade.be/base/fs"
	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
	qutil "brocade.be/qtechng/lib/util"
	qvfs "brocade.be/qtechng/lib/vfs"
)

var releaseCache = new(sync.Map)

// Release models a Brocade version
type Release struct {
	r        string
	readonly bool
	FS       func(s ...string) qvfs.QFs
}

// New constructs a new Release
func (Release) New(r string, readonly bool) (release *Release, err error) {
	r = Canon(r)

	rid := r
	if readonly {
		rid = r + " R"
	}

	rel, _ := releaseCache.Load(rid)
	if rel != nil {
		return rel.(*Release), nil
	}
	if ok, _ := regexp.MatchString(`^[0-9]+\.[0-9][0-9]$`, r); !ok {
		err = &qerror.QError{
			Ref:     []string{"release.new"},
			Version: r,
			Msg:     []string{"Release invalid format"},
		}
		return
	}
	release = &Release{r, readonly, fs(r, readonly)}

	rel, _ = releaseCache.LoadOrStore(rid, release)
	if rel != nil {
		return rel.(*Release), nil
	}
	err = &qerror.QError{
		Ref:     []string{"release.new.cache"},
		Version: r,
		Msg:     []string{"Cannot create a release"},
	}
	return
}

// String of a release: release fulfills the Stringer interface
func (release Release) String() string {
	return release.r
}

// Root base directory
func (release Release) Root() string {
	x, _ := release.FS("/").RealPath("")
	return x
}

// ReadOnly returns true if the release is ReadOnly
func (release Release) ReadOnly() bool {
	return release.readonly
}

// Exists returns true if  a path exists in a release.
// If the path is empty, the function returns if the release itself exists
func (release Release) Exists(p ...string) (bool, error) {
	x := "/source/data"
	if len(p) != 0 {
		x = path.Join(p...)
	}
	exists, err := release.FS("/").Exists(x)
	return exists, err
}

// Init creates a release on disk. The release should be a valid n.nn construction
func (release Release) Init() (err error) {
	if release.ReadOnly() {
		err = &qerror.QError{
			Ref:     []string{"release.init.readonly"},
			Version: release.String(),
			Msg:     []string{"Release is readonly"},
		}
		return err

	}
	ok, err := release.Exists()
	if ok {
		err = &qerror.QError{
			Ref:     []string{"release.init.exists"},
			Version: release.String(),
			Msg:     []string{"Release exists already"},
		}
		return err
	}

	qtechType := qregistry.Registry["qtechng-type"]
	repository := qregistry.Registry["qtechng-repository-dir"]
	if repository == "" {
		err = &qerror.QError{
			Ref:     []string{"release.create.repository"},
			Version: release.String(),
			Msg:     []string{"Create a version only when a qtechng-repository-dir is defined"},
		}
		return err
	}
	fs := release.FS("/")
	for _, dir := range []string{"/", "/source/data", "/meta", "/unique", "/tmp", "/object/m4", "/object/l4", "/object/i4", "/admin"} {
		fs.MkdirAll(dir, 0o770)
	}

	if !strings.ContainsRune(qtechType, 'B') || qregistry.Registry["qtechng-git-enable"] == "1" {
		return nil
	}

	// initialises mercurial repository
	cmd := exec.Command("git", "init", "--quiet")
	sourcedir, _ := release.FS("").RealPath("/source")
	cmd.Dir = sourcedir
	cmd.Run()

	cmd = exec.Command("git", "add", "--all")
	cmd.Dir = sourcedir
	cmd.Run()

	cmd = exec.Command("git", "commit", "--quiet", "--message", "Init")
	cmd.Dir = sourcedir
	cmd.Run()

	return err
}

// IsInstallable checks if th erelease can be installed.
func (release Release) IsInstallable() bool {
	qtechType := qregistry.Registry["qtechng-type"]
	if strings.Contains(qtechType, "B") && release.String() != "0.00" {
		return false
	}
	blocktime := qregistry.Registry["qtechng-block-install"]
	if blocktime != "" {
		h := time.Now()
		t := h.Format(time.RFC3339)
		if strings.Compare(blocktime, t) < 0 {
			blocktime = ""
			qregistry.SetRegistry("qtechng-block-install", "")
		}
	}
	if blocktime == "0" {
		blocktime = ""
		qregistry.SetRegistry("qtechng-block-install", "")
	}
	if blocktime != "" {
		return false
	}
	if strings.Contains(qtechType, "B") {
		return true
	}
	r := Canon(qregistry.Registry["brocade-release"])
	return release.String() == r
}

// Backup makes a backup of the sources and the meta information
func (release Release) Backup(tarfile string) error {
	ftar, err := os.Create(tarfile)

	if err != nil {
		return err
	}
	defer ftar.Close()
	errs := make([]error, 0)

	tw := tar.NewWriter(ftar)
	defer tw.Close()

	fs := release.FS("/")
	for _, dir := range []string{"/source/data", "/meta"} {
		source, err := fs.RealPath(dir)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		info, err := os.Stat(source)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		var baseDir string
		if info.IsDir() {
			baseDir = dir[1:]
		}
		err = filepath.Walk(source,
			func(p string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				header, err := tar.FileInfoHeader(info, info.Name())
				if err != nil {
					return err
				}

				if baseDir != "" {
					header.Name = filepath.Join(baseDir, filepath.ToSlash(strings.TrimPrefix(p, source)))
				}

				if err := tw.WriteHeader(header); err != nil {
					return err
				}

				if info.IsDir() {
					return nil
				}

				file, err := os.Open(p)
				if err != nil {
					return err
				}
				defer file.Close()
				_, err = io.Copy(tw, file)
				return err
			})
		if err != nil {
			errs = append(errs, err)
			continue
		}
	}
	tw.Flush()
	tw.Close()
	ftar.Close()
	if len(errs) == 0 {
		return nil
	}
	return qerror.ErrorSlice(errs)
}

// Restore restores a backup and the meta information
func (release Release) Restore(tarfile string, init bool) (previous string, err error) {

	reader, err := os.Open(tarfile)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	// create backup
	h := time.Now()
	t := h.Format(time.RFC3339)[:19]
	t = strings.ReplaceAll(t, ":", "")
	t = strings.ReplaceAll(t, "-", "")
	r := release.String()
	previous = filepath.Join(filepath.Dir(tarfile), "previous-brocade-"+r+"-"+t+".tar")
	err = release.Backup(previous)
	if err != nil {
		return "", err
	}

	// initialises if necessary
	fs := release.FS("/")
	if init {
		fs := release.FS("/")
		for _, dir := range []string{"/source/data", "/meta"} {
			source, _ := fs.RealPath(dir)
			qfs.Rmpath(source)
			if qfs.Exists(source) {
				return previous, fmt.Errorf("cannot remove `%s`", source)
			}
			qfs.Mkdir(source, "qtech")
			if !qfs.IsDir(source) {
				return previous, fmt.Errorf("cannot create directory `%s`", source)
			}
		}
	}

	// restore files

	tarReader := tar.NewReader(reader)
	target, _ := fs.RealPath("/")
	if !qfs.IsDir(target) {
		return previous, fmt.Errorf("cannot write to directory `%s`", target)
	}

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return previous, fmt.Errorf("cannot read tarfile `%s`", tarfile)
		}

		path := filepath.Join(target, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err = qfs.MkdirAll(path, "qtech"); err != nil {
				return previous, fmt.Errorf("cannot make directory `%s`", path)
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return previous, fmt.Errorf("cannot make file `%s` with mode `%s` (error `%s`)", path, info.Mode(), err)
		}
		defer file.Close()
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return previous, fmt.Errorf("cannot write to file `%s` with mode `%s`", path, info.Mode())
		}
		file.Close()
	}
	return previous, nil

}

////////////////////////////// Help functions ///////////////

// Canon maakt een officiele string van de versie
func Canon(r string) string {

	br := qregistry.Registry["brocade-release"]
	br = strings.TrimRight(br, " -_betaBETA")
	rr := strings.TrimRight(r, " -_betaBETA")
	qtechType := qregistry.Registry["qtechng-type"]

	if !strings.Contains(qtechType, "B") {
		if rr == "" || rr == "0.00" {
			return br
		}
		return rr
	}
	if rr == br {
		return "0.00"
	}
	if rr == "" {
		return "0.00"
	}
	return rr

}

func Releases(n int) string {
	place := qregistry.Registry["qtechng-repository-dir"]
	_, dirs, _ := qfs.FilesDirs(place)
	versions := make([]string, 0)
	for _, vi := range dirs {
		v := filepath.Base(vi.Name())
		if v == "0.00" {
			continue
		}
		versions = append(versions, v)
	}
	if len(versions) > (n - 1) {
		versions = versions[:n-1]
	}

	sort.Slice(versions, func(i, j int) bool { return qutil.LowestVersion(versions[i], versions[j]) == versions[i] })
	versions = append(versions, "0.00")
	return strings.Join(versions, " ")
}

func fs(r string, readonly bool) func(s ...string) qvfs.QFs {
	place := qregistry.Registry["qtechng-repository-dir"]
	if place == "" {
		log.Fatal("Registry value `qtechng-repository-dir` should not be empty")
	}
	place = filepath.Join(place, r)
	fsys := afero.NewOsFs()
	if readonly {
		fsys = afero.NewReadOnlyFs(fsys)
	}
	g := func(s ...string) qvfs.QFs {
		p := ""
		if len(s) != 0 {
			p = filepath.Join(place, filepath.Join(s...))
		} else {
			p = filepath.Join(place, "source", "data")
		}
		return qvfs.QFs{
			Afero:    afero.Afero{Fs: afero.NewBasePathFs(fsys, p)},
			ReadOnly: readonly,
		}
	}
	return g
}
