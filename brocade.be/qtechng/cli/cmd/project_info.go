package cmd

import (
	"github.com/spf13/cobra"

	qproject "brocade.be/qtechng/lib/project"
	qreport "brocade.be/qtechng/lib/report"
)

var projectInfoCmd = &cobra.Command{
	Use:     "info",
	Short:   "Information about a project",
	Long:    `Command provides information about a project`,
	Example: "qtechng project info /stdlib/template",
	RunE:    projectInfo,
	PreRun:  func(cmd *cobra.Command, args []string) { preSSH(cmd) },
	Annotations: map[string]string{
		"remote-allowed":    "yes",
		"always-remote-onW": "yes",
		"fill-version":      "yes",
		"with-qtechtype":    "BW",
	},
}

func init() {
	projectCmd.AddCommand(projectInfoCmd)
}

func projectInfo(cmd *cobra.Command, args []string) error {
	result, errs := qproject.Info(Fversion, args)
	if errs != nil {
		Fmsg = qreport.Report(result, errs, Fjq, Fyaml, Funquote)
		return nil
	}
	Fmsg = qreport.Report(result, errs, Fjq, Fyaml, Funquote)
	return nil

}
