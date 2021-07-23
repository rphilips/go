package cmd

import (
	"bufio"
	"fmt"
	"os"

	qclip "brocade.be/clipboard"
	"github.com/spf13/cobra"
)

var clipboardGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieves the content of the system clipboard",
	Long: `Retrieves the content of the system clipboard and writes on stdout.
Works only with text.`,
	Example: "qtechng clipboard get",
	Args:    cobra.NoArgs,
	RunE:    clipboardGet,
}

func init() {
	clipboardCmd.AddCommand(clipboardGetCmd)
}

func clipboardGet(cmd *cobra.Command, args []string) error {
	clip, _ := qclip.ReadAll()
	if Fstdout == "" || Ftransported {
		fmt.Print(clip)
		return nil
	}
	f, err := os.Create(Fstdout)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	fmt.Fprint(w, clip)
	err = w.Flush()
	return err
}
