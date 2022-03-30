package cmd

import (
	"io"
	"os"
	"strings"

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

	sign := verb.Signature() + ":"

	replacements := make(map[string]string)
	replacements["\n    Argumenten: \"\"\n"] = "\n    Argumenten:\n"
	replacements["\n    Modifiers: \"\"\n"] = "\n    Modifiers:\n"

	after := []string{}

	if verb.WithArguments {
		after = append(after, `    print("argumenten:", repr(args))`)
	}
	if verb.WithModifiers {
		after = append(after, `    print("modifiers:", repr(modifiers))`)
		after = append(after, `    print("modifiers():", repr(modifiers()))`)
		after = append(after, `    for modifier in modifiers():`)
		after = append(after, `        print("modifier('" + modifier + "'):", repr(modifiers(modifier)))`)
	}
	if verb.WithVerbose {
		after = append(after, `    print("verbose:", repr(verbose))`)
	}
	if verb.WithDebug {
		after = append(after, `    print("debug:", repr(debug))`)
	}
	after = append(after, `    print("`+verb.Name+` loopt succesvol!")`)

	_, err = qtoolcat.Display(Fstdout, Fcwd, verb, sign, "    ", strings.Join(after, "\n"), replacements, Ftcclip, true)
	return err
}
