package cmd

import (
	"io/ioutil"
	"os"
	"strings"

	qfs "brocade.be/base/fs"
	"github.com/spf13/cobra"
)

func init() {
	stdinCmd.AddCommand(stdinFormatCmd)
	stdinFormatCmd.Flags().BoolVar(&Finplace, "inplace", false, "Replaces stdin")
}

var stdinFormatCmd = &cobra.Command{
	Use:   "format",
	Short: "Formats stdin",
	Long: `Command formats stdin an writes result on stdout.

The argument specifies the type of file: b | d | i | l | m | x
`,
	Args: cobra.MinimumNArgs(1),
	Example: `
  qtechng file format m`,
	RunE: stdinFormat,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func stdinFormat(cmd *cobra.Command, args []string) error {
	ext := args[0]
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	tmpfile, err := qfs.TempFile("", "")
	if err != nil {
		return err
	}
	qfs.Store(tmpfile+ext, data, "")
	args[0] = tmpfile + ext
	return fileFormat(cmd, args)
}
