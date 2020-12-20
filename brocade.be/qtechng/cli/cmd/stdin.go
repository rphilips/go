package cmd

import (
	"github.com/spf13/cobra"
)

var stdinCmd = &cobra.Command{
	Use:     "stdin",
	Short:   "stdin functions",
	Long:    `All kinds of actions on stdin stream`,
	Args:    cobra.NoArgs,
	Example: "qtechng stdin",
}

func init() {
	rootCmd.AddCommand(stdinCmd)
}
