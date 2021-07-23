package cmd

import (
	"encoding/json"
	"path/filepath"
	"sort"
	"strings"

	qclient "brocade.be/qtechng/lib/client"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var fileRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Checks out QtechNG files",
	Long: `This command retrieves the local files with an appropriate (same version, same qpath)
from the central repository and updates the local version.
` + Mfiles,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng file refresh --qpattern=/catalografie/application/bcawedit.m`,
	RunE:    fileRefresh,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BWP",
	},
}

func init() {
	fileRefreshCmd.Flags().BoolVar(&Frecurse, "recurse", false, "Recursively walks through directory and subdirectories")
	fileRefreshCmd.Flags().BoolVar(&Fonlychanged, "changed", false, "Consider only modified files")
	fileRefreshCmd.Flags().StringSliceVar(&Fqpattern, "qpattern", []string{}, "Posix glob pattern (multiple) on qpath")
	fileCmd.AddCommand(fileRefreshCmd)
}

func fileRefresh(cmd *cobra.Command, args []string) error {
	if len(args) == 0 && len(Fqpattern) == 0 {
		Fqpattern = []string{"*"}
	}
	locfils, errlist := qclient.Find(Fcwd, args, "", Frecurse, Fqpattern, Fonlychanged, Finlist, Fnotinlist, nil)
	if errlist != nil {
		Fmsg = qreport.Report(nil, errlist, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	}
	versions := make(map[string]map[string][]string)

	for _, locfil := range locfils {
		if locfil == nil {
			continue
		}
		v := locfil.Release
		dirs, ok := versions[v]
		if !ok {
			dirs = make(map[string][]string)
			versions[v] = dirs
		}
		place := locfil.Place
		dirname := filepath.Dir(place)
		qpaths, ok := dirs[dirname]
		if !ok {
			qpaths = make([]string, 0)
		}
		versions[v][dirname] = append(qpaths, locfil.QPath)
	}

	if len(versions) == 0 {
		Fmsg = qreport.Report("", errlist, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	}

	result := make([]string, 0)
	errs := make([]error, 0)

	for v := range versions {
		qdirs := versions[v]
		if len(qdirs) == 0 {
			continue
		}
		for qdir := range qdirs {
			qpaths := qdirs[qdir]
			if len(qpaths) == 0 {
				continue
			}
			args := make([]string, len(qpaths)+3)
			args[0] = "source"
			args[1] = "co"
			j := 0
			for i, qpath := range qpaths {
				args[i+2] = qpath
				j = i + 2
			}
			args[j+1] = "--version=" + v
			stdout, _, err := qutil.QtechNG(args, "$..qpath", false, qdir)
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
	}
	sort.Strings(result)
	Fmsg = qreport.Report(result, errs, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	return nil
}
