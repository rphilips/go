package cmd

import (
	"log"
	"strings"

	qclient "brocade.be/qtechng/lib/client"
	"github.com/spf13/cobra"
)

var projectCoCmd = &cobra.Command{
	Use:   "co",
	Short: "Check out qtechng files from projects",
	Long: `This command retrieves files in a project from the qtechng repository.
The --copyonly flag updates the local file contents, but does not affect its qtechng status.
This can be used, for instance, to deliberately replace
one repository version of a file with another.`,
	Args: cobra.MinimumNArgs(0),
	Example: `qtechng project co /catalografie
qtechng project co /catalografie/oai3`,
	RunE:   projectCo,
	PreRun: preProjectCo,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BWP",
		"fill-version":   "yes",
	},
}

func init() {
	projectCoCmd.Flags().BoolVar(&Fcopyonly, "copyonly", false, "Check out without updating local repository information")
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
		addData(Fpayload, Fcargo, true, false, "")
	}

	if Ftransported {
		err := qclient.SendCargo(Fcargo)
		if err != nil {
			log.Fatal("cmd/project_co/2:\n", err)
		}
	}
}
