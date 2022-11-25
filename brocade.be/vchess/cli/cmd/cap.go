package cmd

import (
	vstructure "brocade.be/vchess/lib/structure"
	"github.com/spf13/cobra"
)

var capCmd = &cobra.Command{
	Use:   "cap",
	Short: "Information print `vchess`",
	Long:  `Version and build time printrmation print the vchess executable`,

	Args:    cobra.NoArgs,
	Example: `vchess capl`,
	RunE:    cap,
}

func init() {
	rootCmd.AddCommand(capCmd)
}

func cap(cmd *cobra.Command, args []string) (err error) {

	season := new(vstructure.Season)
	season.Init(nil)

	_, err = season.Calendar()

	if err != nil {
		return
	}

	return nil
}
