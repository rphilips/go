package cmd

import (
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:     "test",
	Short:   "test tools",
	Long:    `Functions regarding IIIF testing`,
	Args:    cobra.NoArgs,
	Example: "iiiftool test",
}

func init() {
	rootCmd.AddCommand(testCmd)
}
