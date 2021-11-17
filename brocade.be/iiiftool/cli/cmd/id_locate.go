package cmd

import (
	"errors"
	"fmt"

	identifier "brocade.be/iiiftool/lib/identifier"

	"github.com/spf13/cobra"
)

var idLocateCmd = &cobra.Command{
	Use:   "locate",
	Short: "Locate a IIIF identifier",
	Long: `Given a IIIF identifier locate the appropriate SQLite filepath
	`,

	Args:    cobra.ExactArgs(1),
	Example: `iiiftool id locate dg:ua:1`,
	RunE:    idLocate,
}

func init() {
	idCmd.AddCommand(idLocateCmd)
}

func idLocate(cmd *cobra.Command, args []string) error {
	id := identifier.Identifier(args[0])

	if id.String() == "" {
		return errors.New("argument is empty")
	}
	fmt.Println(id.Location())
	return nil
}
