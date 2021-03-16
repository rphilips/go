package cmd

import (
	"os"

	qerror "brocade.be/qtechng/lib/error"
	"github.com/spf13/cobra"
)

var argFileCmd = &cobra.Command{
	Use:     "file",
	Short:   "Start qtechng with arguments in a file",
	Long:    `Launches qtechng with the arguments as lines in a file. Arguments should not be empty`,
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
