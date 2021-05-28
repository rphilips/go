package cmd

import (
	"log"
	"strings"

	qclient "brocade.be/qtechng/lib/client"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var sourceInstallCmd = &cobra.Command{
	Use:     "install",
	Short:   "Installs sources in the repository",
	Long:    `Installs sources in the repository according to patterns, nature and contents`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng source install --qpattern=/application/*.m`,
	RunE:    sourceInstall,
	PreRun:  preSourceInstall,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BWP",
		"fill-version":   "yes",
	},
}

func init() {
	sourceInstallCmd.PersistentFlags().StringVar(&Frefname, "refname", "install", "Reference to the installation")
	sourceCmd.AddCommand(sourceInstallCmd)
}

func sourceInstall(cmd *cobra.Command, args []string) error {
	qpaths, result := listTransport(Fcargo)
	qutil.EditList(Flist, Ftransported, qpaths)
	Fmsg = qreport.Report(result, nil, Fjq, Fyaml, Funquote)
	return nil
}

func preSourceInstall(cmd *cobra.Command, args []string) {
	if !Ftransported {
		var err error
		Fcargo, err = fetchData(args, Ffilesinproject, nil, false)
		if err != nil {
			log.Fatal("cmd/source_install/1:\n", err)
		}
	}

	if strings.ContainsRune(QtechType, 'B') || strings.ContainsRune(QtechType, 'P') {
		installData(Fpayload, Fcargo, false, "")
	}

	if Ftransported {
		err := qclient.SendCargo(Fcargo)
		if err != nil {
			log.Fatal("cmd/source_install/2:\n", err)
		}
		cmd.RunE = func(cmd *cobra.Command, args []string) error { return nil }
	}
}
