package cmd

import (
	"strconv"
	"time"

	"github.com/spf13/cobra"

	qregistry "brocade.be/base/registry"
	qreport "brocade.be/qtechng/lib/report"
)

var systemBlockdocCmd = &cobra.Command{
	Use:   "blockdoc sec",
	Short: "Block documentation publishing",
	Long: `This command prevents the documentation from being published.
Provide the number of seconds during with the block applies.
This action has to be initiated on the server itself.

Blocking with 0 seconds unblocks the server`,
	Args: cobra.ExactArgs(1),
	Example: `qtechng system blockdoc 3600
qtechng system blockdoc 0`,

	RunE: systemBlockdoc,
	Annotations: map[string]string{
		"with-qtechtype": "BP",
	},
}

func init() {
	systemCmd.AddCommand(systemBlockdocCmd)
}

func systemBlockdoc(cmd *cobra.Command, args []string) error {
	offset := args[0]
	ioffset, err := strconv.Atoi(offset)
	if err != nil {
		err := qregistry.SetRegistry("qtechng-block-doc", "0")
		return err
	}
	msg := ""
	if ioffset != 0 {
		h := time.Now()
		h = h.Add(time.Second * time.Duration(ioffset))
		t := h.Format(time.RFC3339)
		err = qregistry.SetRegistry("qtechng-block-doc", t)
		msg = "Documentation publishing blocked until `" + t + "`"
		if err != nil {
			msg = ""
		}
	} else {
		err = qregistry.SetRegistry("qtechng-block-doc", "0")
		msg = "Documentation is published again!"
		if err != nil {
			msg = ""
		}
	}
	Fmsg = qreport.Report(msg, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	return nil
}
