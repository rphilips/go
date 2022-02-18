package cmd

import (
	"errors"
	"os"
	"time"

	qparallel "brocade.be/base/parallel"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var fsTouchCmd = &cobra.Command{
	Use:   "touch",
	Short: "Touch files",
	Long: `Touch files
The last modification time of the file(s) is changed to the current moment.

The arguments are files or directories.
A directory stand for ALL its files.

These argument scan be expanded/restricted by using the flags:

	- The '--recurse' flag walks recursively in the subdirectories of the argument directories.
	- The '--pattern' flag builds a list of acceptable patterns on the basenames
	- The '--utf8only' flag restricts to files with UTF-8 content

Some remarks:

	- With the '--ask' flag, you can interactively specify the arguments and flags`,
	Args: cobra.MinimumNArgs(0),
	Example: `qtechng fs touch f1.txt f2.txtt --cwd=../workspace
qtechng fs touch --ask`,
	RunE: fsTouch,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func init() {
	fsCmd.AddCommand(fsTouchCmd)
}

func fsTouch(cmd *cobra.Command, args []string) error {
	if Fask {
		askfor := []string{
			"files",
			"recurse:files:" + qutil.UnYes(Frecurse),
			"patterns:files:",
			"utf8only:files:" + qutil.UnYes(Futf8only),
		}
		argums, abort := qutil.AskArgs(askfor, Fcwd)
		if abort {
			Fmsg = qreport.Report(nil, errors.New("command aborted"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-touch-abort")
			return nil
		}
		args = argums["files"].([]string)
		Frecurse = argums["recurse"].(bool)
		Fpattern = argums["patterns"].([]string)
		Futf8only = argums["utf8only"].(bool)
	}
	files := make([]string, 0)
	if len(args) != 0 {
		var err error
		files, err = glob(Fcwd, args, Frecurse, Fpattern, true, true, Futf8only)
		if err != nil {
			Ferrid = "fs-touch-glob"
			return err
		}
	}
	if len(files) == 0 {
		Fmsg = qreport.Report(nil, errors.New("no matching files found"), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-touch-nofiles")
		return nil
	}
	h := time.Now().Local()
	fn := func(n int) (interface{}, error) {
		src := files[n]
		et := os.Chtimes(src, h, h)
		return src, et
	}

	resultlist, errorlist := qparallel.NMap(len(files), -1, fn)
	var errs []error
	var touched []string
	for i := range resultlist {
		src := files[i]
		if errorlist[i] != nil {
			e := &qerror.QError{
				Ref:  []string{"fs.touch"},
				File: src,
				Msg:  []string{errorlist[i].Error()},
			}
			errs = append(errs, e)
			continue
		}
		touched = append(touched, src)
	}

	msg := make(map[string][]string)
	msg["touched"] = touched
	if len(errs) == 0 {
		Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "fs-touch-touched")
	} else {
		Fmsg = qreport.Report(msg, qerror.ErrorSlice(errs), Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	}
	return nil
}
