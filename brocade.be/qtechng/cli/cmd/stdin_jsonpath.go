package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

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
	var reader *bufio.Reader
	if len(args) == 1 {
		reader = bufio.NewReader(os.Stdin)
	} else {
		reader = bufio.NewReader(strings.NewReader(args[0]))
	}

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
