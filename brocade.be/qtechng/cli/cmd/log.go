package cmd

import (
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:     "log",
	Short:   "Log functions",
	Long:    `All kinds of actions on logs`,
	Args:    cobra.NoArgs,
	Example: "qtechng log",
}

func init() {
	rootCmd.AddCommand(logCmd)

}
