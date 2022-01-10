package cmd

import (
	"io"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var pipewriteCmd = &cobra.Command{
	Use:     "write",
	Short:   "set up named pipe for writing",
	Long:    `set up named pipe for writing`,
	Args:    cobra.NoArgs,
	Example: `goyo pipe write --pipe=/tmp/goyo`,
	RunE:    pipewrite,
}

func init() {
	pipeCmd.AddCommand(pipewriteCmd)
}

func pipewrite(cmd *cobra.Command, args []string) error {
	if Fpipe == "" {
		log.Fatal("Missing named pipe")
	}

	stdin, _ := os.OpenFile(Fpipe, os.O_WRONLY, 0600)

	io.Copy(stdin, os.Stdin)

	stdin.WriteString("<[<end>]>\n")
	stdin.Close()
	return nil
}
