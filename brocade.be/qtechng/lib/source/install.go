package source

import (
	"bytes"
	"os"
	"path"
	"path/filepath"
	"strings"

	qfs "brocade.be/base/fs"
	qmumps "brocade.be/base/mumps"
	qparallel "brocade.be/base/parallel"
	qpython "brocade.be/base/python"
	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
	qobject "brocade.be/qtechng/lib/object"
	qproject "brocade.be/qtechng/lib/project"
	qserver "brocade.be/qtechng/lib/server"
	qsync "brocade.be/qtechng/lib/sync"
	qutil "brocade.be/qtechng/lib/util"
)

// Install a list of qpaths
// - a file of nature auto is installed
// - a project is installed
// - a 'brocade.json' file causes the project to be installed
// - a file of type 'install.py' or 'release.py ' causes the project to be installed.
func Install(batchid string, sources []*Source, rsync bool) (err error) {
	if len(sources) == 0 {
		return nil
	}
	sr := sources[0].Release().String()
	r := qserver.Canon(qregistry.Registry["brocade-release"])
	if sr != r && sr != "" && sr != "0.00" {
		return nil
	}
	// synchronises if necessary
	if rsync {
		_, _, err = RSync(sr)
		if err != nil {
			return err
		}
	}

	errs := make([]error, 0)
	badproj := make(map[string]bool)
	// Find all projects
	mproj := make(map[string]*qproject.Project)
	msources := make(map[string]map[string][]string)
	qsources := make(map[string]*Source)
	for _, s := range sources {
		qp := s.String()
		if s.Release().String() != r {
			e := &qerror.QError{
				Ref:     []string{"source.install.version"},
				Version: r,
				QPath:   qp,
				Msg:     []string{"Wrong version"},
			}
			errs = append(errs, e)
			continue
		}
		qsources[qp] = s
		p := s.Project()
		ps := p.String()
		if badproj[ps] {
			continue
		}
		err := p.IsInstallable()
		if err != nil {
			badproj[ps] = true
			errs = append(errs, err)
			continue
		}
		mproj[ps] = p

		ext := path.Ext(qp)
		x := msources[ps]
		if x == nil {
			x = make(map[string][]string)
			msources[ps] = x
		}
		fext := x[ext]
		if fext == nil {
			fext = make([]string, 0)
		}
		msources[ps][ext] = append(fext, qp)
	}

	if len(mproj) == 0 {
		return nil
	}

	projs := make([]*qproject.Project, 0)
	for _, p := range mproj {
		e := p.IsInstallable()
		if e != nil {
			errs = append(errs, e)
			continue
		}
		projs = append(projs, p)
	}
	if len(projs) == 0 {
		return
	}

	projs = qproject.Sort(projs)

	// install releases
	e := installReleasefiles(batchid, projs, qsources, msources)
	if len(e) != 0 {
		errs = append(errs, e...)
	}

	// install m-files
	e = installMfiles(batchid, projs, qsources, msources)
	if len(e) != 0 {
		errs = append(errs, e...)
	}

	// install other auto files
	e = installAutofiles(batchid, projs, qsources, msources)
	if len(e) != 0 {
		errs = append(errs, e...)
	}

	// install projects
	e = installInstallfiles(batchid, projs, qsources, msources)
	if len(e) != 0 {
		errs = append(errs, e...)
	}

	if len(errs) == 0 {
		return nil
	}

	return qerror.ErrorSlice(errs)
}

// RSync synchronises the version with the development server
func RSync(r string) (changed []string, deleted []string, err error) {
	qtechType := qregistry.Registry["qtechng-type"]
	if strings.Contains(qtechType, "B") {
		return nil, nil, nil
	}
	changed, deleted, err = qsync.Sync(r, r, false)
	return changed, deleted, err
}

func installInstallfiles(batchid string, projs []*qproject.Project, qsources map[string]*Source, msources map[string]map[string][]string) (errs []error) {

	tmpdir, e := qfs.TempDir("", "qtechng."+batchid+".")
	if e != nil {
		e := &qerror.QError{
			Ref: []string{"source.install.tmpdir"},
			Msg: []string{"Cannot make temporary directory: " + e.Error()},
		}
		errs = append(errs, e)
		return
	}
	qtos := make(map[string]*Source)
	for q, v := range qsources {
		qtos[q] = v
	}

	for _, proj := range projs {
		r := proj.Release()
		ps := proj.String()
		parts := strings.SplitN(ps, "/", -1)
		parts[0] = tmpdir
		basedir := filepath.Join(parts...)
		qpaths := proj.QPaths(nil, true)
		for _, qp := range qpaths {
			_, ok := qtos[qp]
			if !ok {
				s, e := Source{}.New(r.String(), qp, true)
				if e == nil {
					qtos[qp] = s
				}
			}
		}
		inpy := ps + "/install.py"
		inso, ok := qtos[inpy]
		if !ok {
			continue
		}
		if qfs.IsDir(basedir) {
			continue
		}
		errz := projcopy(proj, qpaths, qtos, tmpdir)
		if len(errz) != 0 {
			errs = append(errs, errz...)
		}

		errz = installInstallsource(basedir, batchid, inso)
		if errs != nil {
			errs = append(errs, errz...)
		}

	}
	return
}

func projcopy(proj *qproject.Project, qpaths []string, qsources map[string]*Source, tmpdir string) []error {
	done := make(map[string]bool)
	where := make(map[string]string)
	for _, qp := range qpaths {
		parts := strings.SplitN(qp, "/", -1)
		if len(parts) == 1 {
			continue
		}
		parts[0] = tmpdir
		subdir := filepath.Join(parts[:len(parts)-1]...)
		where[qp] = filepath.Join(subdir, parts[len(parts)-1])
		_, ok := done[subdir]
		if ok {
			continue
		}
		os.MkdirAll(subdir, 0770)
		done[subdir] = true
	}

	fn := func(n int) (interface{}, error) {
		qp := qpaths[n]
		qps := qsources[qp]

		content, err := qps.Fetch()
		if err != nil {
			return "", err
		}
		env := qps.Env()
		notreplace := qps.NotReplace()
		objectmap := make(map[string]qobject.Object)
		buf := new(bytes.Buffer)
		_, err = ResolveText(env, content, "rilm", notreplace, objectmap, nil, buf, "")
		if err != nil {
			return "", err
		}

		place := where[qp]

		err = qfs.Store(place, buf, "process")
		if err != nil {
			return "", err
		}
		return "", err
	}
	_, errorlist := qparallel.NMap(len(qpaths), -1, fn)

	return errorlist
}

func installInstallsource(tdir string, batchid string, inso *Source) (errs []error) {
	finso := inso.Path()
	py := qutil.GetPy(finso)

	extra := []string{
		"VERSION__='" + inso.Release().String() + "'",
		"PROJECT__='" + inso.Project().String() + "'",
		"QPATH__='" + inso.String() + "'",
		"ID__='" + batchid + "'",
	}
	_, serr := qpython.Run(finso, py == "py3", nil, extra, tdir)
	serr = strings.TrimSpace(serr)
	serr = string(qutil.Ignore([]byte(serr)))
	return
}

func installReleasefiles(batchid string, projs []*qproject.Project, qsources map[string]*Source, msources map[string]map[string][]string) (errs []error) {
	for _, proj := range projs {
		ps := proj.String()
		repy := ps + "/release.py"
		reso, ok := qsources[repy]
		if !ok {
			continue
		}
		err := installReleasesource(batchid, reso)
		if err != nil {
			errs = append(errs, err...)
		}
	}
	return
}

func installReleasesource(batchid string, reso *Source) (errs []error) {
	freso := reso.Path()
	py := qutil.GetPy(freso)
	tdir := qregistry.Registry["scratch-dir"]

	extra := []string{
		"VERSION__='" + reso.Release().String() + "'",
		"PROJECT__='" + reso.Project().String() + "'",
		"QPATH__='" + reso.String() + "'",
		"ID__='" + batchid + "'",
	}
	_, serr := qpython.Run(freso, py == "py3", nil, extra, tdir)
	serr = strings.TrimSpace(serr)
	serr = string(qutil.Ignore([]byte(serr)))
	return
}

func installMfiles(batchid string, projs []*qproject.Project, qsources map[string]*Source, msources map[string]map[string][]string) (errs []error) {
	mostype := qregistry.Registry["m-os-type"]
	if mostype == "" {
		return errs
	}
	for _, proj := range projs {
		ps := proj.String()
		files := msources[ps][".m"]
		if len(files) == 0 {
			continue
		}
		err := installMsources(batchid, files, qsources)
		if err != nil {
			errs = append(errs, err...)
		}
	}
	return
}

func installMsources(batchid string, files []string, qsources map[string]*Source) (errs []error) {
	roudir := qregistry.Registry["gtm-rou-dir"]
	fn := func(n int) (interface{}, error) {
		qp := files[n]
		qps := qsources[qp]
		nature := qps.Natures()
		buf := new(bytes.Buffer)
		if !nature["auto"] {
			return buf, nil
		}
		qps.MFileToMumps(batchid, buf)
		if roudir != "" {
			_, b := qutil.QPartition(qp)
			target := filepath.Join(roudir, b)
			qfs.Store(target, buf, "process")
		}
		return buf, nil
	}

	if roudir != "" {
		qparallel.NMap(len(files), -1, fn)
	}
	return
}

func installAutofiles(batchid string, projs []*qproject.Project, qsources map[string]*Source, msources map[string]map[string][]string) (errs []error) {

	for _, proj := range projs {
		ps := proj.String()
		files := make([]string, 0)
		for _, ext := range []string{".l", ".x", ".b"} {
			sources := msources[ps][ext]
			if len(sources) > 0 {
				files = append(files, sources...)
			}
		}
		if len(files) == 0 {
			continue
		}
		err := installAutosources(batchid, files, qsources)
		if err != nil {
			errs = append(errs, err...)
		}
	}
	return

}

func installAutosources(batchid string, files []string, qsources map[string]*Source) (errs []error) {
	mostype := qregistry.Registry["m-os-type"]

	if mostype == "" {
		return errs
	}

	fn := func(n int) (interface{}, error) {
		qp := files[n]
		qps := qsources[qp]
		nature := qps.Natures()
		buf := new(bytes.Buffer)
		if !nature["auto"] {
			return buf, nil
		}
		ext := filepath.Ext(qps.String())
		switch ext {
		case ".l":
			qps.LFileToMumps(batchid, buf)
		case ".x":
			qps.XFileToMumps(batchid, buf)
		case ".b":
			qps.BFileToMumps(batchid, buf)
		}
		return buf, nil
	}

	bufs, _ := qparallel.NMap(len(files), -1, fn)

	buffers := make([]*bytes.Buffer, len(bufs))
	for i, b := range bufs {
		buffers[i] = b.(*bytes.Buffer)
	}
	e := qmumps.PipeTo("", buffers)
	if e != nil {
		e := &qerror.QError{
			Ref: []string{"source.install.auto.exec"},
			Msg: []string{"Exec problem with m-import-auto-exe: `" + e.Error() + "`"},
		}
		errs = append(errs, e)
		return
	}

	return
}
