package cmd

import (
	"github.com/spf13/cobra"
)

var argCmd = &cobra.Command{
	Use:   "arg",
	Short: "Alternative ways to start goyo",
	Long: `A a CLI application *goyo* can be started by specifying the the arguments and flags on the command line.
These are not always the most convenient ways. *arg* specifies several alternatives.`,
	Args:    cobra.NoArgs,
	Example: "goya arg",
}

func init() {
	rootCmd.AddCommand(argCmd)
}
