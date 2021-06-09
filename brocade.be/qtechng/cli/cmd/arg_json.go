package cmd

import (
	"encoding/json"

	qerror "brocade.be/qtechng/lib/error"
	"github.com/spf13/cobra"
)

var argJSONCmd = &cobra.Command{
	Use:     "json",
	Short:   "Start qtechng with arguments in JSON",
	Long:    `Launches qtechng with the arguments in a JSON string`,
	Args:    cobra.ExactArgs(1),
	Example: `qtechng arg json '["system", "info"]'`,
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
		//Fmsg = qreport.Report(nil, errorlist, Fjq, Fyaml, Funquote, Fjoiner, Fsilent)
	}

	argums := make([]string, 0)

	err := json.Unmarshal([]byte(jarg), &argums)

	return err

}
