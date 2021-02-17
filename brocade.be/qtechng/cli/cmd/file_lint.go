package cmd

import (
	"path/filepath"
	"strings"

	qfs "brocade.be/base/fs"
	qparallel "brocade.be/base/parallel"
	qerror "brocade.be/qtechng/lib/error"
	qofile "brocade.be/qtechng/lib/file/ofile"
	qobject "brocade.be/qtechng/lib/object"
	"github.com/spf13/cobra"
)

var fileLintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Checks a file",
	Long:  `Command checks file file on its well-formedness`,
	Args:  cobra.MinimumNArgs(0),
	Example: `qtechng file lint cwd=../strings
qtechng file lint --cwd=../strings --remote
qtechng file lint mymfile.d
qtechng file lint /stdlib/strings/mymfile.d --remote --version=5.10`,
	RunE:   fileLint,
	PreRun: func(cmd *cobra.Command, args []string) { preSSH(cmd) },
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BW",
	},
}

func init() {
	fileCmd.AddCommand(fileLintCmd)
}

func fileLint(cmd *cobra.Command, args []string) error {
	files := args
	if len(args) == 0 && strings.ContainsRune(QtechType, 'W') {
		start := "."
		if Fcwd != "" {
			start = Fcwd
		}
		files, _ = qfs.Find(start, []string{"*.[dlimbx]"}, Frecurse, true, false)
	}
	lint := func(n int) (interface{}, error) {
		fname := files[n]
		ext := filepath.Ext(fname)
		var err error
		var objfile qobject.OFile
		switch ext {
		case ".d":
			objfile = new(qofile.DFile)
		case ".i":
			objfile = new(qofile.IFile)
		case ".l":
			objfile = new(qofile.LFile)
		}
		if objfile != nil {
			objfile.SetEditFile(fname)
			err = qobject.Lint(objfile, nil, nil)
		}
		return err == nil, err
	}
	_, errorlist := qparallel.NMap(len(files), -1, lint)
	elist := qerror.FlattenErrors(qerror.ErrorSlice(errorlist))
	if len(elist) == 0 {
		return nil
	}
	return qerror.ErrorSlice(errorlist)
}
