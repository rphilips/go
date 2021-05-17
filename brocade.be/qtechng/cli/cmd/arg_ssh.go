package cmd

import (
	qerror "brocade.be/qtechng/lib/error"
	"github.com/spf13/cobra"
)

var argSSHCmd = &cobra.Command{
	Use:     "ssh",
	Short:   "Start qtechng with arguments retrieved by SSH",
	Long:    `Launches qtechng with the arguments retrieved by SSH`,
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
