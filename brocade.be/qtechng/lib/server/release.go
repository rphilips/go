package server

import (
	"encoding/json"
	"log"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/spf13/afero"

	qfs "brocade.be/base/fs"
	qparallel "brocade.be/base/parallel"
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
	for _, dir := range []string{"/", "/source/data", "/meta", "/unique", "/tmp", "/object/t4", "/object/m4", "/object/l4", "/object/i4", "/object/r4", "/admin", "/log"} {
		fs.MkdirAll(dir, 0o770)
	}

	release.InitGit()

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

func (release Release) ObjectCount() map[string]int {
	stat := make(map[string]int)
	fs := release.FS("/")
	for _, ty := range []string{"i4", "l4", "m4", "r4", "b4"} {
		dir, _ := fs.RealPath("/object/" + ty)
		all, _ := qfs.Find(dir, []string{"obj.json"}, true, true, false)
		stat[ty] = len(all)
	}
	return stat

}

func (release Release) ProjectCount() map[string]int {
	stat := make(map[string]int)
	fs := release.FS("/")
	for _, ty := range []string{"/source/data"} {
		dir, _ := fs.RealPath(ty)
		all, _ := qfs.Find(dir, []string{"brocade.json"}, true, true, false)
		stat[ty] = len(all)
	}
	return stat
}

func (release Release) SourceCount() map[string]int {
	stat := make(map[string]int)
	fs := release.FS("/")
	for _, ty := range []string{"/source/data"} {
		dir, _ := fs.RealPath(ty)
		all, _ := qfs.Find(dir, nil, true, true, false)
		for _, s := range all {
			ext := filepath.Ext(s)
			stat[ext]++
		}
	}
	return stat
}

func (release *Release) ObjectStore(objname string, obj json.Marshaler) (changed bool, before []byte, after []byte, err error) {
	fs, place := release.ObjectPlace(objname)
	return fs.Store(place, obj, "")
}

func (release *Release) UniqueStore(qpath string) error {
	fs, place := release.UniquePlace(qpath)
	m := map[string]string{"path": qpath}
	_, _, _, err := fs.Store(place, m, "")
	return err
}

func (release *Release) UniqueUnlink(qpath string) {
	fs, place := release.UniquePlace(qpath)
	fs.Waste(place)
}

func (release Release) ReInit() error {
	fs := release.FS("/")
	for _, ty := range []string{"/unique", "/object/m4", "/object/l4", "/object/i4", "/object/r4"} {
		dir, _ := fs.RealPath(ty)
		qfs.Rmpath(dir)
		fs.MkdirAll(ty, 0o770)
	}
	return nil
}

func (release Release) QPaths() []string {
	fs := release.FS()
	dir, _ := fs.RealPath("/")
	files, _ := qfs.Find(dir, nil, true, true, false)
	qpaths := make([]string, len(files))
	for i, file := range files {
		qpath, _ := filepath.Rel(dir, file)
		qpath = filepath.ToSlash(qpath)
		qpath = strings.Trim(qpath, "/")
		qpath = "/" + qpath
		qpaths[i] = qpath
	}
	return qpaths
}

func (release Release) Modifications(after time.Time, mode string) (result map[string][]string, err error) {
	result = make(map[string][]string)
	stamp := time.Now()
	dir := ""
	switch mode {
	case "source":
		dir, _ = release.FS("/").RealPath("/source/data")
	case "meta":
		dir, _ = release.FS("/").RealPath("/meta")
	case "i4", "m4", "r4", "l4", "t4":
		dir, _ = release.FS("/").RealPath("/object/" + mode)
	}
	root, _ := release.FS("/").RealPath("/")
	if dir != "" {
		changed, err := qfs.ChangedAfter(dir, after, nil)
		if err != nil {
			return nil, err
		}
		if len(changed) == 0 {
			return nil, nil
		}
		changed2 := make([]string, len(changed))
		for i, p := range changed {
			rel, _ := filepath.Rel(root, p)
			changed2[i] = rel
		}
		sort.Strings(changed2)
		result[mode] = changed2
		return result, nil
	}

	modes := []string{"source", "meta", "i4", "m4", "r4", "t4", "l4"}
	fn := func(n int) (result interface{}, err error) {
		mode := modes[n]
		return release.Modifications(after, mode)
	}
	all, errorlist := qparallel.NMap(len(modes), -1, fn)
	for _, e := range errorlist {
		if e != nil {
			return nil, e
		}
	}
	for _, res := range all {
		x := res.(map[string][]string)
		for sub, value := range x {
			if len(value) != 0 {
				result[sub] = value
			}
		}
	}
	result["context"] = []string{stamp.Format(time.RFC3339Nano), root}
	return result, nil
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
	if n > 0 && len(versions) > (n-1) {
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
