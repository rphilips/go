package cmd

import (
	"github.com/spf13/cobra"
)

var argStdinCmd = &cobra.Command{
	Use:   "stdin",
	Short: "Start goyo with arguments read from stdin",
	Long: `Launches goyo with the arguments as lines on stdin. 
	
If the first non-whitespace character is a *[*, the contents should be a JSON array.

If the input is *NOT* a JSON array, the following restriction apply:
	- Arguments are read line-by-line from *stdin*
	- Whitespace is stripped at the beginning and the end of each line
	- Empty lines are skipped
	- The first line should be *goyo*
	
If the input *IS* a JSON array, the following applies:
	- The first element should always be *goyo*
	- Whitespace is never stripped
	- Empty arguments remain in the argument list`,
	Args:    cobra.NoArgs,
	Example: `goyo arg stdin < commands.txt`,
	RunE:    argStdin,
}

func init() {
	argCmd.AddCommand(argStdinCmd)
}

func argStdin(cmd *cobra.Command, args []string) error {
	return nil
}
