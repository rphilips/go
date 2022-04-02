package cmd

import (
	"github.com/spf13/cobra"
)

var toolcatCmd = &cobra.Command{
	Use:     "toolcat",
	Short:   "Toolcat functions",
	Long:    `Producing docstrings for toolcat application`,
	Args:    cobra.NoArgs,
	Example: "qtechng toolcat",
}

var Ftcclip bool = false

func init() {
	toolcatCmd.PersistentFlags().BoolVar(&Ftcclip, "clipboard", false, "Put in clipboard")
	rootCmd.AddCommand(toolcatCmd)

}
