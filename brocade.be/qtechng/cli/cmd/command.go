package cmd

import (
	"github.com/spf13/cobra"
)

var commandCmd = &cobra.Command{
	Use:     "command",
	Short:   "Command functions",
	Long:    `Working with qtechng commands`,
	Example: "qtechng command list",
}

func init() {

	rootCmd.AddCommand(commandCmd)
}
