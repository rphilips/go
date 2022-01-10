package cmd

import (
	"github.com/spf13/cobra"
)

var pipeCmd = &cobra.Command{
	Use:     "pipe",
	Short:   "Working with named pipes",
	Long:    `Working with named pipes`,
	Args:    cobra.NoArgs,
	Example: "goya pipe",
}

var Fpipe = ""

func init() {
	pipeCmd.PersistentFlags().StringVar(&Fpipe, "pipe", "", "location of named pipe")
	rootCmd.AddCommand(pipeCmd)
}
