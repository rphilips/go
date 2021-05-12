package cmd

import (
	qerror "brocade.be/qtechng/lib/error"
	"github.com/spf13/cobra"
)

var argURLCmd = &cobra.Command{
	Use:     "url",
	Short:   "Start qtechng with arguments retrieved by URL",
	Long:    `Launches qtechng with the arguments retrieved by URL`,
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
		//Fmsg = qreport.Report(nil, errorlist, Fjq, Fyaml)
	}
	return nil

}
