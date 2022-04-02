package cmd

import (
	"log"
	"strings"

	qclient "brocade.be/qtechng/lib/client"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var sourceListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List sources in the repository",
	Long:    `This command lists sources in the repository according to patterns, nature and contents`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng source list --qpattern=/application/*.m`,
	RunE:    sourceList,
	PreRun:  preSourceList,
	Annotations: map[string]string{
		"with-qtechtype": "BWP",
		"fill-version":   "yes",
	},
}

func init() {
	sourceCmd.AddCommand(sourceListCmd)
}

func sourceList(cmd *cobra.Command, args []string) error {
	qpaths, result := listTransport(Fcargo)
	qutil.EditList(Flist, Ftransported, qpaths)
	Fmsg = qreport.Report(result, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
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

	if strings.ContainsAny(QtechType, "PB") {
		addData(Fpayload, Fcargo, false, false, "")
	}

	if Ftransported {
		err := qclient.SendCargo(Fcargo)
		if err != nil {
			log.Fatal("cmd/source_list/2:\n", err)
		}
		cmd.RunE = func(cmd *cobra.Command, args []string) error { return nil }
	}
}
