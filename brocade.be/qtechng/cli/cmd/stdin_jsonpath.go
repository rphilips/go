package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"

	qutil "brocade.be/qtechng/lib/util"
	"github.com/spf13/cobra"
)

var stdinJsonpathCmd = &cobra.Command{
	Use:     "jsonpath",
	Short:   "Filter with jsonpath",
	Long:    `This command filters a JSON string from stdin through jsonpath and writes on stdout`,
	Example: "qtechng stdin jsonpath '$.store.book[*].author'",
	Args:    cobra.NoArgs,
	RunE:    stdinJsonpath,
}

func init() {
	stdinCmd.AddCommand(stdinJsonpathCmd)
}

func stdinJsonpath(cmd *cobra.Command, args []string) (err error) {
	reader := bufio.NewReader(os.Stdin)

	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	output, err := qutil.Transform(data, Fjq, Fyaml)
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
