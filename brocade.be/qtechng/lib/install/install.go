package install

import (
	"bytes"
	"path"

	qfs "brocade.be/base/fs"
	qparallel "brocade.be/base/parallel"
	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
	qproject "brocade.be/qtechng/lib/project"
	qserver "brocade.be/qtechng/lib/server"
	qsource "brocade.be/qtechng/lib/source"
	qutil "brocade.be/qtechng/lib/util"
)

// Install a list of qpaths
// - a file of nature auto is installed
// - a project is installed
// - a 'brocade.json' file causes the project to be installed
// - a file of type 'install.py' or 'release.py ' causes the project to be installed.
func Install(batchid string, sources []*qsource.Source, rsync bool) (err error) {
	r := qserver.Canon("")
	// synchronises if necessary
	if rsync {
		err = RSync(r)
		if err != nil {
			return err
		}
	}
	errs := make([]error, 0)
	// Find all projects
	mproj := make(map[string]*qproject.Project)
	msources := make(map[string]map[string][]string)
	qsources := make(map[string]*qsource.Source)
	for _, s := range sources {
		qp := s.String()
		if s.Release().String() != r {
			e := &qerror.QError{
				Ref:     []string{"source.install.version"},
				Version: r,
				File:    qp,
				Msg:     []string{"Wrong version"},
			}
			errs = append(errs, e)
			continue
		}
		qsources[qp] = s
		p := s.Project()
		ps := p.String()
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

	projs := make([]qproject.Project, len(mproj))
	for _, p := range mproj {
		projs = append(projs, *p)
	}

	projs = qproject.Sort(projs)

	// install m-files
	installMfiles(batchid, projs, qsources, msources)

	return nil
}

// RSync synchronises the version with the development server
func RSync(r string) (err error) {
	return nil
}

func installMfiles(batchid string, projs []qproject.Project, qsources map[string]*qsource.Source, msources map[string]map[string][]string) (errs []error) {
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

func installMsources(batchid string, files []string, qsources map[string]*qsource.Source) (errs []error) {

	roudir := qregistry.Registry["gtm-rou-dir"]
	fn := func(n int) (interface{}, error) {
		qp := files[n]
		qps := qsources[qp]
		buf := new(bytes.Buffer)
		qps.MFileToMumps(batchid, buf)
		if roudir != "" {
			_, b := qutil.QPartition(qp)
			target := path.Join(roudir, b)
			qfs.Store(target, buf, "process")
		}
		return buf, nil
	}

	if roudir != "" {
		qparallel.NMap(len(files), -1, fn)
	}
	return

}
