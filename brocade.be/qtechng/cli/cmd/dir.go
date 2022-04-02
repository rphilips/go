package cmd

import (
	"github.com/spf13/cobra"
)

// Finplace replace the file contents

var dirCmd = &cobra.Command{
	Use:     "dir",
	Short:   "Directory functions",
	Long:    `All kinds of actions on directories`,
	Args:    cobra.NoArgs,
	Example: "qtechng dir",
}

func init() {
	rootCmd.AddCommand(dirCmd)
}
