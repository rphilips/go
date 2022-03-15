package cmd

import (
	"io"
	"os"

	qtoolcat "brocade.be/qtechng/lib/toolcat"
	"github.com/spf13/cobra"
)

var toolcatAppCmd = &cobra.Command{
	Use:   "app",
	Short: "Docstring for a toolcatng app",
	Long: `This command generates a docstring for a toolcatng App to be used in
a python module.

The information is provided as a JSON object

If there is an argument it is this JSON string
Without arguments the JSON string is read from stdin.

`,
	Args:    cobra.MaximumNArgs(1),
	Example: `qtechng toolcat app`,
	RunE:    toolcatApp,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BWP",
	},
}

func init() {
	toolcatCmd.AddCommand(toolcatAppCmd)
}

func toolcatApp(cmd *cobra.Command, args []string) error {
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
	app := &qtoolcat.App{}
	err := app.Load(jsono)
	if err != nil {
		return err
	}

	after := `
from anet.core import base
from anet.toolcatng import toolcat

`

	_, err = qtoolcat.Display(Fstdout, Fcwd, app, "", "", after, nil, Ftcclip, true)
	return err
}
