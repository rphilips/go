package cmd

import (
	"log"
	"strings"

	qclient "brocade.be/qtechng/lib/client"
	qreport "brocade.be/qtechng/lib/report"
	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var sourceLintCmd = &cobra.Command{
	Use:     "lint",
	Short:   "Lints sources in the repository",
	Long:    `Lints sources in the repository according to patterns, nature and contents`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng source lint --qpattern=/application/*.m`,
	RunE:    sourceLint,
	PreRun:  preSourceLint,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BWP",
		"fill-version":   "yes",
	},
}

var Fwarnings bool

func init() {
	sourceCmd.AddCommand(sourceLintCmd)
	sourceLintCmd.Flags().BoolVar(&Fwarnings, "warnings", false, "Include warnings")
}

func sourceLint(cmd *cobra.Command, args []string) error {
	_, result := lintTransport(Fcargo)
	qps := make([]string, 0)
	for _, r := range result {
		if r.Info != "OK" {
			qps = append(qps, r.QPath)
		}
	}
	qutil.EditList(Flist, Ftransported, qps)
	Fmsg = qreport.Report(result, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	return nil
}

func preSourceLint(cmd *cobra.Command, args []string) {
	if !Ftransported {
		var err error
		Fcargo, err = fetchData(args, Ffilesinproject, nil, false)
		if err != nil {
			log.Fatal("cmd/source_lint/1:\n", err)
		}
	}

	if strings.ContainsRune(QtechType, 'B') || strings.ContainsRune(QtechType, 'P') {
		if Fbatchid == "" {
			Fbatchid = "lint"
		}
		if Fwarnings {
			Fbatchid = "w:" + Fbatchid
		}

		addData(Fpayload, Fcargo, false, true, Fbatchid)
	}

	if Ftransported {
		err := qclient.SendCargo(Fcargo)
		if err != nil {
			log.Fatal("cmd/source_lint/2:\n", err)
		}
		cmd.RunE = func(cmd *cobra.Command, args []string) error { return nil }
	}
}
