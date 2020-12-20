package cmd

import (
	"github.com/spf13/cobra"

	qerror "brocade.be/qtechng/lib/error"
	qmeta "brocade.be/qtechng/lib/meta"
	qproject "brocade.be/qtechng/lib/project"
)

var projectNewCmd = &cobra.Command{
	Use:     "new",
	Short:   "Creates a new project",
	Long:    `Command creates a new project on the development server`,
	Args:    cobra.MinimumNArgs(1),
	Example: "qtechng project new /stdlib/template\nqtechng project create /stdlib/template  --version=5.10",
	RunE:    projectNew,
	PreRun:  func(cmd *cobra.Command, args []string) { preSSH(cmd) },
	Annotations: map[string]string{
		"remote-allowed": "yes",
		"always-remote":  "yes",
		"with-qtechtype": "BW",
		"fill-version":   "yes",
	},
}

func init() {
	projectCmd.AddCommand(projectNewCmd)
}

func projectNew(cmd *cobra.Command, args []string) error {
	meta := qmeta.Meta{
		Mu: FUID,
	}
	result, errs := qproject.InitList(Fversion, args, func(a string) qmeta.Meta { return meta })

	Fmsg = qerror.ShowResult(result, Fjq, errs)
	return nil
}
