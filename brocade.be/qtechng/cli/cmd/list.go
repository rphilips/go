package cmd

import (
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List functions",
	Long:    `All kinds of actions on lists`,
	Args:    cobra.NoArgs,
	Example: "qtechng list",
}

var Fshow bool

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.PersistentFlags().StringVar(&Flist, "list", "", "Lists for convenient editing")
	listCmd.PersistentFlags().BoolVar(&Fshow, "show", true, "Show the qpaths")
}
