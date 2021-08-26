package cmd

import (
	"github.com/spf13/cobra"
)

// Fauto ? writes according to repository
var Fauto bool

// Ftree ? writes according to the hierarchie
var Ftree bool

// Froot ? writes according to the hierarchie
var Froot bool

// Flist identifier of list of the results, if in auto mode
var Flist string
var Finlist string
var Fnotinlist string

var stdoutHidden bool
var stderrHidden bool

var sourceCmd = &cobra.Command{
	Use:     "source",
	Short:   "Source file functions",
	Long:    `All kinds of actions on the source files in the repository`,
	Args:    cobra.NoArgs,
	Example: "qtechng source",
}

func init() {
	rootCmd.AddCommand(sourceCmd)
	sourceCmd.PersistentFlags().StringVar(&Flist, "list", "", "Lists for convenient editing")
	sourceCmd.PersistentFlags().StringVar(&Fversion, "version", "", "Version to work with")
	sourceCmd.PersistentFlags().BoolVar(&Ftree, "tree", false, "Files with the repository hierarchy intact")
	sourceCmd.PersistentFlags().BoolVar(&Fauto, "auto", false, "Files according to the registry")
	sourceCmd.PersistentFlags().BoolVar(&Froot, "root", false, "Files according to the directory structure")
	sourceCmd.PersistentFlags().StringSliceVar(&Fnature, "nature", []string{}, "QtechNG nature of file")
	sourceCmd.PersistentFlags().StringSliceVar(&Fcu, "cuser", []string{}, "UID of creator")
	sourceCmd.PersistentFlags().StringSliceVar(&Fmu, "muser", []string{}, "UID of last modifier")
	sourceCmd.PersistentFlags().StringVar(&Fctbefore, "cbefore", "", "Created before")
	sourceCmd.PersistentFlags().StringVar(&Fctafter, "cafter", "", "Created after")
	sourceCmd.PersistentFlags().StringVar(&Fmtbefore, "mbefore", "", "Modified before")
	sourceCmd.PersistentFlags().StringVar(&Fmtafter, "mafter", "", "Modified after")
	sourceCmd.PersistentFlags().StringSliceVar(&Fneedle, "needle", []string{}, "Find substring")
	sourceCmd.PersistentFlags().StringSliceVar(&Fqpattern, "qpattern", []string{}, "Posix glob pattern on qpaths")
	sourceCmd.PersistentFlags().BoolVar(&Ffilesinproject, "neighbours", false, "Indicate if all files in project are selected")
	sourceCmd.PersistentFlags().StringVar(&Fqdir, "qdir", "", "Qpath of a directory under a project")
	sourceCmd.PersistentFlags().BoolVar(&Fperline, "perline", false, "Searches per line")
	sourceCmd.PersistentFlags().BoolVar(&Frecurse, "recurse", false, "Recursively traves directory and subdirectories")
	sourceCmd.PersistentFlags().BoolVar(&Fregexp, "regexp", false, "Searches as a regular expression")
	sourceCmd.PersistentFlags().BoolVar(&Ftolower, "tolower", false, "Transforms to lowercase")
	sourceCmd.PersistentFlags().BoolVar(&Fsmartcaseoff, "smartcaseoff", false, "Forbids smartcase transformation")
	sourceCoCmd.Flags().StringVar(&Flist, "list", "", "List with qpaths, if in auto mode")
}

var Msources = `

How are sources in the QtechNG repository specified?

Sources are identified in two different (but related) ways:

    - the (version, qpath) pair: 'qpath' is like an absolute path on Unix
	  with the root in the base of the source tree
	- the triple (version, qdir, basename): 'qdir' is the dirname of 'qpath'
	  (qpath = qdir/basename)

The version is determined by the '--version=...' flag.
On 'P' servers this value is always 'brocade-release',
Otherwise, most of the time, this value is deduced by the contents of
the current working directory: if all files in the directory
(and checked out of the repository) are of the same version, this value is taken.
Otherwise, on a 'B' machine, the value '0.00' is taken and on a 'W' workstation,
the value of 'qtechng-version' is taken.

The qpaths are calculated by the arguments.
These lead to a list of qpaths which are filtered by a number of restrictions.`
