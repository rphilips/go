package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	pmanuscript "brocade.be/pbladng/lib/manuscript"
	ptools "brocade.be/pbladng/lib/tools"
)

var formatCmd = &cobra.Command{
	Use:   "format",
	Short: "Format `gopblad`",
	Long:  "Format `gopblad`",

	Args:    cobra.ExactArgs(1),
	Example: `gopblad format myfile.pb`,
	RunE:    format,
}

func init() {

	rootCmd.AddCommand(formatCmd)
}

func format(cmd *cobra.Command, args []string) error {
	fname := args[0]
	file, err := os.Open(fname)
	if err != nil {
		return ptools.Error("manuscript-notexist", 0, err)
	}
	source := bufio.NewReader(file)
	m, err := pmanuscript.Parse(source)
	if err != nil {
		return err
	}
	fmt.Print(m)
	return nil
}
