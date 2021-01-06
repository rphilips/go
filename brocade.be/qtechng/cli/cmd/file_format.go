package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	qfs "brocade.be/base/fs"
	qparallel "brocade.be/base/parallel"
	qerror "brocade.be/qtechng/lib/error"
	qbfile "brocade.be/qtechng/lib/file/bfile"
	qdfile "brocade.be/qtechng/lib/file/dfile"
	qifile "brocade.be/qtechng/lib/file/ifile"
	qlfile "brocade.be/qtechng/lib/file/lfile"
	qmfile "brocade.be/qtechng/lib/file/mfile"
	qxfile "brocade.be/qtechng/lib/file/xfile"

	"github.com/spf13/cobra"
)

func init() {
	fileCmd.AddCommand(fileFormatCmd)
	fileFormatCmd.Flags().BoolVar(&Finplace, "inplace", false, "Replaces file")

}

var fileFormatCmd = &cobra.Command{
	Use:   "format files",
	Short: "Formats a file",
	Long: `Command  formats files.

If no arguments are given, all files, matching "*.[dlixmb]" are considered.
The '--recurse' flag walks the tree.
With the '--inplace' modifiers, files are modified 'inplace'.
With only one argument and no '--inplace' flag, the result is written on stdout.
`,
	Args: cobra.MinimumNArgs(0),
	Example: `
  qtechng file format cwd=../strings
  qtechng file format --cwd=../strings --remote
  qtechng file format mymfile.d --inplace`,
	RunE:   fileFormat,
	PreRun: func(cmd *cobra.Command, args []string) { preSSH(cmd) },
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BW",
	},
}

func fileFormat(cmd *cobra.Command, args []string) error {
	files := args
	if len(args) == 0 && strings.ContainsRune(QtechType, 'W') {
		start := "."
		if Fcwd != "" {
			start = Fcwd
		}
		files, _ = qfs.Find(start, "*.[dlixmb]", Frecurse)
	}

	format := func(n int) (result interface{}, err error) {
		fname := files[n]
		ext := filepath.Ext(fname)
		buffer := new(bytes.Buffer)
		switch ext {
		case ".d":
			err = qdfile.Format(fname, nil, buffer)
		case ".m":
			err = qmfile.Format(fname, nil, buffer)
		case ".x":
			err = qxfile.Format(fname, nil, buffer)
		case ".b":
			err = qbfile.Format(fname, nil, buffer)
		case ".i":
			err = qifile.Format(fname, nil, buffer)
		case ".l":
			err = qlfile.Format(fname, nil, buffer)
		default:
			err = &qerror.QError{
				Ref:    []string{"file.format.unknown"},
				File:   fname,
				Lineno: -1,
				Type:   "Error",
				Msg:    []string{"Do not know how to format file"},
			}
			return false, err
		}
		if Finplace && err == nil {
			qfs.Store(fname, buffer, "process")
		}
		if !Finplace && err == nil && n == 0 {
			if Fstdout == "" || Ftransported {
				fmt.Print(buffer.String())
				return true, nil
			}
			f, err := os.Create(Fstdout)
			if err != nil {
				return false, err
			}
			defer f.Close()
			w := bufio.NewWriter(f)
			fmt.Fprint(w, buffer.String())
			err = w.Flush()
			if err != nil {
				return false, err
			}
		}
		return err == nil, err
	}
	_, errorlist := qparallel.NMap(len(files), -1, format)
	elist := qerror.FlattenErrors(qerror.ErrorSlice(errorlist))
	if len(elist) == 0 {
		return nil
	}

	return qerror.ErrorSlice(errorlist)

}
