package cmd

import (
	"log"
	"strings"

	qclient "brocade.be/qtechng/lib/client"
	qreport "brocade.be/qtechng/lib/report"
	"github.com/spf13/cobra"
)

var sourceDeleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Deletes sources in the repository",
	Long:    `Deletes sources in the repository according to patterns, nature and contents`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng source delete --pattern=/application/*.m`,
	RunE:    sourceDelete,
	PreRun:  preSourceDelete,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BW",
	},
}

func init() {
	sourceCmd.AddCommand(sourceDeleteCmd)
}

func sourceDelete(cmd *cobra.Command, args []string) error {
	_, result := listTransport(Fcargo)
	Fmsg = qreport.Report(result, nil, Fjq, Fyaml, Funquote)
	return nil
}

func preSourceDelete(cmd *cobra.Command, args []string) {
	if !Ftransported {
		var err error
		Fcargo, err = fetchData(args, Ffilesinproject, nil, false)
		if err != nil {
			log.Fatal("cmd/source_delete/1:\n", err)
		}
	}

	var errs error = nil

	if strings.ContainsRune(QtechType, 'B') {
		errs = delData(Fpayload, Fcargo)
	}

	if Ftransported {
		if errs == nil {
			err := qclient.SendCargo(Fcargo)
			if err != nil {
				log.Fatal("cmd/source_delete/2:\n", err)
			}
		} else {
			Fmsg = qreport.Report(nil, Fcargo.Error, Fjq, Fyaml, Funquote)
		}
		cmd.RunE = func(cmd *cobra.Command, args []string) error { return nil }
	}

}
