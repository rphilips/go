package cmd

import (
	qerror "brocade.be/qtechng/lib/error"
	"github.com/spf13/cobra"
)

var argSSHCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Start qtechng with arguments retrieved by SSH",
	Long: `Launches qtechng with the arguments retrieved by SSH.
	
The argument is a full path to a file on the development server. 
This file is retrieved and the arguments are extracted.

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
	Example: `qtechng arg ssh /library/tmp/run.txt`,
	RunE:    argSSH,
}

func init() {
	argCmd.AddCommand(argSSHCmd)
}

func argSSH(cmd *cobra.Command, args []string) error {
	jarg := args[0]

	if jarg == "" {
		err := &qerror.QError{
			Ref:  []string{"arg.ssh.empty"},
			Type: "Error",
			Msg:  []string{"Argument is empty"},
		}
		return err
	}
	return nil

}
