package cmd

import (
	"github.com/spf13/cobra"
)

var mumpsCmd = &cobra.Command{
	Use:     "mumps",
	Short:   "Mumps iteractions",
	Long:    `All kinds of actions on M`,
	Args:    cobra.NoArgs,
	Example: "qtechng mumps",
}

func init() {
	rootCmd.AddCommand(mumpsCmd)
}
