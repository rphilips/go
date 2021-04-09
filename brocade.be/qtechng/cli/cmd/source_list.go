package cmd

import (
	"log"
	"strings"

	qclient "brocade.be/qtechng/lib/client"
	qreport "brocade.be/qtechng/lib/report"
	"github.com/spf13/cobra"
)

var sourceListCmd = &cobra.Command{
	Use:     "list",
	Short:   "Lists sources in the repository",
	Long:    `Lists sources in the repository according to patterns, nature and contents`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng source list --pattern=/application/*.m`,
	RunE:    sourceList,
	PreRun:  preSourceList,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BWP",
	},
}

func init() {
	sourceCmd.AddCommand(sourceListCmd)
}

func sourceList(cmd *cobra.Command, args []string) error {
	result := listTransport(Fcargo)
	Fmsg = qreport.Report(result, nil, Fjq, Fyaml)
	return nil
}

func preSourceList(cmd *cobra.Command, args []string) {
	if !Ftransported {
		var err error
		Fcargo, err = fetchData(args, Ffilesinproject, nil, false)
		if err != nil {
			log.Fatal("cmd/source_list/1:\n", err)
		}
	}

	if strings.ContainsRune(QtechType, 'B') || strings.ContainsRune(QtechType, 'P') {
		addData(Fpayload, Fcargo, false, "")
	}

	if Ftransported {
		err := qclient.SendCargo(Fcargo)
		if err != nil {
			log.Fatal("cmd/source_list/2:\n", err)
		}
		cmd.RunE = func(cmd *cobra.Command, args []string) error { return nil }
	}
}
