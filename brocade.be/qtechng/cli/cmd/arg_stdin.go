package cmd

import (
	"github.com/spf13/cobra"
)

var argStdinCmd = &cobra.Command{
	Use:     "stdin",
	Short:   "Start qtechng with arguments read from stdin",
	Long:    `Launches qtechng with the arguments as lines on stdin. Arguments should not be empty`,
	Args:    cobra.NoArgs,
	Example: `qtechng arg stdin`,
	RunE:    argStdin,
}

func init() {
	argCmd.AddCommand(argStdinCmd)
}

func argStdin(cmd *cobra.Command, args []string) error {
	return nil
}
