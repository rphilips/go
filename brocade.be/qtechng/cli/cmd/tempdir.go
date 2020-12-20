package cmd

import (
	"github.com/spf13/cobra"
)

var tempdirCmd = &cobra.Command{
	Use:   "tempdir",
	Short: "Creates temporary directories",
	Long:  `Creates temporary directories`,
}

func init() {
	rootCmd.AddCommand(tempdirCmd)
}
