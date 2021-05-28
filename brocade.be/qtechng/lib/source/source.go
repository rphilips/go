package source

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	qfnmatch "brocade.be/base/fnmatch"
	qparallel "brocade.be/base/parallel"
	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
	qofile "brocade.be/qtechng/lib/file/ofile"
	qmeta "brocade.be/qtechng/lib/meta"
	qobject "brocade.be/qtechng/lib/object"
	qproject "brocade.be/qtechng/lib/project"
	qserver "brocade.be/qtechng/lib/server"
	qutil "brocade.be/qtechng/lib/util"
)

var sourceCache = new(sync.Map)

// Source models a Brocade source file
type Source struct {
	s       string
	r       *qserver.Release
	project *qproject.Project
	blob    []byte
	natures map[string]bool
}

// New constructs a new Source object
func (Source) New(r string, s string, readonly bool) (source *Source, err error) {
	s = qutil.Canon(s)

	version, err := qserver.Release{}.New(r, readonly)
	if err != nil {
		e := &qerror.QError{
			Ref:     []string{"source.new"},
			Version: r,
			QPath:   s,
			Msg:     []string{"Cannot instantiate version"},
		}
		err = qerror.QErrorTune(err, e)
		return
	}
	r = version.String()
	pid := r + " " + s
	if readonly {
		pid = r + " " + s + " R"
	}
	so, _ := sourceCache.Load(pid)
	if so != nil {
		return so.(*Source), nil
	}
	proj := qproject.GetProject(r, s, readonly)
	if proj == nil {
		err = &qerror.QError{
			Ref:     []string{"source.new.noproject"},
			Version: r,
			QPath:   s,
			Msg:     []string{"Cannot identify project"},
		}
		return
	}
	sou := Source{s, version, proj, nil, nil}

	so, _ = sourceCache.LoadOrStore(pid, &sou)
	if so != nil {
		return so.(*Source), nil
	}
	err = &qerror.QError{
		Ref:     []string{"source.new.cache"},
		Version: r,
		QPath:   s,
		Msg:     []string{"Cannot create a source"},
	}
	return
}

func updateCache(r string, qdir string) {
	prefix := r + " " + qdir + "/"
	f := func(key, value interface{}) bool {
		if strings.HasPrefix(key.(string), prefix) {
			value.(*Source).natures = nil
		}
		return true
	}
	sourceCache.Range(f)
}

// String of a source: its qpath!
func (source Source) String() string {
	return source.s
}

// Path absolute filepath of file
func (source Source) Path() string {
	version := source.Release()
	fs := version.FS()
	x, _ := fs.RealPath(source.String())
	return x
}

// ReadOnly returns true if the release is ReadOnly
func (source Source) ReadOnly() bool {
	return source.project.ReadOnly()
}

// Project returns source Project
func (source Source) Project() *qproject.Project {
	return source.project
}

// Rel returns relative to project
func (source Source) Rel() string {
	proj := source.project.String()
	return source.s[len(proj)+1:]
}

// Release returns pointer to version
func (source Source) Release() *qserver.Release {
	return source.r
}

// Fetch haalt de data op
func (source *Source) Fetch() (content []byte, err error) {
	if source.blob == nil {
		version := source.Release()
		fs := version.FS()
		blob, e := fs.ReadFile(source.String())
		if e != nil {
			err = &qerror.QError{
				Ref:     []string{"source.fetch"},
				Version: version.String(),
				QPath:   source.String(),
				Msg:     []string{"Cannot retrieve data: " + e.Error()},
			}
			return nil, err
		}
		source.blob = blob
	}
	return source.blob, nil
}

// UnlinkObjects removes all objects associated with the file
func (source Source) UnlinkObjects() {
	natures := source.Natures()
	if natures["config"] {
		return
	}
	if natures["objectfile"] {
		return
	}
	blob, err := source.Fetch()
	if err == nil {
		source.StoreObjects(blob, []byte{})
	}
}

// StoreObjects stores a reference to the basename
func (source Source) StoreObjects(before []byte, actual []byte) {
	name := source.String()
	r := source.Release().String()
	qobject.StoreLinks(r, name, before, actual)
}

// Waste removes a source
func (source *Source) Waste() (err error) {
	version := source.Release()
	r := version.String()
	content, err := source.Fetch()
	if err != nil && os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		e := &qerror.QError{
			Ref:     []string{"source.waste"},
			Version: r,
			QPath:   source.String(),
			Msg:     []string{"Cannot remove file"},
		}
		err = qerror.QErrorTune(err, e)
		return
	}
	// hebben we te maken met een configuratiefile ?
	natures := source.Natures()
	s := source.String()
	// Configuration file
	if natures["config"] {
		fs := source.project.FS()
		if len(fs.Dir("/", false, false)) != 1 {
			err = &qerror.QError{
				Ref:     []string{"source.waste.config"},
				Version: r,
				QPath:   s,
				Msg:     []string{"Cannot remove configuration file: there are other files or directories"},
			}
			return
		}
	}

	if natures["objectfile"] {
		var objfile qobject.OFile
		switch {
		case natures["dfile"]:
			objfile = new(qofile.DFile)
		case natures["ifile"]:
			objfile = new(qofile.IFile)
		case natures["lfile"]:
			objfile = new(qofile.LFile)
		}
		objfile.SetEditFile(source.String())
		objfile.SetRelease(r)

		err = qobject.Loads(objfile, content, true)
		if err != nil {
			e := &qerror.QError{
				Ref:     []string{"source.waste.load.object"},
				Version: r,
				QPath:   source.String(),
				Msg:     []string{"Cannot load objects"},
			}
			err = qerror.QErrorTune(err, e)
			return err
		}

		objectlist := objfile.Objects()
		if len(objectlist) != 0 {
			deps := Deleteable(objectlist...)
			for obj, dps := range deps {
				if dps == nil {
					continue
				}
				err = &qerror.QError{
					Ref:     []string{"source.waste.dependent"},
					Version: r,
					QPath:   s,
					Msg:     []string{"Dependent on `" + obj + "`: `" + strings.Join(dps, ", ") + "`"},
				}
				return err
			}
			WasteObjList(objectlist)
		}

	}

	// unlink meta
	qmeta.Unlink(r, s)

	// Unlink Objects
	source.UnlinkObjects()

	// Unlink file
	fs := version.FS()
	fs.Waste(s)

	// Unlink 'unique'
	UniqueUnlink(version, s)

	// Cache
	pid := r + " " + s
	sourceCache.Delete(pid)
	pid = r + " " + s + " R"
	sourceCache.Delete(pid)
	return
}

// Store stores content
func (source *Source) Store(meta qmeta.Meta, data interface{}, reset bool) (nmeta *qmeta.Meta, changed bool, chobjs map[string]bool, err error) {
	version := source.Release()
	// hebben we te maken met een configuratiefile ?
	s := source.String()
	natures := source.Natures()
	// Valid configuration file
	if natures["config"] {
		b, e := qutil.MakeBytes(data)
		if e != nil {
			err = &qerror.QError{
				Ref:     []string{"source.store.config.bytes"},
				Version: version.String(),
				QPath:   source.String(),
				Msg:     []string{"Cannot transform to bytes: `" + e.Error() + "`"},
			}
			return nmeta, false, chobjs, err
		}
		if !qproject.IsValidConfig(b) {
			err = &qerror.QError{
				Ref:     []string{"source.store.config.invalid"},
				Version: version.String(),
				QPath:   source.String(),
				Msg:     []string{"Not a valid configuration file"},
			}
			return nmeta, false, chobjs, err
		}
		cfg := qproject.Config{}
		json.Unmarshal(b, &cfg)
		source.project.UpdateConfig(cfg)
		qdir, _ := qutil.QPartition(s)
		updateCache(version.String(), qdir)
	}

	// unique
	ext := filepath.Ext(s)
	uniques := strings.SplitN(qregistry.Registry["qtechng-unique-ext"], " ", -1)
	unique := true
	for _, unq := range uniques {
		if unq == ext {
			unique = false
			break
		}
	}
	if !unique && IsUnique(version, s) {
		unique = true
	}
	if !unique {
		config, _ := source.project.LoadConfig()
		notuniques := config.NotUnique
		if len(notuniques) > 0 {
			relpath := s[len(source.project.String()):]
			for _, nu := range notuniques {
				if qfnmatch.Match(nu, relpath) {
					unique = true
					break
				}
			}
		}
	}
	if !unique {
		err = &qerror.QError{
			Ref:     []string{"source.store.notunique"},
			Version: version.String(),
			QPath:   source.String(),
			Msg:     []string{"Is not unique"},
		}
		return nmeta, false, chobjs, err
	}

	// meta object

	nmeta, err = qmeta.Meta{}.New(version.String(), source.String())
	if err != nil {
		e := &qerror.QError{
			Ref:     []string{"source.store.meta.new"},
			Version: version.String(),
			QPath:   source.String(),
			Msg:     []string{"Cannot create meta object"},
		}
		err = qerror.QErrorTune(err, e)
		return
	}

	fs := version.FS()
	var after []byte
	var before []byte
	changed = false
	if !reset {
		var e error
		changed, before, after, e = fs.Store(source.String(), data, meta.Digest)
		if e != nil {
			err = &qerror.QError{
				Ref:     []string{"source.store.forbidden"},
				Version: version.String(),
				QPath:   source.String(),
				Msg:     []string{"Cannot store on disk: " + e.Error()},
			}
			return
		}
	} else {
		after = data.([]byte)
		before = []byte{}
		changed = true
	}

	if !changed {
		return
	}
	source.blob = after
	// unique

	UniqueStore(version, s)

	if !natures["objectfile"] {
		source.StoreObjects(before, after)
	}

	// meta
	h := time.Now()
	t := h.Format(time.RFC3339)
	digest := qutil.Digest(after)
	if meta.Mt == "" {
		meta.Mt = t
	}
	if !reset {
		nmeta.Update(meta)
		nmeta, err = nmeta.Store(version.String(), source.String())
		nmeta.Digest = digest
		if err != nil {
			e := &qerror.QError{
				Ref:     []string{"source.store.meta.store"},
				Version: version.String(),
				QPath:   source.String(),
				Msg:     []string{"Cannot store meta object"},
			}
			err = qerror.QErrorTune(err, e)
			return
		}
	}

	if !natures["objectfile"] {
		return
	}

	// create maps

	// objects
	var objfile qobject.OFile
	switch {
	case natures["dfile"]:
		objfile = new(qofile.DFile)
	case natures["ifile"]:
		objfile = new(qofile.IFile)
	case natures["lfile"]:
		objfile = new(qofile.LFile)
	}
	objfile.SetEditFile(source.String())
	objfile.SetRelease(version.String())
	err = qobject.Loads(objfile, after, true)
	if err != nil {
		e := &qerror.QError{
			Ref:     []string{"source.store.load.object"},
			Version: version.String(),
			QPath:   source.String(),
			Msg:     []string{"Cannot load objects"},
		}
		err = qerror.QErrorTune(err, e)
		return
	}

	objectlist := objfile.Objects()

	changedmap, errorlist := qobject.StoreList(objectlist)
	if len(changedmap) == 0 {
		changedmap = nil
	}

	if len(errorlist) == 0 {
		return nmeta, changed, changedmap, nil
	}
	errslice := qerror.NewErrorSlice()
	for _, e := range errorlist {
		if e != nil {
			errslice = append(errslice, e)
		}
	}

	if len(errslice) == 0 {
		return nmeta, changed, changedmap, nil
	}

	return nmeta, changed, changedmap, errslice
}

// Neighbours add all sources from the same project
func (source *Source) Neighbours() []*Source {
	project := source.project
	if project == nil {
		return nil
	}
	qpaths := project.QPaths(nil, false)
	if qpaths == nil {
		return nil
	}
	r := source.Release().String()
	ronly := source.ReadOnly()
	sources := make([]*Source, len(qpaths))
	for i, qpath := range qpaths {
		s, _ := Source{}.New(r, qpath, ronly)
		sources[i] = s
	}
	return sources
}

// ToMumps writes to mumps
func (source *Source) ToMumps(batchid string, buf *bytes.Buffer) {
	qpath := source.String()
	ext := filepath.Ext(qpath)

	switch ext {
	case ".l":
		source.LFileToMumps(batchid, buf)
	case ".b":
		source.BFileToMumps(batchid, buf)
	case ".m":
		source.MFileToMumps(batchid, buf)
	case ".x":
		source.XFileToMumps(batchid, buf)
	}
}

// StoreTree installs a tree of projects
func StoreTree(batchid string, version string, basedir string, fmeta func(string) qmeta.Meta) (results map[string]*qmeta.Meta, errs error) {
	release, err := qserver.Release{}.New(version, true)
	if err != nil {
		return nil, err
	}

	ok, err := release.Exists()
	if !ok && err == nil {
		err := &qerror.QError{
			Ref:     []string{"source.storetree.version.notexists"},
			Version: release.String(),
			Msg:     []string{"Version `" + version + "` does not exists"},
		}
		return nil, err
	}
	if err != nil {
		err := &qerror.QError{
			Ref:     []string{"source.storetree.version.notexists.error"},
			Version: release.String(),
			Msg:     []string{err.Error()},
		}
		return nil, err
	}
	// Look for brocade.json
	configs := make([]string, 0)
	sources := make([]string, 0)

	f := func(fname string, finfo os.FileInfo, err error) error {
		if err != nil {
			return &qerror.QError{
				Ref:     []string{"source.storetree.filewalk"},
				Version: version,
				QPath:   fname,
				Msg:     []string{err.Error()},
			}
		}
		if finfo.IsDir() {
			return nil
		}
		fname, _ = filepath.Rel(basedir, fname)
		sources = append(sources, qutil.Canon(fname))
		if filepath.Base(fname) != "brocade.json" {
			return nil
		}
		if finfo.IsDir() {
			return nil
		}
		configs = append(configs, fname)
		return nil
	}
	err = filepath.Walk(basedir, f)

	if err != nil {
		return nil, err
	}

	// keep only parent projects
	sort.Strings(configs)
	projects := make([]string, 0)
	prev := ""
	for i, config := range configs {
		current := qutil.Canon(filepath.Dir(config))
		if i != 0 && strings.HasPrefix(current, prev+"/") {
			continue
		}
		prev = current
		projects = append(projects, current)
	}

	// create projects
	_, errs = qproject.InitList(version, projects, fmeta)
	if errs != nil {
		return nil, errs
	}
	// create sources

	fdata := func(p string) ([]byte, error) {
		fname := filepath.Join(basedir, filepath.FromSlash(p[1:]))
		blob, err := os.ReadFile(fname)
		if err != nil {
			blob = nil
		}
		return blob, nil
	}

	results, errs = StoreList(batchid, version, sources, false, fmeta, fdata)

	return

}

type storeeffect struct {
	pmeta  *qmeta.Meta
	chobjs map[string]bool
}

// StoreList creates a list of projects.
func StoreList(batchid string, version string, paths []string, reset bool, fmeta func(string) qmeta.Meta, fdata func(string) ([]byte, error)) (results map[string]*qmeta.Meta, errs error) {
	if batchid == "" {
		batchid = "install"
	}

	if len(paths) == 0 {
		return
	}

	release, err := qserver.Release{}.New(version, true)
	if err != nil {
		return results, err
	}

	ok, err := release.Exists()
	if !ok && err == nil {
		err := &qerror.QError{
			Ref:     []string{"source.storelist.version.notexists"},
			Version: release.String(),
			Msg:     []string{"Version `" + version + "` does not exists"},
		}
		return results, err
	}
	if err != nil {
		err := &qerror.QError{
			Ref:     []string{"source.storelist.version.notexists.error"},
			Version: release.String(),
			Msg:     []string{err.Error()},
		}
		return results, err
	}
	results = make(map[string]*qmeta.Meta)
	oresults := make(map[string]map[string]bool)

	// Handle configuration files first

	configs := make([]string, 0)
	notconfigs := make([]string, 0)

	for _, p := range paths {
		if strings.HasSuffix(p, "/brocade.json") {
			configs = append(configs, p)
		} else {
			notconfigs = append(notconfigs, p)
		}

	}

	work := make([]string, 0)

	fn := func(n int) (interface{}, error) {
		p := work[n]
		source, err := Source{}.New(version, p, false)
		if err != nil {
			return nil, err
		}
		var met qmeta.Meta
		blob, e := fdata(p)

		if e != nil {
			return nil, e
		}
		if !reset {
			met = fmeta(p)
		}
		nmeta, _, chobjs, y := source.Store(met, blob, reset)

		return storeeffect{nmeta, chobjs}, y
	}

	if len(configs) > 0 {
		sort.Strings(configs)
		work = append(work, configs...)
		resultlist, errorlist := qparallel.NMap(len(configs), 1, fn)
		for i, r := range resultlist {
			p := configs[i]
			if errorlist[i] != nil {
				results[p] = nil
				oresults[p] = nil
			} else {
				results[p] = r.(storeeffect).pmeta
				oresults[p] = r.(storeeffect).chobjs
			}
		}

		errslice := make([]error, 0)

		for _, e := range errorlist {
			if e == nil {
				continue
			}
			errslice = append(errslice, e)
		}

		if len(errslice) == 0 {
			errs = nil
		} else {
			errs = qerror.ErrorSlice(errslice)
			return
		}

	}

	work = work[:0]
	work = append(work, notconfigs...)
	resultlist, errorlist := qparallel.NMap(len(notconfigs), -1, fn)

	for i, r := range resultlist {
		if r == nil {
			continue
		}
		p := notconfigs[i]
		results[p] = r.(storeeffect).pmeta
		oresults[p] = r.(storeeffect).chobjs
	}

	errslice := make([]error, 0)

	for _, e := range errorlist {
		if e == nil {
			continue
		}
		errslice = append(errslice, e)
	}
	if len(errslice) == 0 {
		errs = nil
	} else {
		errs = qerror.ErrorSlice(errslice)
	}

	// installation

	if !release.IsInstallable() {
		return
	}

	sourcesfound := make(map[string]bool)
	for qp := range results {
		sourcesfound[qp] = true
	}
	objs := make([]string, 0)
	objsfound := make(map[string]bool)
	for qp := range oresults {
		chobjs := oresults[qp]
		for ob := range chobjs {
			if objsfound[ob] {
				continue
			}
			objs = append(objs, ob)
			objsfound[ob] = true
		}
	}
	mqpaths, err := qobject.GetDependenciesDeep(release, objs...)
	if err != nil {
		errs = err
	}

	for _, qpaths := range mqpaths {
		for _, qp := range qpaths {
			if !strings.HasPrefix(qp, "/") {
				continue
			}
			sourcesfound[qp] = true
		}
	}

	sources := make([]*Source, len(sourcesfound))
	i := 0
	for qp := range sourcesfound {
		source, _ := Source{}.New(version, qp, true)
		sources[i] = source
		i++
	}
	e := Install(batchid, sources, true)
	if e != nil {
		errslice = append(errslice, e)
		errs = qerror.ErrorSlice(errslice)
	}

	return
}

// TestForWasteList test of een lijst mag worden geschrapt
func TestForWasteList(version string, paths []string) (err error) {

	fn := func(n int) (interface{}, error) {
		deps := make(map[string]bool)
		p := paths[n]
		source, err := Source{}.New(version, p, true)
		if err != nil {
			return deps, err
		}
		natures := source.Natures()
		if natures["config"] {
			project := source.Project()
			qpaths := project.QPaths(nil, false)
			for _, q := range qpaths {
				deps[q] = true
			}
			return deps, err
		}
		if natures["objectfile"] {
			var objfile qobject.OFile
			switch {
			case natures["dfile"]:
				objfile = new(qofile.DFile)
			case natures["ifile"]:
				objfile = new(qofile.IFile)
			case natures["lfile"]:
				objfile = new(qofile.LFile)
			}
			objfile.SetEditFile(source.String())
			objfile.SetRelease(version)
			blob, err := source.Fetch()
			if err != nil {
				return deps, err
			}
			err = qobject.Loads(objfile, blob, true)
			if err != nil {
				return deps, err
			}
			ds := Deleteable(objfile.Objects()...)
			for _, depos := range ds {
				for _, d := range depos {
					deps[d] = true
				}
			}
		}
		return deps, nil
	}

	resultlist, errorlist := qparallel.NMap(len(paths), -1, fn)

	elist := qerror.NewErrorSlice()
	for _, e := range errorlist {
		if e == nil {
			continue
		}
		elist = append(elist, e)
	}

	if len(elist) != 0 {
		return elist
	}
	deps := make(map[string]bool)

	for _, resu := range resultlist {
		resm := resu.(map[string]bool)
		for d := range resm {
			deps[d] = true
		}
	}

	for _, p := range paths {
		delete(deps, p)
	}

	if len(deps) != 0 {

		touched := make([]string, len(deps))
		i := 0
		for d := range deps {
			touched[i] = d
			i++
		}
		sort.Strings(touched)

		err := &qerror.QError{
			Ref:     []string{"source.wastelist.dependencies"},
			Version: version,
			Msg:     []string{"Dependencies: `" + strings.Join(touched, ", ")},
		}
		return err
	}

	return nil

}

// WasteList unlinks a number of paths
func WasteList(version string, paths []string) (errs error) {

	if len(paths) == 0 {
		return
	}

	// test of de lijst kan worden geschrapt
	errs = TestForWasteList(version, paths)
	if errs != nil {
		return
	}

	release, err := qserver.Release{}.New(version, true)
	if err != nil {
		return err
	}

	ok, err := release.Exists()
	if !ok && err == nil {
		err := &qerror.QError{
			Ref:     []string{"source.wastelist.version.notexists"},
			Version: release.String(),
			Msg:     []string{"Version `" + version + "` does not exists"},
		}
		return err
	}
	if err != nil {
		err := &qerror.QError{
			Ref:     []string{"source.wastelist.version.notexists.error"},
			Version: release.String(),
			Msg:     []string{err.Error()},
		}
		return err
	}

	configs := make([]*Source, 0)
	cjsons := make([]string, 0)
	notconfigs := make([]*Source, 0)

	for _, p := range paths {
		source, _ := Source{}.New(version, p, false)
		natures := source.Natures()
		if natures["config"] {
			cjsons = append(cjsons, p)
		} else {
			notconfigs = append(notconfigs, source)
		}
	}

	sort.Sort(sort.Reverse(sort.StringSlice(cjsons)))

	for _, p := range cjsons {
		source, _ := Source{}.New(version, p, false)
		configs = append(configs, source)
	}

	fn1 := func(n int) (interface{}, error) {
		source := notconfigs[n]
		return false, source.Waste()
	}

	_, errorlist1 := qparallel.NMap(len(notconfigs), -1, fn1)

	for _, source := range configs {
		source.Waste()
	}

	errslice := qerror.NewErrorSlice()

	for _, e := range errorlist1 {
		if e == nil {
			continue
		}
		errslice = append(errslice, e)
	}

	if len(errslice) != 0 {
		errs = errslice
		return
	}

	return nil
}

// FetchList gets a number of paths
func FetchList(version string, paths []string) (bodies [][]byte, metas []*qmeta.Meta, errs error) {

	if len(paths) == 0 {
		return
	}

	release, err := qserver.Release{}.New(version, true)
	if err != nil {
		return nil, nil, err
	}

	ok, err := release.Exists()
	if !ok && err == nil {
		err := &qerror.QError{
			Ref:     []string{"source.fetchlist.version.notexists"},
			Version: release.String(),
			Msg:     []string{"Version `" + version + "` does not exists"},
		}
		return nil, nil, err
	}
	if err != nil {
		err := &qerror.QError{
			Ref:     []string{"source.fetchlist.version.notexists.error"},
			Version: release.String(),
			Msg:     []string{err.Error()},
		}
		return nil, nil, err
	}

	type fetchdata struct {
		content []byte
		pmeta   *qmeta.Meta
	}

	fn := func(n int) (interface{}, error) {
		p := paths[n]
		source, err := Source{}.New(version, p, true)
		if err != nil {
			err := &qerror.QError{
				Ref:     []string{"source.fetchlist.path.nosource"},
				Version: release.String(),
				QPath:   p,
				Msg:     []string{"Path `" + p + "` does not exists"},
			}
			return nil, err
		}
		content, err := source.Fetch()
		if err != nil {
			err := &qerror.QError{
				Ref:     []string{"source.fetchlist.path.noread"},
				Version: release.String(),
				QPath:   p,
				Msg:     []string{"Path `" + p + "` unreadable"},
			}
			return nil, err
		}

		pmeta, err := qmeta.Meta{}.New(version, p)
		if err != nil {
			err := &qerror.QError{
				Ref:     []string{"source.fetchlist.path.nometa"},
				Version: release.String(),
				QPath:   p,
				Msg:     []string{"Math of path `" + p + "` not retrievable"},
			}
			return nil, err
		}
		pmeta.Digest = qutil.Digest(content)
		return fetchdata{content, pmeta}, nil
	}

	result, errorlist := qparallel.NMap(len(paths), -1, fn)
	bodies = make([][]byte, len(result))
	metas = make([]*qmeta.Meta, len(result))

	for i, res := range result {
		if errorlist[i] != nil {
			bodies[i] = nil
			metas[i] = nil
		} else {
			fres := res.(fetchdata)
			bodies[i] = fres.content
			metas[i] = fres.pmeta
		}
	}

	errslice := qerror.NewErrorSlice()

	for _, e := range errorlist {
		if e == nil {
			continue
		}
		errslice = append(errslice, e)
	}

	if len(errslice) != 0 {
		errs = errslice
		return
	}

	return
}
