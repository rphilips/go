package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	pmanuscript "brocade.be/pbladng/lib/manuscript"
	ptools "brocade.be/pbladng/lib/tools"
)

var htmlCmd = &cobra.Command{
	Use:   "html",
	Short: "HTML `gopblad`",
	Long:  "HTML `gopblad`",

	Args:    cobra.ExactArgs(1),
	Example: `gopblad html myfile.pb`,
	RunE:    html,
}

var Fcolofon bool

func init() {
	htmlCmd.PersistentFlags().BoolVar(&Fimage, "extern", false, "Check external documents (like images)")
	htmlCmd.PersistentFlags().BoolVar(&Fcolofon, "colofon", false, "Include colofon ?")
	rootCmd.AddCommand(htmlCmd)
}

func html(cmd *cobra.Command, args []string) error {
	fname := args[0]
	dir := filepath.Dir(fname)
	manifest := filepath.Join(dir, "manifest.json")
	file, err := os.Open(fname)
	if err != nil {
		return ptools.Error("manuscript-notexist", 0, err)
	}
	source := bufio.NewReader(file)
	m, err := pmanuscript.Parse(source, Fimage, manifest)
	if err != nil {
		return err
	}
	fmt.Print(m.HTML(Fcolofon, manifest))
	return nil
}
