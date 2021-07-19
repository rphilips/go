package cmd

import (
	"strings"

	qreport "brocade.be/qtechng/lib/report"

	"github.com/spf13/cobra"
)

var commandListCmd = &cobra.Command{
	Use:   "list",
	Short: "Available commands",
	Long: `
Displays a list of all available qtechng commands`,
	Example: "qtechng command list",
	RunE:    commandList}

func init() {

	commandCmd.AddCommand(commandListCmd)
}

func commandList(cmd *cobra.Command, args []string) error {
	msg := map[string]map[string]string{}
	for _, command := range rootCmd.Commands() {
		msg[command.Use] = map[string]string{}
		for _, subCommand := range command.Commands() {
			parts := strings.SplitN(subCommand.Use, " ", 2)
			msg[command.Use][parts[0]] = subCommand.Short
		}
	}
	Fmsg = qreport.Report(msg, nil, Fjq, Fyaml, Funquote, Fjoiner, Fsilent, "")
	return nil
}
