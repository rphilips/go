package cmd

import (
	"github.com/spf13/cobra"

	qregistry "brocade.be/base/registry"
	qreport "brocade.be/qtechng/lib/report"
)

var systemUnblockdocCmd = &cobra.Command{
	Use:     "unblockdoc",
	Short:   "Unblock documentation",
	Long:    `This command unblocks publishing documentation by changing the registry`,
	Args:    cobra.NoArgs,
	Example: "qtechng system unblockdoc",
	RunE:    systemUnblockdoc,
	Annotations: map[string]string{
		"with-qtechtype": "BP",
	},
}

func init() {
	systemCmd.AddCommand(systemUnblockdocCmd)
}

func systemUnblockdoc(cmd *cobra.Command, args []string) error {

	err := qregistry.SetRegistry("qtechng-block-doc", "0")
	msg := "Documentation publishing is unblocked!"
	if err != nil {
		msg = ""
	}
	Fmsg = qreport.Report(msg, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	return nil
}
