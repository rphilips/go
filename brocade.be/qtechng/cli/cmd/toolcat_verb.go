package cmd

import (
	"io"
	"os"

	"brocade.be/qtechng/lib/toolcat"
	qtoolcat "brocade.be/qtechng/lib/toolcat"
	"github.com/spf13/cobra"
)

var toolcatVerbCmd = &cobra.Command{
	Use:   "verb",
	Short: "toolcat verb",
	Long: `This command generates the outline for a toolcatng verb to be used in
a python module.

The information is provided as a JSON object

If there is an argument it is this JSON string
Without arguments the JSON string is read from stdin.

`,
	Args:    cobra.MaximumNArgs(1),
	Example: `qtechng toolcat verb`,
	RunE:    toolcatVerb,
	Annotations: map[string]string{
		"remote-allowed": "no",
		"with-qtechtype": "BWP",
	},
}

func init() {
	toolcatCmd.AddCommand(toolcatVerbCmd)
}

func toolcatVerb(cmd *cobra.Command, args []string) error {
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
	verb := &qtoolcat.Verb{}
	err := verb.Load(jsono)
	if err != nil {
		return err
	}

	sign := "def " + verb.Signature() + ":"

	replacements := make(map[string]string)
	replacements["\n    Argumenten: \"\"\n"] = "\n    Argumenten:\n"
	replacements["\n    Modifiers: \"\"\n"] = "\n    Modifiers:\n"

	return toolcat.Display(Fstdout, Fcwd, verb, sign, "    ", "    pass\n", replacements, Ftcclip, true)
}
