package cmd

import (
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List functions",
	Long: `All kinds of actions on lists.
A list is a collection of qpaths. They are maintained on workstations in files in
a subdirectory of the support directory.
They are identified by a name and this name can be used in operations.`,
	Args:    cobra.NoArgs,
	Example: "qtechng list",
}

var Fshow bool

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.PersistentFlags().StringVar(&Flist, "list", "", "Lists for convenient editing")
	listCmd.PersistentFlags().BoolVar(&Fshow, "show", true, "Show the qpaths")
}
