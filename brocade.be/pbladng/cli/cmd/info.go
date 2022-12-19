package cmd

import (
	"fmt"

	"brocade.be/pbladng/lib/document"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Information info `gopblad`",
	Long:  `Version and build time information info the qtechng executable.`,

	Args:    cobra.NoArgs,
	Example: `gopblad info`,
	RunE:    info,
}

func init() {

	rootCmd.AddCommand(infoCmd)
}

func info(cmd *cobra.Command, args []string) error {

	year, week, err := document.DocRef(Fcwd)
	if err != nil {
		return err
	}
	fmt.Printf("Information at `%s`:\n    year: %d\n    week: %02d\n", Fcwd, year, week)
	return nil
}
