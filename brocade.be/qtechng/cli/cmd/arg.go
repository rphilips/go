package cmd

import (
	"github.com/spf13/cobra"
)

var argCmd = &cobra.Command{
	Use:   "arg",
	Short: "Alternative ways to start qtechng",
	Long: `As a CLI application *qtechng* can be started by specifying the arguments and flags on the command line.
These are not always the most convenient ways. *arg* specifies several alternatives.`,
	Args:    cobra.NoArgs,
	Example: "qtechng arg",
}

func init() {
	rootCmd.AddCommand(argCmd)
}
