package cmd

import (
	"io"
	"os"

	"brocade.be/qtechng/lib/toolcat"
	qtoolcat "brocade.be/qtechng/lib/toolcat"
	"github.com/spf13/cobra"
)

var toolcatArgCmd = &cobra.Command{
	Use:   "arg",
	Short: "toolcat arg",
	Long: `This command generates the outline for a toolcatng arg to be used in
a python module.

The information is provided as a JSON object

If there is an argument it is this JSON string
Without arguments the JSON string is read from stdin.

`,
	Args:    cobra.MaximumNArgs(1),
	Example: `qtechng toolcat arg`,
	RunE:    toolcatArg,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BWP",
	},
}

func init() {
	toolcatCmd.AddCommand(toolcatArgCmd)
}

func toolcatArg(cmd *cobra.Command, args []string) error {
	jsono := ""
	if len(args) != 0 {
		jsono = args[0]
	} else {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		jsono = string(data)
	}
	arg := &qtoolcat.Arg{}
	err := arg.Load(jsono)
	if err != nil {
		return err
	}

	err = toolcat.Display(Fstdout, Fcwd, arg, "", "        ", "", nil, Ftcclip, false)
	return err
}
