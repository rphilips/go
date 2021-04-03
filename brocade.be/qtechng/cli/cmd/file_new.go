package cmd

import (
	"path"
	"path/filepath"
	"strings"

	qfs "brocade.be/base/fs"
	qclient "brocade.be/qtechng/lib/client"
	qerror "brocade.be/qtechng/lib/error"
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

func init() {
	fileNewCmd.Flags().StringVar(&Fqdir, "qdir", "", "Directory the file belongs to in repository")
	fileCmd.AddCommand(fileNewCmd)
}

func fileNew(cmd *cobra.Command, args []string) error {
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
		Fmsg = qerror.ShowResult("", Fjq, err, Fyaml)
		return nil
	}
	if Fqdir == "" || Fqdir == "/" || Fqdir == "." {
		err := &qerror.QError{
			Ref:  []string{"file.add.qdir"},
			Type: "Error",
			Msg:  []string{"Do not know how to deduce directory in repository"},
		}
		Fmsg = qerror.ShowResult("", Fjq, err, Fyaml)
		return nil
	}
	if Fcwd == "" {
		err := &qerror.QError{
			Ref:  []string{"file.add.cwd"},
			Type: "Error",
			Msg:  []string{"Do not know where to place the files"},
		}
		Fmsg = qerror.ShowResult("", Fjq, err, Fyaml)
		return nil
	}
	result := make([]adder, 0)
	errorlist := make([]error, 0)
	for _, arg := range args {
		if done[arg] {
			continue
		}
		done[arg] = true
		place := arg
		if !path.IsAbs(place) {
			place = path.Join(Fcwd, place)
		}

		if qfs.IsDir(place) {
			err := &qerror.QError{
				Ref:  []string{"file.add.dir"},
				Type: "Error",
				Msg:  []string{"Cannot add a directory: `" + place + "`"},
			}
			errorlist = append(errorlist, err)
			continue
		}
		dir := filepath.Dir(place)

		if !direxists[dir] {
			direxists[dir] = true
			qfs.Mkdir(dir, "process")
		}
		if !qfs.IsFile(place) {
			e := qfs.Store(place, "", "")
			if e != nil {
				err := &qerror.QError{
					Ref:  []string{"file.add.create"},
					Type: "Error",
					Msg:  []string{"Cannot create file: `" + place + "`"},
				}
				errorlist = append(errorlist, err)
				continue
			}
		}

		rel, _ := filepath.Rel(Fcwd, place)
		rel = filepath.ToSlash(rel)
		if strings.HasPrefix(rel, "./") {
			if rel == "./" {
				rel = ""
			} else {
				rel = rel[2:]
			}
		}

		rel = strings.Trim(rel, "/")

		d := new(qclient.Dir)
		d.Dir = dir
		locfil := qclient.LocalFile{
			Release: Fversion,
			QPath:   Fqdir + "/" + rel,
		}
		d.Add(locfil)
		result = append(result, adder{arg, Fversion, place, Fqdir + "/" + rel})
	}
	if len(errorlist) == 0 {
		Fmsg = qerror.ShowResult(result, Fjq, nil, Fyaml)
	} else {
		Fmsg = qerror.ShowResult(result, Fjq, qerror.ErrorSlice(errorlist), Fyaml)
	}
	return nil
}
