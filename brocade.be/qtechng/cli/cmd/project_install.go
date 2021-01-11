package cmd

import (
	"log"
	"strings"

	qclient "brocade.be/qtechng/lib/client"
	qerror "brocade.be/qtechng/lib/error"
	"github.com/spf13/cobra"
)

var projectInstallCmd = &cobra.Command{
	Use:     "install",
	Short:   "Installs projects in the repository",
	Long:    `Installs projects in the repository according to patterns, nature and contents`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng project install /catalografie/application`,
	RunE:    projectInstall,
	PreRun:  preProjectInstall,
	Annotations: map[string]string{
		"with-qtechtype": "BWP",
		"fill-version":   "yes",
	},
}

func init() {
	projectCmd.AddCommand(projectInstallCmd)
}

func projectInstall(cmd *cobra.Command, args []string) error {
	result := listTransport(Fcargo)
	Fmsg = qerror.ShowResult(result, Fjq, nil)
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
