package cmd

import (
	"github.com/spf13/cobra"

	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
)

var systemEnableCmd = &cobra.Command{
	Use:     "enable",
	Short:   "Enable actions from a workstations",
	Long:    `Enable actions from a workstations: changes registry`,
	Args:    cobra.NoArgs,
	Example: "qtechng system enable",
	RunE:    systemEnable,
	Annotations: map[string]string{
		"with-qtechtype": "B",
	},
}

func init() {
	systemCmd.AddCommand(systemEnableCmd)
}

func systemEnable(cmd *cobra.Command, args []string) error {

	err := qregistry.SetRegistry("qtechng-disable-qtechng", "0")
	msg := "QtechNG enabled!"
	if err != nil {
		msg = ""
	}
	Fmsg = qerror.ShowResult(msg, Fjq, err)
	return nil
}
