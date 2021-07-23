package cmd

import (
	"github.com/spf13/cobra"
)

var commandCmd = &cobra.Command{
	Use:     "command",
	Short:   "Command functionality",
	Long:    `Working with the QtechNG commands`,
	Example: "qtechng command",
}

func init() {

	rootCmd.AddCommand(commandCmd)
}
