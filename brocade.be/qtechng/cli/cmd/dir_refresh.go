package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	qfs "brocade.be/base/fs"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var dirRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Check out qtechng directories",
	Long: `Command to refresh the files in directories.

The arguments of the command are the directories to be refreshed.
If no arguments are given, the current working directory is used.

There is a subtle difference between this command and refreshing
the files from a directory!

Refreshing files can only refresh existing files,
refreshing a directory can also bring new files in the
directory and even new subdirectories: checking out from
the repository is always with '--recurse'.

Version and paths are inferred from the available files or the position
of the directory in relation to the 'qtechng-work-dir' registry value.`,
	Args: cobra.MinimumNArgs(0),
	Example: `qtechng dir refresh ../collections/application
qtechng dir refresh`,
	RunE: dirRefresh,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BWP",
	},
}

func init() {
	dirCmd.AddCommand(dirRefreshCmd)
}

func dirRefresh(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		args = []string{Fcwd}
	}

	dirs := make([]string, 0)
	errlist := make([]error, 0)
	for _, arg := range args {
		arg := qutil.AbsPath(arg, Fcwd)
		if qfs.IsDir(arg) {
			dirs = append(dirs, arg)
			continue
		}
		errlist = append(errlist, fmt.Errorf("`%s` is not a directory", arg))

	}

	if len(errlist) != 0 {
		Fmsg = qreport.Report("", errlist, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}
	if len(dirs) == 0 {
		Fmsg = qreport.Report("", errlist, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}

	result := make([]string, 0)
	errs := make([]error, 0)
	for _, dir := range dirs {
		version, qdir := dirProps(dir)
		args := make([]string, 5)
		args[0] = "source"
		args[1] = "co"
		args[2] = "--root"
		args[3] = "--qpattern=" + qdir + "/*"
		args[4] = "--version=" + version
		stdout, _, err := qutil.QtechNG(args, []string{"$..qpath"}, false, dir)
		if err != nil {
			errs = append(errs, err)
		}
		stdout = strings.TrimSpace(stdout)
		if !strings.HasPrefix(stdout, "[") {
			continue
		}
		slice := make([]string, 0)
		e := json.Unmarshal([]byte(stdout), &slice)
		if e != nil {
			continue
		}
		result = append(result, slice...)

	}

	sort.Strings(result)
	Fmsg = qreport.Report(result, errs, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	return nil
}
