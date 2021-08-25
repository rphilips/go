package cmd

import (
	qclip "brocade.be/clipboard"
	"github.com/spf13/cobra"
)

var clipboardSetCmd = &cobra.Command{
	Use:   "set text",
	Short: "Set the content of the system clipboard",
	Long: `Stores text in the system clipboard.
The text to be set is in the first argument.
If this argument is missing, the clipboard is emptied.`,
	Example: `qtechng clipboard set \"Hello World\"
qtechng clipboard set`,
	Args: cobra.MaximumNArgs(1),
	RunE: clipboardSet,
}

func init() {
	clipboardCmd.AddCommand(clipboardSetCmd)

}

func clipboardSet(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return qclip.WriteAll("")
	}
	return qclip.WriteAll(args[0])
}
