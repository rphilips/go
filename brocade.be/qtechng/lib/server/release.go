package server

import (
	"log"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/spf13/afero"

	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
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
	if ok, _ := regexp.MatchString(`^[0-9][0-9]?\.[0-9][0-9]$`, r); !ok {
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
		x = filepath.Join(p...)
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
	for _, dir := range []string{"/", "/source/data", "/meta", "/unique", "/tmp", "/objects/m4", "/objects/l4", "/objects/i4", "/admin"} {
		fs.MkdirAll(dir, 0o770)
	}

	if !strings.ContainsRune(qtechType, 'B') || qregistry.Registry["qtechng-vc"] == "" {
		return nil
	}

	// initialises mercurial repository
	cmd := exec.Command("git", "init")
	sourcedir, _ := release.FS("").RealPath("/source")
	cmd.Dir = sourcedir
	cmd.Run()

	// initialise hg webserver
	data := `[extensions]
hgext.highlight=

[web]
pygments_style = default`

	backup := qregistry.Registry["qtechng-hg-backup"]
	if backup != "" {
		backup = strings.Replace(backup, "{version}", release.String(), -1)
		data += "\n\n[paths]\ndefault-push = " + backup + "\n"
	}
	_, _, _, err = release.FS("source").Store("/.hg/hgrc", data, "")
	if err != nil {
		return
	}

	cmd = exec.Command("hg", "add", "--quiet")
	cmd.Dir = sourcedir
	cmd.Run()

	cmd = exec.Command("hg", "commit", "--quiet", "--message", "Init")
	cmd.Dir = sourcedir
	cmd.Run()

	return err
}

////////////////////////////// Help functions ///////////////

// Canon maakt een officiele string van de versie
func Canon(r string) string {
	if r == "" || r == "0.00" {
		r = qregistry.Registry["brocade-release"]
	}
	r = strings.TrimRight(r, " -_betaBETA")
	if r == "" || r == "0.00" {
		r = qregistry.Registry["qtechng-version"]
	}
	return r
}

func fs(r string, readonly bool) func(s ...string) qvfs.QFs {
	place := qregistry.Registry["qtechng-repository-dir"]
	if place == "" {
		log.Fatal("Registry value `qtechng-repository-dir` should not be empty")
	}
	place = path.Join(place, r)
	fsys := afero.NewOsFs()
	if readonly {
		fsys = afero.NewReadOnlyFs(fsys)
	}
	g := func(s ...string) qvfs.QFs {
		p := place
		if len(s) != 0 {
			p = path.Join(place, path.Join(s...))
		} else {
			p = path.Join(place, "source", "data")
		}
		return qvfs.QFs{
			Afero:    afero.Afero{Fs: afero.NewBasePathFs(fsys, p)},
			ReadOnly: readonly,
		}
	}
	return g
}
