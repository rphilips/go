package cmd

import (
	"github.com/spf13/cobra"
)

var clipboardCmd = &cobra.Command{
	Use:   "clipboard",
	Short: "Works with system clipboard",
	Long:  `Works with system clipboard: both setting and retrieving text data`,
}

func init() {
	rootCmd.AddCommand(clipboardCmd)
}
