package cmd

import (
	"log"
	"strings"

	qclient "brocade.be/qtechng/lib/client"
	"github.com/spf13/cobra"
)

var projectCoCmd = &cobra.Command{
	Use:     "co",
	Short:   "Checks out QtechNG files from projects",
	Long:    `Command to retrieve files in a project from the QtechNG repository`,
	Args:    cobra.MinimumNArgs(0),
	Example: `qtechng source co /catalografie/application`,
	RunE:    sourceCo,
	PreRun:  preSourceCo,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BWP",
		"fill-version":   "yes",
	},
}

func init() {
	projectCmd.AddCommand(projectCoCmd)
}

var projectCo = sourceCo

func preProjectCo(cmd *cobra.Command, args []string) {
	if !Ftransported {
		var err error
		Fcargo, err = fetchData(args, true, nil, false)
		if err != nil {
			log.Fatal("cmd/project_co/1:\n", err)
		}
	}

	if strings.ContainsRune(QtechType, 'B') || strings.ContainsRune(QtechType, 'P') {
		addData(Fpayload, Fcargo, true, "")
	}

	if Ftransported {
		err := qclient.SendCargo(Fcargo)
		if err != nil {
			log.Fatal("cmd/project_co/2:\n", err)
		}
	}
}
