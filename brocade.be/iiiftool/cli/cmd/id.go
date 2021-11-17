package cmd

import (
	"github.com/spf13/cobra"
)

var idCmd = &cobra.Command{
	Use:   "id",
	Short: "Identifier tools",
	Long: `Functions regarding IIIF identifiers`,
	Args:    cobra.NoArgs,
	Example: "iiiftool id",
}

func init() {
	rootCmd.AddCommand(idCmd)
}
