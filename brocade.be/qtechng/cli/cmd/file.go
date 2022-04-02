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
	fileCmd.PersistentFlags().StringVar(&Flist, "list", "", "Lists for convenient editing")
	fileCmd.PersistentFlags().StringVar(&Finlist, "inlist", "", "Qpath should be in this list")
	fileCmd.PersistentFlags().StringVar(&Fnotinlist, "notinlist", "", "Qpath should not be in this list")
}

var Mfiles = `

How are the local qtechng files, which are of interest to us, determined?

Local files are always specified against the current working directory (cwd):
paths are determined against this directory (change the current working
directory by means of the '--cwd=...' flag).

Local files are selected by:

	- a list of arguments (if the list of paths is empty, the cwd is added
	  to this list)
	- a version indicator
	- the '--recurse' flag
	- a list of qpatterns (a qpattern is a wildcard on a qpath
	  construction)
	- the '--changed' flag
	- the '--inlist=...' flag
	- the '--notinlist=...' flag

If the list of arguments is empty, the cwd is added to this list.

The file selection works in two steps:

    - The candidates are all the files in the arguments. If the arguments
	  contain directories, the files in these directories are added instead.
	  If the --recurse' flag is given, the files in the subdirectories
	  are added too.
    - This candidate list is then filtered according to the version,
	  the '--changed' flag and the list of accepted qpatterns (one matching
	  pattern is sufficient). If there is an *inlist*, then only candidates
	  with a qpath in this list are accepted. If there is a *notinlist*,
	  then only candidates with a qpath not in that list are accepted.

The *inlist* and the *notinlist* flags refer to files in the 'lists' subdirectory
of the support directory (without the *.lst* extension).
`
