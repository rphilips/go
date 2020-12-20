package cmd

import (
	"github.com/spf13/cobra"

	qerror "brocade.be/qtechng/lib/error"
	qproject "brocade.be/qtechng/lib/project"
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
		Fmsg = qerror.ShowResult(result, Fjq, errs)
		return nil
	}
	Fmsg = qerror.ShowResult(result, Fjq, errs)
	return nil

}
