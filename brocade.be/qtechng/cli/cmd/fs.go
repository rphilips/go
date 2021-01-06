package cmd

import (
	"github.com/spf13/cobra"
)

var fsCmd = &cobra.Command{
	Use:     "fs",
	Short:   "File functions",
	Long:    `All kinds of actions on local filesystem`,
	Args:    cobra.NoArgs,
	Example: "qtechng fs",
}

func init() {
	rootCmd.AddCommand(fsCmd)

}
