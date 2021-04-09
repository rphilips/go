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
	"github.com/spf13/cobra"
)

var projectInstallCmd = &cobra.Command{
	Use:     "install",
	Short:   "Installs projects in the repository",
	Long:    `Installs projects in the repository according to patterns, nature and contents`,
	Args:    cobra.MinimumNArgs(1),
	Example: `qtechng project install /catalografie/application`,
	RunE:    projectInstall,
	PreRun:  preProjectInstall,
	Annotations: map[string]string{
		"with-qtechtype": "BP",
	},
}

func init() {
	projectInstallCmd.PersistentFlags().StringVar(&Finstallref, "installref", "", "Reference to the installation")
	projectCmd.AddCommand(projectInstallCmd)
}

func projectInstall(cmd *cobra.Command, args []string) error {
	current := qserver.Canon(qregistry.Registry["brocade-release"])
	if current == "" {
		err := &qerror.QError{
			Ref: []string{"install.project"},
			Msg: []string{"Registry value `brocade-release` should be a valid release"},
		}
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml)
		return nil
	}
	if Finstallref == "" {
		Finstallref = "install-" + current
	}

	if !strings.Contains(QtechType, "B") {
		qsync.Sync("", "", true)
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

	err := qsource.Install(Finstallref, sources, false)

	if err != nil {
		Fmsg = qreport.Report(nil, err, Fjq, Fyaml)
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
	Fmsg = qreport.Report(msg, nil, Fjq, Fyaml)
	return nil
}

func preProjectInstall(cmd *cobra.Command, args []string) {
	if !Ftransported {
		var err error
		Fcargo, err = fetchData(args, true, nil, false)
		if err != nil {
			log.Fatal("cmd/project_install/1:\n", err)
		}
	}

	if strings.ContainsRune(QtechType, 'B') || strings.ContainsRune(QtechType, 'P') {
		installData(Fpayload, Fcargo, false, "")
	}

	if Ftransported {
		err := qclient.SendCargo(Fcargo)
		if err != nil {
			log.Fatal("cmd/project_install/2:\n", err)
		}
		cmd.RunE = func(cmd *cobra.Command, args []string) error { return nil }
	}
}
