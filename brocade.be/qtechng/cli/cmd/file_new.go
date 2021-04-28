package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	qfs "brocade.be/base/fs"
	qclient "brocade.be/qtechng/lib/client"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var fileNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Adds files to QtechNG",
	Long:  `Command adds a new/existing file to QtechNG. Version and project is necessary`,
	Args:  cobra.MinimumNArgs(0),
	Example: `qtechng file new application/bcawedit.m install.py --version=5.10 --qdir=/catalografie
qtechng file new application/bcawedit.m install.py cwd=../catalografie
qtechng file new bcawedit.m install.py cwd=../application
qtechng file new bcawedit.m install.py
	`,
	RunE:   fileNew,
	PreRun: func(cmd *cobra.Command, args []string) { preSSH(cmd) },
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BW",
		"fill-version":   "yes",
		"fill-qdir":      "yes",
	},
}

// Fcreate create a new file
var Fcreate bool

// Fhint for new files
var Fhint string

func init() {
	fileNewCmd.Flags().StringVar(&Fqdir, "qdir", "", "Directory the file belongs to in repository")
	fileNewCmd.Flags().BoolVar(&Frecurse, "recurse", false, "Recursively walks through directory and subdirectories")
	fileNewCmd.Flags().StringVar(&Fhint, "hint", "", "Hint for new files")
	fileNewCmd.Flags().BoolVar(&Fcreate, "create", false, "Create a new file")
	fileCmd.AddCommand(fileNewCmd)
}

func fileNew(cmd *cobra.Command, args []string) error {

	if Fcreate {
		Frecurse = false
		for _, fname := range args {
			if !filepath.IsAbs(fname) {
				fname = filepath.Join(Fcwd, fname)
			}
			if qfs.IsFile(fname) {
				err := &qerror.QError{
					Ref:  []string{"file.create.isfile"},
					Type: "Error",
					Msg:  []string{fmt.Sprintf("File `%s` exists already", fname)},
				}
				Fmsg = qreport.Report(nil, err, Fjq, Fyaml)
				return nil
			}
			if qfs.IsDir(fname) {
				err := &qerror.QError{
					Ref:  []string{"file.create.isdir"},
					Type: "Error",
					Msg:  []string{fmt.Sprintf("`%s` is the name of a directory", fname)},
				}
				Fmsg = qreport.Report(nil, err, Fjq, Fyaml)
				return nil
			}
			dirname := filepath.Dir(fname)
			if !qfs.IsDir(dirname) {
				err := &qerror.QError{
					Ref:  []string{"file.create.notdir"},
					Type: "Error",
					Msg:  []string{fmt.Sprintf("Directory `%s` does not exist", dirname)},
				}
				Fmsg = qreport.Report(nil, err, Fjq, Fyaml)
				return nil
			}
			e := qutil.FileCreate(fname, Fhint)
			if e != nil {
				err := &qerror.QError{
					Ref:  []string{"file.create.create"},
					Type: "Error",
					Msg:  []string{fmt.Sprintf("Error in creating `%s`: `%s`", fname, e.Error())},
				}
				Fmsg = qreport.Report(nil, err, Fjq, Fyaml)
				return nil
			}

		}

	}

	type adder struct {
		Name    string `json:"arg"`
		Release string `json:"version"`
		QPath   string `json:"qpath"`
		Place   string `json:"file"`
	}
	direxists := make(map[string]bool)
	done := make(map[string]bool)
	if Fversion == "" {
		err := &qerror.QError{
			Ref:  []string{"file.add.version"},
			Type: "Error",
			Msg:  []string{"Do not know how to deduce version"},
		}
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml)
		return nil
	}
	if Fqdir == "" || Fqdir == "/" || Fqdir == "." {
		err := &qerror.QError{
			Ref:  []string{"file.add.qdir"},
			Type: "Error",
			Msg:  []string{"Do not know how to deduce directory in repository"},
		}
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml)
		return nil
	}
	if Fcwd == "" {
		err := &qerror.QError{
			Ref:  []string{"file.add.cwd"},
			Type: "Error",
			Msg:  []string{"Do not know where to place the files"},
		}
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml)
		return nil
	}
	result := make([]adder, 0)
	errorlist := make([]error, 0)
	argums := make([]string, 0)
	if Frecurse {
		if len(args) == 0 {
			args = append(args, Fcwd)
		}
	}
	for _, arg := range args {
		arg = qutil.AbsPath(arg, Fcwd)
		if !qfs.IsDir(arg) {
			argums = append(argums, arg)
			continue
		}
		if !Frecurse {
			err := &qerror.QError{
				Ref:  []string{"file.add.dir"},
				Type: "Error",
				Msg:  []string{"Cannot add a directory: `" + arg + "`"},
			}
			errorlist = append(errorlist, err)
			continue
		}
		a, _ := qfs.Find(arg, nil, true, true, false)
		for _, p := range a {
			argums = append(argums, qutil.AbsPath(p, arg))
		}
	}

	for _, arg := range argums {
		if done[arg] {
			continue
		}
		done[arg] = true

		dir := filepath.Dir(arg)
		if !direxists[dir] {
			direxists[dir] = true
			qfs.Mkdir(dir, "qtech")
		}
		if !qfs.IsFile(arg) {
			e := qfs.Store(arg, "", "qtech")
			if e != nil {
				err := &qerror.QError{
					Ref:  []string{"file.add.create"},
					Type: "Error",
					Msg:  []string{"Cannot create file: `" + arg + "`"},
				}
				errorlist = append(errorlist, err)
				continue
			}
		}
		rel, _ := filepath.Rel(Fcwd, arg)
		rel = filepath.ToSlash(rel)
		if strings.HasPrefix(rel, "./") {
			if rel == "./" {
				rel = ""
			} else {
				rel = rel[2:]
			}
		}
		rel = strings.Trim(rel, "/")
		fmt.Println("rel:", rel)
		if strings.HasPrefix(rel, "..") {
			err := &qerror.QError{
				Ref:  []string{"file.add.noqdir"},
				Type: "Error",
				Msg:  []string{"Cannot determine path: `" + arg + "`"},
			}
			errorlist = append(errorlist, err)
			continue
		}
		d := new(qclient.Dir)
		d.Dir = dir
		locfil := qclient.LocalFile{
			Release: Fversion,
			QPath:   Fqdir + "/" + rel,
		}
		d.Add(locfil)
		result = append(result, adder{arg, Fversion, Fqdir + "/" + rel, arg})

	}

	if len(errorlist) == 0 {
		Fmsg = qreport.Report(result, nil, Fjq, Fyaml)
	} else {
		Fmsg = qreport.Report(result, qerror.ErrorSlice(errorlist), Fjq, Fyaml)
	}
	return nil
}
