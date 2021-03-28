package cmd

import (
	"github.com/spf13/cobra"

	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
)

var systemDisableCmd = &cobra.Command{
	Use:     "disable",
	Short:   "Disable actions from a workstations",
	Long:    `Disable actions from a workstations: changes registry`,
	Args:    cobra.NoArgs,
	Example: "qtechng system disable",
	RunE:    systemDisable,
	Annotations: map[string]string{
		"with-qtechtype": "B",
	},
}

func init() {
	systemCmd.AddCommand(systemDisableCmd)
}

func systemDisable(cmd *cobra.Command, args []string) error {
	err := qregistry.SetRegistry("qtechng-disable-qtechng", "1")
	msg := "QtechNG disabled!"
	if err != nil {
		msg = ""
	}
	Fmsg = qerror.ShowResult(msg, Fjq, err, Fyaml)
	return nil
}
