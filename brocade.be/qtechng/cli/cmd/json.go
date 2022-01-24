package cmd

import (
	"github.com/spf13/cobra"
)

var jsonCmd = &cobra.Command{
	Use:     "json",
	Short:   "JSON functions",
	Long:    `All kinds of actions on JSON`,
	Args:    cobra.NoArgs,
	Example: "qtechng json",
}

func init() {
	rootCmd.AddCommand(jsonCmd)
	jsonCmd.PersistentFlags().BoolVar(&Frecurse, "recurse", false, "Recursively traverse directories")
	jsonCmd.PersistentFlags().BoolVar(&Futf8only, "utf8only", false, "Is this a file with UTF-8 content?")
	jsonCmd.PersistentFlags().BoolVar(&Fask, "ask", false, "Ask for arguments")
	jsonCmd.PersistentFlags().StringVar(&Fext, "ext", "", "Additional extension for result file")
	jsonCmd.PersistentFlags().StringArrayVar(&Fpattern, "pattern", []string{}, "Posix glob pattern on the basenames")
}
