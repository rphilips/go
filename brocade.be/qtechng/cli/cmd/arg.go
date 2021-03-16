package cmd

import (
	"github.com/spf13/cobra"
)

var argCmd = &cobra.Command{
	Use:     "arg",
	Short:   "Alternative ways to start qtechng",
	Long:    `Alternative ways to start qtechng: intended for use in other software`,
	Args:    cobra.NoArgs,
	Example: "  qtechng arg",
}

func init() {
	rootCmd.AddCommand(argCmd)
}
