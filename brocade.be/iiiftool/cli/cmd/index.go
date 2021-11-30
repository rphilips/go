package cmd

import (
	"github.com/spf13/cobra"
)

var indexCmd = &cobra.Command{
	Use:     "index",
	Short:   "Index tools",
	Long:    `Functions regarding IIIF index`,
	Args:    cobra.NoArgs,
	Example: "iiiftool index",
}

func init() {
	rootCmd.AddCommand(indexCmd)
}
