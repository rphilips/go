package cmd

import (
	"github.com/spf13/cobra"
)

var tempdirCmd = &cobra.Command{
	Use:   "tempdir",
	Short: "Temporary directories functions",
	Long:  `All kinds of actions on temporary directories`,
}

func init() {
	rootCmd.AddCommand(tempdirCmd)
}
