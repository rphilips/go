package cmd

import (
	"strconv"
	"time"

	"github.com/spf13/cobra"

	qregistry "brocade.be/base/registry"
	qreport "brocade.be/qtechng/lib/report"
)

var systemBlockinstallCmd = &cobra.Command{
	Use:   "blockinstall sec",
	Short: "Blocks installation of software",
	Long: `
Give a number of seconds during with the block applies.
This action has to be initiated on the servers itself.

Blocking with 0 sec., unblocks the server`,
	Args: cobra.ExactArgs(1),
	Example: `
  qtechng system blockinstall 3600
  qtechng system blockinstall 0`,

	RunE: systemBlockinstall,
	Annotations: map[string]string{
		"with-qtechtype": "BP",
	},
}

func init() {
	systemCmd.AddCommand(systemBlockinstallCmd)
}

func systemBlockinstall(cmd *cobra.Command, args []string) error {
	offset := args[0]
	ioffset, err := strconv.Atoi(offset)
	if err != nil {
		err := qregistry.SetRegistry("qtechng-block-install", "0")
		return err
	}
	msg := ""
	if ioffset != 0 {
		h := time.Now()
		h = h.Add(time.Second * time.Duration(ioffset))
		t := h.Format(time.RFC3339)
		err = qregistry.SetRegistry("qtechng-block-install", t)
		msg = "Installation blocked until `" + t + "`"
		if err != nil {
			msg = ""
		}
	} else {
		err = qregistry.SetRegistry("qtechng-block-install", "0")
		msg = "Installation is possible again!"
		if err != nil {
			msg = ""
		}
	}
	Fmsg = qreport.Report(msg, err, Fjq, Fyaml, Funquote, Fsilent)
	return nil
}
