package cmd

import (
	"github.com/spf13/cobra"
)

var lockCmd = &cobra.Command{
	Use:     "lock",
	Short:   "Lock functions",
	Long:    `All kind of lock (based on files) actions `,
	Args:    cobra.NoArgs,
	Example: "qtechng lock",
}

func init() {
	rootCmd.AddCommand(lockCmd)

}
