package cmd

import (
	"encoding/json"

	qerror "brocade.be/qtechng/lib/error"
	"github.com/spf13/cobra"
)

var argJSONCmd = &cobra.Command{
	Use:   "json",
	Short: "Start qtechng with arguments in JSON",
	Long: `Launches qtechng with the arguments specified in a JSON string.
	
The command works with exactly one argument: a string containing a *JSON array*.

The following applies:
    - The first element should always be *qtechng*
    - Whitespace is never stripped
    - Empty arguments remain in the argument list
`,
	Args:    cobra.ExactArgs(1),
	Example: `qtechng arg json '["qtechng", "system", "info"]'`,
	RunE:    argJSON,
}

func init() {
	argCmd.AddCommand(argJSONCmd)
}

func argJSON(cmd *cobra.Command, args []string) error {
	jarg := args[0]

	if jarg == "" {
		err := &qerror.QError{
			Ref:  []string{"arg.json.empty"},
			Type: "Error",
			Msg:  []string{"Argument is empty"},
		}
		return err
	}

	argums := make([]string, 0)

	err := json.Unmarshal([]byte(jarg), &argums)

	return err

}
