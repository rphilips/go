package cmd

import (
	"github.com/spf13/cobra"
)

var fsCmd = &cobra.Command{
	Use:   "fs",
	Short: "File functions",
	Long: `All kinds of actions on the local filesystem,
including UNIX-style commands like AWK or SED, which thus become available on all platforms`,
	Args:    cobra.NoArgs,
	Example: "qtechng fs",
}

func init() {
	rootCmd.AddCommand(fsCmd)

}
