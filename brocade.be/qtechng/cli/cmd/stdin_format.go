package cmd

import (
	"io"
	"os"
	"strings"

	qfs "brocade.be/base/fs"
	"github.com/spf13/cobra"
)

func init() {
	stdinCmd.AddCommand(stdinFormatCmd)
	stdinFormatCmd.Flags().BoolVar(&Finplace, "inplace", false, "Replace stdin")
}

var stdinFormatCmd = &cobra.Command{
	Use:   "format",
	Short: "Format stdin",
	Long: `This command formats stdin an writes result on stdout.

The argument specifies the type of file: b | d | i | l | m | x
`,
	Args: cobra.MinimumNArgs(1),
	Example: `
  qtechng stdin format m`,
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
	data, err := io.ReadAll(os.Stdin)

	if err != nil {
		return err
	}
	tmpfile, err := qfs.TempFile("", "")
	if err != nil {
		return err
	}
	tmpfile = tmpfile + ext
	defer qfs.Rmpath(tmpfile)
	err = qfs.Store(tmpfile, data, "qtech")
	if err != nil {
		return err
	}
	args[0] = tmpfile
	return fileFormat(cmd, args)
}
