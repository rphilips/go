package cmd

import (
	"github.com/spf13/cobra"
)

var manifestCmd = &cobra.Command{
	Use:     "manifest",
	Short:   "Manifest tools",
	Long:    `Functions regarding IIIF manifests`,
	Args:    cobra.NoArgs,
	Example: "iiiftool manifest",
}

func init() {
	rootCmd.AddCommand(manifestCmd)
}
