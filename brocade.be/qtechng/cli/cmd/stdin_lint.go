package cmd

import (
	"io"
	"os"
	"strings"

	qfs "brocade.be/base/fs"
	"github.com/spf13/cobra"
)

func init() {
	stdinCmd.AddCommand(stdinLintCmd)
	stdinLintCmd.Flags().BoolVar(&Finplace, "inplace", false, "Replaces stdin")
	stdinLintCmd.Flags().BoolVar(&Fforce, "force", false, "Lint even if the file is not in repository")
	stdinLintCmd.Flags().StringVar(&Frefname, "refname", "", "Reference name instead of actual filename")
}

var stdinLintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Lints stdin",
	Long: `Command lints stdin an writes result on stdout.

The argument specifies the type of file: b | d | i | l | m | x
`,
	Args: cobra.MinimumNArgs(1),
	Example: `
  qtechng stdin lint m`,
	RunE: stdinLint,
	Annotations: map[string]string{
		"remote-allowed": "no",
	},
}

func stdinLint(cmd *cobra.Command, args []string) error {
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
	defer qfs.Rmpath(tmpfile)
	qfs.Store(tmpfile+ext, data, "")
	args[0] = tmpfile + ext
	return fileLint(cmd, args)
}
