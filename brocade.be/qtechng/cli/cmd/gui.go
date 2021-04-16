package cmd

import (
	"github.com/spf13/cobra"
)

// Finplace replace the file contents

var guiCmd = &cobra.Command{
	Use:     "gui",
	Short:   "GUI functions",
	Long:    `All kinds of GUI functions`,
	Args:    cobra.NoArgs,
	Example: "qtechng gui",
}

func init() {
	rootCmd.AddCommand(guiCmd)
}
