package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	pdocument "brocade.be/pbladng/lib/document"
	perror "brocade.be/pbladng/lib/error"
	pfs "brocade.be/pbladng/lib/fs"
	pregistry "brocade.be/pbladng/lib/registry"
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
			args = append(args, filepath.Join(Fcwd, "week.md"))
		} else {
			args = append(args, pfs.FName("workspace/week.md"))
		}
	}
	fname := args[0]

	file, err := os.Open(fname)
	if err != nil {
		return perror.Error("document-notexist", 0, err)
	}
	source := bufio.NewReader(file)
	doc, _, _, err := pdocument.Parse(source, Fcwd)
	if err == nil {
		fmt.Print(doc.String())
	}
	return err
}
