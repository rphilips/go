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

var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "lint `pblad`",
	Long:  "lint `pblad`",

	Args:    cobra.MaximumNArgs(1),
	Example: `pblad lint myfile.pb`,
	RunE:    lint,
}

func init() {
	rootCmd.AddCommand(lintCmd)
}

func lint(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		if Fdebug {
			Fcwd = filepath.Join(pregistry.Registry["source-dir"].(string), "brocade.be", "pbladng", "test")
			args = append(args, filepath.Join(Fcwd, "parochieblad.ed"))
		} else {
			args = append(args, pfs.FName("workspace/parochieblad.ed"))
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
	errmsg := ""
	err := doc.Load(source)
	if err != nil {
		errmsg = err.Error()
	}
	fmt.Println(errmsg)
	fmt.Fprint(os.Stderr, errmsg)
	return nil
}
