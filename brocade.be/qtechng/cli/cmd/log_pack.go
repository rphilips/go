package cmd

import (
	qlog "brocade.be/qtechng/lib/log"
	qreport "brocade.be/qtechng/lib/report"
	"github.com/spf13/cobra"
)

var logPackCmd = &cobra.Command{
	Use:   "pack",
	Short: "Pack log files",
	Long: `This command Packs log files.
All log files - earlier than today - are packed into one file
in the logging directory.

This operation runs automatically with the first qtechng action of the day.
`,
	Args:    cobra.NoArgs,
	Example: "qtechng log pack",
	RunE:    logPack,
}

func init() {
	logCmd.AddCommand(logPackCmd)
}

func logPack(cmd *cobra.Command, args []string) error {
	msg, err := qlog.Pack()
	Fmsg = qreport.Report(msg, err, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	return nil
}
