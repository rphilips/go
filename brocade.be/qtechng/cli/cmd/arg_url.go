package cmd

import (
	qerror "brocade.be/qtechng/lib/error"
	"github.com/spf13/cobra"
)

var argURLCmd = &cobra.Command{
	Use:   "url",
	Short: "Start qtechng with arguments retrieved by URL",
	Long: `Launches qtechng with the arguments retrieved by URL.

If the first non-whitespace character is a *[*, the contents should be a JSON array.

If the input is *NOT* a JSON array, the following restrictions apply:
    - The first line should always be *qtechng*
	- Arguments are read line-by-line from the input
	- Whitespace is stripped at the beginning and the end of each line
	- Empty lines are skipped
	- The first line should be *qtechng*

If the input *IS* a JSON array, the following applies:
	- The first element should always be *qtechng*
	- Whitespace is never stripped
	- Empty arguments remain in the argument list`,
	Args:    cobra.ExactArgs(1),
	Example: `qtechng arg url https://dev.anet.be/about.html`,
	RunE:    argURL,
}

func init() {
	argCmd.AddCommand(argURLCmd)
}

func argURL(cmd *cobra.Command, args []string) error {
	jarg := args[0]

	if jarg == "" {
		err := &qerror.QError{
			Ref:  []string{"arg.url.empty"},
			Type: "Error",
			Msg:  []string{"Argument is empty"},
		}
		return err
	}
	return nil

}
