package cmd

import (
	"log"
	"strings"

	qclient "brocade.be/qtechng/lib/client"
	qerror "brocade.be/qtechng/lib/error"
	"github.com/spf13/cobra"
)

var objectListCmd = &cobra.Command{
	Use:     "list",
	Short:   "Lists objects in the repository",
	Long:    `Lists objects by name or by pattern`,
	Args:    cobra.MinimumNArgs(1),
	Example: `qtechng object list m4_getCatIsbdTitles m4_CO`,
	RunE:    objectList,
	PreRun:  preObjectList,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BWP",
	},
}

func init() {
	objectCmd.AddCommand(objectListCmd)
}

func objectList(cmd *cobra.Command, args []string) error {
	result := listTransport(Fcargo)
	Fmsg = qerror.ShowResult(result, Fjq, nil)
	return nil
}

func preObjectList(cmd *cobra.Command, args []string) {
	if !Ftransported {
		var err error
		Fcargo, err = fetchObjectData(args)
		if err != nil {
			log.Fatal("cmd/object_list/1:\n", err)
		}
	}

	if strings.ContainsRune(QtechType, 'B') || strings.ContainsRune(QtechType, 'P') {
		addObjectData(Fpayload, Fcargo, false, "")
	}

	if Ftransported {
		err := qclient.SendCargo(Fcargo)
		if err != nil {
			log.Fatal("cmd/object_list/2:\n", err)
		}
		cmd.RunE = func(cmd *cobra.Command, args []string) error { return nil }
	}
}
