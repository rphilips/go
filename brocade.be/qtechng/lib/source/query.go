package source

import (
	"bufio"
	"bytes"
	"regexp"
	"sort"
	"strings"

	qfnmatch "brocade.be/base/fnmatch"
	qparallel "brocade.be/base/parallel"
	qmeta "brocade.be/qtechng/lib/meta"
	qobject "brocade.be/qtechng/lib/object"
	qproject "brocade.be/qtechng/lib/project"
	qserver "brocade.be/qtechng/lib/server"
	qutil "brocade.be/qtechng/lib/util"
)

// Query zoeken naar bestanden
type Query struct {
	Release        string                                   `json:"release"`
	CmpRelease     string                                   `json:"cmprelease"`
	QDirs          []string                                 `json:"qdirs"`
	Patterns       []string                                 `json:"patterns"`
	Objects        []string                                 `json:"objects"`
	Natures        []string                                 `json:"natures"`
	Cu             []string                                 `json:"cu"`
	Mu             []string                                 `json:"mu"`
	CtBefore       string                                   `json:"ctbefore"`
	CtAfter        string                                   `json:"ctafter"`
	MtBefore       string                                   `json:"mtbefore"`
	MtAfter        string                                   `json:"mtafter"`
	ToLower        bool                                     `json:"tolower"`
	Regexp         bool                                     `json:"regexp"`
	PerLine        bool                                     `json:"perline"`
	SmartCase      bool                                     `json:"smartcase"`
	FilesInProject bool                                     `json:"filesinproject"`
	Contains       []string                                 `json:"contains"`
	Mumps          bool                                     `json:"mumps"`
	Any            [](func(qpath string, blob []byte) bool) `json:"_any"`
	All            [](func(qpath string, blob []byte) bool) `json:"_all"`
	regexp         []*regexp.Regexp
	bmtableb       map[string][][]int
	bmtableg       map[string][]int
	bmtablef       map[string][]int
	harmonised     bool
	natures        map[string]bool
	muid           map[string]bool
	cuid           map[string]bool
	pipe           []func(source *Source) bool
}

// SQuery tussenliggend formaat
type SQuery struct {
	Release        string   `json:"release"`
	CmpRelease     string   `json:"cmprelease"`
	QDirs          []string `json:"qdirs"`
	Patterns       []string `json:"patterns"`
	Objects        []string `json:"objects"`
	Natures        []string `json:"natures"`
	Cu             []string `json:"cu"`
	Mu             []string `json:"mu"`
	CtBefore       string   `json:"ctbefore"`
	CtAfter        string   `json:"ctafter"`
	MtBefore       string   `json:"mtbefore"`
	MtAfter        string   `json:"mtafter"`
	ToLower        bool     `json:"tolower"`
	Regexp         bool     `json:"regexp"`
	PerLine        bool     `json:"perline"`
	SmartCase      bool     `json:"smartcase"`
	FilesInProject bool     `json:"filesinproject"`
	Contains       []string `json:"contains"`
	Mumps          bool     `json:"mumps"`
}

// Copy SQuery simple query
func (squery SQuery) Copy() (query Query) {
	query.Release = squery.Release
	query.CmpRelease = squery.CmpRelease
	query.QDirs = squery.QDirs
	query.Patterns = squery.Patterns
	query.Objects = squery.Objects
	query.Natures = squery.Natures
	query.Cu = squery.Cu
	query.Mu = squery.Mu
	query.CtBefore = squery.CtBefore
	query.CtAfter = squery.CtAfter
	query.MtBefore = squery.MtBefore
	query.MtAfter = squery.MtAfter
	query.ToLower = squery.ToLower
	query.Regexp = squery.Regexp
	query.PerLine = squery.PerLine
	query.SmartCase = squery.SmartCase
	query.Contains = squery.Contains
	query.FilesInProject = squery.FilesInProject
	query.Mumps = squery.Mumps
	return
}

// Harmonise past de query aan
func (query *Query) Harmonise() {
	if query.harmonised {
		return
	}
	query.harmonised = true

	if len(query.Objects) != 0 {
		objects := make([]string, 0)
		for _, obj := range query.Objects {
			if strings.HasPrefix(obj, "l4_") && strings.Count(obj, "_") == 2 {
				parts := strings.SplitN(obj, "_", 3)
				obj = "l4_" + parts[2]
			}
			objects = append(objects, obj)
		}
		query.Objects = objects
	}

	patterns := query.Patterns
	if len(patterns) == 0 {
		patterns = []string{"/*"}
	}
	for i, pattern := range patterns {
		if !strings.HasPrefix(pattern, "/") {
			pattern = "/" + pattern
			patterns[i] = pattern
		}
	}
	query.Patterns = patterns

	release, err := qserver.Release{}.New(query.Release, true)
	if err == nil {
		patterns := query.Patterns
		oks := make([]string, 0)
		fs := release.FS()
		for _, pattern := range patterns {
			k := strings.IndexAny(pattern, "*[?")
			if k != -1 {
				dir := pattern[:k]
				k := strings.LastIndex(dir, "/")
				if k == 0 {
					dir = "/"
				} else {
					dir = pattern[:k]
				}
				ok, _ := fs.DirExists(dir)
				if ok {
					oks = append(oks, pattern)
				}
			} else {

				if ok, _ := fs.IsDir(pattern); ok {
					if strings.HasSuffix(pattern, "/") {
						oks = append(oks, pattern+"*")
					} else {
						oks = append(oks, pattern+"/*")
					}
					continue
				}
				if ok, _ := fs.Exists(pattern); ok {
					oks = append(oks, pattern)
				}
			}
		}
		query.Patterns = oks
		if len(oks) == 0 {
			return
		}
	}

	query.natures = make(map[string]bool)
	for _, n := range query.Natures {
		query.natures[n] = true
	}

	query.cuid = make(map[string]bool)
	for _, n := range query.Cu {
		query.cuid[n] = true
	}

	query.muid = make(map[string]bool)
	for _, n := range query.Mu {
		query.muid[n] = true
	}

	if query.CtBefore != "" {
		query.CtBefore = qutil.Time(query.CtBefore)
	}
	if query.MtBefore != "" {
		query.MtBefore = qutil.Time(query.MtBefore)

	}

	if query.CtAfter != "" {
		query.CtAfter = qutil.Time(query.CtAfter)
	}
	if query.MtAfter != "" {
		query.MtAfter = qutil.Time(query.MtAfter)
	}
	if query.Release != "" {
		query.Release = qserver.Canon(query.Release)
	}
	if query.CmpRelease != "" {
		query.CmpRelease = qserver.Canon(query.CmpRelease)
	}

	if len(query.Contains) < 2 {
		query.PerLine = false
	}
	if query.SmartCase && !query.Regexp && len(query.Contains) != 0 {
		allower := true
		for _, c := range query.Contains {
			if strings.ToLower(c) != c {
				allower = false
				break
			}
		}
		query.ToLower = allower
	}

	contains := make([]string, 0)
	for _, s := range query.Contains {
		if len(s) == 0 {
			continue
		}
		if query.Regexp {
			contains = append(contains, s)
			continue
		}
		if query.ToLower {
			s = strings.ToLower(s)
		}
		contains = append(contains, s)
		bad, good, full := qutil.BMCreateTable([]byte(s))
		if query.bmtableb == nil {
			query.bmtableb = make(map[string][][]int)
		}
		query.bmtableb[s] = bad
		if query.bmtableg == nil {
			query.bmtableg = make(map[string][]int)
		}
		query.bmtableg[s] = good
		if query.bmtablef == nil {
			query.bmtablef = make(map[string][]int)
		}
		query.bmtablef[s] = full
	}
	query.Contains = contains
	if query.Regexp && len(query.Contains) != 0 {
		query.regexp = make([]*regexp.Regexp, len(query.Contains))
		for i := range query.Contains {
			query.regexp[i] = regexp.MustCompile(query.Contains[i])
		}
		query.Contains = nil
	}

	// build the pipeline

	pipe := make([]func(source *Source) bool, 0)

	// release
	if query.Release != "" {
		f := func(source *Source) bool {
			version := source.Release()
			sversion := version.String()
			return sversion == query.Release
		}
		pipe = append(pipe, f)
	}

	// patterns
	f := func(source *Source) bool {
		qpath := source.String()
		ok := false
		for _, pattern := range query.Patterns {
			ok = qfnmatch.Match(pattern, qpath)
			if ok {
				break
			}
		}
		return ok
	}
	pipe = append(pipe, f)

	// Natures
	if len(query.Natures) != 0 {
		f := func(source *Source) bool {
			natures := source.Natures()
			for nature := range natures {
				if query.natures[nature] {
					return true
				}
			}
			return false
		}
		pipe = append(pipe, f)
	}

	// Mumps
	if query.Mumps {
		f := func(source *Source) bool {
			natures := source.Natures()
			if natures["mumps"] {
				return true
			}
			return false
		}
		pipe = append(pipe, f)
	}

	// meta
	if query.CtAfter != "" || query.MtBefore != "" || query.MtAfter != "" || len(query.Cu) != 0 || len(query.Mu) != 0 {
		f := func(source *Source) bool {
			version := source.Release()
			sversion := version.String()
			qpath := source.String()
			meta, err := qmeta.Meta{}.New(sversion, qpath)
			if err != nil {
				return false
			}
			if len(query.Cu) != 0 {
				if !query.cuid[meta.Cu] {
					return false
				}
			}
			if len(query.Mu) != 0 {
				if !query.muid[meta.Cu] {
					return false
				}
			}
			if query.CtBefore != "" && strings.Compare(query.CtBefore, meta.Ct) != 1 {
				return false
			}
			if query.CtAfter != "" && strings.Compare(query.CtAfter, meta.Ct) != -1 {
				return false
			}
			if query.MtBefore != "" && strings.Compare(query.MtBefore, meta.Mt) != 1 {
				return false
			}
			if query.MtAfter != "" && strings.Compare(query.MtAfter, meta.Mt) != -1 {
				return false
			}
			return true
		}
		pipe = append(pipe, f)
	}

	if query.CmpRelease == "" && len(query.Contains) == 0 && len(query.regexp) == 0 && len(query.Any) == 0 && len(query.All) == 0 {
		query.pipe = pipe
		return
	}

	if query.CmpRelease != "" {
		f := func(source *Source) bool {
			s, _ := Source{}.New(source.String(), query.CmpRelease, false)
			blobs, errs := s.Fetch()
			blob, err := source.Fetch()
			if errs != nil && err != nil {
				return false
			}
			if errs != nil || err != nil {
				return true
			}
			if bytes.Compare(blob, blobs) == 0 {
				return false
			}
			return true
		}
		pipe = append(pipe, f)
	}

	if len(query.Contains) != 0 {
		f := func(source *Source) bool {
			blob, err := source.Fetch()
			if err != nil {
				return false
			}
			if query.ToLower {
				blob = bytes.ToLower(blob)
			}
			if !query.PerLine {
				for _, needle := range query.Contains {
					ok := qutil.BMSearch(blob, []byte(needle), query.bmtableb[needle], query.bmtableg[needle], query.bmtablef[needle])
					if !ok {
						return false
					}
				}
				return true
			}
			dlm := byte('\n')
			r := bufio.NewReader(bytes.NewReader(blob))
			ok := false
			for {
				line, err := r.ReadSlice(dlm)
				ok = false
				for _, needle := range query.Contains {
					ok = qutil.BMSearch(line, []byte(needle), query.bmtableb[needle], query.bmtableg[needle], query.bmtablef[needle])
					if !ok {
						break
					}
				}
				if ok {
					return true
				}
				if err != nil {
					break
				}
			}
			return false
		}
		pipe = append(pipe, f)
	}
	if len(query.regexp) != 0 {
		f := func(source *Source) bool {
			blob, err := source.Fetch()
			if err != nil {
				return false
			}
			if query.ToLower {
				blob = bytes.ToLower(blob)
			}
			if !query.PerLine {
				for _, needle := range query.regexp {
					ok := needle.Match(blob)
					if !ok {
						return false
					}
				}
				return true
			}
			dlm := byte('\n')
			r := bufio.NewReader(bytes.NewReader(blob))
			ok := false
			for {
				line, err := r.ReadSlice(dlm)
				ok = false
				for _, needle := range query.regexp {
					ok = needle.Match(line)
					if !ok {
						break
					}
				}
				if ok {
					return true
				}
				if err != nil {
					break
				}
			}
			return false
		}
		pipe = append(pipe, f)
	}

	if len(query.All) != 0 {

		g := func(f func(string, []byte) bool) func(*Source) bool {
			return func(source *Source) bool {
				blob, err := source.Fetch()
				if err != nil {
					return false
				}
				return f(source.String(), blob)
			}
		}
		for i := range query.All {
			pipe = append(pipe, g(query.All[i]))
		}
	}

	if len(query.Any) != 0 {
		f := func(source *Source) bool {
			blob, err := source.Fetch()
			if err != nil {
				return false
			}
			qpath := source.String()
			for _, h := range query.Any {
				if h(qpath, blob) {
					return true
				}
			}
			return false
		}
		pipe = append(pipe, f)
	}

	query.pipe = pipe
	return
}

// Test source against a query
func (source *Source) Test(query *Query) bool {
	query.Harmonise()
	if len(query.Patterns) == 0 {
		return false
	}
	for _, f := range query.pipe {
		if !f(source) {
			return false
		}
	}
	return true
}

// Run search for sources fitting the query
func (query *Query) Run() []*Source {
	query.Harmonise()
	if len(query.Patterns) == 0 {
		return nil
	}
	release, err := qserver.Release{}.New(query.Release, true)
	if err != nil {
		return nil
	}

	patterns := query.Patterns
	starters := make(map[string]bool)
	sures := make([]string, 0)
	fs := release.FS()
	dirs := make([]string, len(query.QDirs))
	for i, qdir := range query.QDirs {
		dirs[i] = qutil.Canon(qdir)
	}

	for _, pattern := range patterns {
		k := strings.IndexAny(pattern, "*[?")
		if k == -1 {
			if len(dirs) == 0 {
				sures = append(sures, pattern)
				continue
			}
			for _, qdir := range dirs {
				if qutil.EMatch(qdir, pattern) {
					sures = append(sures, pattern)
					break
				}
			}
			continue
		}
		starters[pattern[:k]+"*"] = true
	}
	if len(dirs) == 0 {
		startdirs := make(map[string]bool)

		if len(starters) != 0 {
			for starter := range starters {
				k := strings.LastIndex(starter, "/")
				if k > 0 {
					startdirs[starter[:k]] = true
					delete(starters, starter)
				}
			}
		}

		if len(starters) != 0 {
			dirs := fs.Dir("/", false, true)
			for _, dir := range dirs {
				for starter := range starters {
					if qfnmatch.Match(starter, dir) {
						startdirs[dir] = true
					}
				}
			}
		}

		for dir := range startdirs {
			dirs = append(dirs, dir)
		}
	}
	sort.Strings(dirs)
	prev := ""
	nstarters := make([]string, 0)
	for _, start := range dirs {
		if prev == "" {
			prev = start
			nstarters = append(nstarters, start)
			continue
		}
		if strings.HasPrefix(start, prev+"/") {
			continue
		}
		prev = start
		nstarters = append(nstarters, start)
	}

	f1 := func(n int) (interface{}, error) {
		dir := nstarters[n]
		return fs.Glob(dir, patterns, false), nil
	}
	resultlist, _ := qparallel.NMap(len(nstarters), -1, f1)
	qpaths := make([]string, 0)
	for _, p := range resultlist {
		ps := p.([]string)
		qpaths = append(qpaths, ps...)
	}

	f2 := func(n int) (interface{}, error) {
		qpath := qpaths[n]
		source, e := Source{}.New(release.String(), qpath, true)
		if e != nil {
			return false, nil
		}
		return source.Test(query), nil
	}
	found, _ := qparallel.NMap(len(qpaths), -1, f2)

	msures := make(map[string]*Source)
	for _, qpath := range sures {
		source, e := Source{}.New(release.String(), qpath, true)
		if e != nil {
			continue
		}
		if !source.Test(query) {
			continue
		}
		msures[qpath] = source
	}
	sources := make([]*Source, 0)
	r := release.String()
	for i, ok := range found {
		if ok.(bool) {
			source, _ := Source{}.New(r, qpaths[i], true)
			sources = append(sources, source)
			_, ok := msures[source.String()]
			if !ok {
				delete(msures, source.String())
			}
		}
	}
	for _, source := range msures {
		sources = append(sources, source)
	}

	if query.FilesInProject && len(sources) != 0 {
		found := make(map[string]bool)
		keep := make([]string, 0)
		for _, source := range sources {
			keep = append(keep, source.Project().String())
			found[source.String()] = true
		}
		sort.Strings(keep)
		prev := "<>"
		projects := make([]string, 0)
		for _, p := range keep {
			if strings.HasPrefix(p, prev) {
				continue
			}
			projects = append(projects, p)
			prev = p + "/"
		}

		fn := func(n int) (interface{}, error) {
			p := projects[n]
			project, err := qproject.Project{}.New(query.Release, p, true)
			if err != nil {
				return nil, err
			}
			qpaths := project.QPaths(nil, true)
			return qpaths, nil
		}
		result, _ := qparallel.NMap(len(projects), -1, fn)
		for _, iqpaths := range result {
			qpaths := iqpaths.([]string)
			for _, qpath := range qpaths {
				if found[qpath] {
					continue
				}
				found[qpath] = true
				source, _ := Source{}.New(query.Release, qpath, true)
				sources = append(sources, source)
			}
		}
	}
	return sources
}

// RunObject search for sources fitting the query
func (query *Query) RunObject() map[string]*qobject.Uber {
	query.Harmonise()
	if len(query.Objects) == 0 {
		return nil
	}
	r, err := qserver.Release{}.New(query.Release, true)
	if err != nil {
		return nil
	}
	objmap := qobject.InfoObjectList(r.String(), query.Objects)
	length := len(objmap)
	if length == 0 {
		return nil
	}
	ubermap := make(map[string]*qobject.Uber)
	for k, v := range objmap {
		if v != nil {
			ubermap[k] = v.(*qobject.Uber)
		}
	}
	return ubermap
}
