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
import json

from anet.core import base
from anet.toolcatng import toolcat


@toolcat.toolcat
def about():
    r'''
    Titel: Informatie omtrent deze toolcat applicatie

    Beschrijving: |-
        Deze functie verschaft repository informatie omtrent deze toolcat applicatie.

        Deze informatie wordt opgehaald door gebruik te maken van *qtechng*

    Triggers: about

    Voorbeelden:
        - {APPNAME} about

    Argumenten: Geen argumenten
    '''
    qpath = "{APPQPATH}"
    valid = qpath.startswith("/")
    if valid:
        cp = base.catch("qtechng", args=["source", "list", qpath])
    	props = json.loads(cp.stdout)
        valid = "DATA" in props and props["DATA"]
        if valid:
            print(json.dumps(props["DATA"][0], indent=4))
    if not valid:
        print("Geen informatie gevonden betreffende", qpath)
`
	_, err = qtoolcat.Display(Fstdout, Fcwd, app, "", "", after, nil, Ftcclip, true)
	return err
}
