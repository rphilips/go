package cmd

import (
	"github.com/spf13/cobra"
)

var fileCmd = &cobra.Command{
	Use:     "file",
	Short:   "File tools",
	Long:    `Functions regarding files`,
	Args:    cobra.NoArgs,
	Example: "iiiftool file",
}

func init() {
	rootCmd.AddCommand(fileCmd)
}
