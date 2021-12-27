package cmd

import (
	"sort"
	"strings"

	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
	qreport "brocade.be/qtechng/lib/report"
	qserver "brocade.be/qtechng/lib/server"
	qsource "brocade.be/qtechng/lib/source"
	qsync "brocade.be/qtechng/lib/sync"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var projectInstallCmd = &cobra.Command{
	Use:     "install",
	Short:   "Install projects in the repository",
	Long:    `This command installs projects in the repository according to patterns, nature and contents`,
	Args:    cobra.MinimumNArgs(1),
	Example: `qtechng project install /catalografie/application`,
	RunE:    projectInstall,
	PreRun:  preProjectInstall,
	Annotations: map[string]string{
		"with-qtechtype": "BPW",
		"fill-version":   "yes",
	},
}

func init() {
	projectInstallCmd.PersistentFlags().StringVar(&Frefname, "refname", "", "Reference to the installation")
	projectInstallCmd.Flags().BoolVar(&Fwarnings, "warnings", false, "Include warnings")
	projectCmd.AddCommand(projectInstallCmd)
}

func projectInstall(cmd *cobra.Command, args []string) error {
	current := qserver.Canon(qregistry.Registry["brocade-release"])
	if current == "" {
		err := &qerror.QError{
			Ref: []string{"install.project"},
			Msg: []string{"Registry value `brocade-release` should be a valid release"},
		}
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}
	if Frefname == "" {
		Frefname = "projectinstall-" + qutil.Timestamp(true)
	}

	patterns := make([]string, len(args))

	for i, arg := range args {
		patterns[i] = arg + "/*"
	}

	query := &qsource.Query{
		Release:  current,
		Patterns: patterns,
	}

	sources := query.Run()

	err := qsource.Install(Frefname, sources, Fwarnings, nil)

	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
		return nil
	}
	msg := make(map[string][]string)
	if len(sources) != 0 {
		qpaths := make([]string, len(sources))
		for i, s := range sources {
			qpaths[i] = s.String()
		}
		sort.Strings(qpaths)
		msg["installed"] = qpaths
	}
	Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
	return nil
}

func preProjectInstall(cmd *cobra.Command, args []string) {
	if Frefname == "" {
		Frefname = "projectinstall-" + qutil.Timestamp(true)
	}
	if strings.Contains(QtechType, "P") {
		qsync.Sync("", "", true, false)
	}

	if !strings.ContainsAny(QtechType, "BP") {
		preSSH(cmd, nil)
	}
}
