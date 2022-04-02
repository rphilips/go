package cmd

import (
	"io"
	"os"

	qtoolcat "brocade.be/qtechng/lib/toolcat"
	"github.com/spf13/cobra"
)

var toolcatModifierCmd = &cobra.Command{
	Use:   "modifier",
	Short: "toolcat modifier",
	Long: `This command generates the outline for a toolcatng modifier to be used in
a python module.

The information is provided as a JSON object

If there is an argument it is this JSON string
Without arguments the JSON string is read from stdin.

`,
	Args:    cobra.MaximumNArgs(1),
	Example: `qtechng toolcat modifier`,
	RunE:    toolcatModifier,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BWP",
	},
}

func init() {
	toolcatCmd.AddCommand(toolcatModifierCmd)
}

func toolcatModifier(cmd *cobra.Command, args []string) error {
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
	modifier := &qtoolcat.Modifier{}
	err := modifier.Load(jsono)
	if err != nil {
		return err
	}

	_, err = qtoolcat.Display(Fstdout, Fcwd, modifier, "", "    ", "", nil, Ftcclip, false)

	return err
}
