package cmd

import (
	"github.com/spf13/cobra"
)

var clipboardCmd = &cobra.Command{
	Use:   "clipboard",
	Short: "Manipulating the system clipboard",
	Long:  `Manipulating the system clipboard: this commands allows both setting and retrieving text data`,
}

func init() {
	rootCmd.AddCommand(clipboardCmd)
}
