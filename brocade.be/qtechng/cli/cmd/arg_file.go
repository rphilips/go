package cmd

import (
	"os"

	qerror "brocade.be/qtechng/lib/error"
	"github.com/spf13/cobra"
)

var argFileCmd = &cobra.Command{
	Use:   "file",
	Short: "Start qtechng with arguments in a file",
	Long: `Launches qtechng with the arguments specified as lines in a file. 
If the first non-whitespace character in the file is a *[*, the contents should be a *JSON array*.

If the file is *NOT* a JSON array, the following restriction apply:
    - Arguments are read line-by-line from the named file.
    - Whitespace is stripped at the beginning and the end of each line
	- Empty lines are skipped
    - The first line should be *qtechng*
	
If the file *IS* a JSON array, the following applies:
    - The first element should always be *qtechng*
    - Whitespace is never stripped
    - Empty arguments remain in the argument list
`,

	Args:    cobra.ExactArgs(1),
	Example: `qtechng arg file myargs.txt`,
	RunE:    argFile,
}

func init() {
	argCmd.AddCommand(argFileCmd)
}

func argFile(cmd *cobra.Command, args []string) error {
	jarg := args[0]

	if jarg == "" {
		err := &qerror.QError{
			Ref:  []string{"arg.file.empty"},
			Type: "Error",
			Msg:  []string{"Argument is empty"},
		}
		return err
	}
	_, err := os.ReadFile(jarg)

	return err
}
