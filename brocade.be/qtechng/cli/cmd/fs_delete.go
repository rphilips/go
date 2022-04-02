package cmd

import (
	"errors"
	"fmt"

	qfs "brocade.be/base/fs"
	qparallel "brocade.be/base/parallel"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var fsDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete files",
	Long: `This action allows for deleting of files.

Warning! This command is very powerful and can permanently alter your files.

The arguments are files or directories.
A directory stand for ALL its files.

These argument scan be expanded/restricted by using the flags:

	- The '--recurse' flag walks recursively in the subdirectories of the argument directories.
	- The '--pattern' flag builds a list of acceptable patterns on the basenames
	- The '--utf8only' flag restricts to files with UTF-8 content

There are 3 special flags to add functionality:

	- With the '--ask' flag, you can interactively specify the arguments and flags
	- With the '--confirm' flag, you can inspect the FIRST replacement BEFORE it is
		executed.`,
	Args: cobra.MinimumNArgs(0),
	Example: `qtechng fs delete . --recurse --pattern='*.bak'
qtechng fs delete --ask`,
	RunE: fsDelete,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	fsDeleteCmd.Flags().BoolVar(&Fconfirm, "confirm", false, "Ask the first time for confirmation")
	fsCmd.AddCommand(fsDeleteCmd)
}

func fsDelete(cmd *cobra.Command, args []string) error {
	if Fask {
		askfor := []string{
			"files",
			"recurse:files:" + qutil.UnYes(Frecurse),
			"patterns:files:",
			"utf8only:files:" + qutil.UnYes(Futf8only),
			"confirm:files:" + qutil.UnYes(Fconfirm),
		}
		argums, abort := qutil.AskArgs(askfor, Fcwd)
		if abort {
			Fmsg = qreport.Report(nil, errors.New("command aborted"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-delete-abort")
			return nil
		}
		args = argums["files"].([]string)
		Frecurse = argums["recurse"].(bool)
		Fpattern = argums["patterns"].([]string)
		Futf8only = argums["utf8only"].(bool)
		Fconfirm = argums["confirm"].(bool)
	}

	files := make([]string, 0)
	if len(args) != 0 {
		var err error
		files, err = glob(Fcwd, args, Frecurse, Fpattern, true, false, Futf8only)
		if err != nil {
			Ferrid = "fs-delete-glob"
			return err
		}
	}
	if len(files) == 0 {
		Fmsg = qreport.Report(nil, errors.New("no files found to delete"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-delete-nofiles")
		return nil
	}

	if Fconfirm {
		Fconfirm = false
		prompt := fmt.Sprintf("Delete `%s` ? ", files[0])
		confirm := qutil.Confirm(prompt)
		if !confirm {
			Fmsg = qreport.Report(nil, errors.New("did not pass confirmation"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-delete-confirm")
			return nil
		}
	}

	fn := func(n int) (interface{}, error) {

		src := files[n]
		err := qfs.Rmpath(src)
		return src, err
	}

	errs := make([]error, 0)
	resultlist, errorlist := qparallel.NMap(len(files), -1, fn)
	var changed []string
	for i, src := range resultlist {
		if errorlist[i] != nil {
			e := &qerror.QError{
				Ref:  []string{"fs.delete"},
				File: src.(string),
				Msg:  []string{errorlist[i].Error()},
			}
			errs = append(errs, e)
			continue
		}
		changed = append(changed, src.(string))
	}

	msg := make(map[string][]string)
	msg["deleted"] = changed
	if len(errs) == 0 {
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	} else {
		Fmsg = qreport.Report(msg, qerror.ErrorSlice(errs), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	}
	return nil
}
