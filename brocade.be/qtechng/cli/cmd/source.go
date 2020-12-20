package cmd

import (
	"github.com/spf13/cobra"
)

// Fauto ? writes according to repository
var Fauto bool

// Ftree ? writes according to the hierarchie
var Ftree bool

var stdoutHidden bool
var stderrHidden bool

var sourceCmd = &cobra.Command{
	Use:     "source",
	Short:   "Source files activities",
	Long:    `Commands working on the source files in the repository`,
	Args:    cobra.NoArgs,
	Example: "qtechng source",
}

func init() {
	rootCmd.AddCommand(sourceCmd)
	sourceCmd.PersistentFlags().StringVar(&Fversion, "version", "", "Version to work with")
	sourceCmd.PersistentFlags().BoolVar(&Ftree, "tree", false, "Files with hierarchy intact")
	sourceCmd.PersistentFlags().BoolVar(&Fauto, "auto", false, "Files according to the registry")
	sourceCmd.PersistentFlags().StringSliceVar(&Fnature, "nature", []string{}, "QtechNG nature of file")
	sourceCmd.PersistentFlags().StringSliceVar(&Fcu, "cuser", []string{}, "UID of creator")
	sourceCmd.PersistentFlags().StringSliceVar(&Fmu, "muser", []string{}, "UID of last modifier")
	sourceCmd.PersistentFlags().StringVar(&Fctbefore, "cbefore", "", "Created before")
	sourceCmd.PersistentFlags().StringVar(&Fctafter, "cafter", "", "Created after")
	sourceCmd.PersistentFlags().StringVar(&Fmtbefore, "mbefore", "", "Modified before")
	sourceCmd.PersistentFlags().StringVar(&Fmtafter, "mafter", "", "Modified after")
	sourceCmd.PersistentFlags().StringSliceVar(&Fneedle, "needle", []string{}, "Find substring")
	sourceCmd.PersistentFlags().StringSliceVar(&Fpattern, "pattern", []string{}, "Posix glob pattern")
	sourceCmd.PersistentFlags().BoolVar(&Ffilesinproject, "neighbours", false, "Indicate if all files in project are selected")
	sourceCmd.PersistentFlags().StringVar(&Fqdir, "qdir", "", "qpath of a directory under a project")
	sourceCmd.PersistentFlags().BoolVar(&Fperline, "perline", false, "searches per line")
	sourceCmd.PersistentFlags().BoolVar(&Frecurse, "recurse", false, "recursively walks through directory and subdirectories")
	sourceCmd.PersistentFlags().BoolVar(&Fregexp, "regexp", false, "searches as a regular expression")
	sourceCmd.PersistentFlags().BoolVar(&Ftolower, "tolower", false, "transforms to lowercase")
	sourceCmd.PersistentFlags().BoolVar(&Fsmartcaseoff, "smartcaseoff", false, "Forbids smartcase transformation")
}
