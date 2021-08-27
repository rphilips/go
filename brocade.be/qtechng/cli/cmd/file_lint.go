package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	qfs "brocade.be/base/fs"
	qparallel "brocade.be/base/parallel"
	qclient "brocade.be/qtechng/lib/client"
	qerror "brocade.be/qtechng/lib/error"
	qofile "brocade.be/qtechng/lib/file/ofile"
	qobject "brocade.be/qtechng/lib/object"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var fileLintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Lint a file",
	Long:  `Command performs linting of one or more files, i.e. it checks their well-formedness`,
	Args:  cobra.MinimumNArgs(0),
	Example: `qtechng file lint cwd=../strings
qtechng file lint --cwd=../strings --remote
qtechng file lint mymfile.d
qtechng file lint /stdlib/strings/mymfile.d --version=5.10`,
	RunE:   fileLint,
	PreRun: func(cmd *cobra.Command, args []string) { preSSH(cmd, nil) },
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BW",
		"fill-version":   "yes",
	},
}

// Frefname is a reference name

func init() {
	fileLintCmd.Flags().BoolVar(&Frecurse, "recurse", false, "Recursively walk through directory and subdirectories")
	fileLintCmd.Flags().BoolVar(&Fforce, "force", false, "Lint even if the file is not in repository")
	fileLintCmd.Flags().StringArrayVar(&Fqpattern, "qpattern", []string{}, "Posix glob pattern (multiple) on qpath")
	fileLintCmd.Flags().StringVar(&Frefname, "refname", "", "Reference name instead of actual filename")
	fileCmd.AddCommand(fileLintCmd)
}

func fileLint(cmd *cobra.Command, args []string) error {
	errlist := make([]error, 0)

	plocfils, elist := qclient.Find(Fcwd, args, Fversion, Frecurse, Fqpattern, false, Finlist, Fnotinlist, nil)
	if elist != nil {
		errlist = append(errlist, elist)
		return qerror.ErrorSlice(errlist)
	}
	files := make([]string, 0)
	for _, plocfil := range plocfils {
		files = append(files, plocfil.Place)
	}
	if len(files) == 0 {
		for _, arg := range args {
			arg = qutil.AbsPath(arg, Fcwd)
			if qfs.Exists(arg) {
				files = append(files, arg)
			}
		}
	}
	if len(files) == 0 {
		return nil
	}

	lint := func(n int) (interface{}, error) {
		fname := files[n]
		refname := fname
		if Frefname != "" {
			refname = Frefname
		}
		ext := filepath.Ext(fname)
		blob, err := os.ReadFile(fname)
		if err != nil {
			e := &qerror.QError{
				Ref:    []string{"file.lint.read"},
				File:   refname,
				Lineno: 1,
				Type:   "Error",
				Msg:    []string{err.Error()},
			}
			return false, e
		}
		// check utf8
		_, result, e := qutil.NoUTF8(bytes.NewReader(blob))
		if e != nil || len(result) > 0 {
			lineno := -1
			if len(result) > 1 {
				lineno = result[0][0]
			}
			err := &qerror.QError{
				Ref:    []string{"file.lint.utf8"},
				File:   refname,
				Lineno: lineno,
				Type:   "Error",
				Msg:    []string{"UTF-8 error in file"},
			}
			return false, err
		}
		// About line
		switch ext {
		case ".b", ".d", ".i", ".l", ".m", ".x":
			about := qutil.About(blob)
			aboutline := qutil.AboutLine(about)
			if ext == ".m" && len(aboutline) < 2 {
				basename := filepath.Base(refname)
				if strings.HasPrefix(basename, "z") || strings.HasPrefix(basename, "w") {
					aboutline = "xx"
				}
			}
			if len(aboutline) < 2 {
				err := &qerror.QError{
					Ref:    []string{"file.lint.about"},
					File:   refname,
					Lineno: -1,
					Type:   "Error",
					Msg:    []string{"`About:` is missing or empty"},
				}
				return false, err
			}
		}

		var objfile qobject.OFile
		switch ext {
		case ".b":
			objfile = new(qofile.BFile)
		case ".d":
			objfile = new(qofile.DFile)
		case ".i":
			objfile = new(qofile.IFile)
		case ".l":
			objfile = new(qofile.LFile)
		case ".x":
			objfile = new(qofile.XFile)
		}
		if objfile != nil {
			objfile.SetEditFile(refname)
			err = qobject.Loads(objfile, blob, true)
			if err != nil {
				return false, err
			}
			errlist := qobject.LintObjects(objfile)
			if errlist != nil {
				return false, errlist
			}
		}
		return true, nil
	}
	_, errorlist := qparallel.NMap(len(files), -1, lint)

	Fmsg = qreport.Report(nil, errorlist, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	return nil
}
