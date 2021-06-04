package cmd

import (
	qreport "brocade.be/qtechng/lib/report"

	"github.com/spf13/cobra"
)

var commandsCmd = &cobra.Command{
	Use:   "commands",
	Short: "Available commands",
	Long: `
Displays a list of all available qtechng commands`,
	Example: "  qtechng commands",
	RunE:    commands}

func init() {

	rootCmd.AddCommand(commandsCmd)
}

func commands(cmd *cobra.Command, args []string) error {
	msg := map[string]map[string]string{}
	for _, command := range rootCmd.Commands() {
		msg[command.Use] = map[string]string{}
		for _, subCommand := range command.Commands() {
			msg[command.Use][subCommand.Use] = subCommand.Short
		}
	}
	Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fsilent)
	return nil
}
