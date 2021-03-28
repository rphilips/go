package cmd

import (
	"log"
	"strings"

	qclient "brocade.be/qtechng/lib/client"
	qerror "brocade.be/qtechng/lib/error"
	"github.com/spf13/cobra"
)

// Fbatchid string die het batchid opgeeft
var Fbatchid = "batchid"

var sourceMumpsCmd = &cobra.Command{
	Use:     "mumps",
	Short:   "Retrieves the data sent to M",
	Long:    `Retrieves the data sent to M`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng source mumps /catalografie/application/bcawedit.m`,
	RunE:    sourceMumps,
	PreRun:  preSourceMumps,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BWP",
		"fill-version":   "yes",
	},
}

func init() {
	sourceMumpsCmd.PersistentFlags().StringVar(&Fbatchid, "batchid", "", "batchid for the M stream")
	sourceCmd.AddCommand(sourceMumpsCmd)
}

func sourceMumps(cmd *cobra.Command, args []string) error {
	if Fcargo.Error != nil {
		Fmsg = qerror.ShowResult("", Fjq, Fcargo.Error, Fyaml)
		return nil
	}
	Fmsg = Fcargo.Buffer.String()
	return nil
}

func preSourceMumps(cmd *cobra.Command, args []string) {
	if Fbatchid == "" {
		Fbatchid = "batchid"
	}
	if !Ftransported {
		var err error
		Fcargo, err = fetchData(args, Ffilesinproject, nil, true)
		if err != nil {
			log.Fatal("cmd/source_mumps/1:\n", err)
		}
	}

	if strings.ContainsRune(QtechType, 'B') || strings.ContainsRune(QtechType, 'P') {
		addData(Fpayload, Fcargo, true, "m:"+Fbatchid)

	}

	if Ftransported {
		err := qclient.SendCargo(Fcargo)
		if err != nil {
			log.Fatal("cmd/source_mumps/2:\n", err)
		}
	}
}
