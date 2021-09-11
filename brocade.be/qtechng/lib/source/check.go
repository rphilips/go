package source

import (
	"errors"
	"log"
	"path/filepath"
	"strings"

	qfs "brocade.be/base/fs"
	qpython "brocade.be/base/python"
	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
	qproject "brocade.be/qtechng/lib/project"
	qserver "brocade.be/qtechng/lib/server"
	qutil "brocade.be/qtechng/lib/util"
)

// Install a list of qpaths
// - a file of nature auto is installed
// - a project is installed
// - a 'brocade.json' file causes the project to be installed
// - a file of type 'install.py' or 'release.py ' causes the project to be installed.

func Check(batchid string, sources []*Source, warnings bool, logme *log.Logger) (err error) {
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
	for _, s := range sources {
		qp := s.String()
		if s.Release().String() != r {
			e := &qerror.QError{
				Ref:     []string{"source.check.version"},
				Version: r,
				QPath:   qp,
				Msg:     []string{"Wrong version"},
			}
			errs = append(errs, e)
			continue
		}
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
	}

	if len(mproj) == 0 {
		return nil
	}

	if len(errs) != 0 {
		return qerror.ErrorSlice(errs)
	}

	projs := make([]*qproject.Project, 0)
	for _, p := range mproj {
		projs = append(projs, p)
	}

	// sort project in sequence of installation

	projs = qproject.Sort(projs)

	if logme != nil {
		logme.Printf("Projects sorted: %d projects\n", len(projs))
	}

	// install projects
	count, e := checkfiles(batchid, projs, logme)
	if len(e) != 0 {
		errs = append(errs, e...)
	}
	if logme != nil {
		logme.Printf("check.py's executed: %d check.py\n", count)
	}

	if len(errs) == 0 {

		return nil
	}

	return qerror.ErrorSlice(errs)
}

func checkfiles(batchid string, projs []*qproject.Project, logme *log.Logger) (count int, errs []error) {

	tmpdir, e := qfs.TempDir("", batchid+".")
	if e != nil {
		e := &qerror.QError{
			Ref: []string{"source.check.tmpdir"},
			Msg: []string{"Cannot make temporary directory: " + e.Error()},
		}
		errs = append(errs, e)
		return
	}

	for _, proj := range projs {
		r := proj.Release().String()
		ps := proj.String()
		parts := strings.SplitN(ps, "/", -1)
		parts[0] = tmpdir
		inpy := ps + "/check.py"
		count++

		qpaths := proj.QPaths(nil, true)
		qtos := make(map[string]*Source)
		for _, qp := range qpaths {
			qtos[qp], _ = Source{}.New(r, qp, true)
		}

		errz, projplace := projcopy(proj, qpaths, qtos, tmpdir)
		if len(errz) != 0 {
			errs = append(errs, errz...)
			continue
		}
		inso := qtos[inpy]
		err := checksource(projplace, batchid, inso)
		if err != nil {
			errs = append(errs, err)
			if qfs.IsDir(projplace) {
				qfs.Store(filepath.Join(projplace, "__checkerror__"), err.Error(), "qtech")
			}
			if logme != nil {
				logme.Printf("Error in checking `%s`\n", inso)
				logme.Printf("    see: %s\n", filepath.Join(projplace, "__checkerror__"))
			}
		} else {
			qfs.RmpathUntil(projplace, tmpdir)
		}

	}
	if len(errs) == 0 {
		errs = nil
		qfs.Rmpath(tmpdir)
	}
	return
}

func checksource(tdir string, batchid string, inso *Source) (err error) {
	r := inso.Release()
	fs, place := r.SourcePlace(inso.String())
	place, err = fs.RealPath(place)
	if err != nil {
		return
	}
	py := qutil.GetPy(place, filepath.Dir(place))
	finso := filepath.Join(tdir, "check.py")

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
	if !strings.Contains(strings.ToUpper(serr), "ERROR") && !strings.Contains(strings.ToUpper(sout), "ERROR") && !strings.Contains(strings.ToUpper(serr), "FAIL") && !strings.Contains(strings.ToUpper(sout), "FAIL") {
		serr = ""
		sout = ""
	}
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
