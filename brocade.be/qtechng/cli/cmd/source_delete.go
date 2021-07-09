package cmd

import (
	"fmt"

	qreport "brocade.be/qtechng/lib/report"
	"github.com/spf13/cobra"
)

var sourceDeleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Deletes sources in the repository",
	Long:    `Deletes sources in the repository according to patterns, nature and contents`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng source delete --qpattern=/application/*.m`,
	RunE:    sourceDelete,
	PreRun:  func(cmd *cobra.Command, args []string) { preSSH(cmd) },
	Annotations: map[string]string{
		"remote-allowed": "yes",
		"with-qtechtype": "BW",
		"fill-version":   "yes",
	},
}

func init() {
	sourceCmd.AddCommand(sourceDeleteCmd)
}

// func sourceDelete(cmd *cobra.Command, args []string) error {
// 	if Fcargo.Error != nil {
// 		Fmsg = qreport.Report(nil, Fcargo.Error, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
// 	} else {
// 		err, result := listTransport(Fcargo)
// 		Fmsg = qreport.Report(result, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
// 	}
// 	return nil
// }

// func preSourceDelete(cmd *cobra.Command, args []string) {
// 	if !Ftransported {
// 		var err error
// 		Fcargo, err = fetchData(args, Ffilesinproject, nil, false)
// 		if err != nil {
// 			log.Fatal("cmd/source_delete/1:\n", err)
// 		}
// 	}

// 	if strings.ContainsRune(QtechType, 'B') {
// 		delData(Fpayload, Fcargo)
// 	}

// 	if Ftransported {
// 		if Fcargo.Error == nil {
// 			err := qclient.SendCargo(Fcargo)
// 			if err != nil {
// 				log.Fatal("cmd/source_delete/2:\n", err)
// 			}
// 		} else {
// 			Fmsg = qreport.Report(nil, Fcargo.Error, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
// 		}
// 		cmd.RunE = func(cmd *cobra.Command, args []string) error { return nil }
// 	}

// }

func sourceDelete(cmd *cobra.Command, args []string) error {

	squery := buildSQuery(args, Ffilesinproject, nil, false)
	qpaths, errs := delData(squery)
	if qpaths == nil && errs == nil {
		errs = fmt.Errorf("no matching sources found to delete")
	}
	Fmsg = qreport.Report(qpaths, errs, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")

	return nil
}
