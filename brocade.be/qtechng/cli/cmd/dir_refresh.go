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
	Use:     "refresh",
	Short:   "Checks out QtechNG directories",
	Long:    `Command to retrieve files from the QtechNG repository`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng dir refresh --qpattern=/catalografie/application/bcawedit.m`,
	RunE:    dirRefresh,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BWP",
	},
}

func init() {
	dirCmd.AddCommand(dirRefreshCmd)
}

func dirRefresh(cmd *cobra.Command, args []string) error {
	if len(args) == 0 && len(Fqpattern) == 0 {
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
		stdout, _, err := qutil.QtechNG(args, "$..qpath", false, dir)
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
