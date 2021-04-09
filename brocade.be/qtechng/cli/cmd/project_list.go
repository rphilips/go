package cmd

import (
	"log"
	"strings"

	qclient "brocade.be/qtechng/lib/client"
	qreport "brocade.be/qtechng/lib/report"
	"github.com/spf13/cobra"
)

var projectListCmd = &cobra.Command{
	Use:     "list",
	Short:   "List projects and its files",
	Long:    `Command lists all project matching a given pattern`,
	Example: "qtechng project list /catalografie ",
	RunE:    projectList,
	PreRun:  preProjectList,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BWP",
		"fill-version":   "yes",
	},
}

func init() {
	projectCmd.AddCommand(projectListCmd)
}

func preProjectList(cmd *cobra.Command, args []string) {
	if !Ftransported {
		var err error
		Fcargo, err = fetchData(args, true, nil, false)
		if err != nil {
			log.Fatal("cmd/project_list/1:\n", err)
		}
	}

	if strings.ContainsRune(QtechType, 'B') || strings.ContainsRune(QtechType, 'P') {
		addData(Fpayload, Fcargo, false, "")
	}

	if Ftransported {
		err := qclient.SendCargo(Fcargo)
		if err != nil {
			log.Fatal("cmd/project_list/2:\n", err)
		}
	}
}

func projectList(cmd *cobra.Command, args []string) error {

	result := projlistTransport(Fcargo)
	Fmsg = qreport.Report(result, nil, Fjq, Fyaml)
	return nil
}

func projlistTransport(pcargo *qclient.Cargo) []projlister {
	result := make([]projlister, 0)
	done := make(map[string]bool)
	if pcargo != nil && len(pcargo.Transports) != 0 {
		for _, transport := range Fcargo.Transports {
			locfil := transport.LocFile
			p := locfil.Project
			r := locfil.Release
			inx := r + " " + p
			if done[inx] {
				continue
			}
			done[inx] = true
			result = append(result, projlister{
				Release: locfil.Release,
				Project: locfil.Project,
			})
		}
	}
	return result
}
