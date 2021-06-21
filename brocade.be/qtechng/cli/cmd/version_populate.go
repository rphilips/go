package cmd

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	qfs "brocade.be/base/fs"
	qclient "brocade.be/qtechng/lib/client"
	qerror "brocade.be/qtechng/lib/error"
	qmeta "brocade.be/qtechng/lib/meta"
	qproject "brocade.be/qtechng/lib/project"
	qreport "brocade.be/qtechng/lib/report"
	qserver "brocade.be/qtechng/lib/server"
	qutil "brocade.be/qtechng/lib/util"
)

var versionPopulateCmd = &cobra.Command{
	Use:     "populate",
	Short:   "Populates version 0.00",
	Long:    `Populates 0.00 with projects and files from the current working directory`,
	Args:    cobra.NoArgs,
	Example: "qtechng version populate",
	RunE:    versionPopulate,
	Annotations: map[string]string{
		"with-qtechtype": "B",
	},
}

func init() {
	versionCmd.AddCommand(versionPopulateCmd)
}

func versionPopulate(cmd *cobra.Command, args []string) error {

	release, _ := qserver.Release{}.New("0.00", true)
	exists, _ := release.Exists()
	if !exists {
		err := &qerror.QError{
			Ref:     []string{"release.populate.notexists"},
			Version: release.String(),
			Msg:     []string{"Release does not exists"},
		}
		return err
	}

	sourcedir := release.FS().Dir("/", false, true)

	if len(sourcedir) != 0 {
		err := &qerror.QError{
			Ref:     []string{"release.populate.populated"},
			Version: release.String(),
			Msg:     []string{"Release is already populated"},
		}
		return err
	}

	Ftreeprefix = "/"
	argums, err := qfs.Find(Fcwd, []string{"brocade.json"}, true, true, false)
	if err != nil {
		return err
	}
	qdirs := make(map[string]string)
	for _, arg := range argums {
		dirname := filepath.Dir(arg)
		for _, p := range []string{".qtech", ".qtechng", ".marked"} {
			qfs.Rmpath(filepath.Join(dirname, p))
		}
		rel, _ := filepath.Rel(Fcwd, dirname)
		rel = filepath.Clean(rel)
		rel = filepath.ToSlash(rel)
		rel = strings.TrimPrefix(rel, "./")
		rel = strings.TrimSuffix(rel, "/.")
		rel = strings.TrimRight(rel, "/")
		rel = strings.TrimLeft(rel, "/")
		qproject := Ftreeprefix + rel
		args = append(args, qproject)
		qdirs[dirname] = qproject
	}
	meta := qmeta.Meta{
		Mu: FUID,
	}
	Fversion = "0.00"
	result, errs := qproject.InitList(Fversion, args, func(a string) qmeta.Meta { return meta })
	if errs != nil {
		Fmsg = qreport.Report(result, errs, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}

	files, err := qfs.Find(Fcwd, nil, true, true, false)
	if err != nil {
		return err
	}

	qpaths := make(map[string][]string)

	for _, fname := range files {
		dirname := fname
		qdir := ""
		for qdir == "" {
			dirname = filepath.Dir(dirname)
			if dirname == Fcwd {
				break
			}
			qproj := qdirs[dirname]
			if qproj == "" {
				continue
			}
			qdir = dirname
		}
		if qdir == "" {
			continue
		}
		rel, _ := filepath.Rel(qdir, fname)
		rel = filepath.Dir(rel)
		rel = filepath.Clean(rel)
		rel = filepath.ToSlash(rel)
		rel = strings.TrimPrefix(rel, "./")
		rel = strings.TrimSuffix(rel, "/.")
		rel = strings.TrimRight(rel, "/")
		rel = strings.TrimLeft(rel, "/")
		qdir = qdirs[qdir] + "/" + rel
		qdir = strings.TrimSuffix(qdir, "/.")
		qdir = strings.TrimRight(qdir, "/")
		if qdir == "" {
			qdir = "/"
		}

		qps, ok := qpaths[qdir]
		if !ok {
			qps = make([]string, 0)
			qpaths[qdir] = qps
		}
		qpaths[qdir] = append(qps, fname)
	}

	paths := make([]string, 0)

	for qdir := range qpaths {
		paths = append(paths, qdir)
	}
	sort.Strings(paths)

	for _, qp := range paths {
		qdir := qp
		files := qpaths[qdir]
		locfils := make([]qclient.LocalFile, len(files))
		for k, file := range files {
			locfils[k] = qclient.LocalFile{
				Release: "0.00",
				QPath:   qdir + "/" + filepath.Base(file),
			}
		}
		d := new(qclient.Dir)
		dir := filepath.Dir(files[0])
		d.Dir = dir
		d.Add(locfils...)
		qutil.QtechNG([]string{"file", "ci", "--version=0.00", "--quiet"}, "", false, dir)
	}
	return nil
}
