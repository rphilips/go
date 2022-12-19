package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	pdocument "brocade.be/pbladng/lib/document"
	pfs "brocade.be/pbladng/lib/fs"
	pregistry "brocade.be/pbladng/lib/registry"
	ptools "brocade.be/pbladng/lib/tools"
)

var Fdir string

var parseCmd = &cobra.Command{
	Use:   "parse",
	Short: "parse `gopblad`",
	Long:  "parse `gopblad`",

	Example: `gopblad parse myfile.pb`,
	RunE:    parse,
}

func init() {
	parseCmd.PersistentFlags().BoolVar(&Fdebug, "debug", false, "put in debug mode")
	rootCmd.AddCommand(parseCmd)
}

func parse(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		if Fdebug {
			Fcwd = filepath.Join(pregistry.Registry["source-dir"].(string), "brocade.be", "pbladng", "test")
			args = append(args, filepath.Join(Fcwd, "week.md"))
		} else {
			args = append(args, pfs.FName("workspace/week.md"))
		}
	}
	fname := args[0]
	d, err := os.Open(fname)
	if err != nil {
		return ptools.Error("document-notexist", 2, err)
	}
	if Fdir == "" {
		Fdir = filepath.Dir(fname)
	}

	_, codes, alts, err := pdocument.Parse(d, Fcwd)

	if Fdebug {
		fmt.Println("\n\n\nCodes:")
		for k := range codes {
			fmt.Println(k)
		}

		fmt.Println("\n\n\nAlts:")
		for k, v := range alts {
			fmt.Println(k, v)
		}
	}
	return err
}
