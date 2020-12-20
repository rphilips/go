package cmd

import (
	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:     "project",
	Short:   "Project functions",
	Long:    `All kinds of actions on projects`,
	Args:    cobra.NoArgs,
	Example: "qtechng project",
}

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.PersistentFlags().StringVar(&Fversion, "version", "", "Version to work with")
	projectCmd.PersistentFlags().BoolVar(&Ftree, "tree", false, "Files with hierarchy intact")
	projectCmd.PersistentFlags().BoolVar(&Fauto, "auto", false, "Files according to the registry")
	projectCmd.PersistentFlags().StringSliceVar(&Fnature, "nature", []string{}, "QtechNG nature of file")
	projectCmd.PersistentFlags().StringSliceVar(&Fcu, "cuser", []string{}, "UID of creator")
	projectCmd.PersistentFlags().StringSliceVar(&Fmu, "muser", []string{}, "UID of last modifier")
	projectCmd.PersistentFlags().StringVar(&Fctbefore, "cbefore", "", "Created before")
	projectCmd.PersistentFlags().StringVar(&Fctafter, "cafter", "", "Created after")
	projectCmd.PersistentFlags().StringVar(&Fmtbefore, "mbefore", "", "Modified before")
	projectCmd.PersistentFlags().StringVar(&Fmtafter, "mafter", "", "Modified after")
}
