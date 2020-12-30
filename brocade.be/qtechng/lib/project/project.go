package project

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode"

	qfnmatch "brocade.be/base/fnmatch"
	qparallel "brocade.be/base/parallel"
	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
	qmeta "brocade.be/qtechng/lib/meta"
	qserver "brocade.be/qtechng/lib/server"
	qutil "brocade.be/qtechng/lib/util"
	qvfs "brocade.be/qtechng/lib/vfs"
)

var projectCache = new(sync.Map)

// Project models a Brocade project
type Project struct {
	p        string
	readonly bool
	r        *qserver.Release
}

// New constructs a new Project
func (Project) New(r string, p string, readonly bool) (project *Project, err error) {
	p = qutil.Canon(p)
	if !ValidProjectString(p) {
		err = &qerror.QError{
			Ref:     []string{"project.new.id"},
			Version: r,
			Project: p,
			Msg:     []string{"Project string is not valid"},
		}
		return
	}

	version, err := qserver.Release{}.New(r, readonly)
	if err != nil {
		e := &qerror.QError{
			Ref:     []string{"project.New"},
			Version: r,
			Project: p,
			Msg:     []string{"Cannot instantiate version"},
		}
		err = qerror.QErrorTune(err, e)
		return
	}
	r = version.String()
	pid := r + " " + p
	if readonly {
		pid = r + " " + p + " R"
	}

	proj, _ := projectCache.Load(pid)
	if proj != nil {
		return proj.(*Project), nil
	}

	project = &Project{p, readonly, version}

	if ok, _ := version.Exists("/source/data", p, "brocade.json"); ok {
		prj, _ := projectCache.LoadOrStore(pid, project)
		if prj != nil {
			return prj.(*Project), nil
		}
	}

	err = ValidProjectRelease(p, *version)
	if err != nil {
		return
	}
	prj, _ := projectCache.LoadOrStore(pid, project)
	if prj != nil {
		return prj.(*Project), nil
	}
	err = &qerror.QError{
		Ref:     []string{"project.new.cache"},
		Version: r,
		Msg:     []string{"Cannot create a project"},
	}
	return
}

// Orden calculates a string for a project to indicates its ordening under all projects
func (project Project) Orden() (sort string) {
	r := project.Release().String()
	p := project.String() + "/brocade.json"
	seq, _ := Sequence(r, p, true)
	sort = "1"
	if project.IsCore() {
		sort = "0"
	}
	for _, p := range seq {
		cfg, e := p.LoadConfig()
		prio := 10000
		if e == nil && cfg.Priority != 0 {
			prio = cfg.Priority
			if prio > 99999 {
				prio = 99999
			}
			if prio < 1 {
				prio = 1
			}
		}
		prio = 1000000 - prio
		sort += strconv.Itoa(prio)
	}
	return sort
}

// String of a release: release fulfills the Stringer interface
func (project Project) String() string {
	return project.p
}

// ReadOnly returns true if the release is ReadOnly
func (project Project) ReadOnly() bool {
	return project.readonly
}

// Release returns pointer to version
func (project Project) Release() qserver.Release {
	return *(project.r)
}

// Exists returns true if  a path exists in a release.
// If the path is empty, the function returns if the release itself exists
func (project Project) Exists(p string) bool {
	if p == "" {
		p = "brocade.json"
	}
	if _, err := project.Release().FS().Stat(p); err != nil {
		return !os.IsNotExist(err)
	}
	return true
}

// FS gives the filesystem associated with the project
func (project Project) FS() qvfs.QFs {
	return project.Release().FS("/source/data", project.String())
}

// Init creates a project on disk. The release should be a valid n.nn construction
func (project Project) Init(meta qmeta.Meta) (err error) {
	sversion := project.Release().String()
	sproject := project.String()
	if project.ReadOnly() {
		err = &qerror.QError{
			Ref:     []string{"project.init.readonly"},
			Version: sversion,
			Project: sproject,
			Msg:     []string{"Project is readonly"},
		}
		return err
	}
	fs := project.FS()
	fname := "/brocade.json"
	exists, _ := fs.Exists(fname)
	if exists {
		err := &qerror.QError{
			Ref:     []string{"project.init.exists"},
			Version: sversion,
			Project: sproject,
			Msg:     []string{"Project exists already"},
		}
		return err
	}

	err = ValidProjectRelease(sproject, project.Release())
	if err != nil {
		return err
	}

	data := `{
	"$schema": "https://dev.anet.be/brocade/schema/qtechng.schema.json"
}`
	fs.Store(fname, data, "")

	pmet, _ := qmeta.Meta{}.New(sversion, sproject+fname)
	pmet.Update(meta)

	_, err = pmet.Store(sversion, sproject+fname)

	return nil
}

// Store creates a project on disk. The release should be a valid n.nn construction
func (project Project) Store(fname string, data interface{}) (changed bool, err error) {
	if project.ReadOnly() {
		err = &qerror.QError{
			Ref:     []string{"project.store.readonly"},
			Version: project.Release().String(),
			Project: project.String(),
			Msg:     []string{"Project is readonly"},
		}
		return false, err
	}
	fs := project.FS()
	if !strings.HasPrefix(fname, "/") {
		fname = "/" + fname
	}
	changed, _, _, e := fs.Store(fname, data, "")

	if e != nil {
		err = &qerror.QError{
			Ref:     []string{"project.store"},
			Version: project.Release().String(),
			Project: project.String(),
			File:    fname,
			Msg:     []string{"Error on write: `" + e.Error() + "`"},
		}
		return false, err
	}
	return false, nil
}

// Fetch fetches a file from a project
func (project Project) Fetch(fname string) (blob []byte, err error) {
	fname = qutil.Canon(fname)
	fs := project.FS()

	blob, e := fs.ReadFile(fname)
	if e != nil {
		err := &qerror.QError{
			Ref:     []string{"project.fetch"},
			Version: project.Release().String(),
			Project: project.String(),
			File:    fname,
			Msg:     []string{"Error on read: `" + e.Error() + "`"},
		}
		return blob, err
	}
	return blob, nil
}

// Unlink removes a project
func (project Project) Unlink() (err error) {
	r := project.Release().String()
	p := project.String()
	fs := project.FS()
	if len(fs.Dir("/", false, false)) > 1 {
		err = &qerror.QError{
			Ref:     []string{"project.unlink"},
			Version: r,
			Project: p,
			Msg:     []string{"Cannot remove project: there are still files"},
		}
		return
	}
	fs = project.Release().FS()
	_, e := fs.Waste(project.String() + "/brocade.json")
	if e != nil {
		err = &qerror.QError{
			Ref:     []string{"project.unlink.config"},
			Version: r,
			Project: p,
			File:    "brocade.json",
			Msg:     []string{"Cannot remove `brocade.json`: " + e.Error()},
		}
		return err
	}
	pid := r + " " + p
	projectCache.Delete(pid)
	configCache.Delete(pid)
	pid = r + " " + p + " R"
	projectCache.Delete(pid)
	configCache.Delete(pid)
	return nil
}

// IsInstallable vist uit of een project installeerbaar is.
func (project Project) IsInstallable() error {

	v := project.Release().String()
	p := project.String()

	err := &qerror.QError{
		Ref:     []string{"project.installable"},
		Version: v,
		Project: p,
	}

	r := qserver.Canon("")
	if r != v {
		err.Msg = []string{fmt.Sprintf("`%s` is not the current version", v)}
		return err
	}

	sequence, e := Sequence(v, project.String(), true)

	if e != nil || len(sequence) == 0 || project.String() != sequence[len(sequence)-1].String() {
		err.Msg = []string{fmt.Sprintf("Structure of `%s` is invalid", p)}
		return err
	}

	for _, proj := range sequence {
		project := proj
		config, e := project.LoadConfig()
		if e != nil {
			err.Msg = []string{fmt.Sprintf("Project `%s` has invalid `brocade.json`: %s", proj.String(), e.Error())}
			return err
		}
		// Passive
		if config.Passive {
			err.Msg = []string{fmt.Sprintf("Project `%s` is not active", proj.String())}
			return err
		}
		// Mumps
		if len(config.Mumps) > 0 && find(config.Mumps, "", true) == -1 {
			m := qregistry.Registry["m-os-type"]
			if find(config.Mumps, m, false) == -1 {
				err.Msg = []string{fmt.Sprintf("Project `%s` does not work with `%s`", proj.String(), m)}
				return err
			}
		}
		// Groups
		if len(config.Groups) > 0 {
			g := qregistry.Registry["system-group"]
			if find(config.Groups, g, true) == -1 {
				err.Msg = []string{fmt.Sprintf("Project `%s` does not install on `%s`", proj.String(), g)}
				return err
			}
		}
		// Name
		if len(config.Names) > 0 {
			g := qregistry.Registry["system-name"]
			if find(config.Names, g, true) == -1 {
				err.Msg = []string{fmt.Sprintf("Project `%s` does not install on `%s`", proj.String(), g)}
				return err
			}
		}
		// Roles
		if len(config.Roles) > 0 {
			r := qregistry.Registry["system-roles"]
			rs := strings.FieldsFunc(r, func(c rune) bool {
				return !unicode.IsLetter(c) && !unicode.IsNumber(c)
			})
			for _, role := range config.Roles {
				if find(rs, role, false) == -1 {
					err.Msg = []string{fmt.Sprintf("Project `%s` does not install with role `%s`", proj.String(), role)}
					return err
				}
			}
		}
		// VersionLower
		if config.VersionLower != "" {
			r := qregistry.Registry["brocade-release"]
			if strings.Compare(config.VersionLower, r) != -1 {
				err.Msg = []string{fmt.Sprintf("Version should be higher or equal than `%s` for project `%s`", config.VersionLower, proj.String())}
				return err
			}
		}
		// VersionUpper
		if config.VersionUpper != "" {
			r := qregistry.Registry["brocade-release"]
			if strings.Compare(r, config.VersionUpper) != -1 {
				err.Msg = []string{fmt.Sprintf("Version should be lower or equal than `%s` for project `%s`", config.VersionUpper, proj.String())}
				return err
			}
		}
	}
	return nil
}

// IsCore vist uit of een project een kern project is.
func (project Project) IsCore() bool {
	sequence, err := Sequence(project.Release().String(), project.String(), true)
	if err != nil {
		return false
	}
	if len(sequence) == 0 {
		return false
	}
	if project.String() != sequence[len(sequence)-1].String() {
		return false
	}
	for _, proj := range sequence {
		project := proj
		config, err := project.LoadConfig()
		if err != nil {
			return false
		}
		// Core
		if config.Core {
			return true
		}
	}
	return false
}

// IsConfig vist uit of een path een configuratiefile is
func (project Project) IsConfig(s string) bool {
	if !strings.HasSuffix(s, "/brocade.json") {
		return false
	}
	relpath := s[len(project.String())+1:]
	if relpath == "brocade.json" {
		return true
	}
	return false
}

// QPaths finds all paths in project
func (project Project) QPaths(patterns []string, matchonlybasename bool) (qpaths []string) {
	fs := project.FS()
	paths := fs.Glob("/", patterns, matchonlybasename)
	if len(paths) == 0 {
		return nil
	}
	qpaths = make([]string, len(paths))
	prefix := project.String()
	for i, p := range paths {
		qpaths[i] = prefix + p
	}
	return qpaths
}

// InitList creates a list of projects.
func InitList(version string, projects []string, fmeta func(string) qmeta.Meta) (result []string, errs error) {

	release, err := qserver.Release{}.New(version, true)
	if err != nil {
		return nil, err
	}

	ok, err := release.Exists()
	if !ok && err == nil {
		err := &qerror.QError{
			Ref:     []string{"project.init.version.notexists"},
			Version: release.String(),
			Msg:     []string{"Version `" + version + "` does not exists"},
		}
		return nil, err
	}
	if err != nil {
		err := &qerror.QError{
			Ref:     []string{"project.init.version.notexists.error"},
			Version: release.String(),
			Msg:     []string{err.Error()},
		}
		return nil, err
	}

	fn := func(n int) (interface{}, error) {
		project, err := Project{}.New(version, projects[n], false)
		if err == nil {
			meta := fmeta(project.String())
			err = project.Init(meta)
		}
		if err != nil {
			return "", err
		}
		return project.String(), nil
	}

	resultlist, errorlist := qparallel.NMap(len(projects), -1, fn)

	for _, p := range resultlist {
		if p != "" {
			result = append(result, p.(string))
		}
	}

	errslice := qerror.NewErrorSlice()

	for _, e := range errorlist {
		if e != nil {
			errslice = append(errslice, e)
		}
	}

	if len(errslice) != 0 {
		errs = errslice
	} else {
		errs = nil
	}
	return
}

// Info verzamelt informatie over projecten
func Info(version string, patterns []string) (result map[string]map[string][]string, err error) {
	projs, err := List(version, patterns)
	if err != nil {
		return
	}
	result = make(map[string]map[string][]string)
	for _, proj := range projs {
		parents := make([]string, 0)
		seq, _ := Sequence(version, proj, false)
		for _, p := range seq {
			if p.String() != proj {
				parents = append(parents, p.String())
			}
		}

		pattern := proj + "/*"
		children, _ := List(version, []string{pattern})
		result[proj] = make(map[string][]string)
		if len(children) == 0 {
			result[proj]["children"] = nil
		} else {
			result[proj]["children"] = children
		}
		if len(parents) == 0 {
			result[proj]["parents"] = nil
		} else {
			result[proj]["parents"] = parents
		}
	}
	return
}

// List searches projects
func List(version string, patterns []string) (result []string, err error) {
	release, err := qserver.Release{}.New(version, true)
	if err != nil {
		return nil, err
	}
	ok, e := release.Exists()
	if e != nil {
		err := &qerror.QError{
			Ref:     []string{"project.list.version.notexists.error"},
			Version: release.String(),
			Msg:     []string{e.Error()},
		}
		return nil, err
	}
	if !ok {
		err := &qerror.QError{
			Ref:     []string{"project.list.version.notexists"},
			Version: release.String(),
			Msg:     []string{"Version `" + version + "` does not exists"},
		}
		return nil, err
	}
	fs := release.FS()
	if len(patterns) == 0 {
		patterns = []string{"*"}
	}
	pats := make([]string, 0)

	for _, pattern := range patterns {
		pats = append(pats, pattern+"/brocade.json")
	}
	projects := fs.Glob("/", pats, false)

	if len(projects) == 0 {
		return nil, nil
	}

	result = make([]string, len(projects))

	for i, p := range projects {
		result[i] = filepath.Dir(p)
	}
	sort.Strings(result)

	return result, nil
}

// ValidProjectString checks if a string can be a project
func ValidProjectString(p string) bool {
	if !strings.HasPrefix(p, "/") {
		return false
	}
	if len(p) == 1 {
		return false
	}
	parts := strings.SplitN(p, "/", -1)
	for c, part := range parts {
		if c == 0 {
			continue
		}
		if len(part) == 0 {
			return false
		}
		ch := part[0]
		if ch < 65 {
			return false
		}
		if ch > 122 {
			return false
		}
		if strings.Trim(part, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz_1234567890") != "" {
			return false
		}
	}
	return true
}

// ValidProjectRelease checks if a string can be a valid project within a release
func ValidProjectRelease(p string, r qserver.Release) (err error) {
	if !ValidProjectString(p) {
		err = &qerror.QError{
			Ref:     []string{"project.valid.id"},
			Version: r.String(),
			Project: p,
			Msg:     []string{"Project invalid id"},
		}
		return
	}
	return nil
}

// GetProject find the parentproject of a path
func GetProject(v string, s string, ronly bool) (project *Project) {
	version, _ := qserver.Release{}.New(v, ronly)
	fs := version.FS()
	s = qutil.Canon(s)
	r := version.String()
	pid := r + " " + s
	if ronly {
		pid = r + " " + s + " R"
	}
	proj, _ := projectCache.Load(pid)
	if proj != nil {
		return proj.(*Project)
	}

	p1, _ := qutil.QPartition(s)
	pid = r + " " + p1
	if ronly {
		pid = r + " " + p1 + " R"
	}
	proj, _ = projectCache.Load(pid)
	if proj != nil {
		return proj.(*Project)
	}

	parts := strings.SplitN(s, "/", -1)
	if len(parts) == 0 {
		return nil
	}
	start := ""
	notconfig := make([]string, 0)
	for _, part := range parts {
		if part == "" {
			continue
		}
		start += "/" + part
		cfg := start + "/brocade.json"
		if len(notconfig) > 0 {
			found := false
			for _, nc := range notconfig {
				if nc == cfg {
					found = true
					break
				}
			}
			if found {
				continue
			}
		}
		if exists, _ := fs.Exists(cfg); exists {

			proj, _ := Project{}.New(v, start, ronly)
			config, err := proj.LoadConfig()
			if err != nil {
				return nil
			}
			notconfig = config.NotConfig
			project = proj
		}
	}
	return project
}

// Sequence produces a list of projects to which p belongs
func Sequence(v string, p string, ronly bool) (projects []Project, err error) {
	projects = make([]Project, 0)
	p = qutil.Canon(p)
	version, _ := qserver.Release{}.New(v, ronly)
	fs := version.FS()
	parts := strings.SplitN(p, "/", -1)
	start := ""
	notconfig := make([]string, 0)
	for _, part := range parts {
		if part == "" {
			continue
		}
		start += "/" + part
		if !ValidProjectString(start) {
			break
		}
		cfg := start + "/brocade.json"
		if len(notconfig) > 0 {
			found := false
			for _, nc := range notconfig {
				if nc == cfg {
					found = true
					break
				}
			}
			if found {
				continue
			}
		}
		if exists, _ := fs.Exists(cfg); exists {
			project, _ := Project{}.New(v, start, ronly)
			projects = append(projects, *project)
			config, err := project.LoadConfig()
			if err != nil {
				return nil, err
			}
			notconfig = config.NotConfig
		}
	}
	return projects, nil
}

func find(slice []string, val string, wildcard bool) int {
	if len(slice) == 0 {
		return -1
	}
	if !wildcard {
		for i, item := range slice {
			if item == val {
				return i
			}
		}
	}
	if wildcard {
		for i, item := range slice {
			if qfnmatch.Match(item, val) {
				return i
			}
		}
	}
	return -1
}

// Sort Sorteer projecten in volgorde van installeerbaarheid
func Sort(projects []Project) []Project {

	ordens := make([]string, len(projects))
	for i, p := range projects {
		ordens[i] = p.Orden()
	}
	less := func(i, j int) bool {
		return ordens[i] < ordens[j]
	}
	sort.SliceStable(projects, less)
	return projects
}
