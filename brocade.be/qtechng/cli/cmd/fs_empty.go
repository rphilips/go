package cmd

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	qfs "brocade.be/base/fs"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var fsEmptyCmd = &cobra.Command{
	Use:   "empty",
	Short: "Empty files",
	Long: `This action allows for emptying of directories

Warning! This command is very powerful and can permanently alter your files.

The argument is one directory.

With the '--confirm' flag, you can inspect the FIRST replacement BEFORE it is deleted
With the '--slashes' flag, you can indicate the minimum numbers of slashes in the files
executed.`,
	Args:    cobra.ExactArgs(1),
	Example: `qtechng fs empty /library/tmp --confirm --slashes=3`,
	RunE:    fsEmpty,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}
var Fslashes = 0

func init() {
	fsEmptyCmd.Flags().BoolVar(&Fconfirm, "confirm", false, "Ask the first time for confirmation")
	fsEmptyCmd.Flags().IntVar(&Fslashes, "slashes", 0, "Number of slashes")
	fsCmd.AddCommand(fsEmptyCmd)
}

func fsEmpty(cmd *cobra.Command, args []string) error {
	files := make([]string, 0)
	files1, err := glob(Fcwd, args, true, Fpattern, true, true, false)
	if err != nil {
		Ferrid = "fs-empty-glob"
		return err
	}
	stop := ""
	for _, f := range files1 {
		if Fslashes == 0 {
			files = append(files, f)
			continue
		}
		g := filepath.ToSlash(f)
		if strings.Count(g, "/") < Fslashes {
			stop = f
			break
		}
		files = append(files, f)
	}

	if stop != "" {
		Fmsg = qreport.Report(nil, errors.New("files with not enough slashes (e.g. "+stop+")"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-empty-slashes")
		return nil
	}

	if len(files) == 0 {
		Fmsg = qreport.Report(nil, errors.New("no files found to delete"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-empty-nofiles")
		return nil
	}

	sort.Sort(sort.Reverse(sort.StringSlice(files)))

	if Fconfirm {
		Fconfirm = false
		prompt := fmt.Sprintf("Delete `%s` ? ", files[len(files)-1])
		confirm := qutil.Confirm(prompt)
		if !confirm {
			Fmsg = qreport.Report(nil, errors.New("did not pass confirmation"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-empty-confirm")
			return nil
		}
	}
	errs := make([]error, 0)
	changed := make([]string, 0)
	for _, f := range files {
		err := qfs.Rmpath(f)
		if err != nil {
			e := &qerror.QError{
				Ref:  []string{"fs.empty"},
				File: f,
				Msg:  []string{err.Error()},
			}
			errs = append(errs, e)
			continue
		} else {
			changed = append(changed, f)
		}
	}

	msg := make(map[string][]string)
	msg["deleted"] = changed
	if len(errs) == 0 {
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	} else {
		Fmsg = qreport.Report(msg, qerror.ErrorSlice(errs), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-empty-errors")
	}
	return nil
}
