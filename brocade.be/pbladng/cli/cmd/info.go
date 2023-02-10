package cmd

import (
	"fmt"

	pstructure "brocade.be/pbladng/lib/structure"
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

	year, week, mailed, err := pstructure.DocRef("")
	if err != nil {
		return err
	}
	fmt.Printf("Information at `%s`:\n    year: %d\n    week: %02d\n    mail: %s", Fcwd, year, week, mailed)
	return nil
}
