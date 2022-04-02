package cmd

import (
	"log"
	"sort"
	"strings"

	qclient "brocade.be/qtechng/lib/client"
	qreport "brocade.be/qtechng/lib/report"
	"github.com/spf13/cobra"
)

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List projects",
	Long: `This command lists all projects matching a given pattern.
The projects are displayed in order of installation`,
	Example: "qtechng project list /catalografie",
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
		addData(Fpayload, Fcargo, false, false, "")
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
	Fmsg = qreport.Report(result, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "", "")
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
			s := locfil.Sort
			inx := r + " " + p
			if done[inx] {
				continue
			}
			done[inx] = true
			result = append(result, projlister{
				Release: locfil.Release,
				Project: locfil.Project,
				Sort:    s,
			})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Sort < result[j].Sort
	})
	return result
}
