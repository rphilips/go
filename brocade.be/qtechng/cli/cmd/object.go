package cmd

import (
	"github.com/spf13/cobra"
)

var objectCmd = &cobra.Command{
	Use:     "object",
	Short:   "Object functions",
	Long:    `All kinds of actions on the repository objects`,
	Args:    cobra.NoArgs,
	Example: "qtechng object",
}
var Fobjpattern []string

func init() {
	rootCmd.AddCommand(objectCmd)
	objectCmd.PersistentFlags().StringVar(&Fversion, "version", "", "Version to work with")
}
