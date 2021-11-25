package cmd

import (
	"github.com/spf13/cobra"
)

var digestCmd = &cobra.Command{
	Use:     "digest",
	Short:   "Digest tools",
	Long:    `Functions regarding IIIF digests`,
	Args:    cobra.NoArgs,
	Example: "iiiftool digest",
}

func init() {
	rootCmd.AddCommand(digestCmd)
}
