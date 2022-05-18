package source

import (
	"bytes"
	"errors"
	"fmt"
	"log"
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

func Install(batchid string, sources []*Source, warnings bool, logme *log.Logger) (err error) {
	if len(sources) == 0 {
		return nil
	}
	qtechType := qregistry.Registry["qtechng-type"]
	sr := sources[0].Release().String()

	r := ""
	if strings.ContainsRune(qtechType, 'B') {
		r = "0.00"
	}
	if strings.ContainsRune(qtechType, 'P') {
		r = qserver.Canon(qregistry.Registry["brocade-release"])
	}

	if r != sr {
		return nil
	}

	errs := make([]error, 0)
	badproj := make(map[string]bool)
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

	if len(errs) != 0 {
		return qerror.ErrorSlice(errs)
	}

	projs := make([]*qproject.Project, len(mproj))
	i := 0
	for _, p := range mproj {
		projs[i] = p
		i++
	}

	// sort project in sequence of installation

	projs = qproject.Sort(projs)

	if logme != nil {
		logme.Printf("Projects sorted: %d projects\n", len(projs))
	}

	// install releases
	count, e := installReleasefiles(batchid, projs, qsources, msources)
	if len(e) != 0 {
		errs = append(errs, e...)
	}

	if logme != nil {
		logme.Printf("release.py's executed: %d release.py\n", count)
	}

	// install m-files
	mfiles, e := installMfiles(batchid, projs, qsources, msources)
	if len(e) != 0 {
		errs = append(errs, e...)
	}
	if logme != nil {
		logme.Printf("*.m files installed: %d m-files\n", len(mfiles))
	}

	// install other auto files
	ofiles, e := installAutofiles(batchid, projs, qsources, msources)
	if len(e) != 0 {
		errs = append(errs, e...)
	}
	if logme != nil {
		logme.Printf("*.[blx] files installed: %d files\n", len(ofiles))
	}

	// install projects
	zfiles, count, e := installInstallfiles(batchid, projs, qsources, msources, logme)
	if len(e) != 0 {
		errs = append(errs, e...)
	}
	if logme != nil {
		logme.Printf("install.py's executed: %d install.py\n", count)
	}

	allfiles := append(mfiles, ofiles...)
	allfiles = append(allfiles, zfiles...)
	allfiles = qutil.Uniqify(allfiles)
	if len(allfiles) > 0 && strings.ContainsRune(qregistry.Registry["qtechng-type"], 'B') {
		infos, _, lerrs := LintList(r, allfiles, warnings)
		if lerrs != nil {
			errs = append(errs, lerrs)
		}
		for _, info := range infos {
			if info == nil {
				continue
			}
			if info.Error() == "OK" {
				continue
			}
			errs = append(errs, info)
		}
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
	changed, deleted, err = qsync.Sync(r, r, false, false, false)
	return changed, deleted, err
}

func installInstallfiles(batchid string, projs []*qproject.Project, qsources map[string]*Source, msources map[string]map[string][]string, logme *log.Logger) (installed []string, count int, errs []error) {

	tmpdir, e := qfs.TempDir("", batchid+".")
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
		count++
		errz, projplace := projcopy(proj, qpaths, qtos, tmpdir)
		if len(errz) != 0 {
			errs = append(errs, errz...)
			continue
		}
		err := installInstallsource(projplace, batchid, inso)
		if err != nil {
			errs = append(errs, err)
			if qfs.IsDir(projplace) {
				qfs.Store(filepath.Join(projplace, "__error__"), err.Error(), "qtech")
			}
			if logme != nil {
				logme.Printf("Error in installing `%s`\n", inso)
				logme.Printf("    see: %s\n", filepath.Join(projplace, "__error__"))
			}
		} else {
			if logme != nil {
				logme.Printf("Successfully installed `%s`\n", ps)
			}
			qfs.RmpathUntil(projplace, tmpdir)
			for _, q := range qsources {
				if q.Project().String() == ps {
					installed = append(installed, q.String())
				}
			}
		}

	}
	if len(errs) == 0 {
		errs = nil
		qfs.Rmpath(tmpdir)
	}
	return
}

func projcopy(proj *qproject.Project, qpaths []string, qsources map[string]*Source, tmpdir string) ([]error, string) {
	projparts := strings.SplitN(proj.String(), "/", -1)
	if len(projparts) == 1 {
		return nil, ""
	}
	projparts[0] = tmpdir
	projplace := filepath.Join(projparts...)
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

		objectmap := make(map[string]qobject.Object)
		buf := new(bytes.Buffer)
		decomment := qps.Natures()["auto"]
		err := qps.Resolve("rilm", objectmap, nil, buf, decomment)
		if err != nil {
			return "", err
		}

		place := where[qp]

		err = qfs.Store(place, buf, "temp")
		return "", err
	}
	_, errorlist := qparallel.NMap(len(qpaths), -1, fn)

	errs := make([]error, 0)
	for _, e := range errorlist {
		if e == nil {
			continue
		}
		errs = append(errs, e)
	}

	return errs, projplace
}

func installInstallsource(tdir string, batchid string, inso *Source) (err error) {
	r := inso.Release()
	fs, place := r.SourcePlace(inso.String())
	place, err = fs.RealPath(place)
	if err != nil {
		return
	}
	py := qutil.GetPy(place, filepath.Dir(place))
	finso := filepath.Join(tdir, "install.py")

	extra := []string{
		"VERSION__='" + inso.Release().String() + "'",
		"PROJECT__='" + inso.Project().String() + "'",
		"QPATH__='" + inso.String() + "'",
		"ID__='" + batchid + "'",
	}
	sout, serr := qpython.Run(finso, py == "py3", nil, extra, tdir)
	sout = string(qutil.Ignore([]byte(sout)))
	sout = strings.TrimSpace(sout)
	serr = string(qutil.Ignore([]byte(serr)))
	serr = strings.TrimSpace(serr)
	if serr == "" && sout == "" {

		return nil
	}
	errmsg := ""
	if sout != "" {
		errmsg = inso.String() + " > stdout:\n" + sout
	}

	if serr != "" {
		if errmsg != "" {
			errmsg += "\n\n"
		}
		errmsg += inso.String() + " > stderr:\n" + serr
	}
	return errors.New(errmsg)
}

func installReleasefiles(batchid string, projs []*qproject.Project, qsources map[string]*Source, msources map[string]map[string][]string) (count int, errs []error) {
	for _, proj := range projs {
		ps := proj.String()
		repy := ps + "/release.py"
		reso, ok := qsources[repy]
		if !ok {
			continue
		}
		err := installReleasesource(batchid, reso)
		if err != nil {
			e := &qerror.QError{
				Ref:     []string{"install.release"},
				Version: proj.Release().String(),
				QPath:   repy,
				Msg:     []string{err.Error()},
			}
			errs = append(errs, e)
		}
		count++
	}
	return
}

func installReleasesource(batchid string, reso *Source) (err error) {
	freso := reso.Path()
	py := qutil.GetPy(freso, filepath.Dir(freso))
	tdir := qregistry.Registry["scratch-dir"]

	extra := []string{
		"VERSION__='" + reso.Release().String() + "'",
		"PROJECT__='" + reso.Project().String() + "'",
		"QPATH__='" + reso.String() + "'",
		"ID__='" + batchid + "'",
	}
	sout, serr := qpython.Run(freso, py == "py3", nil, extra, tdir)

	sout = string(qutil.Ignore([]byte(sout)))
	sout = strings.TrimSpace(sout)
	serr = string(qutil.Ignore([]byte(serr)))
	serr = strings.TrimSpace(serr)
	if serr == "" && sout == "" {
		return nil
	}
	errmsg := ""
	if sout != "" {
		errmsg = reso.String() + " > stdout:\n" + sout
	}

	if serr != "" {
		if errmsg != "" {
			errmsg += "\n\n"
		}
		errmsg += reso.String() + " > stderr:\n" + serr
	}
	return errors.New(errmsg)
}

func installMfiles(batchid string, projs []*qproject.Project, qsources map[string]*Source, msources map[string]map[string][]string) (mfiles []string, errs []error) {
	mostype := qregistry.Registry["m-os-type"]
	if mostype == "" {
		return nil, nil
	}
	for _, proj := range projs {
		ps := proj.String()
		files := msources[ps][".m"]
		if len(files) == 0 {
			continue
		}
		ofiles, err := installMsources(batchid, files, qsources)
		if err != nil {
			errs = append(errs, err...)
		}
		if len(ofiles) > 0 {
			mfiles = append(mfiles, ofiles...)
		}
	}
	if len(mfiles) == 0 {
		mfiles = nil
	}
	return
}

func installMsources(batchid string, files []string, qsources map[string]*Source) (installed []string, errs []error) {
	roudir := qregistry.Registry["gtm-rou-dir"]
	fn := func(n int) (interface{}, error) {
		qp := files[n]
		qps := qsources[qp]
		nature := qps.Natures()
		buf := new(bytes.Buffer)
		if !nature["auto"] {
			return nil, nil
		}
		err := qps.MFileToMumps(batchid, buf)
		if roudir != "" && buf.Len() != 0 {
			_, b := qutil.QPartition(qp)
			target := filepath.Join(roudir, b)
			qfs.Store(target, buf, "process")
		}
		if err != nil {
			switch v := err.(type) {
			case qerror.QError:
				v.QPath = qp
				err = v
			case *qerror.QError:
				v.QPath = qp
				err = v
			default:
				err = v
			}
		}
		return qp, err
	}

	if roudir != "" {
		result, errorlist := qparallel.NMap(len(files), -1, fn)
		for _, r := range result {
			rs := r.(string)
			if rs == "" {
				continue
			}
			installed = append(installed, rs)
		}
		for _, e := range errorlist {
			if e == nil {
				continue
			}
			errs = append(errs, e)
		}
	}
	if len(errs) == 0 {
		errs = nil
	}

	return
}

func installAutofiles(batchid string, projs []*qproject.Project, qsources map[string]*Source, msources map[string]map[string][]string) (zfiles []string, errs []error) {

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
		ofiles, err := installAutosources(batchid, files, qsources)
		if err != nil {
			errs = append(errs, err...)
		}
		if len(ofiles) > 0 {
			zfiles = append(zfiles, ofiles...)
		}
	}
	return

}

func installAutosources(batchid string, files []string, qsources map[string]*Source) (installed []string, errs []error) {
	mostype := qregistry.Registry["m-os-type"]

	if mostype == "" {
		return nil, errs
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
		var err error
		switch ext {
		case ".l":
			err = qps.LFileToMumps(batchid, buf)
		case ".x":
			err = qps.XFileToMumps(batchid, buf)
		case ".b":
			err = qps.BFileToMumps(batchid, buf)
		}
		if err != nil {
			switch v := err.(type) {
			case qerror.QError:
				v.QPath = qp
				err = v
			case *qerror.QError:
				v.QPath = qp
				err = v
			default:
				err = v
			}
		}
		return buf, err
	}

	bufs, errorlist := qparallel.NMap(len(files), -1, fn)

	buffers := make([]*bytes.Buffer, 0)
	for n, r := range errorlist {
		if r == nil {
			installed = append(installed, files[n])
			buffers = append(buffers, bufs[n].(*bytes.Buffer))
		} else {
			errs = append(errs, r)
		}
	}
	if len(buffers) != 0 {
		b := bytes.NewBufferString(fmt.Sprintf(`d %%Run^bqtin("%s")`, batchid))
		buffers = append(buffers, b)
	}
	e := qmumps.PipeTo("", buffers)
	if e != nil {
		e := &qerror.QError{
			Ref: []string{"source.install.auto.exec"},
			Msg: []string{"Exec problem with m-import-auto-exe: `" + e.Error() + "`: " + batchid + ": " + strings.Join(files, ", ")},
		}
		errs = append(errs, e)
		return
	}
	if len(errs) == 0 {
		errs = nil
	}
	return
}
