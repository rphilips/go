package cmd

import (
	"github.com/spf13/cobra"

	qregistry "brocade.be/base/registry"
	qerror "brocade.be/qtechng/lib/error"
)

var systemUnblockinstallCmd = &cobra.Command{
	Use:     "unblockinstall",
	Short:   "Unblock installation",
	Long:    `Unblock installation: changes registry`,
	Args:    cobra.NoArgs,
	Example: "qtechng system unblockinstall",
	RunE:    systemUnblockinstall,
	Annotations: map[string]string{
		"with-qtechtype": "BP",
	},
}

func init() {
	systemCmd.AddCommand(systemUnblockinstallCmd)
}

func systemUnblockinstall(cmd *cobra.Command, args []string) error {

	err := qregistry.SetRegistry("qtechng-blocked-install", "0")
	msg := "Installation is unblocked!"
	if err != nil {
		msg = ""
	}
	Fmsg = qerror.ShowResult(msg, Fjq, err, Fyaml)
	return nil
}
