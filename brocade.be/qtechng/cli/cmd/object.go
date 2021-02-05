package cmd

import (
	"github.com/spf13/cobra"
)

var objectCmd = &cobra.Command{
	Use:     "object",
	Short:   "Object activities",
	Long:    `Commands working on the objects in the repository`,
	Args:    cobra.NoArgs,
	Example: "qtechng object",
}

func init() {
	rootCmd.AddCommand(objectCmd)
	objectCmd.PersistentFlags().StringVar(&Fversion, "version", "", "Version to work with")
}
