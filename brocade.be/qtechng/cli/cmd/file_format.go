package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	qfs "brocade.be/base/fs"
	qparallel "brocade.be/base/parallel"
	qclient "brocade.be/qtechng/lib/client"
	qerror "brocade.be/qtechng/lib/error"
	qbfile "brocade.be/qtechng/lib/file/bfile"
	qdfile "brocade.be/qtechng/lib/file/dfile"
	qifile "brocade.be/qtechng/lib/file/ifile"
	qlfile "brocade.be/qtechng/lib/file/lfile"
	qmfile "brocade.be/qtechng/lib/file/mfile"
	qxfile "brocade.be/qtechng/lib/file/xfile"
	qutil "brocade.be/qtechng/lib/util"

	"github.com/spf13/cobra"
)

func init() {
	fileCmd.AddCommand(fileFormatCmd)
	fileFormatCmd.Flags().StringVar(&Fversion, "version", "", "Version to work with")
	fileFormatCmd.Flags().BoolVar(&Frecurse, "recurse", false, "Recursively walk through directory and subdirectories")
	fileFormatCmd.Flags().BoolVar(&Finplace, "inplace", false, "Replaces file")
	fileFormatCmd.Flags().StringArrayVar(&Fqpattern, "qpattern", []string{}, "Posix glob pattern (multiple) on qpath")

}

var fileFormatCmd = &cobra.Command{
	Use:   "format [files]",
	Short: "Format a file",
	Long: `Formats the standard Brocade files, namely those matching "*.[dlixmb]".

With the '--inplace' flag specified, the file content is replaced with the formatted
code.

Without the '--inplace' flag, the formatted output of the first file is printed on stdout.
This is only reliable if the command is executed with one argument.` + Mfiles,
	Args: cobra.MinimumNArgs(0),
	Example: `qtechng file format mymfile.d --inplace
qtechng file format mymfile.d`,
	RunE:   fileFormat,
	PreRun: func(cmd *cobra.Command, args []string) { preSSH(cmd, nil) },
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func fileFormat(cmd *cobra.Command, args []string) error {
	var f func(plocfil *qclient.LocalFile) bool
	if len(Fqpattern) == 0 {
		Fqpattern = []string{"*.[dlixmb]"}
	} else {
		f = func(plocfil *qclient.LocalFile) bool {
			return qutil.EMatch("*.[dlixmb]", plocfil.QPath)
		}
	}
	plocfils, _ := qclient.Find(Fcwd, args, Fversion, Frecurse, Fqpattern, Fonlychanged, Finlist, Fnotinlist, f)
	argums := make([]string, 0)
	for _, plocfil := range plocfils {
		argums = append(argums, plocfil.Place)
	}
	if len(argums) == 0 {
		for _, arg := range args {
			arg = qutil.AbsPath(arg, Fcwd)
			if qfs.Exists(arg) {
				argums = append(argums, arg)
			}
		}
	}

	format := func(n int) (result interface{}, err error) {
		fname := argums[n]
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
			return true, nil
		}

		if Finplace && err == nil {
			qfs.Store(fname, buffer, "qtech")
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
			io.Copy(buffer, f)
		}
		return err == nil, err
	}
	_, errorlist := qparallel.NMap(len(argums), -1, format)
	elist := qerror.FlattenErrors(qerror.ErrorSlice(errorlist))
	if len(elist) == 0 {
		if Flist != "" {
			list := make([]string, 0)
			for _, plocfil := range plocfils {
				list = append(list, plocfil.QPath)
			}
			if len(list) != 0 {
				qutil.EditList(Flist, false, list)
			}
		}
		return nil
	}

	return qerror.ErrorSlice(errorlist)

}
