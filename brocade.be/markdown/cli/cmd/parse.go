package cmd

import (
	"os"

	"github.com/spf13/cobra"

	mparse "brocade.be/markdown/lib/parse"
)

var parseCmd = &cobra.Command{
	Use:   "parse",
	Short: "parse `markdown`",
	Long:  "parse `markdown`",

	Example: `markdown parse myfile.pb`,
	RunE:    parse,
}

func init() {
	rootCmd.AddCommand(parseCmd)
}

func parse(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		args = append(args, "/home/rphilips/go/brocade.be/markdown/test/week.md")
	}
	fname := args[0]
	d, err := os.ReadFile(fname)
	if err != nil {
		return err
	}
	doc := mparse.Parse(d)
	doc.Dump(d, 0)
	return nil
}
