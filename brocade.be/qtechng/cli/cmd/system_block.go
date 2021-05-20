package cmd

import (
	"strconv"
	"time"

	"github.com/spf13/cobra"

	qregistry "brocade.be/base/registry"
	qreport "brocade.be/qtechng/lib/report"
)

var systemBlockCmd = &cobra.Command{
	Use:   "block sec",
	Short: "Blocks actions on workstations",
	Long: `
Blocks workstations to act on the development machine.
Give a number of seconds during with the block applies.
This action has to be initiated on the development machine itself.

Blocking with 0 sec., unblocks teh workstations`,
	Args: cobra.ExactArgs(1),
	Example: `
  qtechng system block 3600
  qtechng system block 0`,

	RunE: systemBlock,
	Annotations: map[string]string{
		"with-qtechtype": "B",
	},
}

func init() {
	systemCmd.AddCommand(systemBlockCmd)
}

func systemBlock(cmd *cobra.Command, args []string) error {
	offset := args[0]
	ioffset, err := strconv.Atoi(offset)
	if err != nil {
		err := qregistry.SetRegistry("qtechng-block-qtechng", "0")
		return err
	}
	msg := ""
	if ioffset != 0 {
		h := time.Now()
		h = h.Add(time.Second * time.Duration(ioffset))
		t := h.Format(time.RFC3339)
		err = qregistry.SetRegistry("qtechng-block-qtechng", t)
		msg = "QtechNG blocked until `" + t + "`"
		if err != nil {
			msg = ""
		}
	} else {
		err = qregistry.SetRegistry("qtechng-block-qtechng", "0")
		msg = "QtechNG unblocked!"
		if err != nil {
			msg = ""
		}
	}
	Fmsg = qreport.Report(msg, err, Fjq, Fyaml, Funquote)
	return nil
}
