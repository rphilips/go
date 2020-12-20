package cmd

import (
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "Version functions",
	Long:    `All kinds of actions on versions`,
	Args:    cobra.NoArgs,
	Example: "qtechng version",
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
