package cmd

import (
	"log"
	"strings"

	qclient "brocade.be/qtechng/lib/client"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var sourceRebuildCmd = &cobra.Command{
	Use:     "rebuild",
	Short:   "Rebuilds sources in the repository",
	Long:    `Rebuilds sources in the repository according to patterns, nature and contents`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng source rebuild --qpattern=/application/*.m --version=0.00`,
	RunE:    sourceRebuild,
	PreRun:  preSourceRebuild,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BWP",
		"fill-version":   "yes",
	},
}

func init() {
	sourceRebuildCmd.PersistentFlags().StringVar(&Frefname, "refname", "rebuild", "Reference to the rebuildation")
	sourceCmd.AddCommand(sourceRebuildCmd)
}

func sourceRebuild(cmd *cobra.Command, args []string) error {
	qpaths, result := listTransport(Fcargo)
	qutil.EditList(Flist, Ftransported, qpaths)
	Fmsg = qreport.Report(result, nil, Fjq, Fyaml, Funquote)
	return nil
}

func preSourceRebuild(cmd *cobra.Command, args []string) {
	if !Ftransported {
		var err error
		Fcargo, err = fetchData(args, Ffilesinproject, nil, false)
		if err != nil {
			log.Fatal("cmd/source_rebuild/1:\n", err)
		}
	}

	if strings.ContainsRune(QtechType, 'B') || strings.ContainsRune(QtechType, 'P') {
		rebuildData(Fpayload, Fcargo, false, "")
	}

	if Ftransported {
		err := qclient.SendCargo(Fcargo)
		if err != nil {
			log.Fatal("cmd/source_rebuild/2:\n", err)
		}
		cmd.RunE = func(cmd *cobra.Command, args []string) error { return nil }
	}
}
