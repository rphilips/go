package cmd

import (
	"fmt"
	"io"
	"os"

	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var stdinJsonpathCmd = &cobra.Command{
	Use:     "jsonpath",
	Short:   "jsonpath selection",
	Long:    `Filters stdin - as a JSON string - through jsonpath and writes on stdout`,
	Example: "  qtechng stdin jsonpath '$.store.book[*].author'",
	Args:    cobra.MaximumNArgs(1),
	RunE:    stdinJsonpath,
}

func init() {
	stdinCmd.AddCommand(stdinJsonpathCmd)
}

func stdinJsonpath(cmd *cobra.Command, args []string) (err error) {
	jsonpath := ""
	if len(args) != 0 {
		jsonpath = args[0]
	}
	data, err := io.ReadAll(os.Stdin)
	output, err := qutil.Transform(data, jsonpath, Fyaml)
	if err != nil {
		return err
	}

	if Fstdout == "" || Ftransported {
		fmt.Print(output)
		return nil
	}

	f, err := os.Create(Fstdout)
	if err != nil {
		return err
	}
	fmt.Fprint(f, output)
	defer f.Close()
	return err
}
