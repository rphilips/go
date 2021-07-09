package cmd

import (
	"github.com/spf13/cobra"
)

var textCmd = &cobra.Command{
	Use:     "text",
	Short:   "Text functions",
	Long:    `All kinds of actions on text`,
	Args:    cobra.NoArgs,
	Example: "qtechng fs",
}

func init() {
	rootCmd.AddCommand(textCmd)

}
