package cmd

import (
	"github.com/spf13/cobra"

	qproject "brocade.be/qtechng/lib/project"
	qreport "brocade.be/qtechng/lib/report"
)

var projectInfoCmd = &cobra.Command{
	Use:     "info",
	Short:   "Provide information about a project",
	Long:    `This command provides information about a project`,
	Example: "qtechng project info /stdlib/template",
	RunE:    projectInfo,
	PreRun:  func(cmd *cobra.Command, args []string) { preSSH(cmd, nil) },
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
		Fmsg = qreport.Report(result, errs, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
		return nil
	}
	Fmsg = qreport.Report(result, errs, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	return nil

}
