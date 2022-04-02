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

var Fext = ""
var Futf8only = false
var Fask = false
var Fsearch = ""

func init() {
	rootCmd.AddCommand(fsCmd)
	fsCmd.PersistentFlags().BoolVar(&Frecurse, "recurse", false, "Recursively traverse directories")
	fsCmd.PersistentFlags().BoolVar(&Futf8only, "utf8only", false, "Is this a file with UTF-8 content?")
	fsCmd.PersistentFlags().BoolVar(&Fask, "ask", false, "Ask for arguments")
	fsCmd.PersistentFlags().StringVar(&Fext, "ext", "", "Additional extension for result file")
	fsCmd.PersistentFlags().StringArrayVar(&Fpattern, "pattern", []string{}, "Posix glob pattern on the basenames")
}
