package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	pfs "brocade.be/pbladng/lib/fs"
	pregistry "brocade.be/pbladng/lib/registry"
	pstructure "brocade.be/pbladng/lib/structure"
)

var formatCmd = &cobra.Command{
	Use:   "format",
	Short: "Format `gopblad`",
	Long:  "Format `gopblad`",

	Args:    cobra.MaximumNArgs(1),
	Example: `gopblad format myfile.pb`,
	RunE:    format,
}

func init() {
	rootCmd.AddCommand(formatCmd)
}

func format(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		if Fdebug {
			Fcwd = filepath.Join(pregistry.Registry["source-dir"].(string), "brocade.be", "pbladng", "test")
			args = append(args, filepath.Join(Fcwd, "week.pb"))
		} else {
			args = append(args, pfs.FName("workspace/week.pb"))
		}
	}
	fname := args[0]
	var source io.Reader
	dir := pfs.FName("workspace")
	if fname == "-" {
		source = os.Stdin
	} else {
		file, err := os.Open(fname)
		if err != nil {
			return err
		}
		dir = filepath.Dir(fname)
		source = bufio.NewReader(file)
	}
	doc := new(pstructure.Document)
	doc.Dir = dir
	err := doc.Load(source)
	if err != nil {
		return err
	}
	if err == nil {
		fmt.Print(doc.String())
	}
	return err
}
