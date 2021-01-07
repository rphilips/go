package install

import (
	"bytes"
	"encoding/json"
	"io"
	"os/exec"
	"path"
	"path/filepath"

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

	projs := make([]*qproject.Project, len(mproj))
	i := 0
	for _, p := range mproj {
		projs[i] = p
		i++
	}

	projs = qproject.Sort(projs)

	// install m-files
	installMfiles(batchid, projs, qsources, msources)

	// install other auto files
	installAutofiles(batchid, projs, qsources, msources)

	return nil
}

// RSync synchronises the version with the development server
func RSync(r string) (err error) {
	return nil
}

func installMfiles(batchid string, projs []*qproject.Project, qsources map[string]*qsource.Source, msources map[string]map[string][]string) (errs []error) {
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
		nature := qps.Natures()
		buf := new(bytes.Buffer)
		if !nature["auto"] {
			return buf, nil
		}
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

func installAutofiles(batchid string, projs []*qproject.Project, qsources map[string]*qsource.Source, msources map[string]map[string][]string) (errs []error) {

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

func installAutosources(batchid string, files []string, qsources map[string]*qsource.Source) (errs []error) {
	mostype := qregistry.Registry["m-os-type"]
	if mostype == "" {
		return errs
	}
	rou := qregistry.Registry["m-import-auto"]
	if rou == "" {
		return errs
	}
	rouparts := make([]string, 0)
	e := json.Unmarshal([]byte(rou), &rouparts)

	if e != nil {
		e := &qerror.QError{
			Ref: []string{"source.install.auto.registry"},
			Msg: []string{"Registry value m-import-auto is not JSON: `" + e.Error() + "`"},
		}
		errs = append(errs, e)
		return
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

	inm := rouparts[0]
	inm, _ = exec.LookPath(inm)

	var cmd *exec.Cmd
	if len(rouparts) == 1 {
		cmd = exec.Command(inm)
	} else {
		cmd = exec.Command(inm, rouparts[1:]...)
	}
	stdin, e := cmd.StdinPipe()
	if e != nil {
		e := &qerror.QError{
			Ref: []string{"source.install.auto.pipe"},
			Msg: []string{"Cannot open pipe to m-import-auto: `" + e.Error() + "`"},
		}
		errs = append(errs, e)
		return
	}
	go func() {
		defer stdin.Close()
		for _, buf := range bufs {
			b := buf.(*bytes.Buffer)
			if b.Len() == 0 {
				continue
			}
			io.Copy(stdin, b)
		}
		io.WriteString(stdin, "\nh\n")
	}()
	_, e = cmd.CombinedOutput()
	if e != nil {
		e := &qerror.QError{
			Ref: []string{"source.install.auto.exec"},
			Msg: []string{"Exec problem with m-import-auto: `" + e.Error() + "`"},
		}
		errs = append(errs, e)
		return
	}

	return
}
