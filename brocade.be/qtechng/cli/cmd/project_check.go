package cmd

import (
	"log"
	"sort"
	"strings"

	qregistry "brocade.be/base/registry"
	qclient "brocade.be/qtechng/lib/client"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	qserver "brocade.be/qtechng/lib/server"
	qsource "brocade.be/qtechng/lib/source"
	qsync "brocade.be/qtechng/lib/sync"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var projectCheckCmd = &cobra.Command{
	Use:     "check",
	Short:   "Check projects in the repository",
	Long:    `This command checks projects in the repository according to patterns, nature and contents`,
	Args:    cobra.MinimumNArgs(1),
	Example: `qtechng project check /catalografie/application`,
	RunE:    projectCheck,
	PreRun:  preProjectCheck,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BP",
		"fill-version":   "yes",
	},
}

func init() {
	projectCheckCmd.PersistentFlags().StringVar(&Frefname, "refname", "", "Reference to the check")
	projectCmd.AddCommand(projectCheckCmd)
}

func projectCheck(cmd *cobra.Command, args []string) error {
	current := qserver.Canon(qregistry.Registry["brocade-release"])
	if current == "" {
		err := &qerror.QError{
			Ref: []string{"check.project"},
			Msg: []string{"Registry value `brocade-release` should be a valid release"},
		}
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}
	if Frefname == "" {
		Frefname = "projectcheck-" + qutil.Timestamp(true)
	}

	if !strings.Contains(QtechType, "B") {
		qsync.Sync("", "", true)
	}

	patterns := make([]string, len(args))

	for i, arg := range args {
		patterns[i] = arg + "/check.py"
	}

	query := &qsource.Query{
		Release:  current,
		Patterns: patterns,
	}

	sources := query.Run()

	err := qsource.Check(Frefname, sources, Fwarnings, nil)

	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}
	msg := make(map[string][]string)
	if len(sources) != 0 {
		qpaths := make([]string, len(sources))
		for i, s := range sources {
			qpaths[i] = s.String()
		}
		sort.Strings(qpaths)
		msg["checked"] = qpaths
	}
	Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	return nil
}
func preProjectCheck(cmd *cobra.Command, args []string) {
	if !Ftransported {
		var err error
		Fcargo, err = fetchData(args, true, nil, false)
		if err != nil {
			log.Fatal("cmd/project_check/1:\n", err)
		}
	}

	if strings.ContainsRune(QtechType, 'B') || strings.ContainsRune(QtechType, 'P') {
		checkData(Fpayload, Fcargo, false, false, "", nil)
	}

	if Ftransported {
		err := qclient.SendCargo(Fcargo)
		if err != nil {
			log.Fatal("cmd/project_check/2:\n", err)
		}
		cmd.RunE = func(cmd *cobra.Command, args []string) error { return nil }
	}
}
