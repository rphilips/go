package cmd

import (
	"github.com/spf13/cobra"
)

// Finplace replace the file contents
var Finplace bool

var fileCmd = &cobra.Command{
	Use:     "file",
	Short:   "File functions",
	Long:    `All kinds of actions on files`,
	Args:    cobra.NoArgs,
	Example: "qtechng file",
}

func init() {
	rootCmd.AddCommand(fileCmd)
	fileCmd.PersistentFlags().StringVar(&Fversion, "version", "", "Version to work with")
	fileCmd.PersistentFlags().BoolVar(&Ftree, "tree", false, "Files with hierarchy intact")
	fileCmd.PersistentFlags().BoolVar(&Fauto, "auto", false, "Files according to the registry")
	fileCmd.PersistentFlags().StringSliceVar(&Fnature, "nature", []string{}, "QtechNG nature of file")
	fileCmd.PersistentFlags().StringSliceVar(&Fcu, "cuser", []string{}, "UID of creator")
	fileCmd.PersistentFlags().StringSliceVar(&Fmu, "muser", []string{}, "UID of last modifier")
	fileCmd.PersistentFlags().StringVar(&Fctbefore, "cbefore", "", "Created before")
	fileCmd.PersistentFlags().StringVar(&Fctafter, "cafter", "", "Created after")
	fileCmd.PersistentFlags().StringVar(&Fmtbefore, "mbefore", "", "Modified before")
	fileCmd.PersistentFlags().StringVar(&Fmtafter, "mafter", "", "Modified after")
}
