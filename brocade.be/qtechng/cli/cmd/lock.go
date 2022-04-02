package cmd

import (
	"github.com/spf13/cobra"
)

var lockCmd = &cobra.Command{
	Use:     "lock",
	Short:   "Lock functions",
	Long:    `All kinds of lock actions for files`,
	Args:    cobra.NoArgs,
	Example: "qtechng lock",
}

func init() {
	rootCmd.AddCommand(lockCmd)

}
