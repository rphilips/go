package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	qfs "brocade.be/base/fs"
	qclient "brocade.be/qtechng/lib/client"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

//Ftell tells what kind of informatiom has to be returned
var fileDiffCmd = &cobra.Command{
	Use:     "diff",
	Short:   "Gives difference between a file and its source",
	Long:    `Gives information about files (to be used in shell scripts)`,
	Example: `  qtechng file diff bcawedit.m`,
	Args:    cobra.ExactArgs(1),
	RunE:    fileDiff,

	Annotations: map[string]string{
		"remote-allowed": "no",
		"fill-version":   "yes",
		"with-qtechtype": "BWP",
	},
}

func init() {
	fileDiffCmd.Flags().StringVar(&Fversion, "version", "", "Version to work with")
	fileCmd.AddCommand(fileDiffCmd)
}

func fileDiff(cmd *cobra.Command, args []string) error {

	plocfils, _ := qclient.Find(Fcwd, args, "", Frecurse, Fqpattern, false)

	if len(plocfils) == 0 {
		err := qerror.QError{
			Ref: []string{"diff.args.0"},
			Msg: []string{"Cannot find a file with these specifications"},
		}
		return err
	}

	if len(plocfils) != 1 {
		err := qerror.QError{
			Ref: []string{"diff.args.2"},
			Msg: []string{"Too many files found: need exactly 1"},
		}
		return err
	}

	locfil := plocfils[0]
	if locfil == nil {
		return nil
	}

	argums := []string{"source", "co", locfil.QPath, "--version=" + Fversion}
	fname, _ := qfs.AbsPath(locfil.Place)
	basename := filepath.Base(fname)
	tmpdir, _ := qfs.TempDir("", "diff-")
	target := filepath.Join(tmpdir, basename)
	_, _, err := qutil.QtechNG(argums, "", false, tmpdir)
	if err != nil {
		Fmsg = qreport.Report("", err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}
	name := ""
	if qfs.IsFile(target) {
		ext := filepath.Ext(basename)
		name = strings.TrimSuffix(fname, ext) + "-" + Fversion + ext
		qfs.CopyFile(target, name, "qtech", false)
		fmt.Println(fname, name)
		qfs.Rmpath(tmpdir)
	} else {
		Fmsg = qreport.Report("", fmt.Errorf("cannot retrieve `%s`", locfil.QPath), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}

	diff, _ := qutil.Patch(fname, name)

	if Fstdout == "" {
		fmt.Println(diff)
		return nil
	}
	f, err := os.Create(Fstdout)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	fmt.Fprintln(f, diff)
	return nil
}
