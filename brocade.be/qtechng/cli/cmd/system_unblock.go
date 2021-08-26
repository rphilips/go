package cmd

import (
	"github.com/spf13/cobra"

	qregistry "brocade.be/base/registry"
	qreport "brocade.be/qtechng/lib/report"
)

var systemUnblockCmd = &cobra.Command{
	Use:     "unblock",
	Short:   "Unblock actions from workstations",
	Long:    `This command unblock actions from workstations by changing the registry`,
	Args:    cobra.NoArgs,
	Example: "qtechng system unblock",
	RunE:    systemUnblock,
	Annotations: map[string]string{
		"with-qtechtype": "B",
	},
}

func init() {
	systemCmd.AddCommand(systemUnblockCmd)
}

func systemUnblock(cmd *cobra.Command, args []string) error {

	err := qregistry.SetRegistry("qtechng-blocked-qtechng", "0")
	msg := "QtechNG unblocked!"
	if err != nil {
		msg = ""
	}
	Fmsg = qreport.Report(msg, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	return nil
}
